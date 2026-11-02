package practice

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"

	"github.com/remotetalk/backend/internal/ai"
)

// Limites de paginacion del historial.
const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// maxScore es el techo (y minimo es 0) que aplicamos a los scores antes de
// persistirlos. Si la IA devuelve algo fuera de rango, lo recortamos.
const maxScore = 100

// Service implementa la logica de practica:
// catalogo, prompts, creacion y consulta de sesiones, feedback asociado,
// y orquestacion de transcripcion + feedback (analyze-audio).
type Service struct {
	repo          Repository
	feedback      ai.FeedbackService
	transcription ai.TranscriptionService
}

// NewService construye el service. feedback y transcription pueden ser
// nil si solo se usan los endpoints de catalogo / sesiones; AnalyzeAudio
// requiere las dos.
func NewService(repo Repository, feedback ai.FeedbackService, transcription ai.TranscriptionService) *Service {
	return &Service{repo: repo, feedback: feedback, transcription: transcription}
}

// ListCategories devuelve todas las categorias del catalogo.
func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	all, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Category, 0, len(all))
	for _, cat := range all {
		if isPrototypeCategory(cat.Code) {
			out = append(out, cat)
		}
	}
	return out, nil
}

func isPrototypeCategory(code string) bool {
	switch code {
	case "hr_interview", "daily_meeting", "it_quiz":
		return true
	default:
		return false
	}
}

// ListPrompts devuelve los prompts de una categoria identificada por su code.
// Si la categoria no existe, devuelve ErrCategoryNotFound.
func (s *Service) ListPrompts(ctx context.Context, categoryCode string) ([]Prompt, error) {
	code := strings.TrimSpace(categoryCode)
	if code == "" {
		return nil, ErrCategoryNotFound
	}
	cat, err := s.repo.GetCategoryByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.repo.ListPromptsByCategory(ctx, cat.ID)
}

// CreateSession crea una sesion de practica del usuario.
//
// Validaciones de negocio:
//   - categoryCode debe existir
//   - si hay promptID, debe existir y pertenecer a la categoria indicada
func (s *Service) CreateSession(ctx context.Context, userID uuid.UUID, in CreateSessionInput) (Session, error) {
	cat, err := s.repo.GetCategoryByCode(ctx, strings.TrimSpace(in.CategoryCode))
	if err != nil {
		return Session{}, err
	}

	if in.PromptID != nil {
		prompt, err := s.repo.GetPromptByID(ctx, *in.PromptID)
		if err != nil {
			return Session{}, err
		}
		if prompt.CategoryID != cat.ID {
			return Session{}, ErrPromptCategoryMismatch
		}
	}

	return s.repo.CreateSession(ctx, CreateSessionParams{
		UserID:         userID,
		CategoryID:     cat.ID,
		PromptID:       in.PromptID,
		UserTranscript: strings.TrimSpace(in.UserTranscript),
		Status:         SessionStatusCompleted,
	})
}

// GetSession devuelve una sesion del usuario y su feedback (si existe).
// Si la sesion no existe o no pertenece al usuario, devuelve ErrSessionNotFound.
func (s *Service) GetSession(ctx context.Context, userID, sessionID uuid.UUID) (SessionDetail, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID, userID)
	if err != nil {
		return SessionDetail{}, err
	}

	detail := SessionDetail{Session: session}

	feedback, err := s.repo.GetFeedbackBySessionID(ctx, session.ID)
	switch {
	case err == nil:
		detail.Feedback = &feedback
	case errors.Is(err, ErrFeedbackNotFound):
		// Sin feedback todavia: respuesta valida con Feedback == nil.
	default:
		return SessionDetail{}, err
	}

	return detail, nil
}

// ListSessions devuelve el historial paginado del usuario.
// Limit default 20, max 100. Offset >= 0.
func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID, params ListSessionsParams) (PaginatedSessions, error) {
	limit, offset := normalizePagination(params)

	items, err := s.repo.ListSessionsByUser(ctx, userID, limit, offset)
	if err != nil {
		return PaginatedSessions{}, err
	}
	return PaginatedSessions{Items: items, Limit: limit, Offset: offset}, nil
}

func normalizePagination(p ListSessionsParams) (int, int) {
	limit := p.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// AnalyzeAudio orquesta el flujo de /analyze-audio:
//
//  1. valida la categoria
//  2. transcribe el audio
//  3. genera feedback de IA
//  4. crea la sesion con el transcript
//  5. persiste el feedback (campos del schema)
//  6. devuelve el resultado al handler con coaching adicional
//
// Errores tipicos:
//   - ErrCategoryNotFound si el categoryCode no existe
//   - errores opacos del provider (transcripcion / feedback) que el
//     handler convierte en internal_error
const maxUserTextLength = 5000

func (s *Service) AnalyzeAudio(ctx context.Context, userID uuid.UUID, in AnalyzeAudioInput) (AnalyzeAudioResult, error) {
	if s.feedback == nil || s.transcription == nil {
		return AnalyzeAudioResult{}, errors.New("practice: ai services not configured")
	}

	cat, err := s.repo.GetCategoryByCode(ctx, strings.TrimSpace(in.CategoryCode))
	if err != nil {
		return AnalyzeAudioResult{}, err
	}

	transcription, err := s.transcription.Transcribe(ctx, ai.TranscriptionInput{
		Reader:   in.Audio.Reader,
		Filename: in.Audio.Filename,
		MimeType: in.Audio.MimeType,
		Size:     in.Audio.Size,
	})
	if err != nil {
		return AnalyzeAudioResult{}, fmt.Errorf("transcription: %w", err)
	}

	return s.analyzeTranscript(ctx, userID, cat, in.PromptText, transcription.Text, 0, 0)
}

// AnalyzeText orquesta el flujo de /analyze-text (writing, daily escrito, etc.).
func (s *Service) AnalyzeText(ctx context.Context, userID uuid.UUID, in AnalyzeTextInput) (AnalyzeAudioResult, error) {
	if s.feedback == nil {
		return AnalyzeAudioResult{}, errors.New("practice: feedback service not configured")
	}

	cat, err := s.repo.GetCategoryByCode(ctx, strings.TrimSpace(in.CategoryCode))
	if err != nil {
		return AnalyzeAudioResult{}, err
	}

	transcript, err := normalizeUserText(in.UserText)
	if err != nil {
		return AnalyzeAudioResult{}, err
	}
	if err := validateQuizResult(cat.Code, in.QuizCorrect, in.QuizTotal); err != nil {
		return AnalyzeAudioResult{}, err
	}

	return s.analyzeTranscript(ctx, userID, cat, in.PromptText, transcript, in.QuizCorrect, in.QuizTotal)
}

// SubmitQuiz guarda el quiz terminado sin llamar a la IA.
// Comprensión y vocabulario usan el porcentaje de aciertos.
func (s *Service) SubmitQuiz(ctx context.Context, userID uuid.UUID, in SubmitQuizInput) (AnalyzeAudioResult, error) {
	if err := validateQuizResult("it_quiz", in.Correct, in.Total); err != nil || in.Total == 0 {
		if err == nil {
			err = ErrInvalidQuizResult
		}
		return AnalyzeAudioResult{}, err
	}

	cat, err := s.repo.GetCategoryByCode(ctx, "it_quiz")
	if err != nil {
		return AnalyzeAudioResult{}, err
	}

	pct, explanation, summary, review := DescribeQuizResult(in.Correct, in.Total, in.MissedTopics)
	transcript := fmt.Sprintf("Quiz de inglés IT: %d de %d.", in.Correct, in.Total)
	session, err := s.repo.CreateSession(ctx, CreateSessionParams{
		UserID:         userID,
		CategoryID:     cat.ID,
		UserTranscript: transcript,
		Status:         SessionStatusCompleted,
	})
	if err != nil {
		return AnalyzeAudioResult{}, fmt.Errorf("create session: %w", err)
	}

	saved, err := s.repo.SaveFeedback(ctx, SaveFeedbackParams{
		SessionID:          session.ID,
		CorrectedText:      summary,
		NaturalVersion:     review,
		ExplanationES:      explanation,
		ComprehensionScore: pct,
		VocabularyScore:    pct,
	})
	if err != nil {
		return AnalyzeAudioResult{}, fmt.Errorf("save feedback: %w", err)
	}

	return AnalyzeAudioResult{
		SessionID:      session.ID,
		CategoryCode:   session.CategoryCode,
		Transcript:     transcript,
		CorrectedText:  saved.CorrectedText,
		NaturalVersion: saved.NaturalVersion,
		ExplanationES:  saved.ExplanationES,
		Scores: Scores{
			Comprehension: intPtr(saved.ComprehensionScore),
			Vocabulary:    intPtr(saved.VocabularyScore),
		},
		CreatedAt: saved.CreatedAt,
	}, nil
}

func normalizeUserText(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", ErrEmptyUserText
	}
	if len(text) > maxUserTextLength {
		return "", ErrUserTextTooLong
	}
	return text, nil
}

func (s *Service) analyzeTranscript(ctx context.Context, userID uuid.UUID, cat Category, promptText, transcript string, quizCorrect, quizTotal int) (AnalyzeAudioResult, error) {
	feedback, err := s.feedback.Generate(ctx, ai.FeedbackInput{
		Transcript:      transcript,
		PracticeContext: buildPracticeContext(cat, promptText),
	})
	if err != nil {
		return AnalyzeAudioResult{}, fmt.Errorf("feedback: %w", err)
	}

	session, err := s.repo.CreateSession(ctx, CreateSessionParams{
		UserID:         userID,
		CategoryID:     cat.ID,
		UserTranscript: transcript,
		Status:         SessionStatusCompleted,
	})
	if err != nil {
		return AnalyzeAudioResult{}, fmt.Errorf("create session: %w", err)
	}

	comprehension, vocabulary, communication := scoresForCategory(cat.Code, feedback, quizCorrect, quizTotal)
	saved, err := s.repo.SaveFeedback(ctx, SaveFeedbackParams{
		SessionID:          session.ID,
		CorrectedText:      feedback.CorrectedText,
		NaturalVersion:     feedback.NaturalVersion,
		ExplanationES:      feedback.ExplanationES,
		ComprehensionScore: comprehension,
		VocabularyScore:    vocabulary,
		CommunicationScore: communication,
	})
	if err != nil {
		return AnalyzeAudioResult{}, fmt.Errorf("save feedback: %w", err)
	}

	return AnalyzeAudioResult{
		SessionID:      session.ID,
		CategoryCode:   session.CategoryCode,
		Transcript:     transcript,
		CorrectedText:  saved.CorrectedText,
		NaturalVersion: saved.NaturalVersion,
		ExplanationES:  saved.ExplanationES,
		Scores: Scores{
			Comprehension: intPtr(saved.ComprehensionScore),
			Vocabulary:    intPtr(saved.VocabularyScore),
			Communication: saved.CommunicationScore,
		},
		Strengths:        feedback.Strengths,
		Improvements:     feedback.Improvements,
		SuggestedPhrases: feedback.SuggestedPhrases,
		CreatedAt:        saved.CreatedAt,
	}, nil
}

// buildPracticeContext arma el "practice context" que recibe la IA:
// categoria + prompt original. Lo dejamos en una funcion separada para
// poder enriquecerlo en el futuro (perfil del usuario, nivel de ingles,
// etc.) sin tocar AnalyzeAudio.
func buildPracticeContext(cat Category, promptText string) string {
	prompt := strings.TrimSpace(promptText)
	if prompt == "" {
		return fmt.Sprintf("category=%s (%s)", cat.Code, cat.Name)
	}
	return fmt.Sprintf("category=%s (%s); prompt=%s", cat.Code, cat.Name, prompt)
}

func clampScore(n int) int {
	if n < 0 {
		return 0
	}
	if n > maxScore {
		return maxScore
	}
	return n
}

func validateQuizResult(categoryCode string, correct, total int) error {
	if categoryCode != "it_quiz" || total == 0 {
		return nil
	}
	if correct < 0 || total < 0 || correct > total || total > 100 {
		return ErrInvalidQuizResult
	}
	return nil
}

// scoresForCategory aplica las competencias que corresponde a cada modulo.
// El quiz usa el porcentaje de aciertos y no puntua comunicacion profesional.
func scoresForCategory(categoryCode string, feedback ai.FeedbackResult, quizCorrect, quizTotal int) (int, int, *int) {
	comprehension := clampScore(feedback.ComprehensionScore)
	vocabulary := clampScore(feedback.VocabularyScore)
	if categoryCode == "it_quiz" && quizTotal > 0 {
		pct := int(math.Round(float64(quizCorrect) * 100 / float64(quizTotal)))
		comprehension = clampScore(pct)
		vocabulary = clampScore(pct)
		return comprehension, vocabulary, nil
	}
	if categoryCode == "it_quiz" {
		return comprehension, vocabulary, nil
	}
	communication := clampScore(feedback.CommunicationScore)
	return comprehension, vocabulary, &communication
}
