# Inglés Profesional IT

Prototipo del Trabajo Final de Grado: sistema web para entrenar inglés profesional en situaciones de IT.

Incluye solo los módulos del trabajo práctico.

## Módulos

- Registro, inicio de sesión y perfil (rol, tecnología, nivel y objetivo).
- Entrevista con Recursos Humanos, con respuesta oral o escrita, transcripción y feedback.
- Daily meeting: avances, próximas tareas y bloqueos.
- Quiz de inglés IT de opción múltiple. El puntaje sale de las respuestas correctas.
- Historial y promedios de tres competencias: comprensión, vocabulario profesional y comunicación profesional.
- Recomendación de refuerzo: si una competencia queda bajo 60 en al menos tres de las últimas cinco prácticas que la evalúan, el backend sugiere un módulo. Con menos de cinco resultados el progreso es provisorio.

La pronunciación y la entonación no se puntúan a partir del texto. La gramática queda como observación dentro del feedback.

## Cómo levantarlo

Requisitos: Docker, Go 1.25 y Node.js.

1. En `backend`, el archivo `.env` ya está copiado desde `.env.example`. `AI_PROVIDER=mock` permite la demo sin clave de OpenAI.
2. Desde `backend`:

```bash
docker compose up --build
```

La API queda en `http://localhost:8080`.

3. Desde `frontend`:

```bash
npm install
npm run dev
```

La interfaz queda en `http://localhost:5173`.

Usuario de demo, creado por la migración:

- Email: `admintfg@gmail.com`
- Contraseña: `AdminTFG`

La primera vez pide completar el perfil. El ingreso es por email y contraseña guardados en PostgreSQL. No usa la cuenta real de Gmail ni el botón de Google.

Para usar transcripción y feedback reales, en `backend/.env` cambiá `AI_PROVIDER=openai`, completá `OPENAI_API_KEY` y subí `HTTP_WRITE_TIMEOUT=90s`.

## Pruebas del backend

```bash
cd backend
gofmt -w .
go test ./...
```
