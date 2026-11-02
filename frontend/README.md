# Inglés Profesional IT — Frontend

SPA en **Vue 3 + Vite + Tailwind CSS + Pinia + Axios**.

## Requisitos

- Node.js 20+
- Backend corriendo en `http://localhost:8080` (Docker o `go run`)

## Configuración

```bash
cp .env.example .env
npm install
```

`.env`:

```env
VITE_API_BASE_URL=http://localhost:8080
```

## Desarrollo

```bash
npm run dev
```

Abrí [http://localhost:5173](http://localhost:5173).

El backend debe tener CORS para `http://localhost:5173` (ya viene por defecto en `CORS_ALLOWED_ORIGINS`).

## Pantallas

| Ruta | Descripción |
|------|-------------|
| `/` | Landing |
| `/login` | Login y registro |
| `/onboarding` | Perfil inicial |
| `/dashboard` | Los tres módulos de práctica |
| `/practice/hr-interview` | Entrevista con Recursos Humanos |
| `/practice/daily/:categoryCode` | Daily meeting |
| `/quiz` | Quiz de inglés IT |
| `/results/:sessionId` | Feedback de la sesión |
| `/history` | Historial de sesiones |

## Flujo principal

1. Registro / login → JWT en `localStorage` (`tfg_sai_token`)
2. Onboarding → `PUT /api/v1/me/profile`
3. Dashboard → `GET /api/v1/practice/categories`
4. Práctica → grabación con `MediaRecorder` → `POST /api/v1/practice/analyze-audio`
5. Resultados → scores, transcripción, versión natural, explicación en español

## Build

```bash
npm run build
npm run preview
```

