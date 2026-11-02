package auth

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost balancea seguridad y latencia. 12 es un valor razonable en 2026:
// ~250 ms en CPU moderna, todavia rapido para login y suficientemente caro
// para frenar fuerza bruta.
const bcryptCost = 12

// MaxPasswordBytes evita el corte silencioso de bcrypt a 72 bytes:
// si la entrada los supera, devolvemos un error explicito en lugar de
// ignorar caracteres y abrir un agujero de seguridad.
const MaxPasswordBytes = 72

// ErrPasswordTooLong se devuelve cuando el password supera el limite de bcrypt.
var ErrPasswordTooLong = errors.New("auth: password exceeds maximum length")

// HashPassword devuelve el hash bcrypt del password en texto plano.
func HashPassword(password string) (string, error) {
	if utf8.RuneCountInString(password) > MaxPasswordBytes {
		return "", ErrPasswordTooLong
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword compara un hash bcrypt contra un password en texto plano.
// Devuelve nil si coincide; si no, ErrInvalidCredentials para que el caller
// pueda mapear sin exponer detalles internos.
func ComparePassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}
