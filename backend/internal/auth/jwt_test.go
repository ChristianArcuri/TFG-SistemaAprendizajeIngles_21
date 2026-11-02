package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-secret-with-at-least-32-chars-long"

func TestTokenIssuerGenerateAndParse(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)
	userID := uuid.New()

	tok, err := issuer.Generate(userID)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if tok.Token == "" {
		t.Fatal("empty token")
	}
	if !tok.ExpiresAt.After(time.Now()) {
		t.Fatal("token expires in the past")
	}

	claims, err := issuer.Parse(tok.Token)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("user id mismatch: got %s, want %s", claims.UserID, userID)
	}
	if claims.Subject != userID.String() {
		t.Fatalf("subject mismatch: got %s, want %s", claims.Subject, userID.String())
	}
}

func TestTokenIssuerRejectsTamperedToken(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)
	tok, err := issuer.Generate(uuid.New())
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	tampered := tok.Token + "x"
	if _, err := issuer.Parse(tampered); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestTokenIssuerRejectsWrongSecret(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)
	tok, err := issuer.Generate(uuid.New())
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	other := NewTokenIssuer("a-totally-different-secret-32-chars!", time.Hour)
	if _, err := other.Parse(tok.Token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestTokenIssuerRejectsExpiredToken(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, -time.Hour)
	tok, err := issuer.Generate(uuid.New())
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if _, err := issuer.Parse(tok.Token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for expired, got %v", err)
	}
}

func TestTokenIssuerRejectsAlgNone(t *testing.T) {
	issuer := NewTokenIssuer(testSecret, time.Hour)

	claims := Claims{
		UserID: uuid.New(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "tfg-sai",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	t1 := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := t1.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none: %v", err)
	}

	if _, err := issuer.Parse(signed); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for alg=none, got %v", err)
	}
}
