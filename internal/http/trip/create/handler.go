package create

import (
	"encoding/json"
	"net/http"

	openapi "github.com/oapi-codegen/runtime/types"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/http/domainerror"
)

type Handler struct {
	uc createUsecase
}

func NewHandler(uc createUsecase) *Handler {
	return &Handler{uc: uc}
}

type tripDataRaw struct {
	UserId     *openapi.UUID   `json:"user_id"`
	DriverId   *openapi.UUID   `json:"driver_id"`
	StartPoint *coordinatesRaw `json:"start_point"`
	EndPoint   *coordinatesRaw `json:"end_point"`
	Price      *int64          `json:"price"`
}

type coordinatesRaw struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, _ api.CreateTripParams) {
	var raw tripDataRaw
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil {
		domainerror.Write(w, r, &trip.ValidationError{Field: "body", Message: "invalid JSON or unknown fields"})
		return
	}

	if err := validateTripDataRequired(raw); err != nil {
		domainerror.Write(w, r, err)
		return
	}

	req := api.CreateTripJSONRequestBody{
		UserId:     *raw.UserId,
		DriverId:   *raw.DriverId,
		StartPoint: api.Coordinates{Latitude: *raw.StartPoint.Latitude, Longitude: *raw.StartPoint.Longitude},
		EndPoint:   api.Coordinates{Latitude: *raw.EndPoint.Latitude, Longitude: *raw.EndPoint.Longitude},
		Price:      *raw.Price,
	}

	if err := validateTripData(req); err != nil {
		domainerror.Write(w, r, err)
		return
	}

	t, err := h.uc.Execute(r.Context(), toUsecaseRequest(req))
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}

	resp := toAPIResponse(t)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/api/v1/trips/"+resp.Id.String())
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
