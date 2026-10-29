package users_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/security"
	"github.com/soshyant-joshaghani/go-svelte/tests/backend/testkit"
)

func TestUsersAdminCRUD(t *testing.T) {
	e := testkit.New(t)
	admin := e.Admin()
	r := e.Call("POST", "/base/users/admin", admin, map[string]any{"email": "b@example.com", "password": "password123", "full_name": "Bee"})
	testkit.Expect(t, r, 200, "")
	id := r.JSON()["id"].(string)
	if r.JSON()["is_active"] != true || r.JSON()["is_superuser"] != false || r.JSON()["full_name"] != "Bee" {
		t.Fatalf("created = %s", r.Body)
	}
	if _, ok := r.JSON()["hashed_password"]; ok {
		t.Fatal("hashed_password leaked")
	}
	stored := e.Users.Rows[id].HashedPassword
	if !strings.HasPrefix(stored, "$2b$") || !security.VerifyPassword("password123", stored) {
		t.Fatalf("stored hash %q", stored)
	}
	testkit.Expect(t, e.Call("POST", "/base/users/admin", admin, map[string]any{"email": "b@example.com", "password": "password123"}), 400, "The user with this email already exists in the system.")
	testkit.Expect(t, e.Call("POST", "/base/users/admin", admin, map[string]any{"email": "bad", "password": "password123"}), 422, "")
	testkit.Expect(t, e.Call("POST", "/base/users/admin", admin, map[string]any{"email": "c@example.com", "password": "short"}), 422, "")

	list := e.Call("GET", "/base/users/admin?skip=0&limit=100", admin, nil)
	testkit.Expect(t, list, 200, "")
	if list.JSON()["count"].(float64) != 2 || len(list.JSON()["data"].([]any)) != 2 {
		t.Fatalf("list = %s", list.Body)
	}
	if first := list.JSON()["data"].([]any)[0].(map[string]any)["email"]; first != "admin@example.com" {
		t.Fatalf("not ordered by email: %v", first)
	}
	page := e.Call("GET", "/base/users/admin?skip=1&limit=1", admin, nil)
	if d := page.JSON()["data"].([]any); len(d) != 1 || d[0].(map[string]any)["email"] != "b@example.com" || page.JSON()["count"].(float64) != 2 {
		t.Fatalf("page = %s", page.Body)
	}
	testkit.Expect(t, e.Call("GET", "/base/users/admin?limit=abc", admin, nil), 422, "")

	r = e.Call("PATCH", "/base/users/"+id+"/admin", admin, map[string]any{"full_name": "Renamed", "password": "newpassword1"})
	testkit.Expect(t, r, 200, "")
	if r.JSON()["full_name"] != "Renamed" || r.JSON()["email"] != "b@example.com" {
		t.Fatalf("patched = %s", r.Body)
	}
	testkit.Expect(t, e.Login("b@example.com", "newpassword1"), 200, "")
	testkit.Expect(t, e.Login("b@example.com", "password123"), 400, "")
	// null clears full_name, absent leaves it
	testkit.Expect(t, e.Call("PATCH", "/base/users/"+id+"/admin", admin, map[string]any{"is_active": true}), 200, "")
	if e.Users.Rows[id].FullName == nil {
		t.Fatal("absent full_name must not clear it")
	}
	r = e.Call("PATCH", "/base/users/"+id+"/admin", admin, map[string]any{"full_name": nil})
	if r.JSON()["full_name"] != nil {
		t.Fatalf("null should clear: %s", r.Body)
	}
	testkit.Expect(t, e.Call("PATCH", "/base/users/"+id+"/admin", admin, map[string]any{"email": "admin@example.com"}), 409, "User with this email already exists")
	testkit.Expect(t, e.Call("PATCH", "/base/users/"+uuid.NewString()+"/admin", admin, map[string]any{}), 404, "The user with this id does not exist in the system")

	testkit.Expect(t, e.Call("GET", "/base/users/"+uuid.NewString()+"/admin", admin, nil), 404, "User not found")
	testkit.Expect(t, e.Call("GET", "/base/users/not-a-uuid/admin", admin, nil), 422, "")
	testkit.Expect(t, e.Call("DELETE", "/base/users/"+uuid.NewString()+"/admin", admin, nil), 404, "User not found")
	r = e.Call("DELETE", "/base/users/"+id+"/admin", admin, nil)
	testkit.Expect(t, r, 200, "")
	if r.JSON()["message"] != "User deleted successfully" {
		t.Fatalf("body = %s", r.Body)
	}
	if _, ok := e.Users.Rows[id]; ok {
		t.Fatal("user still stored")
	}
}

func TestUserRoutesPrivileges(t *testing.T) {
	e := testkit.New(t)
	adminID := ""
	for id, u := range e.Users.Rows {
		if u.IsSuperuser {
			adminID = id
		}
	}
	uid, utok := e.NewUser("u@example.com")
	const priv = "The user doesn't have enough privileges"
	testkit.Expect(t, e.Call("GET", "/base/users/admin", utok, nil), 403, priv)
	testkit.Expect(t, e.Call("POST", "/base/users/admin", utok, map[string]any{"email": "x@example.com", "password": "password123"}), 403, priv)
	testkit.Expect(t, e.Call("PATCH", "/base/users/"+uid+"/admin", utok, map[string]any{"full_name": "x"}), 403, priv)
	testkit.Expect(t, e.Call("DELETE", "/base/users/"+uid+"/admin", utok, nil), 403, priv)
	testkit.Expect(t, e.Call("GET", "/base/users/"+uid+"/admin", utok, nil), 200, "")
	testkit.Expect(t, e.Call("GET", "/base/users/"+adminID+"/admin", utok, nil), 403, priv)
	testkit.Expect(t, e.Call("GET", "/base/users/admin", "", nil), 401, "Not authenticated")
	testkit.Expect(t, e.Call("DELETE", "/base/users/"+adminID+"/admin", e.Admin(), nil), 403, "Super users are not allowed to delete themselves")
}

func TestDeleteUserDeletesNotesFirst(t *testing.T) {
	e := testkit.New(t)
	uid, utok := e.NewUser("n@example.com")
	testkit.Expect(t, e.Call("POST", "/sample/notes", utok, map[string]any{"title": "keep?"}), 201, "")
	if len(e.Notes.Rows) != 1 {
		t.Fatal("note not stored")
	}
	testkit.Expect(t, e.Call("DELETE", "/base/users/"+uid+"/admin", e.Admin(), nil), 200, "")
	if len(e.Notes.Rows) != 0 {
		t.Fatal("notes of the deleted user remain")
	}
}
