// Package system serves /utils/health-check and, only when ENVIRONMENT=local,
// the /private routes.
package system

import (
	"net/http"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/apierr"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/httpx"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/jobs"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/base/users"
)

type privateUserCreate struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	FullName *string `json:"full_name"`
}

// Routes registers the system routes under prefix. local enables /private.
func Routes(mux *http.ServeMux, prefix string, local bool, svc *users.Service, queue jobs.Queue) {
	mux.HandleFunc("GET "+prefix+"/utils/health-check", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, true)
	})
	if !local {
		return
	}

	mux.HandleFunc("GET "+prefix+"/private/ping", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, httpx.Message{Message: "private ok"})
	})

	mux.HandleFunc("POST "+prefix+"/private/users", func(w http.ResponseWriter, r *http.Request) {
		var in privateUserCreate
		if err := httpx.DecodeJSON(w, r, &in); err != nil {
			apierr.Write(w, err)
			return
		}
		user, err := svc.Create(r.Context(), users.UserCreate{Email: in.Email, Password: in.Password, FullName: in.FullName})
		if err != nil {
			apierr.Write(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, user.Public())
	})

	mux.HandleFunc("POST "+prefix+"/private/jobs/ping", func(w http.ResponseWriter, r *http.Request) {
		message := r.URL.Query().Get("message")
		if message == "" {
			message = "ping"
		}
		id, err := queue.Enqueue(r.Context(), "ping", map[string]any{"message": message})
		if err != nil {
			apierr.Write(w, apierr.Unavailable("Redis unavailable: "+err.Error()))
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"job_id": id, "message": message})
	})
}
