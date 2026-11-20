package domainerror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
)

func Problem(err error) (int, api.Problem) {
	var valErr *trip.ValidationError
	if errors.As(err, &valErr) {
		return http.StatusBadRequest, api.Problem{
			Code:   "invalid_request",
			Title:  "Invalid request",
			Status: http.StatusBadRequest,
			Detail: strPtr(fmt.Sprintf("%s: %s", valErr.Field, valErr.Message)),
			Type:   "https://tripgo.example/problems/invalid-request",
		}
	}

	switch {
	case errors.Is(err, trip.ErrTripNotFound):
		return http.StatusNotFound, api.Problem{
			Code:   "trip_not_found",
			Title:  "Trip not found",
			Status: http.StatusNotFound,
			Detail: strPtr("Trip was not found"),
			Type:   "https://tripgo.example/problems/trip-not-found",
		}
	case errors.Is(err, trip.ErrTripCompleted):
		return http.StatusConflict, api.Problem{
			Code:   "trip_completed",
			Title:  "Trip completed",
			Status: http.StatusConflict,
			Detail: strPtr("Operation is not allowed for a completed trip"),
			Type:   "https://tripgo.example/problems/trip-completed",
		}
	case errors.Is(err, trip.ErrDriverBusy):
		return http.StatusConflict, api.Problem{
			Code:   "driver_busy",
			Title:  "Driver busy",
			Status: http.StatusConflict,
			Detail: strPtr("Driver already has an active trip"),
			Type:   "https://tripgo.example/problems/driver-busy",
		}
	default:
		return http.StatusInternalServerError, api.Problem{
			Code:   "internal_error",
			Title:  "Internal Server Error",
			Status: http.StatusInternalServerError,
			Detail: strPtr("Internal server error"),
			Type:   "https://tripgo.example/problems/internal-error",
		}
	}
}

func strPtr(s string) *string { return &s }
