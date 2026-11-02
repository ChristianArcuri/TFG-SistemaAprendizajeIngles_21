-- Queries de practice_categories, practice_prompts, practice_sessions
-- y ai_feedback_results. Por ahora se usan a mano desde
-- internal/practice/repository.go con pgx. Cuando integremos sqlc en
-- una pasada dedicada, sqlc generara el codigo a partir de este archivo.

-- name: ListPracticeCategories :many
SELECT *
FROM practice_categories
ORDER BY name;

-- name: GetPracticeCategoryByCode :one
SELECT *
FROM practice_categories
WHERE code = $1;

-- name: ListPromptsByCategory :many
SELECT *
FROM practice_prompts
WHERE category_id = $1
ORDER BY difficulty, title;

-- name: GetPromptByID :one
SELECT *
FROM practice_prompts
WHERE id = $1;

-- name: CreatePracticeSession :one
INSERT INTO practice_sessions (user_id, category_id, prompt_id, user_transcript, status)
VALUES ($1, $2, $3, NULLIF($4, ''), $5)
RETURNING *;

-- name: GetSessionByIDForUser :one
SELECT s.*, c.code AS category_code
FROM practice_sessions s
JOIN practice_categories c ON c.id = s.category_id
WHERE s.id = $1 AND s.user_id = $2;

-- name: ListSessionsByUser :many
SELECT s.*, c.code AS category_code
FROM practice_sessions s
JOIN practice_categories c ON c.id = s.category_id
WHERE s.user_id = $1
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: SaveAIFeedback :one
INSERT INTO ai_feedback_results (
    session_id, corrected_text, natural_version, explanation_es,
    comprehension_score, vocabulary_score, communication_score
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetFeedbackBySession :one
SELECT *
FROM ai_feedback_results
WHERE session_id = $1;
