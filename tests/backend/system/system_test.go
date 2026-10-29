package system_test

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/config"
	"github.com/soshyant-joshaghani/go-svelte/tests/backend/testkit"
)

func TestHealthCheckIsTrue(t *testing.T) {
	e := testkit.New(t)
	r := e.Call("GET", "/utils/health-check", "", nil)
	testkit.Expect(t, r, 200, "")
	if strings.TrimSpace(string(r.Body)) != "true" {
		t.Fatalf("body = %s", r.Body)
	}
}

func TestDocsAndOpenAPI(t *testing.T) {
	e := testkit.New(t)
	for _, path := range []string{"/docs", "/sdoc"} {
		testkit.Expect(t, e.Do("GET", path, "", "", ""), 200, "")
	}
	r := e.Do("GET", "/sdoc", "", "", "")
	if !strings.Contains(string(r.Body), "elysiajs") || !strings.Contains(string(r.Body), "OAuth2PasswordBearer") {
		t.Fatal("scalar page lacks theme or security scheme")
	}
	spec := e.Do("GET", "/api/v1/openapi.json", "", "", "")
	testkit.Expect(t, spec, 200, "")
	paths := spec.JSON()["paths"].(map[string]any)
	for _, p := range []string{"/base/login/access-token", "/base/users/{id}/admin", "/sample/notes/{id}", "/private/jobs/ping"} {
		if _, ok := paths[p]; !ok {
			t.Fatalf("openapi misses %s", p)
		}
	}
}

func TestPrivateRoutesOnlyWhenLocal(t *testing.T) {
	local := testkit.New(t)
	r := local.Call("GET", "/private/ping", "", nil)
	testkit.Expect(t, r, 200, "")
	if r.JSON()["message"] != "private ok" {
		t.Fatalf("body = %s", r.Body)
	}
	prod := testkit.New(t, func(c *config.Config) {
		c.Environment, c.SecretKey, c.FirstSuperuserPassword = "production", "a-real-secret", "a-real-password"
	})
	for _, path := range []string{"/private/ping", "/private/users", "/private/jobs/ping"} {
		method := "POST"
		if path == "/private/ping" {
			method = "GET"
		}
		testkit.Expect(t, prod.Call(method, path, "", map[string]any{}), 404, "")
	}
	spec := prod.Do("GET", "/api/v1/openapi.json", "", "", "")
	if _, ok := spec.JSON()["paths"].(map[string]any)["/private/ping"]; ok {
		t.Fatal("openapi lists /private outside local")
	}
}

func TestPrivateCreateUser(t *testing.T) {
	e := testkit.New(t)
	body := map[string]any{"email": "p@example.com", "password": "password123", "full_name": "P"}
	r := e.Call("POST", "/private/users", "", body)
	testkit.Expect(t, r, 200, "")
	if r.JSON()["is_superuser"] != false || r.JSON()["email"] != "p@example.com" {
		t.Fatalf("body = %s", r.Body)
	}
	testkit.Expect(t, e.Call("POST", "/private/users", "", body), 400, "The user with this email already exists in the system.")
	testkit.Expect(t, e.Call("POST", "/private/users", "", map[string]any{"email": "q@example.com", "password": "short"}), 422, "")
}

func TestPrivateJobsPing(t *testing.T) {
	e := testkit.New(t)
	r := e.Call("POST", "/private/jobs/ping?message=hello", "", nil)
	testkit.Expect(t, r, 200, "")
	if r.JSON()["message"] != "hello" || r.JSON()["job_id"] == "" {
		t.Fatalf("body = %s", r.Body)
	}
	queued := e.Jobs.Jobs()
	if len(queued) != 1 || queued[0].Task != "ping" || queued[0].Args["message"] != "hello" || queued[0].ID != r.JSON()["job_id"] {
		t.Fatalf("queued = %+v", queued)
	}
	if r := e.Call("POST", "/private/jobs/ping", "", nil); r.JSON()["message"] != "ping" {
		t.Fatalf("default message: %s", r.Body)
	}
	e.Jobs.FailWith(errors.New("connection refused"))
	r = e.Call("POST", "/private/jobs/ping", "", nil)
	testkit.Expect(t, r, 503, "")
	if !strings.HasPrefix(r.Detail(), "Redis unavailable: ") {
		t.Fatalf("detail = %q", r.Detail())
	}
}

func TestCORSAllowsConfiguredOriginsAndFrontendHost(t *testing.T) {
	e := testkit.New(t, func(c *config.Config) { c.FrontendHost = "http://dashboard.test/" })
	for _, origin := range []string{"http://app.test", "http://dashboard.test"} {
		req := httptest.NewRequest("OPTIONS", "/api/v1/sample/notes", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
		rec := httptest.NewRecorder()
		e.H.ServeHTTP(rec, req)
		if rec.Header().Get("Access-Control-Allow-Origin") != origin || rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Fatalf("origin %s not allowed: %v", origin, rec.Header())
		}
	}
	req := httptest.NewRequest("GET", "/api/v1/utils/health-check", nil)
	req.Header.Set("Origin", "http://evil.test")
	rec := httptest.NewRecorder()
	e.H.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unlisted origin was allowed")
	}
}

func TestUnknownRouteIsJSON404(t *testing.T) {
	e := testkit.New(t)
	r := e.Do("GET", "/nope", "", "", "")
	testkit.Expect(t, r, 404, "Not Found")
}
