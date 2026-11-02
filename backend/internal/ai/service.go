package ai

import "context"

// FeedbackService devuelve un FeedbackResult para una respuesta del usuario.
// Implementaciones: MockFeedback (default) y la que llegara con OpenAI.
type FeedbackService interface {
	Generate(ctx context.Context, in FeedbackInput) (FeedbackResult, error)
}

// TranscriptionService devuelve el texto del audio recibido.
// Implementaciones: MockTranscription (default) y la que llegara con OpenAI Whisper.
type TranscriptionService interface {
	Transcribe(ctx context.Context, in TranscriptionInput) (TranscriptionResult, error)
}
