package ai

import (
	"log/slog"

	"github.com/remotetalk/backend/internal/config"
)

// NewServices selecciona FeedbackService y TranscriptionService segun AI_PROVIDER.
func NewServices(cfg *config.Config) (FeedbackService, TranscriptionService) {
	switch cfg.AI.Provider {
	case "openai":
		client := NewOpenAIClient(cfg)
		slog.Info("ai_provider_openai_enabled",
			slog.String("transcription_model", cfg.OpenAI.TranscriptionModel),
			slog.String("feedback_model", cfg.OpenAI.FeedbackModel),
		)
		return NewOpenAIFeedback(client), NewOpenAITranscription(client)
	default:
		slog.Info("ai_provider_mock_enabled")
		return NewMockFeedback(), NewMockTranscription()
	}
}
