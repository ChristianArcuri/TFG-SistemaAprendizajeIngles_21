// Package practice implementa el dominio de practica:
// catalogo de categorias, prompts, sesiones y feedback de IA.
//
// Capas:
//   - models.go      tipos del dominio y DTOs HTTP
//   - errors.go      errores tipados del paquete
//   - repository.go  acceso a Postgres (pgx)
//   - service.go     reglas de negocio
//   - handler.go     handlers HTTP finos
package practice

import (
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/remotetalk/backend/internal/ai"
)

// SessionStatus enumera los estados posibles de una sesion.
// Por ahora todas las sesiones nacen como SessionStatusCompleted; cuando
// integremos audio + IA en la Fase 4 sumaremos los estados intermedios.
type SessionStatus string

const (
	SessionStatusCompleted SessionStatus = "completed"
)

// Category es un tipo de practica del catalogo (HR interview, daily, etc.).
type Category struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Prompt es una consigna asociada a una categoria.
type Prompt struct {
	ID         uuid.UUID `json:"id"`
	CategoryID uuid.UUID `json:"categoryId"`
	Title      string    `json:"title"`
	PromptText string    `json:"promptText"`
	Difficulty string    `json:"difficulty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Session es una sesion de practica concreta de un usuario.
// CategoryCode se hidrata via JOIN para que el cliente no necesite resolverlo.
// Scores se hidrata cuando existe feedback de IA (historial / progreso).
type Session struct {
	ID             uuid.UUID     `json:"id"`
	UserID         uuid.UUID     `json:"userId"`
	CategoryID     uuid.UUID     `json:"categoryId"`
	CategoryCode   string        `json:"categoryCode"`
	PromptID       *uuid.UUID    `json:"promptId,omitempty"`
	UserTranscript string        `json:"userTranscript,omitempty"`
	Status         SessionStatus `json:"status"`
	Scores         *Scores       `json:"scores,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
}

// Feedback es el resultado del coach de IA sobre una sesion (1:1).
type Feedback struct {
	ID                 uuid.UUID `json:"id"`
	SessionID          uuid.UUID `json:"sessionId"`
	CorrectedText      string    `json:"correctedText,omitempty"`
	NaturalVersion     string    `json:"naturalVersion,omitempty"`
	ExplanationES      string    `json:"explanationEs,omitempty"`
	ComprehensionScore int       `json:"comprehensionScore"`
	VocabularyScore    int       `json:"vocabularyScore"`
	CommunicationScore *int      `json:"communicationScore,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
}

// SessionDetail es la respuesta de GET /sessions/:id: la sesion + feedback opcional.
type SessionDetail struct {
	Session
	Feedback *Feedback `json:"feedback,omitempty"`
}

// CreateSessionInput es el payload de POST /sessions.
//
// promptId es opcional (sesiones libres). userTranscript tambien lo es:
// se va a poder llenar mas tarde via /transcribe en la Fase 4. Aca lo
// permitimos para flujos de practica solo-escrita / quick test.
type CreateSessionInput struct {
	CategoryCode   string     `json:"categoryCode"   binding:"required"`
	PromptID       *uuid.UUID `json:"promptId"`
	UserTranscript string     `json:"userTranscript" binding:"omitempty,max=5000"`
}

// ListSessionsParams controla la paginacion del historial.
type ListSessionsParams struct {
	Limit  int
	Offset int
}

// PaginatedSessions es la respuesta de GET /sessions.
type PaginatedSessions struct {
	Items  []Session `json:"items"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// AudioBlob describe el archivo de audio recibido por el endpoint
// /analyze-audio. El service no toma posesion del Reader: el handler lo
// abre y lo cierra. El service solo lee mientras llama a la transcripcion.
type AudioBlob struct {
	Reader   io.Reader
	Filename string
	MimeType string
	Size     int64
}

// AnalyzeAudioInput agrupa los datos del request de /analyze-audio.
type AnalyzeAudioInput struct {
	CategoryCode string
	PromptText   string
	Audio        AudioBlob
}

// SubmitQuizInput es el resultado de un quiz ya terminado.
// El puntaje sale de los aciertos. MissedTopics nombra los temas a reforzar.
type SubmitQuizInput struct {
	Correct      int
	Total        int
	MissedTopics []string
}

// AnalyzeTextInput agrupa los datos del request de /analyze-text.
type AnalyzeTextInput struct {
	CategoryCode string
	PromptText   string
	UserText     string
	QuizCorrect  int
	QuizTotal    int
}

// Scores es la forma plana que devolvemos al cliente en /analyze-audio.
// Scores guarda solo las competencias que la actividad evalua.
// En el quiz, Communication queda nil: no se infiere desde una opcion multiple.
type Scores struct {
	Comprehension *int `json:"comprehension,omitempty"`
	Vocabulary    *int `json:"vocabulary,omitempty"`
	Communication *int `json:"communication,omitempty"`
}

// AnalyzeAudioResult es la respuesta de POST /analyze-audio. Incluye los
// campos persistidos en DB (sessionId, transcript, correctedText, ...) y
// los campos de coaching que NO se persisten (strengths, improvements,
// suggestedPhrases) pero el frontend necesita para la pantalla de
// resultados.
type AnalyzeAudioResult struct {
	SessionID        uuid.UUID            `json:"sessionId"`
	CategoryCode     string               `json:"categoryCode"`
	Transcript       string               `json:"transcript"`
	CorrectedText    string               `json:"correctedText"`
	NaturalVersion   string               `json:"naturalVersion"`
	ExplanationES    string               `json:"explanationEs"`
	Scores           Scores               `json:"scores"`
	Strengths        []string             `json:"strengths"`
	Improvements     []string             `json:"improvements"`
	SuggestedPhrases []ai.SuggestedPhrase `json:"suggestedPhrases"`
	CreatedAt        time.Time            `json:"createdAt"`
}
