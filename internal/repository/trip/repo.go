package trip

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/20moonlight01/go_homeworks/internal/infra/postgres/transaction"
)

type Repository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *Repository {
	return &Repository{
		pool:         pool,
		queryTimeout: queryTimeout,
	}
}

var psqlBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *Repository) exec(ctx context.Context) transaction.DBTX {
	return transaction.GetExecutor(ctx, r.pool)
}
