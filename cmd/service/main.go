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
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("config load failed: %v", err)
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	reqCtx, cancelRequests := context.WithCancel(context.Background())
	defer cancelRequests()

	c, err := container.New(ctx, reqCtx, cfg)
	if err != nil {
		log.Printf("container init failed: %v", err)
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		c.Logger.Info("starting server", "addr", cfg.HTTPAddr)
		if err := c.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.Logger.Error("server error", "error", err)
			errCh <- err
		}
	}()

	var serverErr error
	select {
	case serverErr = <-errCh:
	case <-ctx.Done():
		c.Logger.Info("waiting for active requests on termination signal")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if serverErr == nil {
		if err := c.Server.Shutdown(shutdownCtx); err != nil {
			c.Logger.Error("shutdown error", "error", err)
			cancelRequests()
			_ = c.Server.Close()
		}
	} else {
		cancelRequests()
		_ = c.Server.Close()
	}

	c.Close(shutdownCtx)

	return serverErr
}
