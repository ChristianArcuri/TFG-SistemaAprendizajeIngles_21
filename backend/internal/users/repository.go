package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository define las operaciones de persistencia que necesita el service.
// Definirla como interfaz permite que los tests del service usen un fake
// sin levantar Postgres.
type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash, fullName string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, string, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetProfileByUserID(ctx context.Context, userID uuid.UUID) (Profile, error)
	UpsertProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (Profile, error)
}

// PgRepository es la implementacion de Repository sobre Postgres con pgx.
//
// NOTA: las queries SQL viven tambien en internal/database/queries/users.sql
// para cuando integremos sqlc; mientras tanto se ejecutan a mano.
type PgRepository struct {
	db *pgxpool.Pool
}

// NewPgRepository construye el repository concreto.
func NewPgRepository(db *pgxpool.Pool) *PgRepository {
	return &PgRepository{db: db}
}

// uniqueViolation es el SQLSTATE de Postgres para violacion de UNIQUE.
const uniqueViolation = "23505"

// CreateUser inserta un usuario y devuelve la fila creada (sin password_hash).
func (r *PgRepository) CreateUser(ctx context.Context, email, passwordHash, fullName string) (User, error) {
	const q = `
		INSERT INTO users (email, password_hash, full_name)
		VALUES ($1, $2, NULLIF($3, ''))
		RETURNING id, email, COALESCE(full_name, ''), created_at, updated_at
	`
	var u User
	err := r.db.QueryRow(ctx, q, email, passwordHash, fullName).Scan(
		&u.ID, &u.Email, &u.FullName, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return User{}, ErrEmailAlreadyExists
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetUserByEmail devuelve el usuario y su password_hash. El hash se devuelve
// como string aparte para que solo el service lo vea (nunca llega al handler).
func (r *PgRepository) GetUserByEmail(ctx context.Context, email string) (User, string, error) {
	const q = `
		SELECT id, email, COALESCE(full_name, ''), password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var u User
	var hash string
	err := r.db.QueryRow(ctx, q, email).Scan(&u.ID, &u.Email, &u.FullName, &hash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, "", ErrUserNotFound
		}
		return User{}, "", fmt.Errorf("get user by email: %w", err)
	}
	return u, hash, nil
}

// GetUserByID busca un usuario por su UUID.
func (r *PgRepository) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	const q = `
		SELECT id, email, COALESCE(full_name, ''), created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.Email, &u.FullName, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// GetProfileByUserID devuelve el perfil del usuario o ErrProfileNotFound.
func (r *PgRepository) GetProfileByUserID(ctx context.Context, userID uuid.UUID) (Profile, error) {
	const q = `
		SELECT user_id,
		       COALESCE(english_level, ''),
		       COALESCE(role, ''),
		       COALESCE(main_stack, ''),
		       COALESCE(goal, ''),
		       updated_at
		FROM user_profiles
		WHERE user_id = $1
	`
	var p Profile
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&p.UserID, &p.EnglishLevel, &p.Role, &p.MainStack, &p.Goal, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, ErrProfileNotFound
		}
		return Profile{}, fmt.Errorf("get profile: %w", err)
	}
	return p, nil
}

// UpsertProfile crea o actualiza el perfil del usuario.
// NULLIF se usa para que strings vacios se guarden como NULL.
func (r *PgRepository) UpsertProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (Profile, error) {
	const q = `
		INSERT INTO user_profiles (user_id, english_level, role, main_stack, goal)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''))
		ON CONFLICT (user_id) DO UPDATE
		SET english_level = EXCLUDED.english_level,
		    role          = EXCLUDED.role,
		    main_stack    = EXCLUDED.main_stack,
		    goal          = EXCLUDED.goal,
		    updated_at    = NOW()
		RETURNING user_id,
		          COALESCE(english_level, ''),
		          COALESCE(role, ''),
		          COALESCE(main_stack, ''),
		          COALESCE(goal, ''),
		          updated_at
	`
	var p Profile
	err := r.db.QueryRow(ctx, q,
		userID, in.EnglishLevel, in.Role, in.MainStack, in.Goal,
	).Scan(&p.UserID, &p.EnglishLevel, &p.Role, &p.MainStack, &p.Goal, &p.UpdatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("upsert profile: %w", err)
	}
	return p, nil
}
