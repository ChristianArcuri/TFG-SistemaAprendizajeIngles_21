package ai

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand/v2"
	"strings"
)

// MockFeedback genera un FeedbackResult con datos plausibles y scores
// aleatorios entre 60 y 90, deterministicos por hash del transcript.
// Sirve para desarrollar y testear el flujo completo sin tocar OpenAI.
type MockFeedback struct{}

// NewMockFeedback construye el mock.
func NewMockFeedback() *MockFeedback { return &MockFeedback{} }

// Generate cumple FeedbackService.
func (m *MockFeedback) Generate(_ context.Context, in FeedbackInput) (FeedbackResult, error) {
	rng := newRand(in.Transcript)

	transcript := strings.TrimSpace(in.Transcript)
	if transcript == "" {
		transcript = "I worked on a database related bug yesterday and fixed it."
	}

	return FeedbackResult{
		CorrectedText:  correctText(transcript),
		NaturalVersion: naturalize(transcript),
		ExplanationES: "Tu respuesta se entiende y es directa. Para sonar mas natural, " +
			"evita traducir literal del espanol y usa el pasado simple cuando " +
			"contas algo que ya termino.",
		Strengths: []string{
			"Buena estructura de la respuesta.",
			"Vocabulario tecnico correcto.",
			"Mantenes el foco en el problema y la solucion.",
		},
		Improvements: []string{
			"Usa conectores ('first', 'then', 'finally') para ordenar tu explicacion.",
			"Reemplaza 'I work' por 'I worked' cuando hablas de algo del pasado.",
			"Elegi vocabulario de IT en contexto, no una traduccion literal.",
		},
		ComprehensionScore: scoreInRange(rng),
		VocabularyScore:    scoreInRange(rng),
		CommunicationScore: scoreInRange(rng),
		SuggestedPhrases: []SuggestedPhrase{
			{Spanish: "Trabajé en un bug ayer.", English: "I worked on a bug yesterday."},
			{Spanish: "Identifiqué la causa raíz.", English: "I identified the root cause."},
			{Spanish: "Hice el deploy del fix.", English: "I deployed the fix."},
		},
	}, nil
}

// MockTranscription devuelve un texto simulado que incluye el nombre del
// archivo y su tamano para que sea facil identificar en logs / tests
// que la transcripcion vino del mock.
type MockTranscription struct{}

// NewMockTranscription construye el mock.
func NewMockTranscription() *MockTranscription { return &MockTranscription{} }

// Transcribe cumple TranscriptionService. Consume el Reader (para detectar
// audios vacios) pero no usa el contenido.
func (m *MockTranscription) Transcribe(_ context.Context, in TranscriptionInput) (TranscriptionResult, error) {
	read, _ := io.Copy(io.Discard, in.Reader)
	if read == 0 && in.Size == 0 {
		return TranscriptionResult{}, fmt.Errorf("mock transcription: empty audio")
	}

	filename := in.Filename
	if filename == "" {
		filename = "audio"
	}

	return TranscriptionResult{
		Text: fmt.Sprintf(
			"[mock] %s (%d bytes): I worked on a database related bug yesterday "+
				"and I fixed it after identifying the root cause.",
			filename, in.Size,
		),
		Language: "en",
	}, nil
}

// newRand crea un PRNG deterministico a partir del hash SHA-256 del input.
// Permite que los mismos transcripts generen siempre los mismos scores,
// para que los tests sean predecibles.
func newRand(seed string) *rand.Rand {
	sum := sha256.Sum256([]byte(seed))
	a := binary.BigEndian.Uint64(sum[0:8])
	b := binary.BigEndian.Uint64(sum[8:16])
	if a == 0 {
		a = 1
	}
	if b == 0 {
		b = 1
	}
	return rand.New(rand.NewPCG(a, b))
}

// scoreInRange devuelve un entero entre 40 y 95 inclusive.
// El piso bajo 60 permite que la regla de refuerzo del prototipo pueda activarse.
func scoreInRange(rng *rand.Rand) int {
	const minScore, maxScore = 40, 95
	return minScore + rng.IntN(maxScore-minScore+1)
}

func correctText(s string) string {
	return strings.NewReplacer(
		"i work in", "I worked on",
		"I work in", "I worked on",
		"i work on", "I worked on",
		"a bug from", "a bug related to",
	).Replace(s)
}

func naturalize(s string) string {
	if s == "" {
		return ""
	}
	if !strings.HasSuffix(s, ".") {
		s += "."
	}
	return strings.NewReplacer(
		"i work in", "I worked on",
		"I work in", "I worked on",
	).Replace(s)
}
