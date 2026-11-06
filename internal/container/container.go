package container

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/20moonlight01/go_homeworks/internal/config"
	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/http/domainerror"
	"github.com/20moonlight01/go_homeworks/internal/http/middleware"
	httpService "github.com/20moonlight01/go_homeworks/internal/http/service"
	httpCreate "github.com/20moonlight01/go_homeworks/internal/http/trip/create"
	httpFinish "github.com/20moonlight01/go_homeworks/internal/http/trip/finish"
	httpGet "github.com/20moonlight01/go_homeworks/internal/http/trip/get"
	"github.com/20moonlight01/go_homeworks/internal/infra/postgres"
	"github.com/20moonlight01/go_homeworks/internal/infra/postgres/transaction"
	"github.com/20moonlight01/go_homeworks/internal/infra/telemetry"
	"github.com/20moonlight01/go_homeworks/internal/repository/trip"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/create"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/finish"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/get"
)

type Container struct {
	Config *config.Config
	Logger *slog.Logger
	Pool   *pgxpool.Pool
	Server *http.Server

	shutdownTelemetry func(context.Context) error
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	shutdownTelemetry, err := telemetry.Init(ctx, cfg.OTELServiceName)
	if err != nil {
		return nil, fmt.Errorf("init telemetry: %w", err)
	}

	metrics, err := telemetry.NewMetrics()
	if err != nil {
		return nil, fmt.Errorf("init metrics: %w", err)
	}

	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		lvl = slog.LevelInfo
	}
	logger := telemetry.NewLogger(lvl)

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		_ = shutdownTelemetry(ctx)
		return nil, fmt.Errorf("connection to db failed: %w", err)
	}

	txManager := transaction.NewManager(pool)

	tripRepo := trip.NewRepository(pool, cfg.DBQueryTimeout)

	createUC := create.NewUsecase(tripRepo, txManager, logger, metrics)
	getUC := get.NewUsecase(tripRepo)
	finishUC := finish.NewUsecase(tripRepo, txManager, logger, metrics)

	hCreate := httpCreate.NewHandler(createUC)
	hGet := httpGet.NewHandler(getUC)
	hFinish := httpFinish.NewHandler(finishUC)
	hService := httpService.NewHandler(pool)

	r := chi.NewRouter()
	r.Use(middleware.Recover(logger))
	r.Use(middleware.Tracing)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Metrics)

	apiHandler := &api.ServerInterfaceWrapper{
		Handler: &apiAdapter{
			create:  hCreate,
			get:     hGet,
			finish:  hFinish,
			service: hService,
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			domainerror.Write(w, r, &domain.ValidationError{
				Field:   "path",
				Message: "invalid parameter format",
			})
		},
	}

	r.Get("/health", apiHandler.Health)
	r.Get("/ready", apiHandler.Ready)

	r.Post("/api/v1/trips", apiHandler.CreateTrip)
	r.Get("/api/v1/trips/{tripId}", apiHandler.GetTrip)
	r.Post("/api/v1/trips/{tripId}/finish", apiHandler.FinishTrip)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	return &Container{
		Config:            cfg,
		Logger:            logger,
		Pool:              pool,
		Server:            srv,
		shutdownTelemetry: shutdownTelemetry,
	}, nil
}

func (c *Container) Close() {
	c.Pool.Close()
	if c.shutdownTelemetry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.shutdownTelemetry(ctx)
	}
}
