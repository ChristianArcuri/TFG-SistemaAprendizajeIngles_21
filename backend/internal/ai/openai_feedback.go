package ai

import (
	"context"
	"log/slog"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

// feedbackSystemPrompt es el rol fijo del coach; el user message lleva el
// transcript y el contexto de la practica (RenderFeedbackPrompt).
const feedbackSystemPrompt = `You are an English coach for LATAM software developers.
The user is practicing English for remote IT jobs.
Their native language is Spanish.
They need practical feedback, not academic explanations.
Return valid JSON only. Do not use markdown.`

// OpenAIFeedback genera feedback estructurado via Chat Completions.
type OpenAIFeedback struct {
	client *OpenAIClient
}

// NewOpenAIFeedback construye el servicio.
func NewOpenAIFeedback(client *OpenAIClient) *OpenAIFeedback {
	return &OpenAIFeedback{client: client}
}

// Generate cumple FeedbackService.
func (s *OpenAIFeedback) Generate(ctx context.Context, in FeedbackInput) (FeedbackResult, error) {
	ctx, cancel := s.client.withFeedbackTimeout(ctx)
	defer cancel()

	userPrompt, err := RenderFeedbackPrompt(in)
	if err != nil {
		return FeedbackResult{}, err
	}

	jsonFormat := openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONObject: openai.Ptr(shared.NewResponseFormatJSONObjectParam()),
	}

	resp, err := s.client.raw.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: s.client.feedbackModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(feedbackSystemPrompt),
			openai.UserMessage(userPrompt),
		},
		ResponseFormat: jsonFormat,
		Temperature:    openai.Float(0.4),
	})
	if err != nil {
		logProviderError(ctx, "feedback", err)
		return FeedbackResult{}, ErrProviderFailure
	}

	if len(resp.Choices) == 0 {
		slog.ErrorContext(ctx, "openai_feedback_no_choices",
			slog.String("model", string(s.client.feedbackModel)),
		)
		return FeedbackResult{}, ErrProviderFailure
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		slog.ErrorContext(ctx, "openai_feedback_empty_content",
			slog.String("model", string(s.client.feedbackModel)),
		)
		return FeedbackResult{}, ErrProviderFailure
	}

	result, err := parseFeedbackJSON(content)
	if err != nil {
		slog.ErrorContext(ctx, "openai_feedback_invalid_json",
			slog.String("model", string(s.client.feedbackModel)),
			slog.Int("content_len", len(content)),
			slog.String("error", err.Error()),
		)
		return FeedbackResult{}, ErrProviderFailure
	}

	slog.InfoContext(ctx, "openai_feedback_ok",
		slog.String("model", string(s.client.feedbackModel)),
		slog.Int("transcript_len", len(in.Transcript)),
		slog.Int("response_len", len(content)),
	)

	return result, nil
}
