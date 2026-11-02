// Package httpx contiene helpers para que todos los handlers respondan
// JSON con la misma forma. Mantiene los handlers finos y evita filtrar
// errores internos al cliente.
package httpx

import (
	"github.com/gin-gonic/gin"
)

// SuccessResponse es la forma estándar de respuesta correcta.
type SuccessResponse struct {
	Data any `json:"data"`
}

// ErrorPayload describe un error pensado para el cliente.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse es la forma estándar de respuesta con error.
type ErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

// OK responde 200 con el body envuelto en {"data": ...}.
func OK(c *gin.Context, data any) {
	c.JSON(200, SuccessResponse{Data: data})
}

// JSON responde con el status indicado y el body envuelto en {"data": ...}.
func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, SuccessResponse{Data: data})
}

// Error responde con un código y mensaje pensados para el cliente.
// Nunca debe usarse para reenviar errores internos: el caller arma
// el mensaje seguro y el código (por ejemplo "validation_error").
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, ErrorResponse{
		Error: ErrorPayload{Code: code, Message: message},
	})
}
