package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDB creates a new database connection pool.
func NewDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}