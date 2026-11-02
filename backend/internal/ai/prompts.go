package ai

import (
	"bytes"
	"fmt"
	"text/template"
)

// FeedbackPromptTemplate es el prompt que va a usar la implementacion real
// de FeedbackService cuando integremos OpenAI (Fase 6). Vive aca para
// poder iterarlo sin tocar handlers; la version canonica con notas
// adicionales esta en docs/prompts/feedback.md.
const FeedbackPromptTemplate = `You are an English coach for LATAM software developers.

The user is practicing English for remote IT jobs.
Their native language is Spanish.
They need practical feedback, not academic explanations.

Analyze the following answer:

"{{.Transcript}}"

Practice context:
"{{.PracticeContext}}"

Return a JSON object in Spanish with this structure:

{
  "correctedText": "...",
  "naturalVersion": "...",
  "explanationEs": "...",
  "strengths": ["..."],
  "improvements": ["..."],
  "comprehensionScore": 0,
  "vocabularyScore": 0,
  "communicationScore": 0,
  "suggestedPhrases": [
    { "spanish": "...", "english": "..." }
  ]
}

Rules:
- The explanation must be in Spanish.
- Corrected and natural versions must be in English.
- Score only three competencies from 0 to 100: comprehension of the prompt, professional IT vocabulary, and professional communication.
- Put grammar remarks inside explanationEs. Do not score grammar as its own competency.
- Do not score pronunciation or intonation from text. You may mention them only as a non-scored observation inside explanationEs.
- Focus on HR interviews, daily meetings, and technical IT vocabulary.
- Be encouraging but direct.
- Do not return markdown.
- Return valid JSON only.`

// RenderFeedbackPrompt renderiza el template con los campos de FeedbackInput.
// Lo expone el paquete para que el cliente real (OpenAI) lo arme con la misma
// version del prompt; el mock no lo necesita.
func RenderFeedbackPrompt(in FeedbackInput) (string, error) {
	tmpl, err := template.New("feedback").Parse(FeedbackPromptTemplate)
	if err != nil {
		return "", fmt.Errorf("parse prompt template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, in); err != nil {
		return "", fmt.Errorf("render prompt template: %w", err)
	}
	return buf.String(), nil
}
