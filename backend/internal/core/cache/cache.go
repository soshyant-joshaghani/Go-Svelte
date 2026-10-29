// Package cache is a soft-degrading Redis cache: when Redis is unreachable
// every call is a no-op and the request still succeeds.
package cache

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store interface {
	// Get returns the value and whether it was found.
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string, ttl time.Duration)
	Delete(ctx context.Context, key string)
	// DeletePrefix deletes every key that starts with prefix.
	DeletePrefix(ctx context.Context, prefix string)
}

// Redis is the production Store.
type Redis struct{ client *redis.Client }

func NewRedis(client *redis.Client) *Redis { return &Redis{client: client} }

// Options builds client options with short timeouts so a dead Redis does not stall requests.
func Options(host string, port, db int, password string) *redis.Options {
	return &redis.Options{
		Addr:         host + ":" + itoa(port),
		Password:     password,
		DB:           db,
		DialTimeout:  time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		MaxRetries:   1,
	}
}

func (r *Redis) Get(ctx context.Context, key string) (string, bool) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return value, true
}

func (r *Redis) Set(ctx context.Context, key, value string, ttl time.Duration) {
	_ = r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, key string) {
	_ = r.client.Del(ctx, key).Err()
}

func (r *Redis) DeletePrefix(ctx context.Context, prefix string) {
	var cursor uint64
	for {
		keys, next, err := r.client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return
			}
		}
		cursor = next
		if cursor == 0 {
			return
		}
	}
}

// Memory is an in-process Store for tests.
type Memory struct {
	mu     sync.Mutex
	values map[string]memItem
}

type memItem struct {
	value   string
	expires time.Time
}

func NewMemory() *Memory { return &Memory{values: map[string]memItem{}} }

func (m *Memory) Get(_ context.Context, key string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.values[key]
	if !ok || time.Now().After(item.expires) {
		return "", false
	}
	return item.value, true
}

func (m *Memory) Set(_ context.Context, key, value string, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[key] = memItem{value: value, expires: time.Now().Add(ttl)}
}

func (m *Memory) Delete(_ context.Context, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.values, key)
}

func (m *Memory) DeletePrefix(_ context.Context, prefix string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key := range m.values {
		if strings.HasPrefix(key, prefix) {
			delete(m.values, key)
		}
	}
}

// Keys lists the live keys (tests assert on them).
func (m *Memory) Keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []string
	for key, item := range m.values {
		if time.Now().Before(item.expires) {
			out = append(out, key)
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	return string(digits[i:])
}
