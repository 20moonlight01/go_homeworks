package finish

import (
	"context"
	"uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type (
	tripRepository interface {
		Finish(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
	}

	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}
)
