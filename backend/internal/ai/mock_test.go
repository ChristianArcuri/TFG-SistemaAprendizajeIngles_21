package ai

import (
	"context"
	"strings"
	"testing"
)

func TestMockFeedbackDeterministic(t *testing.T) {
	m := NewMockFeedback()
	in := FeedbackInput{Transcript: "I worked on a bug yesterday."}

	first, err := m.Generate(context.Background(), in)
	if err != nil {
		t.Fatalf("generate 1: %v", err)
	}
	second, err := m.Generate(context.Background(), in)
	if err != nil {
		t.Fatalf("generate 2: %v", err)
	}

	if first.ComprehensionScore != second.ComprehensionScore ||
		first.VocabularyScore != second.VocabularyScore ||
		first.CommunicationScore != second.CommunicationScore {
		t.Fatalf("expected deterministic scores: %+v vs %+v", first, second)
	}
}

func TestMockFeedbackScoreRange(t *testing.T) {
	m := NewMockFeedback()
	for _, transcript := range []string{"a", "b", "I worked", "long sentence here"} {
		out, err := m.Generate(context.Background(), FeedbackInput{Transcript: transcript})
		if err != nil {
			t.Fatalf("generate %q: %v", transcript, err)
		}
		for name, score := range map[string]int{
			"comprehension": out.ComprehensionScore,
			"vocabulary":    out.VocabularyScore,
			"communication": out.CommunicationScore,
		} {
			if score < 40 || score > 95 {
				t.Errorf("%s score %d out of range [40, 95] for %q", name, score, transcript)
			}
		}
		if len(out.Strengths) == 0 || len(out.Improvements) == 0 || len(out.SuggestedPhrases) == 0 {
			t.Errorf("expected coaching fields populated for %q", transcript)
		}
	}
}

func TestMockTranscription(t *testing.T) {
	m := NewMockTranscription()
	out, err := m.Transcribe(context.Background(), TranscriptionInput{
		Reader:   strings.NewReader("fake-audio-bytes"),
		Filename: "answer.webm",
		Size:     16,
	})
	if err != nil {
		t.Fatalf("transcribe: %v", err)
	}
	if out.Text == "" {
		t.Fatal("expected transcription text")
	}
	if out.Language != "en" {
		t.Fatalf("expected language en, got %q", out.Language)
	}
	if !strings.Contains(out.Text, "answer.webm") {
		t.Fatalf("expected filename in transcription text, got %q", out.Text)
	}
}

func TestMockTranscriptionEmpty(t *testing.T) {
	m := NewMockTranscription()
	_, err := m.Transcribe(context.Background(), TranscriptionInput{
		Reader: strings.NewReader(""),
		Size:   0,
	})
	if err == nil {
		t.Fatal("expected error for empty audio")
	}
}

func TestRenderFeedbackPrompt(t *testing.T) {
	out, err := RenderFeedbackPrompt(FeedbackInput{
		Transcript:      "Hello world",
		PracticeContext: "category=hr_interview",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out, "Hello world") {
		t.Fatal("transcript not rendered")
	}
	if !strings.Contains(out, "category=hr_interview") {
		t.Fatal("context not rendered")
	}
	if !strings.Contains(out, "JSON") {
		t.Fatal("rules not rendered")
	}
}
