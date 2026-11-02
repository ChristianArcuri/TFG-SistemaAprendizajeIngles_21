// Package handlers contiene los HTTP handlers del backend.
// Cumplen la regla "handlers finos": parsean la request, llaman a un
// servicio o repositorio, y delegan la respuesta a httpx.
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/remotetalk/backend/internal/httpx"
)

// HealthHandler reporta el estado del servicio y de la base de datos.
type HealthHandler struct {
	db      *pgxpool.Pool
	appName string
	env     string
}

// NewHealthHandler construye el handler con sus dependencias.
func NewHealthHandler(db *pgxpool.Pool, appName, env string) *HealthHandler {
	return &HealthHandler{db: db, appName: appName, env: env}
}

type healthResponse struct {
	Service string `json:"service"`
	Env     string `json:"env"`
	Status  string `json:"status"`
	DB      string `json:"db"`
}

// Check responde 200 si la app está viva. Reporta el estado de la DB
// como "ok" o "down" pero no falla la respuesta para que el endpoint
// pueda usarse como liveness probe del contenedor.
func (h *HealthHandler) Check(c *gin.Context) {
	dbStatus := "ok"
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if h.db == nil {
		dbStatus = "down"
	} else if err := h.db.Ping(ctx); err != nil {
		dbStatus = "down"
	}

	status := http.StatusOK
	if dbStatus != "ok" {
		status = http.StatusServiceUnavailable
	}

	httpx.JSON(c, status, healthResponse{
		Service: h.appName,
		Env:     h.env,
		Status:  "ok",
		DB:      dbStatus,
	})
}
