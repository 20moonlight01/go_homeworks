package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("tripgo-http")

	requestDuration, _ = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of HTTP server requests"),
	)

	activeRequests, _ = meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
	)
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		initialAttrs := metric.WithAttributes(
			attribute.String("http.request.method", r.Method),
		)

		activeRequests.Add(r.Context(), 1, initialAttrs)

		defer func() {
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = r.URL.Path
			}
			finalAttrs := metric.WithAttributes(
				attribute.String("http.request.method", r.Method),
				attribute.String("http.route", route),
			)
			activeRequests.Add(r.Context(), -1, finalAttrs)
		}()

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 0}
		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = r.URL.Path
		}

		attrs := []attribute.KeyValue{
			attribute.String("http.request.method", r.Method),
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", rec.status),
		}
		requestDuration.Record(r.Context(), duration, metric.WithAttributes(attrs...))
	})
}
