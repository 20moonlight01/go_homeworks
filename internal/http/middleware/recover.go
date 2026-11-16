package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/20moonlight01/go_homeworks/internal/http/domainerror"
	"go.opentelemetry.io/otel/trace"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, holder := contextWithSpanHolder(r.Context())
			r = r.WithContext(ctx)

			defer func() {
				if p := recover(); p != nil {
					logCtx := r.Context()
					if holder.span != nil {
						logCtx = trace.ContextWithSpan(logCtx, holder.span)
					}

					logger.ErrorContext(logCtx, "PANIC recovered",
						"panic", p,
						"stack", string(debug.Stack()),
					)
					domainerror.Write(w, r.WithContext(logCtx), fmt.Errorf("panic: %v", p))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
