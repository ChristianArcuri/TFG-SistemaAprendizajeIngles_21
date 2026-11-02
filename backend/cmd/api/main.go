// Command api es el entrypoint del backend de Inglés Profesional IT.
// Solo orquesta: carga config, abre la DB, arma servicios, router y arranca el server.
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/remotetalk/backend/internal/ai"
	"github.com/remotetalk/backend/internal/auth"
	"github.com/remotetalk/backend/internal/config"
	"github.com/remotetalk/backend/internal/database"
	"github.com/remotetalk/backend/internal/practice"
	"github.com/remotetalk/backend/internal/server"
	"github.com/remotetalk/backend/internal/users"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config_load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.New(dbCtx, cfg.DB)
	if err != nil {
		slog.Error("database_connect_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database_connected")

	tokenIssuer := auth.NewTokenIssuer(cfg.JWT.Secret, cfg.JWT.ExpiresIn)

	usersRepo := users.NewPgRepository(pool)
	usersService := users.NewService(usersRepo, tokenIssuer)
	usersHandler := users.NewHandler(usersService, server.CurrentUserID)

	feedbackSvc, transcriptionSvc := ai.NewServices(cfg)

	practiceRepo := practice.NewPgRepository(pool)
	practiceService := practice.NewService(practiceRepo, feedbackSvc, transcriptionSvc)
	practiceHandler := practice.NewHandler(practiceService, server.CurrentUserID, cfg.AI.MaxAudioBytes)

	deps := server.Dependencies{
		DB:              pool,
		TokenIssuer:     tokenIssuer,
		UsersHandler:    usersHandler,
		PracticeHandler: practiceHandler,
	}

	router := server.NewRouter(cfg, deps)
	srv := server.New(cfg, router)

	if cfg.AI.Provider == "openai" {
		minWrite := cfg.OpenAI.TranscriptionTimeout + cfg.OpenAI.FeedbackTimeout + 5*time.Second
		if cfg.HTTP.WriteTimeout < minWrite {
			slog.Warn("http_write_timeout_may_be_too_low_for_openai",
				slog.Duration("current", cfg.HTTP.WriteTimeout),
				slog.Duration("recommended_min", minWrite),
			)
		}
	}

	if err := srv.Run(); err != nil {
		slog.Error("server_run_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
