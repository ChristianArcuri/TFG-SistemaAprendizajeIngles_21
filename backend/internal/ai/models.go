// Package ai aisla la integracion con motores de transcripcion y feedback.
//
// Define interfaces (FeedbackService, TranscriptionService), mocks y
// implementaciones OpenAI. Cambiar AI_PROVIDER=openai en la config
// enchufa Whisper + Chat sin tocar handlers ni servicios de practica.
package ai

import "io"

// SuggestedPhrase es un par espanol/ingles util para que el usuario aprenda
// como decir lo que intentaba decir.
type SuggestedPhrase struct {
	Spanish string `json:"spanish"`
	English string `json:"english"`
}

// FeedbackInput agrupa los datos que el coach de IA necesita para evaluar
// una respuesta. PracticeContext deberia describir categoria y prompt
// original (ej. "category=tech_interview prompt=...").
type FeedbackInput struct {
	Transcript      string
	PracticeContext string
}

// FeedbackResult es la salida del coach. Los scores van de 0 a 100.
type FeedbackResult struct {
	CorrectedText      string            `json:"correctedText"`
	NaturalVersion     string            `json:"naturalVersion"`
	ExplanationES      string            `json:"explanationEs"`
	Strengths          []string          `json:"strengths"`
	Improvements       []string          `json:"improvements"`
	ComprehensionScore int               `json:"comprehensionScore"`
	VocabularyScore    int               `json:"vocabularyScore"`
	CommunicationScore int               `json:"communicationScore"`
	SuggestedPhrases   []SuggestedPhrase `json:"suggestedPhrases"`
}

// TranscriptionInput agrupa el audio recibido por el endpoint y los
// metadatos que el provider puede necesitar. El backend no toma posesion
// del Reader: el caller debe cerrarlo.
type TranscriptionInput struct {
	Reader   io.Reader
	Filename string
	MimeType string
	Size     int64
}

// TranscriptionResult contiene el texto transcripto y, si el provider lo
// expone, el idioma detectado.
type TranscriptionResult struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
}
