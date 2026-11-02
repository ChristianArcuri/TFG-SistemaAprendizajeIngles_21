package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestHashAndComparePassword(t *testing.T) {
	hash, err := HashPassword("Sup3r-Secret!")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if hash == "" {
		t.Fatal("hash is empty")
	}

	if err := ComparePassword(hash, "Sup3r-Secret!"); err != nil {
		t.Fatalf("expected match, got %v", err)
	}

	if err := ComparePassword(hash, "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestHashPasswordTooLong(t *testing.T) {
	long := strings.Repeat("a", MaxPasswordBytes+1)
	if _, err := HashPassword(long); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("expected ErrPasswordTooLong, got %v", err)
	}
}
