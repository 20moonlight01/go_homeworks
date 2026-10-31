package middleware

import (
	"fmt"
	"log"
	"net/http"

	"github.com/20moonlight01/go_homeworks/internal/http/domainerror"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("PANIC: %v\n", p)
				domainerror.Write(w, r, fmt.Errorf("panic: %v", p))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
