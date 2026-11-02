package practice

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/remotetalk/backend/internal/ai"
	"github.com/remotetalk/backend/internal/httpx"
)

// userIDFromContext es la firma del extractor que el middleware de auth
// expone (server.CurrentUserID). Se inyecta para evitar ciclos con el paquete server.
type userIDFromContext func(c *gin.Context) (uuid.UUID, bool)

// Handler agrupa los HTTP handlers de practica.
type Handler struct {
	service       *Service
	currentUser   userIDFromContext
	maxAudioBytes int64
}

// NewHandler construye el handler.
// maxAudioBytes es el limite duro para POST /analyze-audio.
func NewHandler(service *Service, currentUser userIDFromContext, maxAudioBytes int64) *Handler {
	return &Handler{service: service, currentUser: currentUser, maxAudioBytes: maxAudioBytes}
}

// allowedAudioMimes lista los MIME types aceptados en /analyze-audio.
// Cubre lo que graban los navegadores modernos via MediaRecorder.
var allowedAudioMimes = map[string]bool{
	"audio/webm":  true,
	"audio/wav":   true,
	"audio/x-wav": true,
	"audio/wave":  true,
	"audio/mpeg":  true,
	"audio/mp3":   true,
	"audio/ogg":   true,
	"audio/mp4":   true,
	"audio/x-m4a": true,
	"audio/m4a":   true,
	"video/webm":  true, // algunos navegadores envian video/webm para grabaciones solo-audio
}

// normalizeAudioMIME quita parametros (p. ej. "audio/webm;codecs=opus" -> "audio/webm")
// e infiere el tipo desde la extension si el cliente no manda Content-Type.
func normalizeAudioMIME(contentType, filename string) string {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	if ct != "" && ct != "application/octet-stream" {
		return ct
	}

	ext := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(ext, ".webm"):
		return "audio/webm"
	case strings.HasSuffix(ext, ".wav"):
		return "audio/wav"
	case strings.HasSuffix(ext, ".mp3"), strings.HasSuffix(ext, ".mpeg"):
		return "audio/mpeg"
	case strings.HasSuffix(ext, ".ogg"):
		return "audio/ogg"
	case strings.HasSuffix(ext, ".m4a"), strings.HasSuffix(ext, ".mp4"):
		return "audio/mp4"
	default:
		return ct
	}
}

// maxPromptTextLength es el largo maximo del campo promptText recibido en el form.
const maxPromptTextLength = 2000

// ListCategories implementa GET /api/v1/practice/categories.
func (h *Handler) ListCategories(c *gin.Context) {
	cats, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		h.internal(c, "list_categories_failed", err)
		return
	}
	httpx.OK(c, cats)
}

// ListPrompts implementa GET /api/v1/practice/prompts?category=<code>.
func (h *Handler) ListPrompts(c *gin.Context) {
	code := c.Query("category")
	if code == "" {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "El parametro 'category' es obligatorio.")
		return
	}

	prompts, err := h.service.ListPrompts(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			httpx.Error(c, http.StatusNotFound, "not_found", "Categoria no encontrada.")
			return
		}
		h.internal(c, "list_prompts_failed", err)
		return
	}
	httpx.OK(c, prompts)
}

// CreateSession implementa POST /api/v1/practice/sessions.
func (h *Handler) CreateSession(c *gin.Context) {
	userID, ok := h.currentUser(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
		return
	}

	var in CreateSessionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "Datos de la sesion invalidos.")
		return
	}

	session, err := h.service.CreateSession(c.Request.Context(), userID, in)
	if err != nil {
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			httpx.Error(c, http.StatusNotFound, "not_found", "Categoria no encontrada.")
		case errors.Is(err, ErrPromptNotFound):
			httpx.Error(c, http.StatusNotFound, "not_found", "Prompt no encontrado.")
		case errors.Is(err, ErrPromptCategoryMismatch):
			httpx.Error(c, http.StatusBadRequest, "validation_error", "El prompt no pertenece a la categoria indicada.")
		default:
			h.internal(c, "create_session_failed", err)
		}
		return
	}
	httpx.JSON(c, http.StatusCreated, session)
}

// ListSessions implementa GET /api/v1/practice/sessions?limit=&offset=.
func (h *Handler) ListSessions(c *gin.Context) {
	userID, ok := h.currentUser(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
		return
	}

	limit, offset, err := parsePagination(c)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "Parametros de paginacion invalidos.")
		return
	}

	page, err := h.service.ListSessions(c.Request.Context(), userID, ListSessionsParams{Limit: limit, Offset: offset})
	if err != nil {
		h.internal(c, "list_sessions_failed", err)
		return
	}
	httpx.OK(c, page)
}

// Recommend implementa GET /api/v1/practice/recommendation.
func (h *Handler) Recommend(c *gin.Context) {
	userID, ok := h.currentUser(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
		return
	}

	rec, err := h.service.Recommend(c.Request.Context(), userID)
	if err != nil {
		h.internal(c, "recommendation_failed", err)
		return
	}
	httpx.OK(c, rec)
}

// AnalyzeAudio implementa POST /api/v1/practice/analyze-audio.
//
// Recibe multipart/form-data:
//   - audio        archivo (validado por MIME y tamano)
//   - categoryCode string  (obligatorio)
//   - promptText   string  (opcional, max 2000 chars)
//
// Orquesta validacion + service.AnalyzeAudio. El handler queda chico:
// no toca DB, IA ni transcription directo.
func (h *Handler) AnalyzeAudio(c *gin.Context) {
	userID, ok := h.currentUser(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxAudioBytes+1<<20)

	if err := c.Request.ParseMultipartForm(h.maxAudioBytes); err != nil {
		httpx.Error(c, http.StatusRequestEntityTooLarge, "payload_too_large", "El audio supera el limite permitido.")
		return
	}

	categoryCode := strings.TrimSpace(c.PostForm("categoryCode"))
	if categoryCode == "" {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "El campo 'categoryCode' es obligatorio.")
		return
	}

	promptText := strings.TrimSpace(c.PostForm("promptText"))
	if len(promptText) > maxPromptTextLength {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "El campo 'promptText' supera el largo maximo.")
		return
	}

	fileHeader, err := c.FormFile("audio")
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "El campo 'audio' es obligatorio.")
		return
	}
	if fileHeader.Size <= 0 {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "El archivo de audio esta vacio.")
		return
	}
	if fileHeader.Size > h.maxAudioBytes {
		httpx.Error(c, http.StatusRequestEntityTooLarge, "payload_too_large", "El audio supera el limite permitido.")
		return
	}

	mimeType := normalizeAudioMIME(fileHeader.Header.Get("Content-Type"), fileHeader.Filename)
	if !allowedAudioMimes[mimeType] {
		httpx.Error(c, http.StatusUnsupportedMediaType, "unsupported_media", "Formato de audio no permitido.")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.internal(c, "open_audio_failed", err)
		return
	}
	defer file.Close()

	result, err := h.service.AnalyzeAudio(c.Request.Context(), userID, AnalyzeAudioInput{
		CategoryCode: categoryCode,
		PromptText:   promptText,
		Audio: AudioBlob{
			Reader:   file,
			Filename: fileHeader.Filename,
			MimeType: mimeType,
			Size:     fileHeader.Size,
		},
	})
	if err != nil {
		if h.writeAIError(c, err) {
			return
		}
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			httpx.Error(c, http.StatusNotFound, "not_found", "Categoria no encontrada.")
		default:
			h.internal(c, "analyze_audio_failed", err)
		}
		return
	}
	httpx.JSON(c, http.StatusCreated, result)
}

type analyzeTextRequest struct {
	CategoryCode string `json:"categoryCode" binding:"required"`
	PromptText   string `json:"promptText"`
	UserText     string `json:"userText" binding:"required"`
	QuizCorrect  int    `json:"quizCorrect"`
	QuizTotal    int    `json:"quizTotal"`
}

// AnalyzeText implementa POST /api/v1/practice/analyze-text.
func (h *Handler) AnalyzeText(c *gin.Context) {
	userID, ok := h.currentUser(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
		return
	}

	var body analyzeTextRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "Payload invalido.")
		return
	}

	promptText := strings.TrimSpace(body.PromptText)
	if len(promptText) > maxPromptTextLength {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "El campo 'promptText' supera el largo maximo.")
		return
	}

	result, err := h.service.AnalyzeText(c.Request.Context(), userID, AnalyzeTextInput{
		CategoryCode: body.CategoryCode,
		PromptText:   promptText,
		UserText:     body.UserText,
		QuizCorrect:  body.QuizCorrect,
		QuizTotal:    body.QuizTotal,
	})
	if err != nil {
		if h.writeAIError(c, err) {
			return
		}
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			httpx.Error(c, http.StatusNotFound, "not_found", "Categoria no encontrada.")
		case errors.Is(err, ErrEmptyUserText):
			httpx.Error(c, http.StatusBadRequest, "validation_error", "El campo 'userText' es obligatorio.")
		case errors.Is(err, ErrUserTextTooLong):
			httpx.Error(c, http.StatusBadRequest, "validation_error", "El texto supera el largo maximo.")
		case errors.Is(err, ErrInvalidQuizResult):
			httpx.Error(c, http.StatusBadRequest, "validation_error", "El resultado del quiz no es valido.")
		default:
			h.internal(c, "analyze_text_failed", err)
		}
		return
	}
	httpx.JSON(c, http.StatusCreated, result)
}

type submitQuizRequest struct {
	Correct      int      `json:"correct"`
	Total        int      `json:"total" binding:"required,min=1,max=100"`
	MissedTopics []string `json:"missedTopics"`
}

// SubmitQuiz implementa POST /api/v1/practice/quiz-results.
// Guarda el puntaje del quiz sin llamar al servicio de IA.
func (h *Handler) SubmitQuiz(c *gin.Context) {
	userID, ok := h.currentUser(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
		return
	}

	var body submitQuizRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "Payload invalido.")
		return
	}

	result, err := h.service.SubmitQuiz(c.Request.Context(), userID, SubmitQuizInput{
		Correct:      body.Correct,
		Total:        body.Total,
		MissedTopics: body.MissedTopics,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			httpx.Error(c, http.StatusNotFound, "not_found", "Categoria no encontrada.")
		case errors.Is(err, ErrInvalidQuizResult):
			httpx.Error(c, http.StatusBadRequest, "validation_error", "El resultado del quiz no es valido.")
		default:
			h.internal(c, "submit_quiz_failed", err)
		}
		return
	}
	httpx.JSON(c, http.StatusCreated, result)
}

// GetSession implementa GET /api/v1/practice/sessions/:id.
func (h *Handler) GetSession(c *gin.Context) {
	userID, ok := h.currentUser(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
		return
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "ID de sesion invalido.")
		return
	}

	detail, err := h.service.GetSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			httpx.Error(c, http.StatusNotFound, "not_found", "Sesion no encontrada.")
			return
		}
		h.internal(c, "get_session_failed", err)
		return
	}
	httpx.OK(c, detail)
}

// parsePagination extrae limit y offset del querystring.
// Limites finales los aplica el service (defaultListLimit / maxListLimit).
func parsePagination(c *gin.Context) (int, int, error) {
	limit := 0
	offset := 0
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return 0, 0, errors.New("invalid limit")
		}
		limit = n
	}
	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return 0, 0, errors.New("invalid offset")
		}
		offset = n
	}
	return limit, offset, nil
}

func (h *Handler) internal(c *gin.Context, op string, err error) {
	slog.ErrorContext(c.Request.Context(), op, slog.String("error", err.Error()))
	httpx.Error(c, http.StatusInternalServerError, "internal_error", "Algo salio mal. Intentalo de nuevo.")
}

// writeAIError mapea fallas del proveedor de IA a un error claro para el cliente.
// Devuelve true si el error ya fue escrito en la respuesta.
func (h *Handler) writeAIError(c *gin.Context, err error) bool {
	if !errors.Is(err, ai.ErrProviderFailure) {
		return false
	}
	httpx.Error(
		c,
		http.StatusServiceUnavailable,
		"ai_unavailable",
		"El servicio de IA no esta disponible. Verifica tu cuenta OpenAI o usa AI_PROVIDER=mock para desarrollo.",
	)
	return true
}
