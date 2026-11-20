package positions_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/20moonlight01/go_homeworks/internal/model"
	pos "github.com/20moonlight01/go_homeworks/internal/worker/positions"
)

func TestWorker_CancelledByPosition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gateway := NewMockpushGateway(ctrl)
	metrics := NewMockworkerMetrics(ctrl)

	trip := model.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}
	pending := pos.NewPendingSet()
	pending.TryAdd(trip)

	queue := make(chan model.StaleTrip, 1)
	queue <- trip

	pending.Cancel(trip.Id)

	gateway.EXPECT().RequestPosition(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	metrics.EXPECT().IncPositionRequest(gomock.Any(), "cancelled")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	pool := pos.NewWorkerPool(queue, pending, gateway, 1, time.Second, time.Second, slog.Default(), metrics, &wg)
	pool.Run(ctx)

	close(queue)
	wg.Wait()
}

func TestWorkerPool_Concurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const workers = 4
	const tasks = 20

	gateway := NewMockpushGateway(ctrl)
	metrics := NewMockworkerMetrics(ctrl)

	var current, maxConcurrent atomic.Int64

	gateway.EXPECT().
		RequestPosition(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, tripID, driverID uuid.UUID) error {
			cur := current.Add(1)
			for {
				prev := maxConcurrent.Load()
				if cur <= prev || maxConcurrent.CompareAndSwap(prev, cur) {
					break
				}
			}
			defer current.Add(-1)

			select {
			case <-time.After(50 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		}).
		Times(tasks)

	metrics.EXPECT().IncPositionRequest(gomock.Any(), "success").Times(tasks)
	metrics.EXPECT().RecordRequestDuration(gomock.Any(), gomock.Any(), "success").Times(tasks)

	pending := pos.NewPendingSet()
	queue := make(chan model.StaleTrip, tasks)
	for i := 0; i < tasks; i++ {
		trip := model.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}
		pending.TryAdd(trip)
		queue <- trip
	}
	close(queue)

	var wg sync.WaitGroup
	pool := pos.NewWorkerPool(queue, pending, gateway, workers, time.Second, time.Second, slog.Default(), metrics, &wg)
	pool.Run(context.Background())

	wg.Wait()
	assert.LessOrEqual(t, int(maxConcurrent.Load()), workers)
}

func TestWorker_PushError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gateway := NewMockpushGateway(ctrl)
	metrics := NewMockworkerMetrics(ctrl)

	trip1 := model.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}
	trip2 := model.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}
	pending := pos.NewPendingSet()
	pending.TryAdd(trip1)
	pending.TryAdd(trip2)

	queue := make(chan model.StaleTrip, 2)
	queue <- trip1
	queue <- trip2

	gateway.EXPECT().
		RequestPosition(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("push failed")).
		Times(2)

	metrics.EXPECT().IncPositionRequest(gomock.Any(), "error").Times(2)
	metrics.EXPECT().RecordRequestDuration(gomock.Any(), gomock.Any(), "error").Times(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	pool := pos.NewWorkerPool(queue, pending, gateway, 1, time.Second, time.Second, slog.Default(), metrics, &wg)
	pool.Run(ctx)

	close(queue)
	wg.Wait()
}

func TestWorkerPool_DrainOnCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gateway := NewMockpushGateway(ctrl)
	metrics := NewMockworkerMetrics(ctrl)

	const tasks = 5
	processed := make(chan struct{}, tasks)
	defer close(processed)

	pending := pos.NewPendingSet()
	queue := make(chan model.StaleTrip, tasks)
	var ids []uuid.UUID
	for i := 0; i < tasks; i++ {
		trip := model.StaleTrip{Id: uuid.New(), DriverId: uuid.New()}
		ids = append(ids, trip.Id)
		pending.TryAdd(trip)
		queue <- trip
	}

	gateway.EXPECT().
		RequestPosition(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, tripID, driverID uuid.UUID) error {
			processed <- struct{}{}
			return nil
		}).
		Times(tasks)

	metrics.EXPECT().IncPositionRequest(gomock.Any(), "success").Times(tasks)
	metrics.EXPECT().RecordRequestDuration(gomock.Any(), gomock.Any(), "success").Times(tasks)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	pool := pos.NewWorkerPool(queue, pending, gateway, 1, time.Second, 2*time.Second, slog.Default(), metrics, &wg)

	drainCtx, drainCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer drainCancel()
	pool.SetDrainContext(drainCtx)

	pool.Run(ctx)

	cancel()
	close(queue)
	wg.Wait()

	assert.Len(t, processed, tasks)
	for _, id := range ids {
		assert.False(t, pending.Contains(id), "pending должен быть очищен после обработки")
	}
}
