# Fases de contenido y práctica (frontend)

Plan para implementar **bancos de ejercicios**, **subtipos clicables** y **navegación entre consignas**.  
Se trabaja **uno por uno**, en el orden listado abajo.

**Estado actual (baseline):**

- Cada categoría usa 1 prompt fijo en `frontend/src/lib/categoryPrompts.js`.
- Quiz IT tiene 5 preguntas en `frontend/src/data/quizQuestions.js`.
- Writing / client support muestran chips decorativos (“Email al cliente”, etc.) sin funcionalidad.
- No hay “siguiente ejercicio”, skip ni historial por prompt individual.

**Convención propuesta:** prompts en JSON (o JS que exporte JSON) bajo `frontend/src/data/prompts/`.

---

## Esquema JSON sugerido (writing / speaking)

```json
{
  "categoryCode": "client_support",
  "subtypes": [
    { "id": "client_email", "label": "Email al cliente" },
    { "id": "slack", "label": "Mensaje de Slack" },
    { "id": "ticket_update", "label": "Update de ticket" }
  ],
  "prompts": [
    {
      "id": "cs-001",
      "subtypeId": "client_email",
      "title": "Más contexto del bug",
      "promptText": "Write a professional email asking the client for steps to reproduce the issue.",
      "hintEs": "Usá 'steps to reproduce' y tono calmado.",
      "difficulty": "medium"
    }
  ]
}
```

## Esquema JSON sugerido (quiz)

```json
{
  "id": "quiz-001",
  "topic": "daily_meeting",
  "situationEs": "Querés decir que estás trabajando en un bug de base de datos.",
  "hintEs": "¿Cuál suena más natural en una daily?",
  "options": ["...", "...", "..."],
  "correctIndex": 1,
  "explanationEs": "Para tareas usamos work on."
}
```

## Comportamiento UX común (todas las fases)

- **Siguiente ejercicio:** botón para avanzar sin enviar (no marca como practicado).
- **Aleatorio opcional:** al entrar, elegir prompt no visto recientemente (localStorage o backend futuro).
- **Hecho / practicado:** solo al enviar con IA (`completeCurrent`); badge **· practicado** en UI.
- **Subtipo activo:** filtrar prompts por `subtypeId` cuando el usuario hace click en un chip.

---

## Fase A — Atender consultas (`client_support`) ✅

**Prioridad:** 1 — **Implementado** (MVP frontend)

### Entregado

- `frontend/src/data/prompts/client_support.json` — **51 consignas** (17 email + 17 Slack + 17 ticket).
- `frontend/src/composables/usePracticePrompts.js` — subtipos, siguiente, saltar, vistos en `localStorage`.
- `frontend/src/views/ClientSupportPracticeView.vue` — ruta `/practice/client-support`.
- Chips clicables: Email al cliente · Mensaje de Slack · Update de ticket.

### Pendiente (futuro)

- Migrar banco a DB + `GET /practice/prompts?subtype=`.

---

## Fase B — Entrevista técnica (`tech_interview`) ✅

**Prioridad:** 2 — **Implementado** (MVP frontend)

### Entregado

- `frontend/src/data/prompts/tech_interview.json` — **50 preguntas** (10 por tema).
- `frontend/src/views/TechInterviewPracticeView.vue` — ruta `/practice/tech-interview`.
- Temas: Bugs y debugging · APIs y servicios · Bases de datos · Arquitectura · Diseño y trade-offs.
- Speaking + grabación + **Siguiente pregunta** + badge **· practicado** al enviar audio.

### Pendiente (futuro)

- Misma migración a DB que Fase A cuando corresponda.

---

## Fase C — Entrevista HR (`hr_interview`) ✅

**Prioridad:** 3 — **Implementado** (MVP frontend)

### Entregado

- `frontend/src/data/prompts/hr_interview.json` — **50 preguntas** (10 por bloque).
- `frontend/src/views/HrInterviewPracticeView.vue` — ruta `/practice/hr-interview`.
- Bloques: Sobre vos · Fortalezas/debilidades · Motivación · STAR · Remoto/salario.
- Speaking + **Siguiente pregunta** + **· practicado** al enviar audio.

### Pendiente (futuro)

- Migración a DB cuando corresponda.

---

## Fase D — Quiz de inglés IT (`it_quiz`) ✅

**Prioridad:** 4 — **Implementado** (MVP frontend)

### Entregado

- `frontend/src/data/prompts/it_quiz.json` — **100 preguntas** (10 temas × 10).
- `frontend/src/composables/useQuizSession.js` — sesiones de **10 preguntas**, barra de progreso.
- `QuizView.vue` refactorizado; eliminado `quizQuestions.js`.
- Temas: daily, deploy, bugs, blockers, PR, APIs, DB, incidentes, cliente, vocabulario.
- **Siguiente pregunta** sin responder (cuenta como omitida en resumen).
- Resumen de sesión + **Nueva sesión** + feedback IA.

### Pendiente (futuro)

- Migración a DB cuando corresponda.

---

## Fase E — Redactar emails (`writing_email`) ✅

**Prioridad:** 5 — **Implementado** (MVP frontend)

### Entregado

- `frontend/src/data/prompts/writing_email.json` — **50 ejercicios** (10 por tipo).
- `frontend/src/views/WritingEmailPracticeView.vue` — ruta `/practice/writing-email`.
- Tipos: Follow-up · Aviso de demora · Reporte de bug · Status update · Pedir información.
- Writing + **Siguiente ejercicio** + **· practicado** al corregir con IA.

### Pendiente (futuro)

- Migración a DB cuando corresponda.

---

## Orden de implementación acordado

| # | Fase | Categoría | Cantidad | Modo |
|---|------|-----------|----------|------|
| 1 | A | `client_support` | ~50 | Writing + 3 subtipos |
| 2 | B | `tech_interview` | ~50 | Speaking |
| 3 | C | `hr_interview` | ~50 | Speaking |
| 4 | D | `it_quiz` | ~100 | Quiz |
| 5 | E | `writing_email` | ~50 | Writing |

**Total aproximado:** ~300 ítems de práctica en frontend (MVP local).

---

## Tareas técnicas transversales (cuando toque cada fase)

1. Crear `frontend/src/data/prompts/<category>.json` con schema consistente.
2. Crear `frontend/src/composables/usePracticePrompts.js` (cargar banco, subtipo, random, next, skip, seen).
3. Persistir prompts vistos en `localStorage` (MVP).
4. Actualizar vistas: `WritingPracticeView`, `PracticeView`, `QuizView`.
5. (Futuro) Seed en DB + endpoint con paginación y `subtype` en `practice_prompts`.

---

## Fuera de scope de estas fases

- Pagos, admin panel, gamificación compleja.
- Tracking de progreso por prompt en backend (se puede agregar después).
- Generación dinámica de preguntas con IA (solo banco estático por ahora).

---

## Referencias en el repo

| Recurso | Ubicación |
|---------|-----------|
| Prompt único por categoría | `frontend/src/lib/categoryPrompts.js` |
| Quiz actual (5 preguntas) | `frontend/src/data/quizQuestions.js` |
| Config rutas/modos | `frontend/src/lib/categoryConfig.js` |
| Writing UI | `frontend/src/views/WritingPracticeView.vue` |
| Speaking UI | `frontend/src/views/PracticeView.vue` |
| Quiz UI | `frontend/src/views/QuizView.vue` |
| Mock estático HTML | `remotetalk-static/writing.html`, `quiz.html` |
