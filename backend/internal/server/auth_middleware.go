package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/remotetalk/backend/internal/auth"
	"github.com/remotetalk/backend/internal/httpx"
)

// userIDContextKey es la clave bajo la cual el middleware guarda el UUID
// del usuario autenticado en el *gin.Context.
const userIDContextKey = "user_id"

// AuthMiddleware valida el JWT de la request:
//   - extrae el header Authorization
//   - parsea el Bearer token
//   - guarda el UUID del usuario en el context
//
// Importante: NO se loguea el header ni el token.
func AuthMiddleware(issuer *auth.TokenIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
			return
		}

		claims, err := issuer.Parse(token)
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, "unauthorized", "Token ausente o invalido.")
			return
		}

		c.Set(userIDContextKey, claims.UserID)
		c.Next()
	}
}

// CurrentUserID extrae el UUID del usuario autenticado del context.
// Devuelve (uuid.Nil, false) si no esta seteado o si tiene un tipo invalido.
func CurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(userIDContextKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

// bearerToken extrae el token del header Authorization. Acepta solo el
// esquema "Bearer". Devuelve ok=false si el header esta vacio o no respeta
// el formato.
func bearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}
