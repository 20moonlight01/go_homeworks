//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package create

import (
	"context"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	uc "github.com/20moonlight01/go_homeworks/internal/usecase/trip/position/save"
)

type (
	savePositionUsecase interface {
		Execute(ctx context.Context, req uc.Request) (*trip.Position, error)
	}
)
