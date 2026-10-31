package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	api "github.com/20moonlight01/go_homeworks/internal/generated"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	pinger Pinger
}

func NewHandler(p Pinger) *Handler {
	return &Handler{pinger: p}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.HealthResponse{Status: api.Ok})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	checkCtx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	if err := h.pinger.Ping(checkCtx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(api.HealthResponse{Status: api.Unavailable})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.HealthResponse{Status: api.Ok})
}
