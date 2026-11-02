package practice

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository define las operaciones de persistencia que necesita el service.
// Definirla como interface permite testear el service con un fake en memoria.
type Repository interface {
	ListCategories(ctx context.Context) ([]Category, error)
	GetCategoryByCode(ctx context.Context, code string) (Category, error)
	ListPromptsByCategory(ctx context.Context, categoryID uuid.UUID) ([]Prompt, error)
	GetPromptByID(ctx context.Context, id uuid.UUID) (Prompt, error)

	CreateSession(ctx context.Context, params CreateSessionParams) (Session, error)
	GetSessionByID(ctx context.Context, sessionID, userID uuid.UUID) (Session, error)
	ListSessionsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Session, error)

	SaveFeedback(ctx context.Context, params SaveFeedbackParams) (Feedback, error)
	GetFeedbackBySessionID(ctx context.Context, sessionID uuid.UUID) (Feedback, error)
}

// SaveFeedbackParams agrupa los campos persistidos en ai_feedback_results.
// Strengths, improvements y suggestedPhrases NO se guardan: viajan solo
// en la respuesta inmediata de /analyze-audio.
type SaveFeedbackParams struct {
	SessionID          uuid.UUID
	CorrectedText      string
	NaturalVersion     string
	ExplanationES      string
	ComprehensionScore int
	VocabularyScore    int
	CommunicationScore *int
}

// CreateSessionParams agrupa los campos para insertar una sesion.
// Es separado de CreateSessionInput porque ya viene resuelto: el code
// se mapeo a categoryID y el prompt se valido.
type CreateSessionParams struct {
	UserID         uuid.UUID
	CategoryID     uuid.UUID
	PromptID       *uuid.UUID
	UserTranscript string
	Status         SessionStatus
}

// PgRepository implementa Repository sobre Postgres con pgx.
//
// NOTA: las queries SQL estan en internal/database/queries/practice.sql
// como referencia para sqlc; mientras tanto se ejecutan a mano.
type PgRepository struct {
	db *pgxpool.Pool
}

// NewPgRepository construye el repository concreto.
func NewPgRepository(db *pgxpool.Pool) *PgRepository {
	return &PgRepository{db: db}
}

func (r *PgRepository) ListCategories(ctx context.Context) ([]Category, error) {
	const q = `
		SELECT id, code, name, description, COALESCE(icon, ''), created_at
		FROM practice_categories
		ORDER BY name
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	out := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Icon, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return out, nil
}

func (r *PgRepository) GetCategoryByCode(ctx context.Context, code string) (Category, error) {
	const q = `
		SELECT id, code, name, description, COALESCE(icon, ''), created_at
		FROM practice_categories
		WHERE code = $1
	`
	var c Category
	err := r.db.QueryRow(ctx, q, code).Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Icon, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, fmt.Errorf("get category by code: %w", err)
	}
	return c, nil
}

func (r *PgRepository) ListPromptsByCategory(ctx context.Context, categoryID uuid.UUID) ([]Prompt, error) {
	const q = `
		SELECT id, category_id, title, prompt_text, difficulty, created_at
		FROM practice_prompts
		WHERE category_id = $1
		ORDER BY difficulty, title
	`
	rows, err := r.db.Query(ctx, q, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list prompts: %w", err)
	}
	defer rows.Close()

	out := []Prompt{}
	for rows.Next() {
		var p Prompt
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Title, &p.PromptText, &p.Difficulty, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan prompt: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate prompts: %w", err)
	}
	return out, nil
}

func (r *PgRepository) GetPromptByID(ctx context.Context, id uuid.UUID) (Prompt, error) {
	const q = `
		SELECT id, category_id, title, prompt_text, difficulty, created_at
		FROM practice_prompts
		WHERE id = $1
	`
	var p Prompt
	err := r.db.QueryRow(ctx, q, id).Scan(&p.ID, &p.CategoryID, &p.Title, &p.PromptText, &p.Difficulty, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Prompt{}, ErrPromptNotFound
		}
		return Prompt{}, fmt.Errorf("get prompt: %w", err)
	}
	return p, nil
}

func (r *PgRepository) CreateSession(ctx context.Context, params CreateSessionParams) (Session, error) {
	const q = `
		WITH inserted AS (
			INSERT INTO practice_sessions (user_id, category_id, prompt_id, user_transcript, status)
			VALUES ($1, $2, $3, NULLIF($4, ''), $5)
			RETURNING *
		)
		SELECT i.id, i.user_id, i.category_id, c.code, i.prompt_id,
		       COALESCE(i.user_transcript, ''), i.status, i.created_at
		FROM inserted i
		JOIN practice_categories c ON c.id = i.category_id
	`
	var s Session
	err := r.db.QueryRow(ctx, q,
		params.UserID, params.CategoryID, params.PromptID, params.UserTranscript, params.Status,
	).Scan(
		&s.ID, &s.UserID, &s.CategoryID, &s.CategoryCode, &s.PromptID,
		&s.UserTranscript, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return s, nil
}

func (r *PgRepository) GetSessionByID(ctx context.Context, sessionID, userID uuid.UUID) (Session, error) {
	const q = `
		SELECT s.id, s.user_id, s.category_id, c.code, s.prompt_id,
		       COALESCE(s.user_transcript, ''), s.status, s.created_at
		FROM practice_sessions s
		JOIN practice_categories c ON c.id = s.category_id
		WHERE s.id = $1 AND s.user_id = $2
	`
	var s Session
	err := r.db.QueryRow(ctx, q, sessionID, userID).Scan(
		&s.ID, &s.UserID, &s.CategoryID, &s.CategoryCode, &s.PromptID,
		&s.UserTranscript, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}
	return s, nil
}

func (r *PgRepository) ListSessionsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Session, error) {
	const q = `
		SELECT s.id, s.user_id, s.category_id, c.code, s.prompt_id,
		       COALESCE(s.user_transcript, ''), s.status, s.created_at,
		       f.comprehension_score, f.vocabulary_score, f.communication_score
		FROM practice_sessions s
		JOIN practice_categories c ON c.id = s.category_id
		LEFT JOIN ai_feedback_results f ON f.session_id = s.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	out := []Session{}
	for rows.Next() {
		var s Session
		var comprehension, vocabulary, communication *int
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.CategoryID, &s.CategoryCode, &s.PromptID,
			&s.UserTranscript, &s.Status, &s.CreatedAt,
			&comprehension, &vocabulary, &communication,
		); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		if comprehension != nil || vocabulary != nil || communication != nil {
			s.Scores = &Scores{
				Comprehension: comprehension,
				Vocabulary:    vocabulary,
				Communication: communication,
			}
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}
	return out, nil
}

func (r *PgRepository) SaveFeedback(ctx context.Context, params SaveFeedbackParams) (Feedback, error) {
	const q = `
		INSERT INTO ai_feedback_results (
			session_id, corrected_text, natural_version, explanation_es,
			comprehension_score, vocabulary_score, communication_score
		) VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), $5, $6, $7)
		RETURNING id, session_id,
		          COALESCE(corrected_text, ''),
		          COALESCE(natural_version, ''),
		          COALESCE(explanation_es, ''),
		          COALESCE(comprehension_score, 0),
		          COALESCE(vocabulary_score, 0),
		          communication_score,
		          created_at
	`
	var f Feedback
	err := r.db.QueryRow(ctx, q,
		params.SessionID, params.CorrectedText, params.NaturalVersion, params.ExplanationES,
		params.ComprehensionScore, params.VocabularyScore, params.CommunicationScore,
	).Scan(
		&f.ID, &f.SessionID,
		&f.CorrectedText, &f.NaturalVersion, &f.ExplanationES,
		&f.ComprehensionScore, &f.VocabularyScore, &f.CommunicationScore,
		&f.CreatedAt,
	)
	if err != nil {
		return Feedback{}, fmt.Errorf("save feedback: %w", err)
	}
	return f, nil
}

func (r *PgRepository) GetFeedbackBySessionID(ctx context.Context, sessionID uuid.UUID) (Feedback, error) {
	const q = `
		SELECT id, session_id,
		       COALESCE(corrected_text, ''),
		       COALESCE(natural_version, ''),
		       COALESCE(explanation_es, ''),
		       COALESCE(comprehension_score, 0),
		       COALESCE(vocabulary_score, 0),
		       communication_score,
		       created_at
		FROM ai_feedback_results
		WHERE session_id = $1
	`
	var f Feedback
	err := r.db.QueryRow(ctx, q, sessionID).Scan(
		&f.ID, &f.SessionID,
		&f.CorrectedText, &f.NaturalVersion, &f.ExplanationES,
		&f.ComprehensionScore, &f.VocabularyScore, &f.CommunicationScore,
		&f.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Feedback{}, ErrFeedbackNotFound
		}
		return Feedback{}, fmt.Errorf("get feedback: %w", err)
	}
	return f, nil
}

func intPtr(n int) *int {
	return &n
}
