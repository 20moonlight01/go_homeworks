package middleware

import (
	"net/http"
	"strings"
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
		route := UnknownRoute
		if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.Routes != nil {
			tmpCtx := chi.NewRouteContext()
			if rctx.Routes.Match(tmpCtx, r.Method, r.URL.Path) {
				route = tmpCtx.RoutePattern()
			}
		}
		route = strings.TrimSuffix(route, "/*")
		if route == "" {
			route = UnknownRoute
		}

		attrs := metric.WithAttributes(
			attribute.String("http.request.method", r.Method),
			attribute.String("http.route", route),
		)
		activeRequests.Add(r.Context(), 1, attrs)

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 0}

		defer func() {
			activeRequests.Add(r.Context(), -1, attrs)

			status := rec.status
			if p := recover(); p != nil {
				status = http.StatusInternalServerError
				defer panic(p)
			} else if status == 0 {
				status = http.StatusOK
			}

			requestDuration.Record(r.Context(), time.Since(start).Seconds(),
				metric.WithAttributes(
					attribute.String("http.request.method", r.Method),
					attribute.String("http.route", route),
					attribute.Int("http.response.status_code", status),
				),
			)
		}()

		next.ServeHTTP(rec, r)
	})
}
