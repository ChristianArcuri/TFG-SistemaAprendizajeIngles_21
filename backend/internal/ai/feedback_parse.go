package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// openAIFeedbackJSON es el shape que pedimos a OpenAI en el prompt. Puede
// incluir "transcript" aunque nosotros ya lo tenemos; lo ignoramos al mapear.
type openAIFeedbackJSON struct {
	Transcript         string            `json:"transcript"`
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

// parseFeedbackJSON valida y normaliza la respuesta JSON del modelo.
func parseFeedbackJSON(raw string) (FeedbackResult, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return FeedbackResult{}, fmt.Errorf("empty feedback json")
	}

	var parsed openAIFeedbackJSON
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return FeedbackResult{}, fmt.Errorf("unmarshal feedback json: %w", err)
	}

	if strings.TrimSpace(parsed.CorrectedText) == "" {
		return FeedbackResult{}, fmt.Errorf("missing correctedText")
	}
	if strings.TrimSpace(parsed.ExplanationES) == "" {
		return FeedbackResult{}, fmt.Errorf("missing explanationEs")
	}

	return FeedbackResult{
		CorrectedText:      strings.TrimSpace(parsed.CorrectedText),
		NaturalVersion:     strings.TrimSpace(parsed.NaturalVersion),
		ExplanationES:      strings.TrimSpace(parsed.ExplanationES),
		Strengths:          nonEmptyStrings(parsed.Strengths),
		Improvements:       nonEmptyStrings(parsed.Improvements),
		ComprehensionScore: clampScore(parsed.ComprehensionScore),
		VocabularyScore:    clampScore(parsed.VocabularyScore),
		CommunicationScore: clampScore(parsed.CommunicationScore),
		SuggestedPhrases:   parsed.SuggestedPhrases,
	}, nil
}

func nonEmptyStrings(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func clampScore(n int) int {
	if n < 0 {
		return 0
	}
	if n > 100 {
		return 100
	}
	return n
}
