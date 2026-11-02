// Package users implementa el dominio de usuarios y perfiles.
//
// Capas:
//   - models.go      tipos del dominio y DTOs de entrada/salida
//   - repository.go  acceso a Postgres (pgx)
//   - service.go     reglas de negocio (registro, login, perfil)
//   - handler.go     handlers HTTP finos
//   - validation.go  validacion adicional fuera de los binding tags
package users

import (
	"time"

	"github.com/google/uuid"
)

// User representa el usuario autenticable.
type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Profile es la informacion 1:1 con User (nivel de ingles, rol, stack, objetivo).
type Profile struct {
	UserID       uuid.UUID `json:"userId"`
	EnglishLevel string    `json:"englishLevel,omitempty"`
	Role         string    `json:"role,omitempty"`
	MainStack    string    `json:"mainStack,omitempty"`
	Goal         string    `json:"goal,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// MeResponse es lo que devuelve GET /api/v1/me.
type MeResponse struct {
	User    User     `json:"user"`
	Profile *Profile `json:"profile,omitempty"`
}

// AuthResponse es lo que devuelven register y login.
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	User      User      `json:"user"`
}

// RegisterInput valida lo minimo en el binding; el service hace el resto.
type RegisterInput struct {
	Email    string `json:"email"    binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	FullName string `json:"fullName" binding:"omitempty,min=2,max=150"`
}

// LoginInput se usa en POST /auth/login.
type LoginInput struct {
	Email    string `json:"email"    binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,max=72"`
}

// UpdateProfileInput se usa en PUT /me/profile. Todos los campos son opcionales:
// la combinacion de los enviados se persiste con UPSERT.
type UpdateProfileInput struct {
	EnglishLevel string `json:"englishLevel" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2 unknown"`
	Role         string `json:"role"         binding:"omitempty,max=100"`
	MainStack    string `json:"mainStack"    binding:"omitempty,max=150"`
	Goal         string `json:"goal"         binding:"omitempty,max=150"`
}
