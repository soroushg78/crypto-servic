package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"wallet-observer/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool initializes a PostgreSQL connection pool with connection settings.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection with a short timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	slog.Info("PostgreSQL connection pool established successfully")
	return pool, nil
}

// AutoMigrate ensures standard tables are initialized.
func AutoMigrate(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS payments (
		id VARCHAR(64) PRIMARY KEY,
		order_id VARCHAR(64) UNIQUE NOT NULL,
		customer_ref VARCHAR(128),
		amount NUMERIC(20, 8) NOT NULL,
		currency VARCHAR(16) NOT NULL,
		network VARCHAR(32) NOT NULL,
		pay_address TEXT NOT NULL,
		pay_url TEXT,
		status VARCHAR(32) NOT NULL,
		tx_hash TEXT,
		provider_name VARCHAR(64) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		expires_at TIMESTAMPTZ NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
	CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);
	`
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	slog.Info("Database schema migrations verified")
	return nil
}
