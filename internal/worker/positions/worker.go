package positions

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/20moonlight01/go_homeworks/internal/model"
)

type WorkerPool struct {
	queue           <-chan model.StaleTrip
	pending         *pendingSet
	gateway         pushGateway
	workers         int
	timeout         time.Duration
	shutdownTimeout time.Duration
	logger          *slog.Logger
	metrics         workerMetrics
	wg              *sync.WaitGroup
	activeWorkers   atomic.Int64
	drainMu         sync.RWMutex
	drainCtx        context.Context
}

func NewWorkerPool(
	queue <-chan model.StaleTrip,
	pending *pendingSet,
	gateway pushGateway,
	workers int,
	timeout time.Duration,
	shutdownTimeout time.Duration,
	logger *slog.Logger,
	metrics workerMetrics,
	wg *sync.WaitGroup,
) *WorkerPool {
	return &WorkerPool{
		queue:           queue,
		pending:         pending,
		gateway:         gateway,
		workers:         workers,
		timeout:         timeout,
		shutdownTimeout: shutdownTimeout,
		logger:          logger,
		metrics:         metrics,
		wg:              wg,
	}
}

func (p *WorkerPool) SetMetrics(metrics workerMetrics) {
	p.metrics = metrics
}

func (p *WorkerPool) SetDrainContext(ctx context.Context) {
	p.drainMu.Lock()
	defer p.drainMu.Unlock()
	p.drainCtx = ctx
}

func (p *WorkerPool) getDrainContext() (context.Context, context.CancelFunc) {
	p.drainMu.RLock()
	defer p.drainMu.RUnlock()
	if p.drainCtx != nil {
		return p.drainCtx, nil
	}
	return context.WithTimeout(context.Background(), p.shutdownTimeout)
}

func (p *WorkerPool) ActiveWorkers() int64 {
	return p.activeWorkers.Load()
}

func (p *WorkerPool) Run(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.workerLoop(ctx)
		}()
	}
}

func (p *WorkerPool) workerLoop(ctx context.Context) {
	p.activeWorkers.Add(1)
	defer p.activeWorkers.Add(-1)

	for {
		select {
		case <-ctx.Done():
			p.drain()
			return
		case trip, ok := <-p.queue:
			if !ok {
				return
			}
			p.handle(ctx, trip)
		}
	}
}

func (p *WorkerPool) drain() {
	drainCtx, cancel := p.getDrainContext()
	if cancel != nil {
		defer cancel()
	}

	for {
		select {
		case t, ok := <-p.queue:
			if !ok {
				return
			}
			p.handle(drainCtx, t)
		case <-drainCtx.Done():
			p.logger.Warn("drain timeout exceeded, dropping remaining queue", "error", drainCtx.Err())
			return
		}
	}
}

func (p *WorkerPool) handle(ctx context.Context, trip model.StaleTrip) {
	if !p.pending.Contains(trip.Id) {
		p.metrics.IncPositionRequest(ctx, "cancelled")
		return
	}

	p.pending.Remove(trip.Id)

	reqCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	tracer := otel.Tracer("tripgo-worker")
	ctx, span := tracer.Start(context.Background(), "push.request",
		trace.WithNewRoot(),
		trace.WithLinks(trace.Link{SpanContext: trip.CheckSpanContext}),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("trip_id", trip.Id.String()),
			attribute.String("driver_id", trip.DriverId.String()),
		),
	)
	defer span.End()

	start := time.Now()
	err := p.gateway.RequestPosition(reqCtx, trip.Id, trip.DriverId)
	duration := time.Since(start).Seconds()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		p.metrics.IncPositionRequest(ctx, "error")
		p.metrics.RecordRequestDuration(ctx, duration, "error")
		p.logger.WarnContext(ctx, "push request failed", "trip_id", trip.Id, "error", err)
		return
	}

	p.metrics.IncPositionRequest(ctx, "success")
	p.metrics.RecordRequestDuration(ctx, duration, "success")
	p.logger.InfoContext(ctx, "push request sent", "trip_id", trip.Id)
}
