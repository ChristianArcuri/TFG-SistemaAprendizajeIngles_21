package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/remotetalk/backend/internal/auth"
)

// TokenIssuer es la dependencia de auth que emite JWTs. Definida como
// interface chica para que los tests del service puedan reemplazarla.
type TokenIssuer interface {
	Generate(userID uuid.UUID) (auth.IssuedToken, error)
}

// Service implementa la logica de registro, login y gestion de perfil.
// Los handlers HTTP solo lo invocan; nunca acceden a Repository directamente.
type Service struct {
	repo  Repository
	token TokenIssuer
}

// NewService construye el service.
func NewService(repo Repository, token TokenIssuer) *Service {
	return &Service{repo: repo, token: token}
}

// Register crea un usuario nuevo y devuelve su JWT inicial.
// Errores posibles: ErrEmailAlreadyExists, errores propios de auth.
func (s *Service) Register(ctx context.Context, in RegisterInput) (AuthResponse, error) {
	email := normalizeEmail(in.Email)

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, email, hash, trim(in.FullName))
	if err != nil {
		return AuthResponse{}, err
	}

	return s.issueAuthResponse(user)
}

// Login valida credenciales y devuelve un JWT. Cualquier fallo de
// busqueda o comparacion se convierte en ErrInvalidCredentials para no
// filtrar si el email existe o no.
func (s *Service) Login(ctx context.Context, in LoginInput) (AuthResponse, error) {
	email := normalizeEmail(in.Email)

	user, hash, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return AuthResponse{}, auth.ErrInvalidCredentials
		}
		return AuthResponse{}, err
	}

	if err := auth.ComparePassword(hash, in.Password); err != nil {
		return AuthResponse{}, auth.ErrInvalidCredentials
	}

	return s.issueAuthResponse(user)
}

// GetMe devuelve el usuario autenticado y su perfil (si existe).
func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (MeResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return MeResponse{}, err
	}

	resp := MeResponse{User: user}

	profile, err := s.repo.GetProfileByUserID(ctx, userID)
	switch {
	case err == nil:
		resp.Profile = &profile
	case errors.Is(err, ErrProfileNotFound):
		// Sin perfil cargado todavia: respuesta valida con Profile == nil.
	default:
		return MeResponse{}, err
	}

	return resp, nil
}

// UpdateProfile crea o actualiza el perfil del usuario.
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (Profile, error) {
	in.EnglishLevel = trim(in.EnglishLevel)
	in.Role = trim(in.Role)
	in.MainStack = trim(in.MainStack)
	in.Goal = trim(in.Goal)
	return s.repo.UpsertProfile(ctx, userID, in)
}

func (s *Service) issueAuthResponse(user User) (AuthResponse, error) {
	tk, err := s.token.Generate(user.ID)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("issue token: %w", err)
	}
	return AuthResponse{
		Token:     tk.Token,
		ExpiresAt: tk.ExpiresAt,
		User:      user,
	}, nil
}
