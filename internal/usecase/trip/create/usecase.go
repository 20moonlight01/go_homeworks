package create

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	"github.com/20moonlight01/go_homeworks/internal/infra/telemetry"
)

type Usecase struct {
	repo    tripRepository
	txm     txManager
	logger  *slog.Logger
	metrics *telemetry.Metrics
}

func NewUsecase(repo tripRepository, txm txManager, logger *slog.Logger, metrics *telemetry.Metrics) *Usecase {
	return &Usecase{
		repo:    repo,
		txm:     txm,
		logger:  logger,
		metrics: metrics,
	}
}

type Request struct {
	UserId         uuid.UUID
	DriverId       uuid.UUID
	StartLatitude  float64
	StartLongitude float64
	EndLatitude    float64
	EndLongitude   float64
	Price          int64
}

func (u *Usecase) Execute(ctx context.Context, req Request) (*trip.Trip, error) {
	t, err := trip.NewTrip(
		req.UserId,
		req.DriverId,
		req.StartLatitude,
		req.StartLongitude,
		req.EndLatitude,
		req.EndLongitude,
		req.Price,
	)
	if err != nil {
		return nil, err
	}

	err = u.txm.Do(ctx, func(ctx context.Context) error {
		return u.repo.Save(ctx, t)
	})
	if err != nil {
		return nil, err
	}

	u.metrics.TripsCreated.Add(ctx, 1)

	newActiveCount := telemetry.ActiveTripsCount.Add(1)
	u.metrics.TripsActive.Record(ctx, newActiveCount)

	u.logger.InfoContext(ctx, "trip created", "trip_id", t.Id)

	return t, nil
}
