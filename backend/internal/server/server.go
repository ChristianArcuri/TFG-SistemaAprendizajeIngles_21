package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/remotetalk/backend/internal/config"
)

// Server envuelve un *http.Server para correrlo con graceful shutdown.
type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

// New construye el Server a partir de la config y un *gin.Engine ya armado.
func New(cfg *config.Config, handler *gin.Engine) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.HTTP.Port,
			Handler:      handler,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
		},
		shutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}
}

// Run arranca el servidor HTTP y bloquea hasta recibir SIGINT o SIGTERM.
// Cuando llega la señal, intenta un shutdown ordenado dentro del timeout
// configurado.
func (s *Server) Run() error {
	errCh := make(chan error, 1)

	go func() {
		slog.Info("http_server_listening", slog.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen: %w", err)
			return
		}
		errCh <- nil
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		slog.Info("shutdown_signal_received", slog.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	slog.Info("http_server_stopped")
	return nil
}
