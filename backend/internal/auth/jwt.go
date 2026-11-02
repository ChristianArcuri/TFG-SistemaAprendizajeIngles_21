package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const tokenIssuer = "tfg-sai"

// Claims es el payload del JWT del backend.
// Subject siempre va a ser el UUID del usuario en string, redundante con
// UserID pero util para tooling externo (jwt.io, debug).
type Claims struct {
	UserID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

// TokenIssuer emite y valida JWTs firmados con HS256.
type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenIssuer construye el emisor con el secreto de la config.
func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret), ttl: ttl}
}

// IssuedToken es el resultado de Generate: el string firmado y su expiracion.
type IssuedToken struct {
	Token     string
	ExpiresAt time.Time
}

// Generate emite un JWT firmado para el usuario indicado.
func (t *TokenIssuer) Generate(userID uuid.UUID) (IssuedToken, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(t.ttl)

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(t.secret)
	if err != nil {
		return IssuedToken{}, fmt.Errorf("sign jwt: %w", err)
	}
	return IssuedToken{Token: signed, ExpiresAt: expiresAt}, nil
}

// Parse valida un token y devuelve los claims. Mapea cualquier error a
// ErrInvalidToken para no filtrar detalles al middleware.
func (t *TokenIssuer) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(tk *jwt.Token) (any, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return t.secret, nil
	}, jwt.WithIssuer(tokenIssuer), jwt.WithLeeway(30*time.Second))

	if err != nil || parsed == nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UserID == uuid.Nil {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
