package middleware

import (
	"fmt"
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

		setSpanInHolder(ctx, span)

		rec := &statusRecorder{ResponseWriter: w, status: 0}
		r = r.WithContext(ctx)

		defer func() {
			status := rec.status
			if p := recover(); p != nil {
				status = http.StatusInternalServerError
				span.RecordError(fmt.Errorf("panic: %v", p))
				defer panic(p)
			} else if status == 0 {
				status = http.StatusOK
			}

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
		}()

		next.ServeHTTP(rec, r)
	})
}
