package domainerror_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/http/domainerror"
)

func TestProblem(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantCode       string
		wantTitle      string
		wantDetailPart string
		wantType       string
	}{
		{
			name:           "trip not found",
			err:            trip.ErrTripNotFound,
			wantStatus:     http.StatusNotFound,
			wantCode:       "trip_not_found",
			wantTitle:      "Trip not found",
			wantDetailPart: "Trip was not found",
			wantType:       "https://tripgo.example/problems/trip-not-found",
		},
		{
			name:           "trip completed",
			err:            trip.ErrTripCompleted,
			wantStatus:     http.StatusConflict,
			wantCode:       "trip_completed",
			wantTitle:      "Trip completed",
			wantDetailPart: "Operation is not allowed for a completed trip",
			wantType:       "https://tripgo.example/problems/trip-completed",
		},
		{
			name:           "driver busy",
			err:            trip.ErrDriverBusy,
			wantStatus:     http.StatusConflict,
			wantCode:       "driver_busy",
			wantTitle:      "Driver busy",
			wantDetailPart: "Driver already has an active trip",
			wantType:       "https://tripgo.example/problems/driver-busy",
		},
		{
			name:           "validation error",
			err:            &trip.ValidationError{Field: "price", Message: "must be non-negative"},
			wantStatus:     http.StatusBadRequest,
			wantCode:       "invalid_request",
			wantTitle:      "Invalid request",
			wantDetailPart: "price: must be non-negative",
			wantType:       "https://tripgo.example/problems/invalid-request",
		},
		{
			name:           "unknown error",
			err:            errors.New("db connection lost"),
			wantStatus:     http.StatusInternalServerError,
			wantCode:       "internal_error",
			wantTitle:      "Internal Server Error",
			wantDetailPart: "Internal server error",
			wantType:       "https://tripgo.example/problems/internal-error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, prob := domainerror.Problem(tt.err)

			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantCode, prob.Code)
			assert.Equal(t, tt.wantTitle, prob.Title)
			assert.Equal(t, tt.wantType, prob.Type)

			require.NotNil(t, prob.Detail)
			assert.Contains(t, *prob.Detail, tt.wantDetailPart)
		})
	}
}

func TestWrite(t *testing.T) {
	t.Run("success write", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/123", nil)
		rec := httptest.NewRecorder()

		domainerror.Write(rec, req, trip.ErrTripNotFound)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

		var prob api.Problem
		err := json.Unmarshal(rec.Body.Bytes(), &prob)
		require.NoError(t, err)

		assert.Equal(t, "trip_not_found", prob.Code)
		assert.Equal(t, http.StatusNotFound, int(prob.Status))

		require.NotNil(t, prob.Instance)
		assert.Equal(t, "/api/v1/trips/123", *prob.Instance)
	})
}
