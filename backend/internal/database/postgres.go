// Package database expone la conexión a PostgreSQL del backend.
// Encapsula pgxpool para que el resto de la app reciba un *pgxpool.Pool
// listo para usar y no tenga que conocer los detalles de configuración.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/remotetalk/backend/internal/config"
)

// New crea un pool de conexiones a Postgres y verifica que responda
// con un Ping. Si falla, devuelve error y el caller decide cómo seguir.
func New(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		poolCfg.MaxConns = cfg.MaxOpenConns
	}
	if cfg.MaxIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxIdleTime
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout(cfg.ConnectTimeout))
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func pingTimeout(d time.Duration) time.Duration {
	if d <= 0 {
		return 5 * time.Second
	}
	return d
}
