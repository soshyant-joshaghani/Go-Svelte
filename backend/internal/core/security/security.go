// Package security has bcrypt password hashing and HS256 access tokens.
package security

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// bcrypt only looks at the first 72 bytes.
func prepare(password string) []byte {
	b := []byte(password)
	if len(b) > 72 {
		b = b[:72]
	}
	return b
}

// HashPassword returns a "$2b$" hash, the format Fast, Rust and .NET write.
// Go's bcrypt emits "$2a$"; the two prefixes are the same algorithm.
func HashPassword(password string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(prepare(password), cost)
	if err != nil {
		return "", err
	}
	return "$2b$" + strings.TrimPrefix(string(hash), "$2a$"), nil
}

func VerifyPassword(password, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), prepare(password)) == nil
}

func CreateAccessToken(subject, secret string, expireMinutes int) (string, error) {
	return CreateTokenWithExp(subject, secret, time.Now().Add(time.Duration(expireMinutes)*time.Minute))
}

// CreateTokenWithExp signs a token with an explicit expiry (tests use it).
func CreateTokenWithExp(subject, secret string, exp time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(exp),
	})
	return token.SignedString([]byte(secret))
}

// DecodeAccessToken returns the "sub" claim of a valid, unexpired HS256 token.
func DecodeAccessToken(token, secret string) (string, error) {
	var c jwt.RegisteredClaims
	parsed, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		return "", err
	}
	if !parsed.Valid || c.Subject == "" {
		return "", errors.New("invalid token")
	}
	return c.Subject, nil
}
