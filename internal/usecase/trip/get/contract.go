package get

import (
	"context"
	"uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type (
	tripRepository interface {
		GetByID(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
	}
)
