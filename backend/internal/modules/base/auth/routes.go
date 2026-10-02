package auth

import (
	"net/http"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/apierr"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/httpx"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/security"
)

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// Routes registers POST /base/login/access-token and GET /base/login/me under prefix.
func Routes(mux *http.ServeMux, prefix string, a *Authenticator) {
	mux.HandleFunc("POST "+prefix+"/base/login/access-token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			apierr.Write(w, apierr.Validation("Invalid form body: "+err.Error()))
			return
		}
		username, password := r.PostForm.Get("username"), r.PostForm.Get("password")
		if username == "" || password == "" {
			apierr.Write(w, apierr.Validation("username and password are required"))
			return
		}
		user, err := a.Users.Authenticate(r.Context(), username, password)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		token, err := security.CreateAccessToken(user.ID, a.Secret, a.ExpireMinutes)
		if err != nil {
			apierr.Write(w, apierr.Internal(err))
			return
		}
		httpx.WriteJSON(w, http.StatusOK, Token{AccessToken: token, TokenType: "bearer"})
	})

	mux.HandleFunc("GET "+prefix+"/base/login/me", func(w http.ResponseWriter, r *http.Request) {
		user, err := a.CurrentUser(r)
		if err != nil {
			apierr.Write(w, err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, user.Public())
	})
}
