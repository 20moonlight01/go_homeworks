package create

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type Usecase struct {
	repo    tripRepository
	txm     txManager
	logger  *slog.Logger
	metrics metrics
}

func NewUsecase(repo tripRepository, txm txManager, logger *slog.Logger, metrics metrics) *Usecase {
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

	u.metrics.IncTripsCreated(ctx)

	u.logger.InfoContext(ctx, "trip created", "trip_id", t.Id)

	return t, nil
}
