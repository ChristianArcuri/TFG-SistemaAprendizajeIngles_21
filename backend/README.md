# Inglés Profesional IT — Backend

API del prototipo del TFG. Cubre registro, perfil, entrevista con Recursos Humanos, daily meeting, quiz de inglés IT, feedback lingüístico, historial y recomendación de refuerzo.

La IA se usa solo para transcribir audio y redactar el feedback. Las tres competencias (comprensión, vocabulario profesional y comunicación profesional) y la regla de refuerzo se calculan en este backend.

## Stack

- Go 1.25 + [Gin](https://github.com/gin-gonic/gin)
- PostgreSQL 16
- [pgx/v5](https://github.com/jackc/pgx) + `pgxpool`
- [sqlc](https://docs.sqlc.dev/) (configurado, sin queries todavía)
- [golang-migrate](https://github.com/golang-migrate/migrate) (carpeta `migrations/` lista)
- JWT (secret cargado en config, sin middleware todavía)
- Docker + docker-compose

## Estructura

```
backend/
├── cmd/api/                    # Entry point
├── internal/
│   ├── config/                 # Carga de variables de entorno
│   ├── database/               # Pool de PostgreSQL (pgxpool)
│   ├── server/                 # Router Gin + http.Server con graceful shutdown
│   ├── handlers/               # HTTP handlers finos
│   ├── httpx/                  # Helpers de respuesta JSON consistente
│   ├── auth/                   # JWT + bcrypt
│   ├── users/                  # Registro, login, perfil
│   ├── practice/               # Categorías, sesiones, analyze-audio
│   └── ai/                     # Mock + OpenAI (transcripción + feedback)
├── migrations/                 # SQL para golang-migrate
├── docs/
│   ├── api.md                  # Contrato HTTP del MVP
│   ├── roadmap.md              # Plan por fases
│   ├── flows.md                # (Fase 8) Flujos end-to-end — pendiente
│   └── postman/                # (Fase 8) Collection Postman — pendiente
├── sqlc.yaml                   # Configuración de sqlc
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

## Cómo correrlo

### 1. Variables de entorno

```bash
cp .env.example .env
```

Editá `.env` y cambiá `JWT_SECRET` por un valor real (>= 32 chars en producción).

### 2. Levantar con docker-compose

```bash
docker compose up --build
```

Levanta Postgres, corre las migraciones (servicio `migrate`) y arranca la API. La API queda en `http://localhost:8080`.

El servicio `migrate` usa `migrate/migrate` y aplica todo lo que esté en `migrations/`. La API depende de que ese paso termine con éxito antes de arrancar.

Para correr migraciones manualmente:

```bash
# Aplicar todas las pendientes
docker compose run --rm migrate -path=/migrations -database=$DATABASE_URL up

# Bajar la última
docker compose run --rm migrate -path=/migrations -database=$DATABASE_URL down 1

# Crear una nueva migración (genera 000003_xxx.up.sql y .down.sql)
docker compose run --rm --entrypoint migrate migrate create -ext sql -dir /migrations -seq nombre_de_la_migracion
```

### 3. Probar el health check

```bash
curl http://localhost:8080/health
```

Respuesta esperada:

```json
{
  "data": {
    "service": "tfg-sai",
    "env": "development",
    "status": "ok",
    "db": "ok"
  }
}
```

### 4. Correr sin Docker (desarrollo local)

Necesitás Postgres corriendo y `DATABASE_URL` apuntando a él.

```bash
go run ./cmd/api
```

## Endpoints

| Método | Path                                              | Auth | Descripción                              |
|--------|---------------------------------------------------|------|------------------------------------------|
| GET    | `/health`                                         | No   | Liveness + estado de DB                  |
| GET    | `/api/v1/health`                                  | No   | Igual que `/health`, bajo el prefijo API |
| POST   | `/api/v1/auth/register`                           | No   | Crea usuario y devuelve JWT              |
| POST   | `/api/v1/auth/login`                              | No   | Login por email/password                 |
| GET    | `/api/v1/me`                                      | JWT  | Usuario autenticado + perfil             |
| PUT    | `/api/v1/me/profile`                              | JWT  | Upsert del perfil                        |
| GET    | `/api/v1/practice/categories`                     | JWT  | Lista categorías de práctica             |
| GET    | `/api/v1/practice/prompts?category=<code>`        | JWT  | Lista prompts de una categoría           |
| POST   | `/api/v1/practice/sessions`                       | JWT  | Crea una sesión                          |
| GET    | `/api/v1/practice/sessions?limit=&offset=`        | JWT  | Historial paginado                       |
| GET    | `/api/v1/practice/sessions/:id`                   | JWT  | Detalle + feedback (si existe)           |
| POST   | `/api/v1/practice/analyze-audio`                  | JWT  | multipart: audio + categoryCode + promptText. Transcribe + feedback (mock u OpenAI). |

La superficie completa del MVP (auth, perfil, práctica, feedback) está documentada en [`docs/api.md`](docs/api.md).

## Modelo de datos

Las tablas iniciales viven en `migrations/000001_init_schema.up.sql`:

- `users` — credenciales y datos básicos.
- `user_profiles` — 1:1 con `users` (nivel de inglés, rol, stack, objetivo).
- `practice_categories` — catálogo (HR interview, daily, etc.).
- `practice_prompts` — preguntas/consignas por categoría.
- `practice_sessions` — sesión de un usuario sobre una categoría/prompt.
- `ai_feedback_results` — feedback de IA 1:1 con la sesión.

`migrations/000002_seed_practice_categories.up.sql` carga las 6 categorías del MVP de forma idempotente.

## Convenciones

- Handlers finos: parsean request, llaman a un servicio, responden con `httpx`.
- Lógica de negocio en `internal/services/` (todavía no creado).
- Acceso a DB en `internal/repositories/` o código generado por `sqlc`.
- Errores internos no se exponen al cliente: usar `httpx.Error` con un mensaje seguro.
- Sin secrets hardcodeados: todo viene de variables de entorno.

## Roadmap

1. ~~Backend base~~ ✅
2. ~~Migraciones iniciales + seed~~ ✅
3. ~~Auth (registro, login, JWT, middleware)~~ ✅
4. ~~Práctica: categorías, prompts, sesiones, historial~~ ✅
5. ~~IA mockeada + endpoint `POST /practice/analyze-audio`~~ ✅
6. ~~IA real (OpenAI Whisper + Chat)~~ ✅
7. ~~Frontend (Vue 3 + Vite + Tailwind + Pinia)~~ ✅ — carpeta [`../frontend`](../frontend).
8. **Collection Postman + documento de flujos** — pendiente (ver abajo).

Detalle por fase en [`docs/roadmap.md`](docs/roadmap.md). Prompt de feedback de IA en [`docs/prompts/feedback.md`](docs/prompts/feedback.md).

---

## Fase 8 — Collection Postman y flujos completos (pendiente)

Objetivo: poder probar y entender el backend sin leer código — una **collection de Postman** con todos los endpoints y un **documento de flujos** que explique cómo encajan (auth, práctica, IA, condiciones de error).

### Entregables previstos

| Archivo | Descripción |
|---------|-------------|
| `docs/postman/tfg-sai.postman_collection.json` | Collection importable con las 12 rutas del MVP, variables de entorno y ejemplos de body. |
| `docs/postman/tfg-sai.postman_environment.json` | Environment `local` (`baseUrl`, `token`, `sessionId`, etc.). |
| `docs/flows.md` | Guía narrativa de flujos end-to-end (ver índice abajo). |

### Collection Postman (borrador)

- **Carpetas:** Health · Auth · Perfil · Práctica (catálogo) · Práctica (sesiones) · Práctica (audio + IA).
- **Variables de collection:** `baseUrl` = `http://localhost:8080`, `apiPrefix` = `/api/v1`.
- **Variables de environment:** `token` (se llena con script tras login/register), `sessionId`, `categoryCode`.
- **Auth:** en requests protegidos, header `Authorization: Bearer {{token}}`.
- **Script post-login:** guardar `data.token` en `token` automáticamente.
- **Requests incluidos:** los 12 endpoints de la tabla [Endpoints](#endpoints); `analyze-audio` como `multipart/form-data` con archivo de ejemplo.

### Documento de flujos (`docs/flows.md`) — índice previsto

1. **Convenciones** — formato `{ data }` / `{ error }`, prefijo `/api/v1`, códigos de error comunes.
2. **Flujo: registro y login** — register → token → `GET /me` → onboarding con `PUT /me/profile`.
3. **Flujo: elegir práctica** — `GET /categories` → `GET /prompts?category=...` (cuando existan prompts en DB).
4. **Flujo: daily meeting** — categoría `daily_meeting`, consigna típica (yesterday / today / blockers), camino con sesión manual vs `analyze-audio`.
5. **Flujo: entrevista técnica / HR** — mismos pasos con `categoryCode` distinto (`tech_interview`, `hr_interview`).
6. **Flujo: analyze-audio (IA)** — multipart, validaciones (MIME, tamaño), orden interno: transcribir → feedback → guardar sesión + `ai_feedback_results`; diferencia `AI_PROVIDER=mock` vs `openai`; requisitos (`OPENAI_API_KEY`, cuota, `HTTP_WRITE_TIMEOUT` ≥ 95s).
7. **Flujo: historial y resultados** — `GET /sessions` → `GET /sessions/:id` (feedback persistido vs campos solo en respuesta de analyze-audio).
8. **Diagrama resumen** — secuencia cliente → API → Postgres / OpenAI.
9. **Checklist de prueba manual** — pasos para validar un deploy local con Docker.

### Condiciones importantes (para el doc de flujos)

- Rutas públicas: solo `health`, `auth/register`, `auth/login`.
- JWT: obligatorio en el resto; expiración según `JWT_EXPIRES_IN`.
- IA mock: sin costo, respuestas determinísticas; ideal para desarrollo.
- IA OpenAI: requiere billing activo; errores de cuota → `500 internal_error` (detalle en logs del contenedor).
- `analyze-audio`: audio máx. `MAX_AUDIO_BYTES` (default 15 MB); MIME permitidos (`audio/webm`, `wav`, `mpeg`, `ogg`, `mp4`, etc.).

Cuando retomemos la Fase 8, generamos la collection y `docs/flows.md` a partir de [`docs/api.md`](docs/api.md) y las pruebas ya hechas contra Docker.
