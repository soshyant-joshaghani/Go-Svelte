package sample

import (
	"net/http"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/apierr"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/httpx"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/base/auth"
)

// Routes registers /sample and /sample/notes under prefix.
func Routes(mux *http.ServeMux, prefix string, svc *Service, a *auth.Authenticator) {
	mux.HandleFunc("GET "+prefix+"/sample", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, httpx.Message{
			Message: "Sample module — see /sample/notes for the canonical CRUD example",
		})
	})

	mux.HandleFunc("GET "+prefix+"/sample/notes", func(w http.ResponseWriter, r *http.Request) {
		user, err := a.CurrentUser(r)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		notes, err := svc.List(r.Context(), user)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, notes)
	})

	mux.HandleFunc("POST "+prefix+"/sample/notes", func(w http.ResponseWriter, r *http.Request) {
		user, err := a.CurrentUser(r)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		var in NoteCreate
		if err := httpx.DecodeJSON(w, r, &in); err != nil {
			apierr.Write(w, err)
			return
		}
		note, err := svc.Create(r.Context(), user, in)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusCreated, note)
	})

	mux.HandleFunc("GET "+prefix+"/sample/notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		user, err := a.CurrentUser(r)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		id, err := httpx.ParseUUID(r.PathValue("id"))
		if err != nil {
			apierr.Write(w, err)
			return
		}
		note, err := svc.Get(r.Context(), user, id)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, note)
	})

	mux.HandleFunc("PATCH "+prefix+"/sample/notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		user, err := a.CurrentUser(r)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		id, err := httpx.ParseUUID(r.PathValue("id"))
		if err != nil {
			apierr.Write(w, err)
			return
		}
		var in NoteUpdate
		if err := httpx.DecodeJSON(w, r, &in); err != nil {
			apierr.Write(w, err)
			return
		}
		note, err := svc.Update(r.Context(), user, id, in)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, note)
	})

	mux.HandleFunc("DELETE "+prefix+"/sample/notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		user, err := a.CurrentUser(r)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		id, err := httpx.ParseUUID(r.PathValue("id"))
		if err != nil {
			apierr.Write(w, err)
			return
		}
		if err := svc.Delete(r.Context(), user, id); err != nil {
			apierr.Write(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
