package create

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/http/domainerror"
)

type Handler struct {
	uc savePositionUsecase
}

func NewHandler(uc savePositionUsecase) *Handler {
	return &Handler{uc: uc}
}

type tripPositionDataRaw struct {
	Coordinates *coordinatesRaw `json:"coordinates"`
	RecordedAt  *time.Time      `json:"recorded_at"`
}

type coordinatesRaw struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func (h *Handler) CreateTripPosition(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	var raw tripPositionDataRaw
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil {
		domainerror.Write(w, r, &trip.ValidationError{Field: "body", Message: "invalid JSON or unknown fields"})
		return
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		domainerror.Write(w, r, &trip.ValidationError{Field: "body", Message: "unexpected trailing data"})
		return
	}

	if err := validateTripPositionDataRequired(raw); err != nil {
		domainerror.Write(w, r, err)
		return
	}

	req := api.CreateTripPositionJSONRequestBody{
		Coordinates: api.Coordinates{Latitude: *raw.Coordinates.Latitude, Longitude: *raw.Coordinates.Longitude},
		RecordedAt:  *raw.RecordedAt,
	}

	p, err := h.uc.Execute(r.Context(), toUsecaseRequest(req, tripId))
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}

	resp := toAPIResponse(p)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
