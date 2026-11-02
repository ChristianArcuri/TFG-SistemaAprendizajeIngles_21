# Inglés Profesional IT

Prototipo del Trabajo Final de Grado: sistema web para entrenar inglés profesional en situaciones de IT.

Incluye solo los módulos del trabajo práctico: entrevista con Recursos Humanos, daily meeting y quiz de inglés IT.

## Cómo funciona

El usuario se registra o ingresa, completa un perfil y practica desde el dashboard. Cada práctica queda guardada. El sistema muestra el historial, el promedio de tres competencias y, cuando corresponde, una recomendación de refuerzo.

La aplicación tiene tres partes:

- **Frontend:** Vue 3, Vite, Tailwind CSS y Pinia. Corre en `http://localhost:5173`.
- **Backend:** API en Go (Gin). Corre en `http://localhost:8080`.
- **Base de datos:** PostgreSQL 16, levantada con Docker. Guarda usuarios, perfiles, sesiones y puntajes.

El ingreso usa email y contraseña. La contraseña se guarda con bcrypt. La sesión del navegador usa un JWT. No hay inicio de sesión con Google ni envío de correo de verificación.

La inteligencia artificial se usa solo en la entrevista y en la daily: transcribe el audio y devuelve feedback lingüístico. El quiz se corrige con la clave de respuestas, sin llamar a la IA.

Hay dos modos de IA, definidos en `backend/.env`:

- `AI_PROVIDER=mock` recorre todo el flujo sin clave de OpenAI. El audio no se transcribe de verdad: el sistema usa un texto de ejemplo y un feedback simulado.
- `AI_PROVIDER=openai` transcribe el audio real y genera el feedback. Hace falta `OPENAI_API_KEY` y `HTTP_WRITE_TIMEOUT=90s`.

## Instalación

Requisitos para la demo: Docker Desktop y Node.js 20 o posterior.

Go 1.25 hace falta solo si se ejecutan las pruebas del backend o si la API se levanta fuera de Docker.

1. Clonar el repositorio y entrar a la carpeta.
2. Crear la configuración del backend a partir del ejemplo:

```bash
cd backend
cp .env.example .env
```

En PowerShell:

```powershell
cd backend
copy .env.example .env
```

Para la demo, dejá `AI_PROVIDER=mock`. El archivo `.env` no se sube al repositorio porque puede contener la clave de OpenAI.

3. Instalar el frontend:

```bash
cd frontend
npm install
```

El frontend usa `http://localhost:8080` aunque no exista `frontend/.env`. Si hace falta cambiar la URL de la API, copiá `frontend/.env.example` a `frontend/.env`.

## Cómo levantarlo

En una terminal, desde `backend`:

```bash
docker compose up --build
```

La primera vez crea la base, aplica las migraciones y carga el usuario de demo. La API queda en `http://localhost:8080`. Podés comprobar que responde en `http://localhost:8080/health`.

En otra terminal, desde `frontend`:

```bash
npm run dev
```

La interfaz queda en `http://localhost:5173`.

Para detener el backend, en la terminal de Docker usá `Ctrl+C` y después `docker compose down`.

Usuario de demo, creado por la migración:

- Email: `admintfg@gmail.com`
- Contraseña: `AdminTFG`

También se puede crear otra cuenta desde **Empezar gratis**. La contraseña debe tener al menos 8 caracteres. La primera vez el sistema pide completar el perfil.

Para usar transcripción y feedback reales, en `backend/.env` cambiá `AI_PROVIDER=openai`, completá `OPENAI_API_KEY` y subí `HTTP_WRITE_TIMEOUT=90s`. Después volvé a levantar el backend.

## Cómo usar cada módulo

### Ingreso y perfil

1. En la página de inicio, entrá con **Ingresar**.
2. Usá el usuario de demo o registrate.
3. Completá el perfil: objetivo, rol, nivel aproximado de inglés y stack principal.
4. **Crear mi dashboard** abre la pantalla de práctica.

El perfil se pide una sola vez. Los objetivos son preparar entrevistas con Recursos Humanos, comunicarse en daily meetings o reforzar vocabulario técnico de IT.

### Dashboard

Muestra el saludo, los promedios de comprensión, vocabulario profesional y comunicación profesional, y los tres módulos en este orden:

1. Entrevista con Recursos Humanos.
2. Daily meeting.
3. Quiz de inglés IT.

Si ya hay historial suficiente, arriba aparece una tarjeta de refuerzo con el módulo sugerido. **Salir** cierra la sesión. **Progreso**, en el menú lateral, abre el historial.

### Entrevista con Recursos Humanos

Sirve para responder preguntas típicas de una entrevista laboral.

1. Entrá con **Practicar entrevista**.
2. A la derecha, elegí el tipo de pregunta: sobre vos, fortalezas, motivación, situaciones STAR, o remoto y cierre.
3. Leé la pregunta en inglés y el tip en español.
4. **Grabar respuesta**, hablá en inglés y **Detener grabación**. El navegador pide permiso para usar el micrófono.
5. **Enviar y ver feedback** guarda la sesión y abre el resultado.
6. **Siguiente pregunta** pasa a otra consigna del mismo tipo.

### Daily meeting

Sirve para contar qué se hizo, qué sigue y si hay bloqueos.

1. Entrá con **Empezar daily**.
2. Leé la consigna. La estructura sugerida es *yesterday*, *today* y *blockers*.
3. Escribí el update en inglés y usá **Corregir texto con IA**.
4. O grabá la daily con **Grabar daily** y **Enviar grabación**.

Las dos formas guardan la sesión y abren el resultado.

### Quiz de inglés IT

Sirve para elegir la frase más natural en situaciones de trabajo: daily, deploy, bugs, blockers, code review, APIs, base de datos, incidentes y comunicación con el cliente.

1. Entrá con **Resolver quiz**.
2. La sesión trae 10 preguntas. Antes de responder, a la derecha se puede cambiar el tema.
3. Elegí una opción. La pantalla marca la respuesta correcta y muestra la explicación en español.
4. **Siguiente pregunta** avanza. Si todavía no respondiste, ese botón omite la pregunta.
5. Al terminar, el resultado se guarda solo. El resumen indica cuántas salieron bien y qué temas conviene reforzar.
6. **Ver resultado** abre el detalle. **Nueva sesión** empieza otro quiz.

El puntaje es el porcentaje de respuestas correctas. Una sesión de 8/10 queda como un buen resultado y, si hubo errores, igual nombra los conceptos a reforzar. Menos de 60 se informa como resultado bajo.

### Resultado de la práctica

Después de la entrevista o la daily, la pantalla muestra:

- puntaje general y las competencias evaluadas;
- lo que se dijo o escribió;
- una versión corregida y una versión más natural;
- la explicación en español.

**Practicar otra vez** vuelve al mismo módulo.

### Progreso y refuerzo

**Progreso** lista las sesiones guardadas, incluidos los quiz. Cada fila abre el resultado de esa sesión.

Las competencias son comprensión, vocabulario profesional y comunicación profesional. La entrevista y la daily puntúan las tres. El quiz puntúa comprensión y vocabulario con el porcentaje de aciertos, y no puntúa comunicación profesional.

La recomendación de refuerzo la calcula el backend:

- mira las últimas cinco prácticas que evalúan cada competencia;
- si una queda bajo 60 en al menos tres de esas cinco, sugiere un módulo;
- comprensión recomienda la daily, vocabulario recomienda el quiz y comunicación recomienda la entrevista;
- si hay más de una, elige el promedio más bajo de esa ventana;
- con menos de cinco resultados válidos en una competencia, el progreso de esa competencia es provisorio y no se marca como dificultad recurrente.

La pronunciación y la entonación no se puntúan. La gramática queda como observación dentro de la explicación, no como una cuarta competencia.

## Pruebas del backend

Con Go instalado, desde `backend`:

```bash
gofmt -w .
go test ./...
```
