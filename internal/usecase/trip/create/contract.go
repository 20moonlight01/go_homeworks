//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package create

import (
	"context"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

type (
	tripRepository interface {
		Save(ctx context.Context, t *trip.Trip) error
	}

	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}
)
