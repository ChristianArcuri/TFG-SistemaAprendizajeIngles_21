package ai

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/remotetalk/backend/internal/config"
)

// OpenAIClient envuelve el SDK oficial con modelos, timeouts y logging seguro.
type OpenAIClient struct {
	raw                  *openai.Client
	transcriptionModel   openai.AudioModel
	feedbackModel        openai.ChatModel
	transcriptionTimeout time.Duration
	feedbackTimeout      time.Duration
	maxAudioBytes        int64
}

// NewOpenAIClient construye el wrapper a partir de la config validada.
func NewOpenAIClient(cfg *config.Config) *OpenAIClient {
	client := openai.NewClient(option.WithAPIKey(cfg.OpenAI.APIKey))
	return &OpenAIClient{
		raw:                  &client,
		transcriptionModel:   openai.AudioModel(cfg.OpenAI.TranscriptionModel),
		feedbackModel:        openai.ChatModel(cfg.OpenAI.FeedbackModel),
		transcriptionTimeout: cfg.OpenAI.TranscriptionTimeout,
		feedbackTimeout:      cfg.OpenAI.FeedbackTimeout,
		maxAudioBytes:        cfg.AI.MaxAudioBytes,
	}
}

func (c *OpenAIClient) withTranscriptionTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.transcriptionTimeout)
}

func (c *OpenAIClient) withFeedbackTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.feedbackTimeout)
}

// logProviderError registra metadatos del fallo sin API key, audio ni transcript.
func logProviderError(ctx context.Context, operation string, err error) {
	attrs := []any{
		slog.String("operation", operation),
		slog.String("error", err.Error()),
	}
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		attrs = append(attrs,
			slog.Int("status_code", apiErr.StatusCode),
			slog.String("error_type", apiErr.Type),
			slog.String("error_code", apiErr.Code),
		)
		if apiErr.Response != nil {
			if reqID := apiErr.Response.Header.Get("x-request-id"); reqID != "" {
				attrs = append(attrs, slog.String("request_id", reqID))
			}
		}
	}
	slog.ErrorContext(ctx, "openai_request_failed", attrs...)
}
