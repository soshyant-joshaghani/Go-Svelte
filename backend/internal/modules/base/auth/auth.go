// Package auth resolves the caller from "Authorization: Bearer <jwt>" and
// serves the login routes.
package auth

import (
	"net/http"
	"strings"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/apierr"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/security"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/base/users"
)

type Authenticator struct {
	Users         *users.Service
	Secret        string
	ExpireMinutes int
}

// bearerToken returns the raw token or 401 "Not authenticated".
func bearerToken(r *http.Request) (string, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	scheme, token, ok := strings.Cut(header, " ")
	token = strings.TrimSpace(token)
	if !ok || !strings.EqualFold(scheme, "bearer") || token == "" {
		return "", apierr.NotAuthenticated()
	}
	return token, nil
}

// CurrentUser resolves the caller.
//
// Missing header: 401 Not authenticated. Bad or expired token, or an unknown
// user: 401 Could not validate credentials. Inactive user: 400 Inactive user.
func (a *Authenticator) CurrentUser(r *http.Request) (*users.User, error) {
	token, err := bearerToken(r)
	if err != nil {
		return nil, err
	}
	sub, err := security.DecodeAccessToken(token, a.Secret)
	if err != nil {
		return nil, apierr.InvalidCredentials()
	}
	user, err := a.Users.GetByID(r.Context(), sub)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apierr.InvalidCredentials()
	}
	if !user.IsActive {
		return nil, apierr.BadRequest("Inactive user")
	}
	return user, nil
}

func (a *Authenticator) CurrentSuperuser(r *http.Request) (*users.User, error) {
	user, err := a.CurrentUser(r)
	if err != nil {
		return nil, err
	}
	if !user.IsSuperuser {
		return nil, users.Privileges()
	}
	return user, nil
}

// Guards adapts the authenticator for the users module.
func (a *Authenticator) Guards() users.Guards {
	return users.Guards{User: a.CurrentUser, Superuser: a.CurrentSuperuser}
}
