-- Esquema inicial de Inglés Profesional IT.
-- gen_random_uuid() ya viene en Postgres 13+, pero dejamos pgcrypto
-- por compatibilidad con entornos donde no este habilitado.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Usuarios autenticables.
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   TEXT         NOT NULL,
    full_name       VARCHAR(150),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Perfil 1:1 con el usuario (nivel de ingles, rol, stack, objetivo).
CREATE TABLE user_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    english_level   VARCHAR(20),
    role            VARCHAR(100),
    main_stack      VARCHAR(150),
    goal            VARCHAR(150),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Catalogo de tipos de practica (HR interview, daily, etc.).
CREATE TABLE practice_categories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(50)  NOT NULL UNIQUE,
    name            VARCHAR(100) NOT NULL,
    description     TEXT         NOT NULL,
    icon            VARCHAR(50),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Prompts asociados a una categoria (preguntas / consignas).
CREATE TABLE practice_prompts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id     UUID         NOT NULL REFERENCES practice_categories(id) ON DELETE CASCADE,
    title           VARCHAR(150) NOT NULL,
    prompt_text     TEXT         NOT NULL,
    difficulty      VARCHAR(20)  NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_practice_prompts_category_id
    ON practice_prompts(category_id);

-- Una sesion de practica de un usuario sobre una categoria/prompt.
CREATE TABLE practice_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id     UUID         NOT NULL REFERENCES practice_categories(id),
    prompt_id       UUID         REFERENCES practice_prompts(id),
    user_transcript TEXT,
    status          VARCHAR(30)  NOT NULL DEFAULT 'completed',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_practice_sessions_user_created
    ON practice_sessions(user_id, created_at DESC);

CREATE INDEX idx_practice_sessions_category
    ON practice_sessions(category_id);

-- Resultado del feedback de IA para una sesion (1:1).
CREATE TABLE ai_feedback_results (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id          UUID NOT NULL UNIQUE REFERENCES practice_sessions(id) ON DELETE CASCADE,
    corrected_text      TEXT,
    natural_version     TEXT,
    explanation_es      TEXT,
    comprehension_score INT,
    vocabulary_score    INT,
    communication_score INT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
