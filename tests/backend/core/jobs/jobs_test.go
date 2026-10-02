package jobs_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/jobs"
)

func TestJobsRetryProtocol(t *testing.T) {
	var pushed []string
	repush := func(p string) error { pushed = append(pushed, p); return nil }
	fail := func(jobs.Job) error { return errors.New("boom") }

	raw := `{"id":"1","task":"flaky","args":{},"enqueued_at":"2026-01-01T00:00:00Z"}`
	for attempt := 1; attempt <= jobs.MaxRetries; attempt++ {
		pushed = nil
		jobs.Handle(raw, fail, repush)
		if len(pushed) != 1 || !strings.Contains(pushed[0], `"attempt":`+string(rune('0'+attempt))) {
			t.Fatalf("attempt %d: pushed %v", attempt, pushed)
		}
		raw = pushed[0]
	}
	pushed = nil
	jobs.Handle(raw, fail, repush)
	if len(pushed) != 0 {
		t.Fatalf("job re-pushed after %d retries", jobs.MaxRetries)
	}
	jobs.Handle(`{"id":"2","task":"nope","args":{}}`, jobs.Run, repush)
	jobs.Handle(`not json`, jobs.Run, repush)
	jobs.Handle(`{"id":"3","task":"ping","args":{"message":"hi"}}`, jobs.Run, repush)
	if len(pushed) != 0 {
		t.Fatalf("unknown, malformed and ping jobs must not be re-pushed: %v", pushed)
	}
	if err := jobs.Run(jobs.NewJob("ping", map[string]any{"message": "x"})); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(jobs.Run(jobs.NewJob("nope", nil)), jobs.ErrUnknownTask) {
		t.Fatal("unknown task not reported")
	}
}

func TestJobPayloadShape(t *testing.T) {
	j := jobs.NewJob("ping", map[string]any{"message": "m"})
	raw, _ := json.Marshal(j)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	for _, k := range []string{"id", "task", "args", "enqueued_at"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("payload lacks %s: %s", k, raw)
		}
	}
	if _, ok := m["attempt"]; ok {
		t.Fatal("first enqueue must not carry attempt")
	}
}
