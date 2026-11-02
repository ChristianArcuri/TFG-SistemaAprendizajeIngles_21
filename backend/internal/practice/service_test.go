package practice

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/remotetalk/backend/internal/ai"
)

// fakeRepo implementa Repository en memoria para testear el service
// sin Postgres. Cada test crea su propia instancia con datos sembrados.
type fakeRepo struct {
	categories     []Category
	prompts        []Prompt
	sessions       map[uuid.UUID]Session
	feedback       map[uuid.UUID]Feedback
	saveFeedbackFn func(SaveFeedbackParams) (Feedback, error)
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		sessions: map[uuid.UUID]Session{},
		feedback: map[uuid.UUID]Feedback{},
	}
}

func (r *fakeRepo) seedCategory(code, name string) Category {
	c := Category{
		ID:          uuid.New(),
		Code:        code,
		Name:        name,
		Description: "desc",
		CreatedAt:   time.Now().UTC(),
	}
	r.categories = append(r.categories, c)
	return c
}

func (r *fakeRepo) seedPrompt(catID uuid.UUID, title, difficulty string) Prompt {
	p := Prompt{
		ID:         uuid.New(),
		CategoryID: catID,
		Title:      title,
		PromptText: "text",
		Difficulty: difficulty,
		CreatedAt:  time.Now().UTC(),
	}
	r.prompts = append(r.prompts, p)
	return p
}

func (r *fakeRepo) ListCategories(_ context.Context) ([]Category, error) {
	out := make([]Category, len(r.categories))
	copy(out, r.categories)
	return out, nil
}

func (r *fakeRepo) GetCategoryByCode(_ context.Context, code string) (Category, error) {
	for _, c := range r.categories {
		if c.Code == code {
			return c, nil
		}
	}
	return Category{}, ErrCategoryNotFound
}

func (r *fakeRepo) ListPromptsByCategory(_ context.Context, catID uuid.UUID) ([]Prompt, error) {
	out := []Prompt{}
	for _, p := range r.prompts {
		if p.CategoryID == catID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *fakeRepo) GetPromptByID(_ context.Context, id uuid.UUID) (Prompt, error) {
	for _, p := range r.prompts {
		if p.ID == id {
			return p, nil
		}
	}
	return Prompt{}, ErrPromptNotFound
}

func (r *fakeRepo) CreateSession(_ context.Context, params CreateSessionParams) (Session, error) {
	var code string
	for _, c := range r.categories {
		if c.ID == params.CategoryID {
			code = c.Code
			break
		}
	}
	s := Session{
		ID:             uuid.New(),
		UserID:         params.UserID,
		CategoryID:     params.CategoryID,
		CategoryCode:   code,
		PromptID:       params.PromptID,
		UserTranscript: params.UserTranscript,
		Status:         params.Status,
		CreatedAt:      time.Now().UTC(),
	}
	r.sessions[s.ID] = s
	return s, nil
}

func (r *fakeRepo) GetSessionByID(_ context.Context, sessionID, userID uuid.UUID) (Session, error) {
	s, ok := r.sessions[sessionID]
	if !ok || s.UserID != userID {
		return Session{}, ErrSessionNotFound
	}
	return s, nil
}

func (r *fakeRepo) ListSessionsByUser(_ context.Context, userID uuid.UUID, limit, offset int) ([]Session, error) {
	all := []Session{}
	for _, s := range r.sessions {
		if s.UserID == userID {
			all = append(all, s)
		}
	}
	if offset >= len(all) {
		return []Session{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *fakeRepo) GetFeedbackBySessionID(_ context.Context, sessionID uuid.UUID) (Feedback, error) {
	f, ok := r.feedback[sessionID]
	if !ok {
		return Feedback{}, ErrFeedbackNotFound
	}
	return f, nil
}

func (r *fakeRepo) SaveFeedback(_ context.Context, params SaveFeedbackParams) (Feedback, error) {
	if r.saveFeedbackFn != nil {
		return r.saveFeedbackFn(params)
	}
	f := Feedback{
		ID:                 uuid.New(),
		SessionID:          params.SessionID,
		CorrectedText:      params.CorrectedText,
		NaturalVersion:     params.NaturalVersion,
		ExplanationES:      params.ExplanationES,
		ComprehensionScore: params.ComprehensionScore,
		VocabularyScore:    params.VocabularyScore,
		CommunicationScore: params.CommunicationScore,
		CreatedAt:          time.Now().UTC(),
	}
	r.feedback[params.SessionID] = f
	return f, nil
}

func TestServiceListCategories(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("hr_interview", "HR Interview")
	repo.seedCategory("daily_meeting", "Daily Meeting")

	svc := NewService(repo, nil, nil)
	cats, err := svc.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
}

func TestServiceListPrompts(t *testing.T) {
	repo := newFakeRepo()
	hr := repo.seedCategory("hr_interview", "HR Interview")
	daily := repo.seedCategory("daily_meeting", "Daily Meeting")
	repo.seedPrompt(hr.ID, "Tell me about yourself", "easy")
	repo.seedPrompt(hr.ID, "Why this company?", "medium")
	repo.seedPrompt(daily.ID, "What did you do yesterday?", "easy")

	svc := NewService(repo, nil, nil)

	prompts, err := svc.ListPrompts(context.Background(), "hr_interview")
	if err != nil {
		t.Fatalf("list prompts: %v", err)
	}
	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompts))
	}

	if _, err := svc.ListPrompts(context.Background(), "missing"); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestServiceCreateSessionWithoutPrompt(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("hr_interview", "HR Interview")

	svc := NewService(repo, nil, nil)
	userID := uuid.New()

	s, err := svc.CreateSession(context.Background(), userID, CreateSessionInput{
		CategoryCode:   "hr_interview",
		UserTranscript: "  Hello world  ",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if s.UserID != userID {
		t.Fatal("user id mismatch")
	}
	if s.CategoryCode != "hr_interview" {
		t.Fatalf("expected category code hr_interview, got %s", s.CategoryCode)
	}
	if s.UserTranscript != "Hello world" {
		t.Fatalf("transcript not trimmed: %q", s.UserTranscript)
	}
	if s.Status != SessionStatusCompleted {
		t.Fatalf("expected status completed, got %s", s.Status)
	}
	if s.PromptID != nil {
		t.Fatalf("expected nil prompt id, got %v", s.PromptID)
	}
}

func TestServiceCreateSessionWithPrompt(t *testing.T) {
	repo := newFakeRepo()
	hr := repo.seedCategory("hr_interview", "HR Interview")
	tech := repo.seedCategory("tech_interview", "Tech Interview")
	hrPrompt := repo.seedPrompt(hr.ID, "Tell me about yourself", "easy")
	techPrompt := repo.seedPrompt(tech.ID, "Explain a bug", "medium")

	svc := NewService(repo, nil, nil)
	userID := uuid.New()

	s, err := svc.CreateSession(context.Background(), userID, CreateSessionInput{
		CategoryCode: "hr_interview",
		PromptID:     &hrPrompt.ID,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if s.PromptID == nil || *s.PromptID != hrPrompt.ID {
		t.Fatalf("expected prompt id %s, got %v", hrPrompt.ID, s.PromptID)
	}

	_, err = svc.CreateSession(context.Background(), userID, CreateSessionInput{
		CategoryCode: "hr_interview",
		PromptID:     &techPrompt.ID,
	})
	if !errors.Is(err, ErrPromptCategoryMismatch) {
		t.Fatalf("expected ErrPromptCategoryMismatch, got %v", err)
	}

	missing := uuid.New()
	_, err = svc.CreateSession(context.Background(), userID, CreateSessionInput{
		CategoryCode: "hr_interview",
		PromptID:     &missing,
	})
	if !errors.Is(err, ErrPromptNotFound) {
		t.Fatalf("expected ErrPromptNotFound, got %v", err)
	}
}

func TestServiceCreateSessionUnknownCategory(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, nil, nil)

	_, err := svc.CreateSession(context.Background(), uuid.New(), CreateSessionInput{CategoryCode: "missing"})
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestServiceGetSessionWithAndWithoutFeedback(t *testing.T) {
	repo := newFakeRepo()
	hr := repo.seedCategory("hr_interview", "HR Interview")
	svc := NewService(repo, nil, nil)
	userID := uuid.New()

	s, err := svc.CreateSession(context.Background(), userID, CreateSessionInput{CategoryCode: "hr_interview"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	detail, err := svc.GetSession(context.Background(), userID, s.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if detail.Feedback != nil {
		t.Fatal("expected no feedback yet")
	}

	repo.feedback[s.ID] = Feedback{
		ID:                 uuid.New(),
		SessionID:          s.ID,
		ExplanationES:      "Bien hecho",
		ComprehensionScore: 80,
		CreatedAt:          time.Now().UTC(),
	}
	_ = hr

	detail, err = svc.GetSession(context.Background(), userID, s.ID)
	if err != nil {
		t.Fatalf("get session 2: %v", err)
	}
	if detail.Feedback == nil {
		t.Fatal("expected feedback")
	}
	if detail.Feedback.ComprehensionScore != 80 {
		t.Fatalf("comprehension score mismatch")
	}
}

func TestServiceGetSessionForeign(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("hr_interview", "HR Interview")
	svc := NewService(repo, nil, nil)

	owner := uuid.New()
	intruder := uuid.New()

	s, err := svc.CreateSession(context.Background(), owner, CreateSessionInput{CategoryCode: "hr_interview"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := svc.GetSession(context.Background(), intruder, s.ID); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound for foreign user, got %v", err)
	}
}

func TestServiceListSessionsPagination(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("hr_interview", "HR Interview")
	svc := NewService(repo, nil, nil)
	userID := uuid.New()

	for i := 0; i < 5; i++ {
		if _, err := svc.CreateSession(context.Background(), userID, CreateSessionInput{CategoryCode: "hr_interview"}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	page, err := svc.ListSessions(context.Background(), userID, ListSessionsParams{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if page.Limit != defaultListLimit {
		t.Fatalf("expected default limit %d, got %d", defaultListLimit, page.Limit)
	}

	page, err = svc.ListSessions(context.Background(), userID, ListSessionsParams{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(page.Items))
	}

	page, err = svc.ListSessions(context.Background(), userID, ListSessionsParams{Limit: 1000, Offset: -5})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if page.Limit != maxListLimit {
		t.Fatalf("expected limit clamp to %d, got %d", maxListLimit, page.Limit)
	}
	if page.Offset != 0 {
		t.Fatalf("expected offset clamp to 0, got %d", page.Offset)
	}
}

// fakeFeedback y fakeTranscription son implementaciones controladas
// (no aleatorias) para testear AnalyzeAudio sin depender de mocks reales.

type fakeFeedback struct {
	out FeedbackCapture
	err error
}

// FeedbackCapture es un alias del FeedbackResult del paquete ai, reusado
// para evitar confusiones con la entidad Feedback persistida.
type FeedbackCapture = ai.FeedbackResult

func (f *fakeFeedback) Generate(_ context.Context, _ ai.FeedbackInput) (ai.FeedbackResult, error) {
	return ai.FeedbackResult(f.out), f.err
}

type fakeTranscription struct {
	text string
	err  error
}

func (f *fakeTranscription) Transcribe(_ context.Context, _ ai.TranscriptionInput) (ai.TranscriptionResult, error) {
	return ai.TranscriptionResult{Text: f.text}, f.err
}

func TestServiceAnalyzeAudioHappyPath(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("hr_interview", "HR Interview")

	fb := &fakeFeedback{
		out: ai.FeedbackResult{
			CorrectedText:      "I worked on a database related bug yesterday.",
			NaturalVersion:     "Yesterday I worked on a database-related bug and fixed it.",
			ExplanationES:      "Tu respuesta es clara, pero podes sonar mas natural.",
			Strengths:          []string{"Buena estructura."},
			Improvements:       []string{"Trabaja la pronunciacion."},
			ComprehensionScore: 150, // fuera de rango: debe ser clamped a 100
			VocabularyScore:    80,
			CommunicationScore: -10, // fuera de rango: debe ser clamped a 0
			SuggestedPhrases: []ai.SuggestedPhrase{
				{Spanish: "Trabajé en un bug.", English: "I worked on a bug."},
			},
		},
	}
	tr := &fakeTranscription{text: "I worked on a database bug yesterday."}

	svc := NewService(repo, fb, tr)

	out, err := svc.AnalyzeAudio(context.Background(), uuid.New(), AnalyzeAudioInput{
		CategoryCode: "hr_interview",
		PromptText:   "Tell me about a bug you fixed recently",
		Audio:        AudioBlob{Reader: strings.NewReader("fake-bytes"), Filename: "a.webm", Size: 10},
	})
	if err != nil {
		t.Fatalf("analyze audio: %v", err)
	}

	if out.SessionID == uuid.Nil {
		t.Fatal("expected session id")
	}
	if out.CategoryCode != "hr_interview" {
		t.Fatalf("category code mismatch: %s", out.CategoryCode)
	}
	if out.Transcript != "I worked on a database bug yesterday." {
		t.Fatalf("transcript mismatch: %q", out.Transcript)
	}
	if out.Scores.Comprehension == nil || *out.Scores.Comprehension != 100 {
		t.Fatalf("expected comprehension clamped to 100, got %v", out.Scores.Comprehension)
	}
	if out.Scores.Communication == nil || *out.Scores.Communication != 0 {
		t.Fatalf("expected communication clamped to 0, got %v", out.Scores.Communication)
	}
	if len(out.Strengths) != 1 || len(out.SuggestedPhrases) != 1 {
		t.Fatal("expected coaching fields populated")
	}

	if _, ok := repo.sessions[out.SessionID]; !ok {
		t.Fatal("session not persisted")
	}
	if _, ok := repo.feedback[out.SessionID]; !ok {
		t.Fatal("feedback not persisted")
	}
}

func TestServiceAnalyzeAudioUnknownCategory(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fakeFeedback{}, &fakeTranscription{text: "x"})

	_, err := svc.AnalyzeAudio(context.Background(), uuid.New(), AnalyzeAudioInput{
		CategoryCode: "missing",
		Audio:        AudioBlob{Reader: strings.NewReader("x"), Size: 1},
	})
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestServiceAnalyzeAudioMissingDeps(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("hr_interview", "HR Interview")
	svc := NewService(repo, nil, nil)

	_, err := svc.AnalyzeAudio(context.Background(), uuid.New(), AnalyzeAudioInput{
		CategoryCode: "hr_interview",
		Audio:        AudioBlob{Reader: strings.NewReader("x"), Size: 1},
	})
	if err == nil {
		t.Fatal("expected error when ai services are nil")
	}
}

func TestServiceAnalyzeTextHappyPath(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("daily_meeting", "Daily")

	fb := &fakeFeedback{
		out: ai.FeedbackResult{
			CorrectedText:      "Yesterday I finished the API fix.",
			NaturalVersion:     "Yesterday I finished the API fix and today I will review the tests.",
			ExplanationES:      "La respuesta cubre avance y proximo paso.",
			ComprehensionScore: 82,
			VocabularyScore:    85,
			CommunicationScore: 78,
		},
	}
	svc := NewService(repo, fb, nil)

	out, err := svc.AnalyzeText(context.Background(), uuid.New(), AnalyzeTextInput{
		CategoryCode: "daily_meeting",
		PromptText:   "What did you do yesterday?",
		UserText:     "Yesterday I finished the API fix.",
	})
	if err != nil {
		t.Fatalf("analyze text: %v", err)
	}
	if out.Transcript != "Yesterday I finished the API fix." {
		t.Fatalf("unexpected transcript: %q", out.Transcript)
	}
	if out.Scores.Communication == nil || *out.Scores.Communication != 78 {
		t.Fatalf("expected communication score, got %v", out.Scores.Communication)
	}
	if out.SessionID == uuid.Nil {
		t.Fatal("expected session id")
	}
}

func TestServiceAnalyzeTextEmpty(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("daily_meeting", "Daily")
	svc := NewService(repo, &fakeFeedback{}, nil)

	_, err := svc.AnalyzeText(context.Background(), uuid.New(), AnalyzeTextInput{
		CategoryCode: "daily_meeting",
		UserText:     "   ",
	})
	if !errors.Is(err, ErrEmptyUserText) {
		t.Fatalf("expected ErrEmptyUserText, got %v", err)
	}
}

func TestServiceAnalyzeQuizUsesAnswerKey(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("it_quiz", "Quiz")
	fb := &fakeFeedback{
		out: ai.FeedbackResult{
			CorrectedText:      "Reviewed quiz answers.",
			NaturalVersion:     "Reviewed quiz answers.",
			ExplanationES:      "El resultado sale de las opciones correctas.",
			ComprehensionScore: 10,
			VocabularyScore:    10,
			CommunicationScore: 99,
		},
	}
	svc := NewService(repo, fb, nil)

	out, err := svc.AnalyzeText(context.Background(), uuid.New(), AnalyzeTextInput{
		CategoryCode: "it_quiz",
		PromptText:   "IT quiz",
		UserText:     "3 of 10 correct",
		QuizCorrect:  3,
		QuizTotal:    10,
	})
	if err != nil {
		t.Fatalf("analyze quiz: %v", err)
	}
	if out.Scores.Comprehension == nil || *out.Scores.Comprehension != 30 {
		t.Fatalf("expected comprehension 30, got %v", out.Scores.Comprehension)
	}
	if out.Scores.Vocabulary == nil || *out.Scores.Vocabulary != 30 {
		t.Fatalf("expected vocabulary 30, got %v", out.Scores.Vocabulary)
	}
	if out.Scores.Communication != nil {
		t.Fatalf("quiz must not score communication, got %d", *out.Scores.Communication)
	}
}

func TestServiceSubmitQuizSavesWithoutAI(t *testing.T) {
	repo := newFakeRepo()
	repo.seedCategory("it_quiz", "Quiz")
	svc := NewService(repo, nil, nil)

	out, err := svc.SubmitQuiz(context.Background(), uuid.New(), SubmitQuizInput{
		Correct:      8,
		Total:        10,
		MissedTopics: []string{"Deploy y release"},
	})
	if err != nil {
		t.Fatalf("submit quiz: %v", err)
	}
	if out.Scores.Comprehension == nil || *out.Scores.Comprehension != 80 {
		t.Fatalf("comprehension %+v", out.Scores.Comprehension)
	}
	if out.Scores.Communication != nil {
		t.Fatal("quiz must not score communication")
	}
	if out.ExplanationES == "" || out.SessionID == uuid.Nil {
		t.Fatalf("expected saved explanation, got %+v", out)
	}
}
