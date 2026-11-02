package users

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/remotetalk/backend/internal/auth"
	"github.com/remotetalk/backend/internal/httpx"
)

// userIDFromContext es la firma del extractor que el middleware de auth
// expone. Lo definimos aca para evitar ciclos con internal/server.
type userIDFromContext func(c *gin.Context) (uuid.UUID, bool)

// Handler agrupa los HTTP handlers de auth y perfil.
type Handler struct {
	service     *Service
	currentUser userIDFromContext
}

// NewHandler construye el handler.
func NewHandler(service *Service, currentUser userIDFromContext) *Handler {
	return &Handler{service: service, currentUser: currentUser}
}

// Register implementa POST /api/v1/auth/register.
func (h *Handler) Register(c *gin.Context) {
	var in RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "Datos de registro invalidos.")
		return
	}

	resp, err := h.service.Register(c.Request.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			httpx.Error(c, http.StatusConflict, "conflict", "Ese email ya esta registrado.")
		case errors.Is(err, auth.ErrPasswordTooLong):
			httpx.Error(c, http.StatusBadRequest, "validation_error", "El password supera el largo maximo permitido.")
		default:
			h.internal(c, "register_failed", err)
		}
		return
	}
	httpx.JSON(c, http.StatusCreated, resp)
}

// Login implementa POST /api/v1/auth/login.
func (h *Handler) Login(c *gin.Context) {
	var in LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "Datos de login invalidos.")
		return
	}

	resp, err := h.service.Login(c.Request.Context(), in)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			httpx.Error(c, http.StatusUnauthorized, "invalid_credentials", "Email o password incorrectos.")
			return
		}
		h.internal(c, "login_failed", err)
		return
	}
	httpx.OK(c, resp)
}

// Me implementa GET /api/v1/me.
func (h *Handler) Me(c *gin.Context) {
	userID, ok := h.currentUser(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
		return
	}

	resp, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
			return
		}
		h.internal(c, "get_me_failed", err)
		return
	}
	httpx.OK(c, resp)
}

// UpdateProfile implementa PUT /api/v1/me/profile.
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, ok := h.currentUser(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
		return
	}

	var in UpdateProfileInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.Error(c, http.StatusBadRequest, "validation_error", "Datos de perfil invalidos.")
		return
	}

	profile, err := h.service.UpdateProfile(c.Request.Context(), userID, in)
	if err != nil {
		h.internal(c, "update_profile_failed", err)
		return
	}
	httpx.OK(c, profile)
}

// internal centraliza el manejo de errores no esperados:
// loguea el error real y responde un mensaje seguro al cliente.
func (h *Handler) internal(c *gin.Context, op string, err error) {
	slog.ErrorContext(c.Request.Context(), op, slog.String("error", err.Error()))
	httpx.Error(c, http.StatusInternalServerError, "internal_error", "Algo salio mal. Intentalo de nuevo.")
}
