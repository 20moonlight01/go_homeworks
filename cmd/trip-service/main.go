package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/20moonlight01/go_homeworks/internal/config"
	"github.com/20moonlight01/go_homeworks/internal/db"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/handler"
	"github.com/20moonlight01/go_homeworks/internal/repository"
	"github.com/20moonlight01/go_homeworks/internal/txmanager"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.New(ctx, cfg)
	if err != nil {
		log.Fatalf("connection to db failed: %v", err)
	}
	defer pool.Close()

	tripRepo := repository.NewTripRepository(pool, cfg.DBQueryTimeout)
	txm := txmanager.NewManager(pool)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	api.HandlerWithOptions(
		&handler.Server{
			Ping:     pool,
			TripRepo: tripRepo,
			Txm:      txm,
		},
		api.ChiServerOptions{
			BaseRouter: r,
			ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				instance := r.URL.Path
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(
					api.Problem{
						Code:     "invalid_request",
						Instance: &instance,
						Status:   http.StatusBadRequest,
						Title:    "Invalid request",
						Type:     "https://tripgo.example/problems/invalid-request",
					},
				)
			},
		})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
	}

	go func() {
		log.Printf("listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("waiting for active requests on termination signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("shutdown: ", err)
	}

	log.Println("service stopped")
}
