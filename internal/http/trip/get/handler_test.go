package get_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/http/trip/get"
)

func TestHandler_GetTrip(t *testing.T) {
	tripID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := NewMockgetUsecase(ctrl)
		expectedTrip := &trip.Trip{Id: tripID, Status: trip.StatusActive}
		mockUC.EXPECT().Execute(gomock.Any(), tripID).Return(expectedTrip, nil)

		h := get.NewHandler(mockUC)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+tripID.String(), nil)
		rec := httptest.NewRecorder()

		h.GetTrip(rec, req, tripID)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp api.Trip
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		assert.Equal(t, tripID.String(), resp.Id.String())
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := NewMockgetUsecase(ctrl)
		mockUC.EXPECT().Execute(gomock.Any(), tripID).Return(nil, trip.ErrTripNotFound)

		h := get.NewHandler(mockUC)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+tripID.String(), nil)
		rec := httptest.NewRecorder()

		h.GetTrip(rec, req, tripID)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		var prob api.Problem
		_ = json.NewDecoder(rec.Body).Decode(&prob)
		assert.Equal(t, "trip_not_found", prob.Code)
	})
}
