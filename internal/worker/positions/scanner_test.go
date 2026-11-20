package positions_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/20moonlight01/go_homeworks/internal/model"
	pos "github.com/20moonlight01/go_homeworks/internal/worker/positions"
)

func TestScanner_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockstaleTripRepository(ctrl)
	metrics := NewMockscannerMetrics(ctrl)
	ticker := NewMockTicker(ctrl)

	tickCh := make(chan time.Time, 2)
	ticker.EXPECT().C().Return(tickCh).AnyTimes()
	ticker.EXPECT().Stop()

	trip := model.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}
	repo.EXPECT().SelectStale(gomock.Any(), time.Minute).Return([]model.StaleTrip{trip}, nil).Times(2)

	queuedDone := make(chan struct{})
	dupDone := make(chan struct{})

	gomock.InOrder(
		metrics.EXPECT().IncStaleCheck(gomock.Any(), "queued").Do(func(context.Context, string) {
			close(queuedDone)
		}),
		metrics.EXPECT().IncStaleCheck(gomock.Any(), "duplicate").Do(func(context.Context, string) {
			close(dupDone)
		}),
	)

	pending := pos.NewPendingSet()
	queue := make(chan model.StaleTrip, 10)
	s := pos.NewScanner(repo, pending, queue, time.Minute, time.Minute, slog.Default(), metrics, func(time.Duration) pos.Ticker {
		return ticker
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.Run(ctx)
	}()

	tickCh <- time.Now()
	<-queuedDone

	tickCh <- time.Now()
	<-dupDone

	cancel()
	<-done
}

func TestScanner_QueueFull(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockstaleTripRepository(ctrl)
	metrics := NewMockscannerMetrics(ctrl)
	ticker := NewMockTicker(ctrl)

	tickCh := make(chan time.Time, 1)
	ticker.EXPECT().C().Return(tickCh).AnyTimes()
	ticker.EXPECT().Stop()

	trip1 := model.StaleTrip{Id: uuid.New()}
	trip2 := model.StaleTrip{Id: uuid.New()}
	repo.EXPECT().SelectStale(gomock.Any(), time.Minute).Return([]model.StaleTrip{trip1, trip2}, nil)

	droppedDone := make(chan struct{})

	gomock.InOrder(
		metrics.EXPECT().IncStaleCheck(gomock.Any(), "queued"),
		metrics.EXPECT().IncStaleCheck(gomock.Any(), "dropped").Do(func(context.Context, string) {
			close(droppedDone)
		}),
	)

	pending := pos.NewPendingSet()
	queue := make(chan model.StaleTrip, 1)
	s := pos.NewScanner(repo, pending, queue, time.Minute, time.Minute, slog.Default(), metrics, func(time.Duration) pos.Ticker {
		return ticker
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.Run(ctx)
	}()

	tickCh <- time.Now()
	<-droppedDone

	assert.False(t, pending.Contains(trip2.Id))

	cancel()
	<-done
}
