package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/20moonlight01/go_homeworks/internal/config"
	"github.com/20moonlight01/go_homeworks/internal/container"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c, err := container.New(ctx, cfg)
	if err != nil {
		log.Printf("container init failed: %v", err)
		return
	}
	defer c.Close()

	go func() {
		c.Logger.Info("starting server", "addr", cfg.HTTPAddr)
		if err := c.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.Logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	c.Logger.Info("waiting for active requests on termination signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := c.Server.Shutdown(shutdownCtx); err != nil {
		c.Logger.Error("shutdown error", "error", err)
	}
	c.Logger.Info("service stopped")
}
