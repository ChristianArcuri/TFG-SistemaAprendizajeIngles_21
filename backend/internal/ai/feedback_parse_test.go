package ai

import (
	"strings"
	"testing"
)

func TestParseFeedbackJSONValid(t *testing.T) {
	raw := `{
		"correctedText": "I worked on a bug.",
		"naturalVersion": "Yesterday I worked on a bug.",
		"explanationEs": "Muy bien, pero mejora el pasado.",
		"strengths": ["Claro"],
		"improvements": ["Pronunciacion"],
		"comprehensionScore": 150,
		"vocabularyScore": 80,
		"communicationScore": -5,
		"suggestedPhrases": [{"spanish": "Hola", "english": "Hi"}]
	}`

	out, err := parseFeedbackJSON(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if out.CorrectedText != "I worked on a bug." {
		t.Fatalf("correctedText: %q", out.CorrectedText)
	}
	if out.ComprehensionScore != 100 {
		t.Fatalf("comprehension clamp: %d", out.ComprehensionScore)
	}
	if out.CommunicationScore != 0 {
		t.Fatalf("communication clamp: %d", out.CommunicationScore)
	}
	if len(out.SuggestedPhrases) != 1 {
		t.Fatalf("expected 1 phrase, got %d", len(out.SuggestedPhrases))
	}
}

func TestParseFeedbackJSONMissingFields(t *testing.T) {
	_, err := parseFeedbackJSON(`{"explanationEs":"solo esto"}`)
	if err == nil || !strings.Contains(err.Error(), "correctedText") {
		t.Fatalf("expected missing correctedText error, got %v", err)
	}

	_, err = parseFeedbackJSON(`{"correctedText":"ok"}`)
	if err == nil || !strings.Contains(err.Error(), "explanationEs") {
		t.Fatalf("expected missing explanationEs error, got %v", err)
	}
}

func TestParseFeedbackJSONInvalid(t *testing.T) {
	_, err := parseFeedbackJSON("not-json")
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}
