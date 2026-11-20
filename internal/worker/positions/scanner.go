package positions

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/20moonlight01/go_homeworks/internal/model"
)

type Scanner struct {
	repo       staleTripRepository
	pending    *pendingSet
	queue      chan<- model.StaleTrip
	interval   time.Duration
	staleAfter time.Duration
	logger     *slog.Logger
	metrics    scannerMetrics
	newTicker  tickerFactory
}

func NewScanner(
	repo staleTripRepository,
	pending *pendingSet,
	queue chan<- model.StaleTrip,
	interval time.Duration,
	staleAfter time.Duration,
	logger *slog.Logger,
	metrics scannerMetrics,
	newTicker tickerFactory,
) *Scanner {
	return &Scanner{
		repo:       repo,
		pending:    pending,
		queue:      queue,
		interval:   interval,
		staleAfter: staleAfter,
		logger:     logger,
		metrics:    metrics,
		newTicker:  newTicker,
	}
}

func (s *Scanner) SetMetrics(metrics scannerMetrics) {
	s.metrics = metrics
}

func (s *Scanner) Run(ctx context.Context) {
	ticker := s.newTicker(s.interval)
	defer ticker.Stop()

	s.logger.InfoContext(ctx, "scanner started", "interval", s.interval, "stale_after", s.staleAfter)

	for {
		select {
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "scanner stopped")
			return
		case <-ticker.C():
			s.scan(ctx)
		}
	}
}

func (s *Scanner) scan(ctx context.Context) {
	tracer := otel.Tracer("tripgo-worker")
	ctx, span := tracer.Start(ctx, "stale_check", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	trips, err := s.repo.SelectStale(ctx, s.staleAfter)
	if err != nil {
		s.logger.ErrorContext(ctx, "select stale trips failed", "error", err)
		return
	}
	spanCtx := trace.SpanContextFromContext(ctx)

	for _, t := range trips {
		t.CheckSpanContext = spanCtx

		if !s.pending.TryAdd(t) {
			s.metrics.IncStaleCheck(ctx, "duplicate")
			continue
		}

		select {
		case s.queue <- t:
			s.metrics.IncStaleCheck(ctx, "queued")
		default:
			s.pending.Remove(t.Id)
			s.metrics.IncStaleCheck(ctx, "dropped")
			s.logger.WarnContext(ctx, "position queue full, trip dropped",
				"trip_id", t.Id,
				"driver_id", t.DriverId,
			)
		}
	}
}
