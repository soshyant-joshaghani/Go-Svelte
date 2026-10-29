package security_test

import (
	"strings"
	"testing"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/security"
)

func TestBcryptIsDollar2b(t *testing.T) {
	h, err := security.HashPassword("hunter2hunter2", 4)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$2b$04$") || !security.VerifyPassword("hunter2hunter2", h) || security.VerifyPassword("other", h) {
		t.Fatalf("hash %q", h)
	}
	// a long password is cut at 72 bytes, like the other kits
	long := strings.Repeat("a", 100)
	lh, err := security.HashPassword(long, 4)
	if err != nil || !security.VerifyPassword(long, lh) {
		t.Fatal("long password")
	}
	// a hash written by another kit (Python bcrypt, "$2b$") verifies here
	const foreign = "$2b$04$IW3BAD0vmdRxCgJC3FIyvuNV5Ny35juGm8TbyKh/.pfFNsZPeaYrG"
	if !security.VerifyPassword("foreign-pass", foreign) || security.VerifyPassword("wrong", foreign) {
		t.Fatal("foreign $2b$ hash")
	}
}
