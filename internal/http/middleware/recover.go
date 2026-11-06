package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/20moonlight01/go_homeworks/internal/http/domainerror"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					logger.ErrorContext(r.Context(), "PANIC recovered", "panic", p)
					domainerror.Write(w, r, fmt.Errorf("panic: %v", p))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
