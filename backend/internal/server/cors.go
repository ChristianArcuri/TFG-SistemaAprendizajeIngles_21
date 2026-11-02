package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// corsMiddleware permite requests del frontend en desarrollo (Vite) y origenes
// configurados via CORS_ALLOWED_ORIGINS (coma-separados).
// Ejemplo: http://localhost:5173,http://127.0.0.1:5173
func corsMiddleware(allowedOrigins string) gin.HandlerFunc {
	origins := parseOrigins(allowedOrigins)
	allowAll := len(origins) == 0

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowAll {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		} else if originAllowed(origin, origins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func parseOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if o := strings.TrimSpace(p); o != "" {
			out = append(out, o)
		}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}
	for _, o := range allowed {
		if strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}
