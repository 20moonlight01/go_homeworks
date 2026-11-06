package middleware

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(next http.Handler) http.Handler {
	tracer := otel.Tracer("tripgo-http")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "HTTP Request", trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		rec := &statusRecorder{ResponseWriter: w, status: 0}
		r = r.WithContext(ctx)

		next.ServeHTTP(rec, r)

		route := chi.RouteContext(ctx).RoutePattern()
		if route == "" {
			route = r.URL.Path
		}

		span.SetName(r.Method + " " + route)
		span.SetAttributes(
			semconv.HTTPRequestMethodKey.String(r.Method),
			semconv.HTTPRoute(route),
			semconv.HTTPResponseStatusCode(rec.status),
		)

		if rec.status >= 400 {
			errorType := http.StatusText(rec.status)
			if errorType == "" {
				errorType = strconv.Itoa(rec.status)
			}
			span.SetAttributes(attribute.String("error.type", errorType))
			span.SetStatus(codes.Error, errorType)
		}
	})
}
