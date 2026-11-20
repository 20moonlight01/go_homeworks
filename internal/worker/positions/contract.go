//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package positions

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/20moonlight01/go_homeworks/internal/model"
)

type (
	staleTripRepository interface {
		SelectStale(ctx context.Context, staleAfter time.Duration) ([]model.StaleTrip, error)
	}

	scannerMetrics interface {
		IncStaleCheck(ctx context.Context, result string)
	}

	pushGateway interface {
		RequestPosition(ctx context.Context, tripId, driverId uuid.UUID) error
	}

	workerMetrics interface {
		IncPositionRequest(ctx context.Context, result string)
		RecordRequestDuration(ctx context.Context, seconds float64, result string)
	}

	Ticker interface {
		C() <-chan time.Time
		Stop()
	}
)
