package telemetry

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// var ActiveTripsCount atomic.Int64

type Metrics struct {
	TripsCreated   metric.Int64Counter
	TripsCompleted metric.Int64Counter
	TripsActive    metric.Int64Gauge

	ActiveTripsCount atomic.Int64
}

func (m *Metrics) IncTripsCreated(ctx context.Context) {
	m.TripsCreated.Add(ctx, 1)
}

func (m *Metrics) IncTripsCompleted(ctx context.Context) {
	m.TripsCompleted.Add(ctx, 1)
}

func (m *Metrics) IncTripsActive(ctx context.Context) {
	n := m.ActiveTripsCount.Add(1)
	m.TripsActive.Record(ctx, n)
}

func (m *Metrics) DecTripsActive(ctx context.Context) {
	n := m.ActiveTripsCount.Add(-1)
	m.TripsActive.Record(ctx, n)
}

func (m *Metrics) SetTripsActive(ctx context.Context, n int64) {
	m.TripsActive.Record(ctx, n)
}

func NewMetrics() (*Metrics, error) {
	meter := otel.Meter("tripgo")

	created, err := meter.Int64Counter("tripgo.trips.created")
	if err != nil {
		return nil, err
	}

	completed, err := meter.Int64Counter("tripgo.trips.completed")
	if err != nil {
		return nil, err
	}

	active, err := meter.Int64Gauge("tripgo.trips.active")
	if err != nil {
		return nil, err
	}

	return &Metrics{
		TripsCreated:   created,
		TripsCompleted: completed,
		TripsActive:    active,
	}, nil
}
