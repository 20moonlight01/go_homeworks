package finish

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

func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	var result *trip.Trip

	err := u.txm.Do(ctx, func(ctx context.Context) error {
		var repoErr error
		result, repoErr = u.repo.Finish(ctx, id)
		return repoErr
	})
	if err != nil {
		return nil, err
	}

	u.metrics.TripsCompleted.Add(ctx, 1)

	newActiveCount := telemetry.ActiveTripsCount.Add(-1)
	u.metrics.TripsActive.Record(ctx, newActiveCount)

	u.logger.InfoContext(ctx, "trip finished", "trip_id", result.Id)

	return result, nil
}
