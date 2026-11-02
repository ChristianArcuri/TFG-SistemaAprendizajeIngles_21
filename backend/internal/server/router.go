// Package server arma el router HTTP y el servidor con graceful shutdown.
// El router monta los handlers y middlewares globales; no contiene logica
// de negocio.
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/remotetalk/backend/internal/auth"
	"github.com/remotetalk/backend/internal/config"
	"github.com/remotetalk/backend/internal/handlers"
	"github.com/remotetalk/backend/internal/practice"
	"github.com/remotetalk/backend/internal/users"
)

// Dependencies agrupa lo que el router necesita para cablear todas las rutas.
// Se construye en cmd/api/main.go y se pasa entero.
type Dependencies struct {
	DB              *pgxpool.Pool
	TokenIssuer     *auth.TokenIssuer
	UsersHandler    *users.Handler
	PracticeHandler *practice.Handler
}

// NewRouter construye el *gin.Engine con middlewares y rutas registradas.
func NewRouter(cfg *config.Config, deps Dependencies) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(cfg.HTTP.CORSOrigins))
	r.Use(requestLogger())

	healthHandler := handlers.NewHealthHandler(deps.DB, cfg.App.Name, cfg.App.Env)
	r.GET("/health", healthHandler.Check)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)

		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", deps.UsersHandler.Register)
			authGroup.POST("/login", deps.UsersHandler.Login)
		}

		protected := v1.Group("")
		protected.Use(AuthMiddleware(deps.TokenIssuer))
		{
			protected.GET("/me", deps.UsersHandler.Me)
			protected.PUT("/me/profile", deps.UsersHandler.UpdateProfile)

			practiceGroup := protected.Group("/practice")
			{
				practiceGroup.GET("/categories", deps.PracticeHandler.ListCategories)
				practiceGroup.GET("/prompts", deps.PracticeHandler.ListPrompts)
				practiceGroup.POST("/sessions", deps.PracticeHandler.CreateSession)
				practiceGroup.GET("/sessions", deps.PracticeHandler.ListSessions)
				practiceGroup.GET("/recommendation", deps.PracticeHandler.Recommend)
				practiceGroup.GET("/sessions/:id", deps.PracticeHandler.GetSession)
				practiceGroup.POST("/analyze-audio", deps.PracticeHandler.AnalyzeAudio)
				practiceGroup.POST("/analyze-text", deps.PracticeHandler.AnalyzeText)
				practiceGroup.POST("/quiz-results", deps.PracticeHandler.SubmitQuiz)
			}
		}
	}

	return r
}
