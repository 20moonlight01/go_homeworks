package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type metricsCtxKey struct{}

type metricsData struct {
	start time.Time
	sql   string
}

type MetricsTracer struct {
	dur metric.Float64Histogram
}

func NewMetricsTracer() *MetricsTracer {
	meter := otel.Meter("tripgo-db")
	dur, _ := meter.Float64Histogram(
		"db.client.operation.duration",
		metric.WithUnit("s"),
	)
	return &MetricsTracer{dur: dur}
}

func (t *MetricsTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, metricsCtxKey{}, metricsData{
		start: time.Now(),
		sql:   data.SQL,
	})
}

func (t *MetricsTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	val, ok := ctx.Value(metricsCtxKey{}).(metricsData)
	if !ok {
		return
	}

	duration := time.Since(val.start).Seconds()
	op, table := parseSQL(val.sql)

	attrs := []attribute.KeyValue{
		attribute.String("db.operation.name", op),
		attribute.String("db.collection.name", table),
	}

	if data.Err != nil {
		attrs = append(attrs, attribute.String("error.type", getErrorType(data.Err)))
	}

	t.dur.Record(ctx, duration, metric.WithAttributes(attrs...))
}
