// Package httpx holds the small HTTP helpers shared by every module.
package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/apierr"
)

type Message struct {
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// DecodeJSON parses a JSON body; failures become 422 {"detail": "..."}.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(dest); err != nil {
		return apierr.Validation("Invalid request body: " + err.Error())
	}
	return nil
}

// ParseUUID validates a path id and returns its canonical lowercase form.
func ParseUUID(raw string) (string, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return "", apierr.Validation("Invalid UUID: " + raw)
	}
	return id.String(), nil
}

// QueryInt reads an optional integer query parameter; a bad value is 422.
func QueryInt(r *http.Request, key string, def int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, apierr.Validation(fmt.Sprintf("%s: value is not a valid integer", key))
	}
	return n, nil
}

// CheckLen counts characters, like pydantic min_length / max_length.
func CheckLen(field, value string, min, max int) error {
	n := utf8.RuneCountInString(value)
	if n < min {
		return apierr.Validation(fmt.Sprintf("%s: must have at least %d character(s)", field, min))
	}
	if n > max {
		return apierr.Validation(fmt.Sprintf("%s: must have at most %d characters", field, max))
	}
	return nil
}

// CheckEmail is a light shape check (a@b, no whitespace) plus the 255 limit.
func CheckEmail(field, value string) error {
	if err := CheckLen(field, value, 3, 255); err != nil {
		return err
	}
	local, domain, ok := strings.Cut(value, "@")
	if !ok || local == "" || domain == "" || strings.Contains(domain, "@") ||
		strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return apierr.Validation(field + ": value is not a valid email address")
	}
	return nil
}
