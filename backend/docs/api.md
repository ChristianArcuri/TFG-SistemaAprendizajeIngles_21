# Inglés Profesional IT — API

Documento de referencia de la superficie HTTP del backend. Lo que ya está implementado y lo que queda pendiente.

Todas las respuestas siguen la forma estándar:

- Éxito: `{ "data": <payload> }`
- Error: `{ "error": { "code": "...", "message": "..." } }`

Prefijo base de la API versionada: `/api/v1`.

---

## Implementado

### Health

| Método | Path                | Auth | Descripción                            |
|--------|---------------------|------|----------------------------------------|
| GET    | `/health`           | No   | Liveness + estado de DB                |
| GET    | `/api/v1/health`    | No   | Igual que `/health` bajo el prefijo v1 |

Respuesta:

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

`db` puede ser `"ok"` o `"down"`. Si la DB está caída, el endpoint responde `503`.

### Auth

| Método | Path                       | Auth | Descripción                                |
|--------|----------------------------|------|--------------------------------------------|
| POST   | `/api/v1/auth/register`    | No   | Crea un usuario y devuelve JWT             |
| POST   | `/api/v1/auth/login`       | No   | Autentica por email/password y devuelve JWT|

`POST /api/v1/auth/register`

```json
{
  "email": "user@example.com",
  "password": "min-8-chars",
  "fullName": "Optional Name"
}
```

Respuesta `201 Created`:

```json
{
  "data": {
    "token": "eyJhbGciOi...",
    "expiresAt": "2026-05-26T20:00:00Z",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "fullName": "Optional Name",
      "createdAt": "...",
      "updatedAt": "..."
    }
  }
}
```

`POST /api/v1/auth/login`: mismo body sin `fullName`, mismo shape de respuesta con `200`.

### Perfil

| Método | Path                 | Auth | Descripción                          |
|--------|----------------------|------|--------------------------------------|
| GET    | `/api/v1/me`         | JWT  | Devuelve `{ user, profile }` (profile puede ser null) |
| PUT    | `/api/v1/me/profile` | JWT  | Upsert del perfil del usuario        |

`PUT /api/v1/me/profile`:

```json
{
  "englishLevel": "B1",
  "role": "Backend Developer",
  "mainStack": "Go, Postgres, Docker",
  "goal": "Daily meetings y entrevistas remotas"
}
```

`englishLevel` debe ser uno de: `A1`, `A2`, `B1`, `B2`, `C1`, `C2`, `unknown`.

### Práctica — catálogo

| Método | Path                                              | Auth | Descripción                                  |
|--------|---------------------------------------------------|------|----------------------------------------------|
| GET    | `/api/v1/practice/categories`                     | JWT  | Lista las categorías de práctica             |
| GET    | `/api/v1/practice/prompts?category=hr_interview`| JWT  | Lista prompts de una categoría               |

`GET /api/v1/practice/prompts?category=<code>`: si no existe la categoría devuelve `404 not_found`.

### Práctica — sesiones

| Método | Path                                          | Auth | Descripción                                          |
|--------|-----------------------------------------------|------|------------------------------------------------------|
| POST   | `/api/v1/practice/sessions`                   | JWT  | Crea una sesión de práctica                          |
| GET    | `/api/v1/practice/sessions?limit=&offset=`    | JWT  | Historial paginado del usuario                       |
| GET    | `/api/v1/practice/sessions/:id`               | JWT  | Detalle de una sesión + feedback asociado            |

`POST /api/v1/practice/sessions`:

```json
{
  "categoryCode": "hr_interview",
  "promptId": "uuid-opcional",
  "userTranscript": "texto opcional (max 5000 chars)"
}
```

Reglas:

- `categoryCode` es obligatorio y debe existir.
- Si `promptId` viene, debe existir y pertenecer a la categoría indicada (si no, `400 validation_error`).
- `userTranscript` es opcional acá; en Fase 4 se completa via `/analyze-audio`.
- Status inicial siempre `"completed"` (en Fase 4 se sumarán estados intermedios).

Respuesta `201 Created`:

```json
{
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "categoryId": "uuid",
    "categoryCode": "hr_interview",
    "promptId": null,
    "userTranscript": "...",
    "status": "completed",
    "createdAt": "..."
  }
}
```

`GET /api/v1/practice/sessions`: paginado.

```json
{
  "data": {
    "items": [ /* Session[] */ ],
    "limit": 20,
    "offset": 0
  }
}
```

`limit` default 20, máx 100. `offset` default 0.

`GET /api/v1/practice/sessions/:id`: detalle. Devuelve `404 not_found` si la sesión no existe **o no pertenece al usuario** (no diferenciamos para no facilitar enumeración de IDs ajenos).

```json
{
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "categoryCode": "hr_interview",
    "userTranscript": "...",
    "status": "completed",
    "createdAt": "...",
    "feedback": {
      "id": "uuid",
      "sessionId": "uuid",
      "correctedText": "...",
      "naturalVersion": "...",
      "explanationEs": "...",
      "comprehensionScore": 80,
      "vocabularyScore": 85,
      "communicationScore": 75,
      "createdAt": "..."
    }
  }
}
```

`feedback` puede ser `null` si la sesión todavía no fue procesada por IA.

### Práctica — analyze-audio

| Método | Path                                | Auth | Descripción                                                |
|--------|-------------------------------------|------|------------------------------------------------------------|
| POST   | `/api/v1/practice/analyze-audio`    | JWT  | Recibe audio + `categoryCode` + `promptText`. Transcribe, genera feedback, crea sesión y persiste todo en una sola llamada. |

**Content-Type**: `multipart/form-data`.

Campos:

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `audio` | file | sí | Archivo de audio. MIME permitidos: `audio/webm`, `audio/wav`, `audio/x-wav`, `audio/wave`, `audio/mpeg`, `audio/mp3`, `audio/ogg`, `audio/mp4`, `audio/x-m4a`, `audio/m4a`. Tamaño máximo: `MAX_AUDIO_BYTES` (default 15 MB). |
| `categoryCode` | string | sí | Code de categoría existente. |
| `promptText` | string | no | Texto del prompt original. Máx 2000 chars. |

Errores:

| Status | `code` | Cuándo |
|--------|--------|--------|
| 400 | `validation_error` | Falta `audio` o `categoryCode`, archivo vacío, `promptText` muy largo |
| 401 | `unauthorized` | Token ausente o inválido |
| 404 | `not_found` | `categoryCode` no existe |
| 413 | `payload_too_large` | Audio supera `MAX_AUDIO_BYTES` |
| 415 | `unsupported_media` | MIME no permitido |
| 500 | `internal_error` | Falla del provider de transcripción / feedback |

Respuesta `201 Created`:

```json
{
  "data": {
    "sessionId": "uuid",
    "categoryCode": "hr_interview",
    "transcript": "[mock] answer.webm (12345 bytes): I worked on a database related bug yesterday and I fixed it...",
    "correctedText": "I worked on a database related bug yesterday.",
    "naturalVersion": "Yesterday I worked on a database-related bug and fixed it.",
    "explanationEs": "Tu respuesta se entiende y es directa...",
    "scores": {
      "comprehension": 78,
      "vocabulary": 84,
      "communication": 72
    },
    "strengths": ["Buena estructura.", "Vocabulario tecnico correcto."],
    "improvements": ["Usa conectores.", "Elegi vocabulario de IT en contexto."],
    "suggestedPhrases": [
      { "spanish": "Identifiqué la causa raíz.", "english": "I identified the root cause." }
    ],
    "createdAt": "2026-05-26T20:15:00Z"
  }
}
```

### Práctica — analyze-text

| Método | Path                             | Auth | Descripción                                                                 |
|--------|----------------------------------|------|-----------------------------------------------------------------------------|
| POST   | `/api/v1/practice/analyze-text`  | JWT  | Recibe texto en inglés + `categoryCode` + `promptText`. Genera feedback, crea sesión y persiste (writing, daily escrito, resumen de quiz). |

**Content-Type**: `application/json`.

```json
{
  "categoryCode": "daily_meeting",
  "promptText": "What did you do yesterday, what will you do today, and do you have any blockers?",
  "userText": "Yesterday I finished the API fix. Today I will review the tests."
}
```

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `categoryCode` | string | sí | Code de categoría existente. |
| `promptText` | string | no | Consigna original. Máx 2000 chars. |
| `userText` | string | sí | Texto del usuario en inglés. Máx 5000 chars. |

Errores: mismos códigos que analyze-audio (sin 413/415 de audio). Respuesta `201`: mismo shape que analyze-audio (`transcript` = `userText` enviado).

Notas:

- `AI_PROVIDER=mock` (default): coach simulado, deterministico por hash del transcript.
- `AI_PROVIDER=openai`: Whisper (`OPENAI_TRANSCRIPTION_MODEL`) + Chat Completions con JSON (`OPENAI_FEEDBACK_MODEL`). Requiere `OPENAI_API_KEY`. Recomendado `HTTP_WRITE_TIMEOUT=90s` o mas.
- Los scores se persisten clamped a `0..100`. Los campos `strengths`, `improvements` y `suggestedPhrases` viajan solo en esta respuesta; **no** se persisten en `ai_feedback_results`. `GET /sessions/:id` solo expone los campos persistidos.
- El header `Authorization` y el contenido del audio nunca se loguean.

---

## Reglas de seguridad de los endpoints

- Todas las rutas bajo `/api/v1/*` (excepto `auth/*` y `health`) requieren JWT.
- El header `Authorization: Bearer <token>` nunca se loguea (regla de `AGENTS.md`).
- Validar MIME y tamaño máximo del audio antes de tocar OpenAI.
- Los errores internos se traducen a un `code` + `message` seguros para el cliente; nunca se reenvía el mensaje del error original.

---

## Códigos de error sugeridos

| `code`              | Cuándo                                              |
|---------------------|-----------------------------------------------------|
| `validation_error`  | Payload inválido / campos faltantes                 |
| `invalid_credentials` | Email o password incorrectos                      |
| `unauthorized`      | Token ausente, vencido o inválido                   |
| `forbidden`         | Token válido pero sin permiso sobre el recurso      |
| `not_found`         | Recurso no existe o no pertenece al usuario         |
| `conflict`          | Email ya registrado, etc.                           |
| `payload_too_large` | Audio supera el límite                              |
| `unsupported_media` | MIME del audio no permitido                         |
| `internal_error`    | Fallback genérico                                   |
