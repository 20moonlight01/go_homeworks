package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/20moonlight01/go_homeworks/internal/http/ctxlog"
)

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 0}

			entry := &ctxlog.Entry{}
			r = r.WithContext(ctxlog.WithEntry(r.Context(), entry))

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
					route = UnknownRoute
				}

				logFunc := logger.InfoContext
				if status >= 500 {
					logFunc = logger.ErrorContext
				}

				attrs := []any{
					"method", r.Method,
					"route", route,
					"status", status,
					"duration_ms", time.Since(start).Milliseconds(),
				}
				if entry.Err != nil {
					attrs = append(attrs, "error", entry.Err.Error())
				}

				logFunc(r.Context(), "request handled", attrs...)
			}()

			next.ServeHTTP(rec, r)
		})
	}
}
