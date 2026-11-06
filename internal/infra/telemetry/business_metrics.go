package telemetry

import (
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var ActiveTripsCount atomic.Int64

type Metrics struct {
	TripsCreated   metric.Int64Counter
	TripsCompleted metric.Int64Counter
	TripsActive    metric.Int64Gauge
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

func NewNoopMetrics() *Metrics {
	meter := noop.NewMeterProvider().Meter("noop")

	created, _ := meter.Int64Counter("tripgo.trips.created")
	completed, _ := meter.Int64Counter("tripgo.trips.completed")
	active, _ := meter.Int64Gauge("tripgo.trips.active")

	return &Metrics{
		TripsCreated:   created,
		TripsCompleted: completed,
		TripsActive:    active,
	}
}
