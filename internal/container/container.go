package container

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/20moonlight01/go_homeworks/internal/config"
	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
	"github.com/20moonlight01/go_homeworks/internal/gateway/push"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/http/domainerror"
	"github.com/20moonlight01/go_homeworks/internal/http/middleware"
	httpService "github.com/20moonlight01/go_homeworks/internal/http/service"
	httpCreate "github.com/20moonlight01/go_homeworks/internal/http/trip/create"
	httpFinish "github.com/20moonlight01/go_homeworks/internal/http/trip/finish"
	httpGet "github.com/20moonlight01/go_homeworks/internal/http/trip/get"
	httpCreatePosition "github.com/20moonlight01/go_homeworks/internal/http/trip/position/create"
	httpListPosition "github.com/20moonlight01/go_homeworks/internal/http/trip/position/list"
	"github.com/20moonlight01/go_homeworks/internal/infra/postgres"
	"github.com/20moonlight01/go_homeworks/internal/infra/postgres/transaction"
	"github.com/20moonlight01/go_homeworks/internal/infra/telemetry"
	"github.com/20moonlight01/go_homeworks/internal/model"
	"github.com/20moonlight01/go_homeworks/internal/repository/trip"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/create"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/finish"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/get"
	listPosition "github.com/20moonlight01/go_homeworks/internal/usecase/trip/position/list"
	savePosition "github.com/20moonlight01/go_homeworks/internal/usecase/trip/position/save"
	"github.com/20moonlight01/go_homeworks/internal/worker/positions"
)

type Container struct {
	Config *config.Config
	Logger *slog.Logger
	Pool   *pgxpool.Pool
	Server *http.Server

	shutdownTelemetry func(context.Context) error

	scanner      *positions.Scanner
	workerCancel context.CancelFunc
	workerWG     sync.WaitGroup
	workerPool   *positions.WorkerPool

	pushGateway *push.Gateway
}

func New(ctx, baseCtx context.Context, cfg *config.Config) (*Container, error) {
	shutdownTelemetry, err := telemetry.Init(ctx, cfg.OTELServiceName)
	if err != nil {
		return nil, fmt.Errorf("init telemetry: %w", err)
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

	metrics, err := telemetry.NewMetrics(tripRepo.CountActive)
	if err != nil {
		_ = shutdownTelemetry(ctx)
		pool.Close()
		return nil, fmt.Errorf("init metrics: %w", err)
	}

	pending := positions.NewPendingSet()

	createUC := create.NewUsecase(tripRepo, txManager, logger, metrics)
	getUC := get.NewUsecase(tripRepo)
	finishUC := finish.NewUsecase(tripRepo, txManager, logger, metrics)
	savePositionUC := savePosition.NewUsecase(tripRepo, txManager, logger, metrics, pending)
	listPositionsUC := listPosition.NewUsecase(tripRepo, tripRepo)

	hCreate := httpCreate.NewHandler(createUC)
	hGet := httpGet.NewHandler(getUC)
	hFinish := httpFinish.NewHandler(finishUC)
	hSavePosition := httpCreatePosition.NewHandler(savePositionUC)
	hListPositions := httpListPosition.NewHandler(listPositionsUC)
	hService := httpService.NewHandler(pool)

	r := chi.NewRouter()
	r.Use(middleware.Recover(logger))
	r.Use(middleware.Tracing)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Metrics)

	apiHandler := &api.ServerInterfaceWrapper{
		Handler: &apiAdapter{
			create:        hCreate,
			get:           hGet,
			finish:        hFinish,
			savePosition:  hSavePosition,
			listPositions: hListPositions,
			service:       hService,
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

	r.Post("/api/v1/trips/{tripId}/positions", apiHandler.CreateTripPosition)
	r.Get("/api/v1/trips/{tripId}/positions", apiHandler.ListTripPositions)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		BaseContext:       func(_ net.Listener) context.Context { return baseCtx },
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	queue := make(chan model.StaleTrip, cfg.PositionQueueSize)
	pushGateway := push.NewGateway(cfg.PushHTTPURL, cfg.PushRequestTimeout)

	scanner := positions.NewScanner(
		tripRepo,
		pending,
		queue,
		cfg.PositionCheckInterval,
		cfg.PositionStaleAfter,
		logger,
		nil,
		positions.RealTickerFactory,
	)

	c := &Container{
		Config:            cfg,
		Logger:            logger,
		Pool:              pool,
		Server:            srv,
		shutdownTelemetry: shutdownTelemetry,
		scanner:           scanner,
		workerWG:          sync.WaitGroup{},
		pushGateway:       pushGateway,
	}

	workerCtx, workerCancel := context.WithCancel(context.Background())
	c.workerCancel = workerCancel
	workerPool := positions.NewWorkerPool(
		queue,
		pending,
		pushGateway,
		int(cfg.PositionWorkers),
		cfg.PushRequestTimeout,
		cfg.ShutdownTimeout,
		logger, nil,
		&c.workerWG,
	)
	c.workerPool = workerPool

	workerMetrics, err := telemetry.NewWorkerMetrics(
		func() int64 { return int64(len(queue)) },
		workerPool.ActiveWorkers,
	)
	if err != nil {
		pushGateway.Close()
		_ = shutdownTelemetry(ctx)
		pool.Close()
		return nil, fmt.Errorf("init worker metrics: %w", err)
	}

	c.scanner.SetMetrics(workerMetrics)
	c.workerPool.SetMetrics(workerMetrics)

	c.workerWG.Add(1)
	go func() {
		defer c.workerWG.Done()
		scanner.Run(workerCtx)
		close(queue)
	}()
	workerPool.Run(workerCtx)

	return c, nil
}

func (c *Container) Close(ctx context.Context) {
	c.workerPool.SetDrainContext(ctx)
	c.workerCancel()

	done := make(chan struct{})
	go func() {
		c.workerWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.Logger.Info("workers stopped gracefully")
	case <-ctx.Done():
		c.Logger.Warn("workers shutdown deadline exceeded", "error", ctx.Err())
	}

	if c.shutdownTelemetry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.shutdownTelemetry(ctx)
	}

	if c.pushGateway != nil {
		c.pushGateway.Close()
		c.Logger.Info("http client for push service stopped")
	}
	c.Pool.Close()
	c.Logger.Info("service stopped")
}
