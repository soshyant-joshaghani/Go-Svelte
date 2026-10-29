package sample_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/soshyant-joshaghani/go-svelte/tests/backend/testkit"
)

func TestSampleRoot(t *testing.T) {
	e := testkit.New(t)
	r := e.Call("GET", "/sample", "", nil)
	testkit.Expect(t, r, 200, "")
	if !strings.HasPrefix(r.JSON()["message"].(string), "Sample module") {
		t.Fatalf("body = %s", r.Body)
	}
	testkit.Expect(t, e.Call("GET", "/sample/notes", "", nil), 401, "Not authenticated")
}

func TestNotesCRUD(t *testing.T) {
	e := testkit.New(t)
	uid, tok := e.NewUser("n@example.com")
	r := e.Call("POST", "/sample/notes", tok, map[string]any{"title": "  Hello  ", "content": "  body  "})
	testkit.Expect(t, r, 201, "")
	note := r.JSON()
	if note["title"] != "Hello" || note["content"] != "body" || note["owner_id"] != uid || len(note) != 6 {
		t.Fatalf("created = %s", r.Body)
	}
	for _, k := range []string{"created_at", "updated_at"} {
		ts, err := time.Parse(time.RFC3339Nano, note[k].(string))
		if err != nil || !strings.HasSuffix(note[k].(string), "Z") || ts.IsZero() {
			t.Fatalf("%s = %v", k, note[k])
		}
	}
	id := note["id"].(string)

	r = e.Call("GET", "/sample/notes/"+id, tok, nil)
	testkit.Expect(t, r, 200, "")
	list := e.Call("GET", "/sample/notes", tok, nil)
	if l := list.List(); len(l) != 1 || l[0].(map[string]any)["id"] != id {
		t.Fatalf("list = %s", list.Body)
	}
	r = e.Call("PATCH", "/sample/notes/"+id, tok, map[string]any{"title": "  Renamed "})
	testkit.Expect(t, r, 200, "")
	if r.JSON()["title"] != "Renamed" || r.JSON()["content"] != "body" {
		t.Fatalf("patched = %s", r.Body)
	}
	if l := e.Call("GET", "/sample/notes", tok, nil).List(); l[0].(map[string]any)["title"] != "Renamed" {
		t.Fatal("list not refreshed after update")
	}
	r = e.Call("DELETE", "/sample/notes/"+id, tok, nil)
	testkit.Expect(t, r, 204, "")
	if len(r.Body) != 0 {
		t.Fatalf("204 body = %q", r.Body)
	}
	testkit.Expect(t, e.Call("GET", "/sample/notes/"+id, tok, nil), 404, "Note not found")
	if l := e.Call("GET", "/sample/notes", tok, nil).List(); len(l) != 0 {
		t.Fatalf("list after delete = %v", l)
	}
	if string(e.Call("GET", "/sample/notes", tok, nil).Body) == "null\n" {
		t.Fatal("empty list must be []")
	}
}

func TestNoteValidation(t *testing.T) {
	e := testkit.New(t)
	_, tok := e.NewUser("v@example.com")
	testkit.Expect(t, e.Call("POST", "/sample/notes", tok, map[string]any{"title": "   "}), 422, "Title cannot be empty")
	testkit.Expect(t, e.Call("POST", "/sample/notes", tok, map[string]any{"title": ""}), 422, "")
	testkit.Expect(t, e.Call("POST", "/sample/notes", tok, map[string]any{"content": "no title"}), 422, "")
	testkit.Expect(t, e.Call("POST", "/sample/notes", tok, map[string]any{"title": strings.Repeat("x", 256)}), 422, "")
	testkit.Expect(t, e.Call("POST", "/sample/notes", tok, map[string]any{"title": "ok", "content": strings.Repeat("x", 10001)}), 422, "")
	testkit.Expect(t, e.Do("POST", "/api/v1/sample/notes", tok, "application/json", "{not json"), 422, "")
	testkit.Expect(t, e.Call("POST", "/sample/notes", tok, map[string]any{"title": strings.Repeat("é", 255)}), 201, "")
	id := e.Call("POST", "/sample/notes", tok, map[string]any{"title": "t"}).JSON()["id"].(string)
	testkit.Expect(t, e.Call("PATCH", "/sample/notes/"+id, tok, map[string]any{"title": "   "}), 422, "Title cannot be empty")
	testkit.Expect(t, e.Call("GET", "/sample/notes/not-a-uuid", tok, nil), 422, "")
}

func TestNoteOwnerIsolation(t *testing.T) {
	e := testkit.New(t)
	_, a := e.NewUser("a@example.com")
	_, b := e.NewUser("b@example.com")
	id := e.Call("POST", "/sample/notes", a, map[string]any{"title": "mine"}).JSON()["id"].(string)
	const denied = "Not allowed to access this note"
	testkit.Expect(t, e.Call("GET", "/sample/notes/"+id, b, nil), 403, denied)
	testkit.Expect(t, e.Call("PATCH", "/sample/notes/"+id, b, map[string]any{"title": "x"}), 403, denied)
	testkit.Expect(t, e.Call("DELETE", "/sample/notes/"+id, b, nil), 403, denied)
	testkit.Expect(t, e.Call("GET", "/sample/notes/"+uuid.NewString(), b, nil), 404, "Note not found")
	if l := e.Call("GET", "/sample/notes", b, nil).List(); len(l) != 0 {
		t.Fatal("b sees a's notes")
	}
}

func TestNoteCacheKeysAndInvalidation(t *testing.T) {
	e := testkit.New(t)
	uid, tok := e.NewUser("c@example.com")
	id := e.Call("POST", "/sample/notes", tok, map[string]any{"title": "t"}).JSON()["id"].(string)
	wantNote := "sample:notes:v1:note:" + uid + ":" + id
	wantList := "sample:notes:v1:list:" + uid
	keys := strings.Join(e.Cache.Keys(), ",")
	if !strings.Contains(keys, wantNote) || strings.Contains(keys, wantList) {
		t.Fatalf("after create keys = %v", e.Cache.Keys())
	}
	e.Call("GET", "/sample/notes", tok, nil)
	if !strings.Contains(strings.Join(e.Cache.Keys(), ","), wantList) {
		t.Fatalf("list not cached: %v", e.Cache.Keys())
	}
	// reads go through the cache: no database read for a cached note
	before := e.Notes.Gets
	testkit.Expect(t, e.Call("GET", "/sample/notes/"+id, tok, nil), 200, "")
	if e.Notes.Gets != before {
		t.Fatal("cached GET hit the database")
	}
	// update loads from the database and drops the list key
	testkit.Expect(t, e.Call("PATCH", "/sample/notes/"+id, tok, map[string]any{"content": "c"}), 200, "")
	if e.Notes.Gets != before+1 {
		t.Fatal("update must load from the database")
	}
	if strings.Contains(strings.Join(e.Cache.Keys(), ","), wantList) {
		t.Fatal("update left the list key")
	}
	testkit.Expect(t, e.Call("DELETE", "/sample/notes/"+id, tok, nil), 204, "")
	if strings.Contains(strings.Join(e.Cache.Keys(), ","), wantNote) {
		t.Fatal("delete left the note key")
	}
}

type deadCache struct{}

func (deadCache) Get(context.Context, string) (string, bool)         { return "", false }
func (deadCache) Set(context.Context, string, string, time.Duration) {}
func (deadCache) Delete(context.Context, string)                     {}
func (deadCache) DeletePrefix(context.Context, string)               {}

func TestNotesWorkWithoutCache(t *testing.T) {
	e := testkit.New(t)
	e.Srv.Cache = deadCache{}
	e.H = e.Srv.Handler()
	_, tok := e.NewUser("d@example.com")
	id := e.Call("POST", "/sample/notes", tok, map[string]any{"title": "t"}).JSON()["id"].(string)
	testkit.Expect(t, e.Call("GET", "/sample/notes/"+id, tok, nil), 200, "")
	if l := e.Call("GET", "/sample/notes", tok, nil).List(); len(l) != 1 {
		t.Fatalf("list = %v", l)
	}
}
