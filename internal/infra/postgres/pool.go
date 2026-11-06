package postgres

import (
	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/20moonlight01/go_homeworks/internal/config"
)

func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	poolCfg.MaxConns = cfg.DBMaxConns
	poolCfg.MinConns = cfg.DBMinConns
	poolCfg.ConnConfig.ConnectTimeout = cfg.DBConnectTimeout
	poolCfg.MaxConnLifetime = cfg.DBMaxConnLifetime

	innerTracer := otelpgx.NewTracer()
	spanTracer := NewSpanTracer(innerTracer)
	metricsTracer := NewMetricsTracer()

	poolCfg.ConnConfig.Tracer = NewMultiTracer(spanTracer, metricsTracer)

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := otelpgx.RecordStats(pool); err != nil {
		return nil, fmt.Errorf("unable to record database stats: %w", err)
	}

	checkCtx, cancel := context.WithTimeout(ctx, cfg.DBConnectTimeout)
	defer cancel()

	if err := pool.Ping(checkCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
