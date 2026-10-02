package users

import (
	"net/http"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/apierr"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/httpx"
)

// Guards resolve the caller. The auth module provides them; they are injected
// so that users and auth do not import each other.
type Guards struct {
	User      func(r *http.Request) (*User, error)
	Superuser func(r *http.Request) (*User, error)
}

// Routes registers /base/users/admin and /base/users/{id}/admin under prefix.
func Routes(mux *http.ServeMux, prefix string, svc *Service, g Guards) {
	fail := func(w http.ResponseWriter, err error) { apierr.Write(w, err) }

	mux.HandleFunc("GET "+prefix+"/base/users/admin", func(w http.ResponseWriter, r *http.Request) {
		if _, err := g.Superuser(r); err != nil {
			fail(w, err)
			return
		}
		skip, err := httpx.QueryInt(r, "skip", 0)
		if err != nil {
			fail(w, err)
			return
		}
		limit, err := httpx.QueryInt(r, "limit", 100)
		if err != nil {
			fail(w, err)
			return
		}
		out, err := svc.List(r.Context(), skip, limit)
		if err != nil {
			fail(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, out)
	})

	mux.HandleFunc("POST "+prefix+"/base/users/admin", func(w http.ResponseWriter, r *http.Request) {
		if _, err := g.Superuser(r); err != nil {
			fail(w, err)
			return
		}
		var in UserCreate
		if err := httpx.DecodeJSON(w, r, &in); err != nil {
			fail(w, err)
			return
		}
		user, err := svc.Create(r.Context(), in)
		if err != nil {
			fail(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, user.Public())
	})

	mux.HandleFunc("GET "+prefix+"/base/users/{id}/admin", func(w http.ResponseWriter, r *http.Request) {
		current, err := g.User(r)
		if err != nil {
			fail(w, err)
			return
		}
		id, err := httpx.ParseUUID(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		user, err := svc.GetFor(r.Context(), current, id)
		if err != nil {
			fail(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, user.Public())
	})

	mux.HandleFunc("PATCH "+prefix+"/base/users/{id}/admin", func(w http.ResponseWriter, r *http.Request) {
		if _, err := g.Superuser(r); err != nil {
			fail(w, err)
			return
		}
		id, err := httpx.ParseUUID(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		var in UserUpdate
		if err := httpx.DecodeJSON(w, r, &in); err != nil {
			fail(w, err)
			return
		}
		user, err := svc.Update(r.Context(), id, in)
		if err != nil {
			fail(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, user.Public())
	})

	mux.HandleFunc("DELETE "+prefix+"/base/users/{id}/admin", func(w http.ResponseWriter, r *http.Request) {
		current, err := g.Superuser(r)
		if err != nil {
			fail(w, err)
			return
		}
		id, err := httpx.ParseUUID(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		if err := svc.Delete(r.Context(), current, id); err != nil {
			fail(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, httpx.Message{Message: "User deleted successfully"})
	})
}
