package service

import (
	db "api-cultura-conecta/internal/db/generated"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func withTx(ctx context.Context, pool *pgxpool.Pool, fn func(q db.Querier) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(db.New(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
