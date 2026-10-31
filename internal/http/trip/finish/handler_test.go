package finish_test

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
	"github.com/20moonlight01/go_homeworks/internal/http/trip/finish"
)

func TestHandler_FinishTrip(t *testing.T) {
	tripID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := NewMockfinishUsecase(ctrl)
		expectedTrip := &trip.Trip{Id: tripID, Status: trip.StatusCompleted}
		mockUC.EXPECT().Execute(gomock.Any(), tripID).Return(expectedTrip, nil)

		h := finish.NewHandler(mockUC)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+tripID.String()+"/finish", nil)
		rec := httptest.NewRecorder()

		h.FinishTrip(rec, req, api.TripId(tripID))

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp api.Trip
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		assert.Equal(t, string(trip.StatusCompleted), string(resp.Status))
	})

	t.Run("trip completed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := NewMockfinishUsecase(ctrl)
		mockUC.EXPECT().Execute(gomock.Any(), tripID).Return(nil, trip.ErrTripCompleted)

		h := finish.NewHandler(mockUC)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+tripID.String()+"/finish", nil)
		rec := httptest.NewRecorder()

		h.FinishTrip(rec, req, api.TripId(tripID))

		assert.Equal(t, http.StatusConflict, rec.Code)
		var prob api.Problem
		_ = json.NewDecoder(rec.Body).Decode(&prob)
		assert.Equal(t, "trip_completed", prob.Code)
	})
}
