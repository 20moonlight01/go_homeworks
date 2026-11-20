package list

import (
	"context"

	"github.com/google/uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type Usecase struct {
	tripRepo     tripRepository
	positionRepo positionRepository
}

func NewUsecase(tripRepo tripRepository, positionRepo positionRepository) *Usecase {
	return &Usecase{tripRepo: tripRepo, positionRepo: positionRepo}
}

func (u *Usecase) Execute(ctx context.Context, tripId uuid.UUID) ([]*trip.Position, error) {
	if _, err := u.tripRepo.GetByID(ctx, tripId); err != nil {
		return nil, err
	}

	return u.positionRepo.ListPositions(ctx, tripId)
}
