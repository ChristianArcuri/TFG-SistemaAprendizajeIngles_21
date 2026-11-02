package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/remotetalk/backend/internal/auth"
)

// fakeRepo implementa Repository en memoria para testear el service
// sin Postgres. Cada test crea su propia instancia.
type fakeRepo struct {
	users    map[uuid.UUID]userRecord
	byEmail  map[string]uuid.UUID
	profiles map[uuid.UUID]Profile
}

type userRecord struct {
	user User
	hash string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users:    map[uuid.UUID]userRecord{},
		byEmail:  map[string]uuid.UUID{},
		profiles: map[uuid.UUID]Profile{},
	}
}

func (r *fakeRepo) CreateUser(_ context.Context, email, hash, fullName string) (User, error) {
	if _, exists := r.byEmail[email]; exists {
		return User{}, ErrEmailAlreadyExists
	}
	id := uuid.New()
	now := time.Now().UTC()
	u := User{ID: id, Email: email, FullName: fullName, CreatedAt: now, UpdatedAt: now}
	r.users[id] = userRecord{user: u, hash: hash}
	r.byEmail[email] = id
	return u, nil
}

func (r *fakeRepo) GetUserByEmail(_ context.Context, email string) (User, string, error) {
	id, ok := r.byEmail[email]
	if !ok {
		return User{}, "", ErrUserNotFound
	}
	rec := r.users[id]
	return rec.user, rec.hash, nil
}

func (r *fakeRepo) GetUserByID(_ context.Context, id uuid.UUID) (User, error) {
	rec, ok := r.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return rec.user, nil
}

func (r *fakeRepo) GetProfileByUserID(_ context.Context, id uuid.UUID) (Profile, error) {
	p, ok := r.profiles[id]
	if !ok {
		return Profile{}, ErrProfileNotFound
	}
	return p, nil
}

func (r *fakeRepo) UpsertProfile(_ context.Context, id uuid.UUID, in UpdateProfileInput) (Profile, error) {
	p := Profile{
		UserID:       id,
		EnglishLevel: in.EnglishLevel,
		Role:         in.Role,
		MainStack:    in.MainStack,
		Goal:         in.Goal,
		UpdatedAt:    time.Now().UTC(),
	}
	r.profiles[id] = p
	return p, nil
}

func newTestService(t *testing.T) (*Service, *fakeRepo) {
	t.Helper()
	repo := newFakeRepo()
	issuer := auth.NewTokenIssuer("test-secret-with-at-least-32-chars-long!", time.Hour)
	return NewService(repo, issuer), repo
}

func TestServiceRegisterAndLogin(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	resp, err := svc.Register(ctx, RegisterInput{
		Email:    "  Test@Example.COM ",
		Password: "Sup3r-Secret!",
		FullName: "  Test User  ",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if resp.User.Email != "test@example.com" {
		t.Fatalf("email not normalized: %q", resp.User.Email)
	}
	if resp.User.FullName != "Test User" {
		t.Fatalf("full_name not trimmed: %q", resp.User.FullName)
	}
	if resp.Token == "" {
		t.Fatal("expected token")
	}

	loginResp, err := svc.Login(ctx, LoginInput{
		Email:    "TEST@example.com",
		Password: "Sup3r-Secret!",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loginResp.User.ID != resp.User.ID {
		t.Fatalf("login returned different user")
	}
}

func TestServiceLoginInvalidCredentials(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	if _, err := svc.Login(ctx, LoginInput{Email: "missing@example.com", Password: "x"}); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for missing user, got %v", err)
	}

	if _, err := svc.Register(ctx, RegisterInput{Email: "u@example.com", Password: "Sup3r-Secret!"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, err := svc.Login(ctx, LoginInput{Email: "u@example.com", Password: "wrong-password"}); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}
}

func TestServiceRegisterDuplicateEmail(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	if _, err := svc.Register(ctx, RegisterInput{Email: "dup@example.com", Password: "Sup3r-Secret!"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, err := svc.Register(ctx, RegisterInput{Email: "DUP@example.com", Password: "Sup3r-Secret!"})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestServiceGetMeWithAndWithoutProfile(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	resp, err := svc.Register(ctx, RegisterInput{Email: "me@example.com", Password: "Sup3r-Secret!"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	me, err := svc.GetMe(ctx, resp.User.ID)
	if err != nil {
		t.Fatalf("get me: %v", err)
	}
	if me.Profile != nil {
		t.Fatalf("expected no profile, got %+v", me.Profile)
	}

	if _, err := svc.UpdateProfile(ctx, resp.User.ID, UpdateProfileInput{
		EnglishLevel: "B1",
		Role:         "Backend",
		MainStack:    "Go, Postgres",
		Goal:         "Daily meetings",
	}); err != nil {
		t.Fatalf("update profile: %v", err)
	}

	me2, err := svc.GetMe(ctx, resp.User.ID)
	if err != nil {
		t.Fatalf("get me 2: %v", err)
	}
	if me2.Profile == nil {
		t.Fatal("expected profile, got nil")
	}
	if me2.Profile.EnglishLevel != "B1" {
		t.Fatalf("english level mismatch: %q", me2.Profile.EnglishLevel)
	}
}

func TestServiceGetMeUnknownUser(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	_, err := svc.GetMe(ctx, uuid.New())
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
