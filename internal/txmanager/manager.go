package txmanager

import (
	"context"
	"fmt"

	"github.com/20moonlight01/go_homeworks/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Manager struct {
	pool *pgxpool.Pool
	opts pgx.TxOptions
}

func NewManager(pool *pgxpool.Pool) *Manager {
	return &Manager{
		pool: pool,
		opts: pgx.TxOptions{IsoLevel: pgx.ReadCommitted},
	}
}

func (m *Manager) Do(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, ok := db.GetTxFromContext(ctx)
	if ok {
		return fn(ctx)
	}

	tx, err = m.pool.BeginTx(ctx, m.opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	if err = fn(db.PutTxIntoContext(ctx, tx)); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
