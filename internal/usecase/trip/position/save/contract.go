//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package save

import (
	"context"

	"github.com/google/uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type (
	positionRepository interface {
		SavePosition(ctx context.Context, p *trip.Position) (*trip.Position, error)
	}

	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}

	metrics interface {
		IncPositionsSaved(ctx context.Context)
	}

	positionCanceller interface {
		Cancel(tripId uuid.UUID)
	}
)
