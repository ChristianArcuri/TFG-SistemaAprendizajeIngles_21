# Roadmap del backend

Plan operativo del MVP, alineado con `AGENTS.md`. Va por fases chicas y testeables, dejando IA mockeada antes que real para no bloquear el desarrollo.

---

## Estado actual

### Fase 1 — Backend base ✅

- `cmd/api/main.go`, `internal/config`, `internal/database`, `internal/server`, `internal/handlers`, `internal/httpx`.
- Conexión a Postgres (`pgxpool`), graceful shutdown, logger sin `Authorization`.
- Endpoint `GET /health` (y `GET /api/v1/health`).
- Migraciones iniciales aplicadas vía servicio `migrate` en `docker-compose`.
- Seed de `practice_categories`.
- `docker-compose.yml` con Postgres + migrate + api.

---

## Fase 2 — Auth (registro, login, JWT)

**Objetivo**: que un usuario pueda registrarse, hacer login y consumir endpoints protegidos.

### Archivos a crear

```
internal/auth/
  password.go           # hashing con bcrypt (HashPassword, ComparePassword)
  jwt.go                # GenerateToken(userID), ParseToken(string) -> claims
  errors.go             # errores tipados del paquete (ErrInvalidCredentials, etc.)

internal/users/
  models.go             # User, Profile, RegisterInput, LoginInput, UpdateProfileInput
  repository.go         # UserRepository sobre el código generado por sqlc
  service.go            # AuthService.Register, AuthService.Login, UserService.GetMe, UpdateProfile
  handler.go            # POST /auth/register, /auth/login, GET /me, PUT /me/profile
  validation.go         # validación de payloads (email, password mínimo, etc.)

internal/server/
  auth_middleware.go    # extrae Bearer, valida JWT, mete userID en context.Context

internal/database/queries/
  users.sql             # queries que consumirá sqlc

internal/database/db/   # carpeta destino del código generado por sqlc (no editar a mano)
```

### Dependencias nuevas

- `golang.org/x/crypto/bcrypt`
- `github.com/golang-jwt/jwt/v5`
- `github.com/go-playground/validator/v10` (ya viene transitiva con Gin)

### Queries sqlc

```sql
-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

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

-- name: GetUserProfile :one
SELECT * FROM user_profiles WHERE user_id = $1;
```

### Endpoints nuevos

| Método | Path                       | Auth | Descripción                                    |
|--------|----------------------------|------|------------------------------------------------|
| POST   | `/api/v1/auth/register`    | No   | Crea usuario, devuelve `{ token, user }`       |
| POST   | `/api/v1/auth/login`       | No   | Valida credenciales, devuelve `{ token, user }` |
| GET    | `/api/v1/me`               | JWT  | Devuelve `{ user, profile }`                   |
| PUT    | `/api/v1/me/profile`       | JWT  | Upsert del perfil                              |

### Reglas

- Password mínimo 8 chars (validar en service).
- `email` se guarda en minúsculas.
- Errores al cliente vía `httpx.Error` con códigos: `validation_error`, `invalid_credentials`, `conflict`, `unauthorized`.
- El handler nunca recibe el password en limpio en logs.
- `JWT_SECRET` ya está cargado en `config.JWT.Secret`.
- Tests unitarios del `AuthService` con un `UserRepository` fake (sin DB).

---

## Fase 3 — Práctica (catálogo, prompts, sesiones, historial)

**Objetivo**: el usuario lista categorías, ve prompts, crea sesiones y consulta su historial. Sin audio ni IA todavía.

### Archivos a crear

```
internal/practice/
  models.go             # Category, Prompt, Session, CreateSessionInput
  repository.go         # PracticeRepository
  service.go            # CategoryService, PromptService, SessionService
  handler.go            # endpoints HTTP

internal/database/queries/
  practice.sql
```

### Queries sqlc

```sql
-- name: ListPracticeCategories :many
SELECT * FROM practice_categories ORDER BY name;

-- name: GetPracticeCategoryByCode :one
SELECT * FROM practice_categories WHERE code = $1;

-- name: ListPromptsByCategory :many
SELECT * FROM practice_prompts
WHERE category_id = $1
ORDER BY difficulty, title;

-- name: CreatePracticeSession :one
INSERT INTO practice_sessions (user_id, category_id, prompt_id, user_transcript, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM practice_sessions WHERE id = $1 AND user_id = $2;

-- name: ListSessionsByUser :many
SELECT * FROM practice_sessions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SaveAIFeedback :one
INSERT INTO ai_feedback_results (
    session_id, corrected_text, natural_version, explanation_es,
    grammar_score, fluency_score, pronunciation_score, vocabulary_score
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetFeedbackBySession :one
SELECT * FROM ai_feedback_results WHERE session_id = $1;
```

### Endpoints nuevos

| Método | Path                                              | Auth | Descripción                            |
|--------|---------------------------------------------------|------|----------------------------------------|
| GET    | `/api/v1/practice/categories`                     | JWT  | Lista las 6 categorías                 |
| GET    | `/api/v1/practice/prompts?category=tech_interview`| JWT  | Lista prompts de una categoría         |
| POST   | `/api/v1/practice/sessions`                       | JWT  | Crea una sesión (sin audio todavía)    |
| GET    | `/api/v1/practice/sessions`                       | JWT  | Historial paginado del usuario         |
| GET    | `/api/v1/practice/sessions/:id`                   | JWT  | Detalle de sesión + feedback asociado  |

---

## Fase 4 — IA mockeada (`/analyze-audio` end-to-end fake)

**Objetivo**: el flujo completo del frontend funciona sin tocar OpenAI. Podemos demostrar la app y testear el camino crítico.

### Archivos a crear

```
internal/ai/
  prompts.go            # template del prompt de feedback (ver docs/prompts/feedback.md)
  models.go             # FeedbackResult, TranscriptionResult, FeedbackInput
  feedback_service.go   # interfaz FeedbackService { Generate(ctx, input) (FeedbackResult, error) }
  transcription_service.go # interfaz TranscriptionService { Transcribe(ctx, file) (string, error) }
  mock.go               # implementaciones MockFeedback, MockTranscription
```

### Endpoint

| Método | Path                                | Auth | Descripción                                                              |
|--------|-------------------------------------|------|--------------------------------------------------------------------------|
| POST   | `/api/v1/practice/analyze-audio`    | JWT  | `multipart/form-data` con `audio` + `categoryCode` + `promptText`. Crea sesión, transcribe, genera feedback y guarda todo. |

### Validaciones del handler

- `audio`: máximo `MAX_AUDIO_BYTES` (default 15 MB), MIME `audio/webm`, `audio/wav`, `audio/mpeg`, `audio/ogg`, `audio/mp4`.
- `categoryCode`: existe en DB.
- `promptText`: 1..2000 chars.
- Toda la lógica vive en `practice.SessionService.AnalyzeAudio(ctx, ...)`. El handler solo parsea, llama, responde con `httpx.OK`.

### Mock realista

`MockFeedback.Generate` devuelve un `FeedbackResult` con texto en español/inglés plausible y scores aleatorios entre 60 y 90 (semilla determinística por hash del transcript para que sea testeable). `MockTranscription.Transcribe` devuelve un string fijo más el tamaño del archivo recibido.

### Variables de entorno nuevas

```
AI_PROVIDER=mock         # mock | openai
MAX_AUDIO_BYTES=15728640
```

`AI_PROVIDER=mock` por default. Cambiar a `openai` activa la implementación real (Fase 5).

---

## Fase 5 — IA real (OpenAI) ✅

**Objetivo**: reemplazar mock por cliente OpenAI sin tocar handlers ni servicios de práctica.

### Archivos a crear

```
internal/ai/
  openai_client.go      # wrapper sobre el SDK con timeout, retry simple, sin loguear key ni audio
  openai_feedback.go    # implementa FeedbackService con OpenAI (chat completions, response_format json)
  openai_transcription.go # implementa TranscriptionService con Whisper (audio.transcriptions)
```

### Cableado

En `cmd/api/main.go`:

```go
var feedback ai.FeedbackService = ai.NewMockFeedback()
var transcription ai.TranscriptionService = ai.NewMockTranscription()

if cfg.AI.Provider == "openai" {
    client := ai.NewOpenAIClient(cfg.OpenAI.APIKey)
    feedback = ai.NewOpenAIFeedback(client)
    transcription = ai.NewOpenAITranscription(client)
}
```

### Reglas

- Nunca loguear `OPENAI_API_KEY`, body de transcripción ni audio en bruto.
- `context.WithTimeout` en cada llamada (default 30s para feedback, 60s para transcripción).
- Si OpenAI devuelve JSON inválido, el service responde `internal_error` al cliente y loguea solo el `request_id` de OpenAI.
- Validar que el JSON parseado tiene los campos requeridos.

### Prompt del feedback

Vive en `docs/prompts/feedback.md` y lo carga `prompts.go` con un template Go (`text/template`). Ver ese archivo para el contenido exacto.

---

## Fase 7 — Frontend ✅

Carpeta: `frontend/` (Vue 3 + Vite + Tailwind + Pinia + Axios).

Implementado:

- Landing, login/registro, onboarding, dashboard, práctica con grabación, resultados, historial.
- JWT en `localStorage`, interceptor Axios, guards de router.
- Integración con todos los endpoints del MVP.
- CORS en backend para `localhost:5173`.

Ver [`frontend/README.md`](../../frontend/README.md).

---

## Fase 7 — Frontend (referencia original)

Stack: Vue 3 + Vite + TailwindCSS + Pinia + Axios.

Pantallas (basadas en `remotetalk-static/`):

- `index.html` → Landing
- `login.html` → Login + Register
- `onboarding.html` → upsert de `user_profile`
- `dashboard.html` → categorías + stats
- `practice.html` → grabación de audio → POST `/practice/analyze-audio`
- `quiz.html`, `writing.html` → variantes de práctica
- `results.html` → muestra `FeedbackResult` (transcript, naturalVersion, explanationEs, scores, suggestedPhrases)

Reglas de UI:

- Idioma: español, salvo el inglés practicado, ejemplos, naturalVersion y suggestedPhrases.
- Token JWT en `localStorage` + interceptor de Axios que setea `Authorization: Bearer ...`.
- Captura de audio con `MediaRecorder` (`audio/webm`).

---

## Fase 8 — Collection Postman y flujos completos ⏳

**Objetivo:** probar y entender el backend sin leer código.

**Entregables:**

- `docs/postman/tfg-sai.postman_collection.json`
- `docs/postman/tfg-sai.postman_environment.json`
- `docs/flows.md`

**Contenido de `flows.md` (índice):** convenciones API · login/onboarding · catálogo · daily meeting · entrevista técnica/HR · `analyze-audio` (mock vs OpenAI, MIME, cuota) · historial · diagrama de secuencia · checklist Docker.

Ver detalle en [`README.md`](../README.md#fase-8--collection-postman-y-flujos-completos-pendiente).

---

## Decisiones transversales

- **sqlc**: las queries van en `internal/database/queries/*.sql`. El código generado en `internal/database/db/`. Re-generar con `sqlc generate` cada vez que se toca un `.sql`. Versión que vamos a usar: `kyleconroy/sqlc` v1.27+ (compatible con `pgx/v5`).
- **golang-migrate**: ya cableado vía servicio `migrate` en `docker-compose.yml`.
- **Errores**: nunca exponer texto interno. Usar `httpx.Error(c, status, code, message)` con códigos de la lista en `docs/api.md`.
- **Tests**: unitarios para services con repos fake; smoke test del router para health y rutas protegidas.
- **OpenAI mock por default**: es la regla operativa que viene del punto 14 del plan original — todo se construye contra interfaces, mock primero, real después.

---

## Orden de trabajo

1. ✅ Fase 1 — backend base.
2. ✅ Migraciones + seed.
3. ✅ Fase 2 — auth.
4. ✅ Fase 3 — práctica (catálogo + sesiones).
5. ✅ Fase 4 — IA mockeada + endpoint `analyze-audio`.
6. ✅ Fase 5 — IA real (OpenAI).
7. ✅ Fase 7 — frontend.
8. ⏳ **Fase 8 — Postman + flujos** (siguiente).
9. ✅ **Fase 9 — Bancos de ejercicios (frontend)** — ver [`docs/practice-content-phases.md`](../../docs/practice-content-phases.md) (A–E completadas).
