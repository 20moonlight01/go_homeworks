package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"uuid"

	"github.com/20moonlight01/go_homeworks/internal/domain"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/repository"
	"github.com/20moonlight01/go_homeworks/internal/txmanager"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Server struct {
	api.Unimplemented
	Ping     Pinger
	TripRepo *repository.TripRepository
	Txm      txmanager.TxManager
}

func (s *Server) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	var req api.CreateTripJSONRequestBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeProblemJSON(w, r, http.StatusBadRequest,
			"invalid_request", "Invalid request", "invalid-request", "JSON decoding failed")
		return
	}

	if detail := validateTripData(req); detail != "" {
		writeProblemJSON(w, r, http.StatusBadRequest,
			"invalid_request", "Invalid request", "invalid-request", detail)
		return
	}

	trip := api.Trip{
		Id:         openapi_types.UUID(uuid.New()),
		DriverId:   req.DriverId,
		UserId:     req.UserId,
		StartPoint: req.StartPoint,
		EndPoint:   req.EndPoint,
		Price:      req.Price,
		Status:     api.Active,
		StartedAt:  time.Now().UTC(),
	}

	var newTrip api.Trip
	err := s.Txm.Do(r.Context(), func(ctx context.Context) error {
		var repoErr error
		newTrip, repoErr = s.TripRepo.Create(ctx, trip)
		return repoErr
	})
	if err != nil {
		if errors.Is(err, domain.ErrDriverBusy) {
			writeProblemJSON(w, r, http.StatusConflict,
				"driver_busy", "Driver busy", "driver-busy", "Driver already has an active trip")
			return
		}

		writeProblemJSON(w, r, http.StatusInternalServerError,
			"internal_error", "Internal Server Error", "internal-error", "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/api/v1/trips/"+newTrip.Id.String())
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTrip)
}

func (s *Server) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	trip, err := s.TripRepo.GetByID(r.Context(), tripId)
	if err != nil {
		if errors.Is(err, domain.ErrTripNotFound) {
			writeProblemJSON(w, r, http.StatusNotFound,
				"trip_not_found", "Trip not found", "trip-not-found", "Trip was not found")
			return
		}

		writeProblemJSON(w, r, http.StatusInternalServerError,
			"internal_error", "Internal Server Error", "internal-error", "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trip)
}

func (s *Server) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	var trip api.Trip
	err := s.Txm.Do(r.Context(), func(ctx context.Context) error {
		var repoErr error
		trip, repoErr = s.TripRepo.Finish(ctx, tripId)
		return repoErr
	})
	if err != nil {
		if errors.Is(err, domain.ErrTripNotFound) {
			writeProblemJSON(w, r, http.StatusConflict,
				"trip_not_found", "Trip not found", "trip-not-found", "Trip was not found")
			return
		}

		if errors.Is(err, domain.ErrTripCompleted) {
			writeProblemJSON(w, r, http.StatusConflict,
				"trip_completed", "Trip completed", "trip-completed", "Operation is not allowed for a completed trip")
			return
		}

		writeProblemJSON(w, r, http.StatusInternalServerError,
			"internal_error", "Internal Server Error", "internal-error", "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trip)
}

func validateTripData(req api.CreateTripJSONRequestBody) string {
	if req.UserId == (openapi_types.UUID{}) {
		return "user_id must be non-empty uuid"
	}

	if req.DriverId == (openapi_types.UUID{}) {
		return "driver_id must be non-empty uuid"
	}

	if !(-90 < req.StartPoint.Latitude && req.StartPoint.Latitude < 90) ||
		!(-180 < req.StartPoint.Longitude && req.StartPoint.Longitude < 180) {
		return "invalid start point coords"
	}

	if !(-90 < req.EndPoint.Latitude && req.EndPoint.Latitude < 90) ||
		!(-180 < req.EndPoint.Longitude && req.EndPoint.Longitude < 180) {
		return "invalid end point coords"
	}

	if req.Price < 0 {
		return "price must be non-negative"
	}

	return ""
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.HealthResponse{Status: api.Ok})
}

func (s *Server) Ready(w http.ResponseWriter, r *http.Request) {
	checkCtx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	if err := s.Ping.Ping(checkCtx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(api.HealthResponse{Status: api.Unavailable})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.HealthResponse{Status: api.Ok})
}
