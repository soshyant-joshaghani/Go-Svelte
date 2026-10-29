// Package testkit is the shared harness for the backend tests: in-memory fakes for the
// repositories, plus an Env that builds the real HTTP handler around them.
package testkit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/cache"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/config"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/jobs"
	"github.com/soshyant-joshaghani/go-svelte/internal/httpserver"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/apps/sample"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/base/users"
)

// ---- in-memory fakes: no Postgres or Redis needed --------------------------

type FakeUsers struct {
	mu   sync.Mutex
	Rows map[string]users.User
}

func (f *FakeUsers) GetByID(_ context.Context, id string) (*users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, ok := f.Rows[id]; ok {
		return &u, nil
	}
	return nil, nil
}
func (f *FakeUsers) GetByEmail(_ context.Context, email string) (*users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.Rows {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, nil
}
func (f *FakeUsers) sorted() []users.User {
	out := make([]users.User, 0, len(f.Rows))
	for _, u := range f.Rows {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Email < out[j].Email })
	return out
}
func (f *FakeUsers) Count(context.Context) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Rows), nil
}
func (f *FakeUsers) List(_ context.Context, skip, limit int) ([]users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := f.sorted()
	if skip > len(all) {
		skip = len(all)
	}
	all = all[skip:]
	if limit < len(all) {
		all = all[:limit]
	}
	return all, nil
}
func (f *FakeUsers) Create(_ context.Context, u users.User) (users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, other := range f.Rows {
		if other.Email == u.Email {
			return users.User{}, users.ErrDuplicateEmail
		}
	}
	f.Rows[u.ID] = u
	return u, nil
}
func (f *FakeUsers) Update(_ context.Context, u users.User) (users.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Rows[u.ID] = u
	return u, nil
}
func (f *FakeUsers) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.Rows, id)
	return nil
}

type FakeNotes struct {
	mu   sync.Mutex
	Rows map[string]sample.Note
	Gets int // GetByID calls: tests use it to tell cache hits from database reads
}

func (f *FakeNotes) GetByID(_ context.Context, id string) (*sample.Note, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Gets++
	if n, ok := f.Rows[id]; ok {
		return &n, nil
	}
	return nil, nil
}
func (f *FakeNotes) ListByOwner(_ context.Context, owner string) ([]sample.Note, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []sample.Note{}
	for _, n := range f.Rows {
		if n.OwnerID == owner {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}
func (f *FakeNotes) Create(_ context.Context, n sample.Note) (sample.Note, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Rows[n.ID] = n
	return n, nil
}
func (f *FakeNotes) Update(_ context.Context, n sample.Note) (sample.Note, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Rows[n.ID] = n
	return n, nil
}
func (f *FakeNotes) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.Rows, id)
	return nil
}
func (f *FakeNotes) DeleteByOwner(_ context.Context, owner string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, n := range f.Rows {
		if n.OwnerID == owner {
			delete(f.Rows, id)
		}
	}
	return nil
}

// ---- harness ---------------------------------------------------------------

type Env struct {
	T     *testing.T
	Srv   *httpserver.Server
	H     http.Handler
	Users *FakeUsers
	Notes *FakeNotes
	Cache *cache.Memory
	Jobs  *jobs.Memory
}

func New(t *testing.T, mutate ...func(*config.Config)) *Env {
	t.Helper()
	cfg := config.ForTests()
	cfg.BackendCORSOrigins = []string{"http://app.test"}
	for _, m := range mutate {
		m(&cfg)
	}
	e := &Env{
		T: t, Users: &FakeUsers{Rows: map[string]users.User{}}, Notes: &FakeNotes{Rows: map[string]sample.Note{}},
		Cache: cache.NewMemory(), Jobs: &jobs.Memory{},
	}
	e.Srv = &httpserver.Server{Config: cfg, Users: e.Users, Notes: e.Notes, Cache: e.Cache, Jobs: e.Jobs}
	if err := e.Srv.UserService().EnsureFirstSuperuser(context.Background(), cfg.FirstSuperuser, cfg.FirstSuperuserPassword); err != nil {
		t.Fatal(err)
	}
	e.H = e.Srv.Handler()
	return e
}

type Resp struct {
	Code   int
	Header http.Header
	Body   []byte
}

func (r Resp) JSON() map[string]any {
	var m map[string]any
	_ = json.Unmarshal(r.Body, &m)
	return m
}
func (r Resp) List() []any {
	var l []any
	_ = json.Unmarshal(r.Body, &l)
	return l
}
func (r Resp) Detail() string { s, _ := r.JSON()["detail"].(string); return s }

func (e *Env) Do(method, path, token, contentType, body string) Resp {
	e.T.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.H.ServeHTTP(rec, req)
	return Resp{Code: rec.Code, Header: rec.Header(), Body: rec.Body.Bytes()}
}

func (e *Env) Call(method, path, token string, body any) Resp {
	raw := ""
	if body != nil {
		b, _ := json.Marshal(body)
		raw = string(b)
	}
	return e.Do(method, "/api/v1"+path, token, "application/json", raw)
}

func (e *Env) Login(email, password string) Resp {
	form := url.Values{"username": {email}, "password": {password}}
	return e.Do("POST", "/api/v1/base/login/access-token", "", "application/x-www-form-urlencoded", form.Encode())
}

func (e *Env) Token(email, password string) string {
	e.T.Helper()
	r := e.Login(email, password)
	if r.Code != 200 {
		e.T.Fatalf("login %s: %d %s", email, r.Code, r.Body)
	}
	return r.JSON()["access_token"].(string)
}

func (e *Env) Admin() string { return e.Token("admin@example.com", "adminpass123") }

// newUser creates a regular user through the admin API and returns its id and token.
func (e *Env) NewUser(email string) (string, string) {
	e.T.Helper()
	r := e.Call("POST", "/base/users/admin", e.Admin(), map[string]any{"email": email, "password": "password123"})
	if r.Code != 200 {
		e.T.Fatalf("create user: %d %s", r.Code, r.Body)
	}
	return r.JSON()["id"].(string), e.Token(email, "password123")
}

func Expect(t *testing.T, r Resp, code int, detail string) {
	t.Helper()
	if r.Code != code {
		t.Fatalf("status = %d, want %d (body %s)", r.Code, code, r.Body)
	}
	if detail != "" && r.Detail() != detail {
		t.Fatalf("detail = %q, want %q", r.Detail(), detail)
	}
}
