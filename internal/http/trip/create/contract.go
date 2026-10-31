package create

import (
	"context"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	"github.com/20moonlight01/go_homeworks/internal/usecase/trip/create"
)

type (
	createUsecase interface {
		Execute(ctx context.Context, req create.Request) (*trip.Trip, error)
	}
)
