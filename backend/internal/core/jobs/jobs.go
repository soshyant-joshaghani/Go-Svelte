// Package jobs is the FoxG Redis-list job protocol shared by the API (enqueue)
// and the worker (BRPOP). No queue library.
//
// Queue: list "foxg:jobs". Payload: {"id","task","args","enqueued_at"} and, on
// retries, "attempt": n. A failed task is re-pushed up to 3 times.
package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	QueueKey   = "foxg:jobs"
	MaxRetries = 3
)

type Job struct {
	ID         string         `json:"id"`
	Task       string         `json:"task"`
	Args       map[string]any `json:"args"`
	EnqueuedAt string         `json:"enqueued_at"`
	Attempt    int            `json:"attempt,omitempty"`
}

func NewJob(task string, args map[string]any) Job {
	return Job{
		ID:         uuid.NewString(),
		Task:       task,
		Args:       args,
		EnqueuedAt: time.Now().UTC().Format("2006-01-02T15:04:05.000000Z"),
	}
}

// Queue enqueues a job and returns its id. The error text is shown to clients
// (503 "Redis unavailable: ...").
type Queue interface {
	Enqueue(ctx context.Context, task string, args map[string]any) (string, error)
}

type RedisQueue struct{ client *redis.Client }

func NewRedisQueue(client *redis.Client) *RedisQueue { return &RedisQueue{client: client} }

func (q *RedisQueue) Enqueue(ctx context.Context, task string, args map[string]any) (string, error) {
	job := NewJob(task, args)
	payload, err := json.Marshal(job)
	if err != nil {
		return "", err
	}
	if err := q.client.LPush(ctx, QueueKey, payload).Err(); err != nil {
		return "", err
	}
	return job.ID, nil
}

// Memory is an in-process Queue for tests.
type Memory struct {
	mu   sync.Mutex
	jobs []Job
	fail error
}

// FailWith makes every Enqueue return err (simulates Redis down).
func (m *Memory) FailWith(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fail = err
}

func (m *Memory) Jobs() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Job(nil), m.jobs...)
}

func (m *Memory) Enqueue(_ context.Context, task string, args map[string]any) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail != nil {
		return "", m.fail
	}
	job := NewJob(task, args)
	m.jobs = append(m.jobs, job)
	return job.ID, nil
}

// ErrUnknownTask marks a task name the worker does not know; the job is dropped.
var ErrUnknownTask = errors.New("unknown task")

// Run executes one task by name.
func Run(job Job) error {
	switch job.Task {
	case "ping":
		message, _ := job.Args["message"].(string)
		if message == "" {
			message = "pong"
		}
		log.Printf("ping job received: %s", message)
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownTask, job.Task)
	}
}

// Handle processes one raw payload. repush re-queues a failed job.
func Handle(raw string, run func(Job) error, repush func(payload string) error) {
	var job Job
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		log.Printf("dropping malformed job payload: %v", err)
		return
	}
	err := run(job)
	switch {
	case err == nil:
		log.Printf("job %s (%s) done", job.ID, job.Task)
	case errors.Is(err, ErrUnknownTask):
		log.Printf("job %s: unknown task %q, dropped", job.ID, job.Task)
	case job.Attempt < MaxRetries:
		job.Attempt++
		log.Printf("job %s (%s) failed: %v; retry %d/%d", job.ID, job.Task, err, job.Attempt, MaxRetries)
		payload, _ := json.Marshal(job)
		if err := repush(string(payload)); err != nil {
			log.Printf("could not re-push job %s: %v", job.ID, err)
		}
	default:
		log.Printf("job %s (%s) failed permanently: %v", job.ID, job.Task, err)
	}
}

// RunWorker blocks on BRPOP foxg:jobs 5 until ctx is cancelled.
func RunWorker(ctx context.Context, client *redis.Client) {
	log.Printf("worker started, waiting for jobs on %s", QueueKey)
	repush := func(payload string) error { return client.LPush(ctx, QueueKey, payload).Err() }
	for ctx.Err() == nil {
		res, err := client.BRPop(ctx, 5*time.Second, QueueKey).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("redis unavailable (%v); retrying in 2s", err)
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
			continue
		}
		if len(res) == 2 {
			Handle(res[1], Run, repush)
		}
	}
	log.Print("worker shutting down")
}
