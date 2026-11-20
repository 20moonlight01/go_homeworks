//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package list

import (
	"context"

	"github.com/google/uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type (
	tripRepository interface {
		GetByID(ctx context.Context, id uuid.UUID) (*trip.Trip, error)
	}

	positionRepository interface {
		ListPositions(ctx context.Context, tripId uuid.UUID) ([]*trip.Position, error)
	}
)
