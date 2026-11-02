package server

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// requestLogger emite un log estructurado por request usando slog.
// Importante: no loguea el header Authorization ni el body, para
// cumplir las reglas de seguridad de AGENTS.md.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		slog.Info("http_request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
