// Package db manages the PostgreSQL connection pool.
package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// New creates a PostgreSQL connection pool and verifies connectivity.
// Returns an error if the pool cannot be created or the database
// does not respond to a ping.
func New(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
