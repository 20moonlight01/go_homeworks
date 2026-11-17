package finish

import (
	"context"
	"log/slog"
	"time"

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

func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (*trip.Trip, error) {
	var result *trip.Trip

	err := u.txm.Do(ctx, func(ctx context.Context) error {
		t, err := u.repo.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}

		prevStatus := t.Status

		if err := t.Finish(time.Now().UTC()); err != nil {
			return err
		}

		if err := u.repo.SaveFinished(ctx, t, prevStatus); err != nil {
			return err
		}

		result = t
		return nil
	})
	if err != nil {
		return nil, err
	}

	u.metrics.IncTripsCompleted(ctx)

	u.logger.InfoContext(ctx, "trip finished", "trip_id", result.Id)

	return result, nil
}
