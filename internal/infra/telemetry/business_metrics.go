package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	TripsCreated   metric.Int64Counter
	TripsCompleted metric.Int64Counter
	PositionsSaved metric.Int64Counter
}

func (m *Metrics) IncTripsCreated(ctx context.Context) {
	m.TripsCreated.Add(ctx, 1)
}

func (m *Metrics) IncTripsCompleted(ctx context.Context) {
	m.TripsCompleted.Add(ctx, 1)
}

func (m *Metrics) IncPositionsSaved(ctx context.Context) {
	m.PositionsSaved.Add(ctx, 1)
}

func NewMetrics(activeTripsFn func(context.Context) (int64, error)) (*Metrics, error) {
	meter := otel.Meter("tripgo")

	created, err := meter.Int64Counter("tripgo.trips.created")
	if err != nil {
		return nil, err
	}

	completed, err := meter.Int64Counter("tripgo.trips.completed")
	if err != nil {
		return nil, err
	}

	positionsSaved, err := meter.Int64Counter("tripgo.positions.saved")
	if err != nil {
		return nil, err
	}

	_, err = meter.Int64ObservableGauge(
		"tripgo.trips.active",
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			count, err := activeTripsFn(ctx)
			if err != nil {
				return err
			}
			o.Observe(count)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		TripsCreated:   created,
		TripsCompleted: completed,
		PositionsSaved: positionsSaved,
	}, nil
}
