//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package finish

import (
	"context"

	"github.com/google/uuid"

	domain "github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type (
	tripRepository interface {
		GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*domain.Trip, error)
		SaveFinished(ctx context.Context, t *domain.Trip, fromStatus domain.Status) error
	}

	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}

	metrics interface {
		IncTripsCompleted(ctx context.Context)
	}
)
