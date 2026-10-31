package create

import (
	"context"
	"uuid"

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

	return t, nil
}
