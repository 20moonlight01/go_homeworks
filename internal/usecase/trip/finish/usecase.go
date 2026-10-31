package finish

import (
	"context"

	"github.com/google/uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type Usecase struct {
	repo tripRepository
	txm  txManager
}

func NewUsecase(repo tripRepository, txm txManager) *Usecase {
	return &Usecase{
		repo: repo,
		txm:  txm,
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

	return result, nil
}
