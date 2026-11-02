-- Queries de users / user_profiles. Por ahora se usan a mano desde
-- internal/users/repository.go con pgx. Cuando integremos sqlc en una
-- pasada dedicada, sqlc generara el codigo a partir de este archivo y
-- el repository pasara a envolver el codigo generado.

-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserProfile :one
SELECT *
FROM user_profiles
WHERE user_id = $1;

-- name: UpsertUserProfile :one
INSERT INTO user_profiles (user_id, english_level, role, main_stack, goal)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id) DO UPDATE
SET english_level = EXCLUDED.english_level,
    role          = EXCLUDED.role,
    main_stack    = EXCLUDED.main_stack,
    goal          = EXCLUDED.goal,
    updated_at    = NOW()
RETURNING *;
