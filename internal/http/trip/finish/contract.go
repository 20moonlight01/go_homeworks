package finish

import (
	"context"
	"uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type (
	finishUsecase interface {
		Execute(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
	}
)
