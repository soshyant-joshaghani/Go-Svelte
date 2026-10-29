// Package apierr renders every error as {"detail": "<message>"}.
package apierr

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Error struct {
	Status int
	Detail string
	// Bearer adds "WWW-Authenticate: Bearer" (401 responses).
	Bearer bool
}

func (e *Error) Error() string { return e.Detail }

func New(status int, detail string) *Error { return &Error{Status: status, Detail: detail} }

func BadRequest(detail string) *Error  { return New(http.StatusBadRequest, detail) }
func Forbidden(detail string) *Error   { return New(http.StatusForbidden, detail) }
func NotFound(detail string) *Error    { return New(http.StatusNotFound, detail) }
func Conflict(detail string) *Error    { return New(http.StatusConflict, detail) }
func Validation(detail string) *Error  { return New(http.StatusUnprocessableEntity, detail) }
func Unavailable(detail string) *Error { return New(http.StatusServiceUnavailable, detail) }

func NotAuthenticated() *Error {
	return &Error{Status: http.StatusUnauthorized, Detail: "Not authenticated", Bearer: true}
}

func InvalidCredentials() *Error {
	return &Error{Status: http.StatusUnauthorized, Detail: "Could not validate credentials", Bearer: true}
}

// Internal logs the real error and answers with a generic 500.
func Internal(err error) *Error {
	log.Printf("internal error: %v", err)
	return New(http.StatusInternalServerError, "Internal Server Error")
}

// Write renders err. Anything that is not an *Error becomes a 500.
func Write(w http.ResponseWriter, err error) {
	var api *Error
	if !errors.As(err, &api) {
		api = Internal(err)
	}
	if api.Bearer {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(api.Status)
	_ = json.NewEncoder(w).Encode(map[string]string{"detail": api.Detail})
}
