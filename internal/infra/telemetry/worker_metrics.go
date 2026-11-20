package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type WorkerMetrics struct {
	staleChecks      metric.Int64Counter
	positionRequests metric.Int64Counter
	requestDuration  metric.Float64Histogram
}

func NewWorkerMetrics(
	queueSizeFn func() int64,
	activeWorkersFn func() int64,
) (*WorkerMetrics, error) {
	meter := otel.Meter("tripgo-worker")

	staleChecks, err := meter.Int64Counter("tripgo.stale_checks")
	if err != nil {
		return nil, err
	}
	_, err = meter.Int64ObservableGauge(
		"tripgo.position_requests.queue.size",
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(queueSizeFn())
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	_, err = meter.Int64ObservableGauge(
		"tripgo.position_requests.workers.active",
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(activeWorkersFn())
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	positionRequests, err := meter.Int64Counter("tripgo.position_requests")
	if err != nil {
		return nil, err
	}
	requestDuration, err := meter.Float64Histogram(
		"tripgo.position_requests.duration",
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return &WorkerMetrics{
		staleChecks:      staleChecks,
		positionRequests: positionRequests,
		requestDuration:  requestDuration,
	}, nil
}

func (m *WorkerMetrics) IncStaleCheck(ctx context.Context, result string) {
	m.staleChecks.Add(ctx, 1, metric.WithAttributes(
		attribute.String("result", result),
	))
}

func (m *WorkerMetrics) IncPositionRequest(ctx context.Context, result string) {
	m.positionRequests.Add(ctx, 1, metric.WithAttributes(
		attribute.String("result", result),
	))
}

func (m *WorkerMetrics) RecordRequestDuration(ctx context.Context, seconds float64, result string) {
	m.requestDuration.Record(ctx, seconds, metric.WithAttributes(
		attribute.String("result", result),
	))
}
