package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 0}

			defer func() {
				status := rec.status
				if p := recover(); p != nil {
					status = http.StatusInternalServerError
					defer panic(p)
				} else if status == 0 {
					status = http.StatusOK
				}

				route := chi.RouteContext(r.Context()).RoutePattern()
				if route == "" {
					route = r.URL.Path
				}

				logFunc := logger.InfoContext
				if rec.status >= 500 {
					logFunc = logger.ErrorContext
				}

				logFunc(r.Context(), "request handled",
					"method", r.Method,
					"route", route,
					"status", rec.status,
					"duration_ms", time.Since(start).Milliseconds(),
				)
			}()

			next.ServeHTTP(rec, r)
		})
	}
}
