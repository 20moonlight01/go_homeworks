package container

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/20moonlight01/go_homeworks/internal/config"
	"github.com/20moonlight01/go_homeworks/internal/infra/postgres"
	"github.com/20moonlight01/go_homeworks/internal/infra/postgres/transaction"
	"github.com/20moonlight01/go_homeworks/internal/repository/trip"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/create"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/finish"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/get"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Config *config.Config
	Logger *slog.Logger
	Pool   *pgxpool.Pool
	Server *http.Server
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connection to db failed: %w", err)
	}

	txManager := transaction.NewManager(pool)

	tripRepo := trip.NewRepository(pool, cfg.DBQueryTimeout)

	createUC := create.NewUsecase(tripRepo, txManager)
	getUC := get.NewUsecase(tripRepo)
	finishUC := finish.NewUsecase(tripRepo, txManager)

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	return &Container{
		Config: cfg,
		Logger: logger,
		Pool:   pool,
		Server: srv,
	}, nil
}

func (c *Container) Close() {
	c.Pool.Close()
}
