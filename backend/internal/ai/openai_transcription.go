package ai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/openai/openai-go/v3"
)

// OpenAITranscription transcribe audio via OpenAI Whisper (u otro AudioModel).
type OpenAITranscription struct {
	client *OpenAIClient
}

// NewOpenAITranscription construye el servicio.
func NewOpenAITranscription(client *OpenAIClient) *OpenAITranscription {
	return &OpenAITranscription{client: client}
}

// Transcribe cumple TranscriptionService.
func (s *OpenAITranscription) Transcribe(ctx context.Context, in TranscriptionInput) (TranscriptionResult, error) {
	ctx, cancel := s.client.withTranscriptionTimeout(ctx)
	defer cancel()

	data, err := readAudioPayload(in.Reader, in.Size, s.client.maxAudioBytes)
	if err != nil {
		return TranscriptionResult{}, err
	}
	if len(data) == 0 {
		return TranscriptionResult{}, fmt.Errorf("empty audio payload")
	}

	filename, mimeType := audioUploadMeta(in.Filename, in.MimeType)

	resp, err := s.client.raw.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		File:     openai.File(bytes.NewReader(data), filename, mimeType),
		Model:    s.client.transcriptionModel,
		Language: openai.String("en"),
	})
	if err != nil {
		logProviderError(ctx, "transcription", err)
		return TranscriptionResult{}, ErrProviderFailure
	}

	tr := resp.AsTranscription()
	text := strings.TrimSpace(tr.Text)
	if text == "" {
		slog.ErrorContext(ctx, "openai_transcription_empty_text",
			slog.String("model", string(s.client.transcriptionModel)),
		)
		return TranscriptionResult{}, ErrProviderFailure
	}

	slog.InfoContext(ctx, "openai_transcription_ok",
		slog.String("model", string(s.client.transcriptionModel)),
		slog.Int("transcript_len", len(text)),
		slog.Int64("audio_bytes", int64(len(data))),
		slog.String("filename", filename),
	)

	return TranscriptionResult{Text: text, Language: "en"}, nil
}

func readAudioPayload(r io.Reader, declaredSize, maxBytes int64) ([]byte, error) {
	limit := maxBytes
	if declaredSize > 0 && declaredSize < limit {
		limit = declaredSize
	}
	if limit <= 0 {
		return nil, fmt.Errorf("invalid audio size limit")
	}

	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read audio: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("audio exceeds maximum size")
	}
	return data, nil
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "audio.webm"
	}
	// Evitar paths en el nombre del archivo del multipart.
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return name
}

// audioUploadMeta prepara filename y MIME para el multipart de Whisper.
// Sin nombre de archivo con extension, OpenAI responde "Unrecognized file format".
func audioUploadMeta(filename, mimeType string) (string, string) {
	name := sanitizeFilename(filename)
	mime := strings.ToLower(strings.TrimSpace(mimeType))
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	switch mime {
	case "", "application/octet-stream":
		mime = "audio/webm"
	case "video/webm":
		mime = "audio/webm"
	}
	if !strings.Contains(name, ".") {
		switch mime {
		case "audio/webm":
			name += ".webm"
		case "audio/wav", "audio/x-wav", "audio/wave":
			name += ".wav"
		case "audio/mpeg", "audio/mp3":
			name += ".mp3"
		case "audio/ogg":
			name += ".ogg"
		case "audio/mp4", "audio/m4a", "audio/x-m4a":
			name += ".m4a"
		}
	}
	return name, mime
}
