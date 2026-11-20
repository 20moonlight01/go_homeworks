package save

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type Usecase struct {
	repo      positionRepository
	txm       txManager
	logger    *slog.Logger
	metrics   metrics
	canceller positionCanceller
}

type Request struct {
	TripId     uuid.UUID
	Latitude   float64
	Longitude  float64
	RecordedAt time.Time
}

func NewUsecase(
	repo positionRepository,
	txm txManager,
	logger *slog.Logger,
	metrics metrics,
	canceller positionCanceller,
) *Usecase {
	return &Usecase{
		repo:      repo,
		txm:       txm,
		logger:    logger,
		metrics:   metrics,
		canceller: canceller,
	}
}

func (u *Usecase) Execute(ctx context.Context, req Request) (*trip.Position, error) {
	now := time.Now().UTC()
	p, err := trip.NewPosition(req.TripId, req.Latitude, req.Longitude, req.RecordedAt, now)
	if err != nil {
		return nil, err
	}

	var result *trip.Position
	err = u.txm.Do(ctx, func(ctx context.Context) error {
		var repoErr error
		result, repoErr = u.repo.SavePosition(ctx, p)
		return repoErr
	})
	if err != nil {
		return nil, err
	}

	u.canceller.Cancel(result.TripId)

	u.metrics.IncPositionsSaved(ctx)
	u.logger.InfoContext(ctx, "position saved", "trip_id", result.TripId, "position_id", result.Id)

	return result, nil
}
