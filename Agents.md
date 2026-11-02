# Inglés Profesional IT - Agent Guidelines

## Product
Prototipo del TFG. Versión acotada para entrenar inglés profesional de IT en tres módulos: entrevista con Recursos Humanos, daily meeting y quiz de inglés IT. Incluye perfil, historial, tres competencias y recomendación de refuerzo.

## Backend Stack
- Go
- Gin
- PostgreSQL
- sqlc
- golang-migrate
- JWT
- OpenAI API
- Docker

## Architecture Rules
- Keep handlers thin.
- Business logic must be in services.
- Database access must be in repositories or sqlc generated code.
- Do not call OpenAI directly from handlers.
- Do not expose internal errors to clients.
- Use context.Context in service and repository methods.
- Return structured JSON responses.
- Validate request payloads.
- Keep code simple and explicit.

## Security Rules
- Never hardcode API keys.
- Read secrets from environment variables.
- Do not log authorization headers.
- Do not log full audio content or sensitive user data.
- Validate file size for audio uploads.
- Validate MIME type for uploaded audio.

## Testing Rules
- Add unit tests for services when possible.
- Do not refactor unrelated files.
- Do not introduce unnecessary abstractions.
- After changes, run:
  - gofmt
  - go test ./...

## MVP Scope
Build only:
- user registration/login
- user profile
- HR interview practice
- daily meeting practice
- IT English quiz
- audio transcription and linguistic feedback
- history, competency progress and reinforcement recommendation

Do not build:
- technical interview, client support or email writing modules
- payments
- avatars
- mobile app
- realtime voice conversation
- complex gamification
- admin panel
- pronunciation or intonation scores inferred from text