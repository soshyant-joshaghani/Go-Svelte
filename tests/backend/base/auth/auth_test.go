package auth_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/security"
	"github.com/soshyant-joshaghani/go-svelte/tests/backend/testkit"
)

func TestLoginAndMe(t *testing.T) {
	e := testkit.New(t)
	r := e.Login("admin@example.com", "adminpass123")
	testkit.Expect(t, r, 200, "")
	if r.JSON()["token_type"] != "bearer" || r.JSON()["access_token"] == "" {
		t.Fatalf("body = %s", r.Body)
	}
	me := e.Call("GET", "/base/login/me", r.JSON()["access_token"].(string), nil)
	testkit.Expect(t, me, 200, "")
	got := me.JSON()
	if got["email"] != "admin@example.com" || got["is_superuser"] != true || got["is_active"] != true || got["full_name"] != nil {
		t.Fatalf("me = %s", me.Body)
	}
	if len(got) != 5 {
		t.Fatalf("UserPublic must have exactly 5 keys: %s", me.Body)
	}
}

func TestLoginFailures(t *testing.T) {
	e := testkit.New(t)
	testkit.Expect(t, e.Login("admin@example.com", "nope"), 400, "Incorrect email or password")
	testkit.Expect(t, e.Login("nobody@example.com", "adminpass123"), 400, "Incorrect email or password")
	testkit.Expect(t, e.Do("POST", "/api/v1/base/login/access-token", "", "application/json", `{"username":"a","password":"b"}`), 422, "")
	id, _ := e.NewUser("off@example.com")
	testkit.Expect(t, e.Call("PATCH", "/base/users/"+id+"/admin", e.Admin(), map[string]any{"is_active": false}), 200, "")
	testkit.Expect(t, e.Login("off@example.com", "password123"), 400, "Inactive user")
}

func TestBearerErrors(t *testing.T) {
	e := testkit.New(t)
	r := e.Call("GET", "/base/login/me", "", nil)
	testkit.Expect(t, r, 401, "Not authenticated")
	if r.Header.Get("WWW-Authenticate") != "Bearer" {
		t.Fatal("missing WWW-Authenticate")
	}
	r = e.Call("GET", "/base/login/me", "not-a-jwt", nil)
	testkit.Expect(t, r, 401, "Could not validate credentials")
	if r.Header.Get("WWW-Authenticate") != "Bearer" {
		t.Fatal("missing WWW-Authenticate")
	}
	// expired, unknown user, wrong secret
	expired, _ := security.CreateTokenWithExp(uuid.NewString(), "test-secret-key", time.Now().Add(-time.Minute))
	testkit.Expect(t, e.Call("GET", "/base/login/me", expired, nil), 401, "Could not validate credentials")
	unknown, _ := security.CreateAccessToken(uuid.NewString(), "test-secret-key", 5)
	testkit.Expect(t, e.Call("GET", "/base/login/me", unknown, nil), 401, "Could not validate credentials")
	forged, _ := security.CreateAccessToken(uuid.NewString(), "other-secret", 5)
	testkit.Expect(t, e.Call("GET", "/base/login/me", forged, nil), 401, "Could not validate credentials")
	req := httptest.NewRequest("GET", "/api/v1/base/login/me", nil)
	req.Header.Set("Authorization", "Basic abc")
	rec := httptest.NewRecorder()
	e.H.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("basic scheme = %d", rec.Code)
	}
}

func TestTokenClaims(t *testing.T) {
	e := testkit.New(t)
	tok := e.Admin()
	sub, err := security.DecodeAccessToken(tok, "test-secret-key")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(sub); err != nil {
		t.Fatalf("sub %q is not a uuid", sub)
	}
	parts := strings.Split(tok, ".")
	header, _ := decodeSegment(parts[0])
	if header["alg"] != "HS256" {
		t.Fatalf("alg = %v", header["alg"])
	}
}

func decodeSegment(seg string) (map[string]any, error) {
	raw, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	return m, json.Unmarshal(raw, &m)
}
