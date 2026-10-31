package create_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/http/trip/create"
)

func TestHandler_CreateTrip(t *testing.T) {
	validBody := `{
		"user_id":"5cb72c04-7650-45c9-a79b-bcdba0631e0c",
		"driver_id":"8860b315-ec86-42eb-a17c-7c163d721ff5",
		"start_point":{"latitude":59.9398,"longitude":30.3146},
		"end_point":{"latitude":59.929,"longitude":30.3626},
		"price":1450
	}`

	newReq := func(body string) *http.Request {
		return httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewBufferString(body))
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := NewMockcreateUsecase(ctrl)
		expectedTrip := &trip.Trip{
			Id:     uuid.New(),
			Status: trip.StatusActive,
		}

		mockUC.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(expectedTrip, nil)

		h := create.NewHandler(mockUC)
		rec := httptest.NewRecorder()
		h.CreateTrip(rec, newReq(validBody), api.CreateTripParams{})

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := NewMockcreateUsecase(ctrl)
		h := create.NewHandler(mockUC)
		rec := httptest.NewRecorder()

		h.CreateTrip(rec, newReq(`{invalid}`), api.CreateTripParams{})

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var prob api.Problem
		_ = json.NewDecoder(rec.Body).Decode(&prob)
		assert.Equal(t, "invalid_request", prob.Code)
	})

	t.Run("missing required fields (null)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := NewMockcreateUsecase(ctrl)
		h := create.NewHandler(mockUC)
		rec := httptest.NewRecorder()

		body := `{"user_id":null,"driver_id":"8860b315-ec86-42eb-a17c-7c163d721ff5","start_point":{"latitude":59.9,"longitude":30.3},"end_point":{"latitude":59.8,"longitude":30.4},"price":1450}`
		h.CreateTrip(rec, newReq(body), api.CreateTripParams{})

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var prob api.Problem
		_ = json.NewDecoder(rec.Body).Decode(&prob)
		assert.Equal(t, "invalid_request", prob.Code)
		assert.Contains(t, *prob.Detail, "user_id")
	})

	t.Run("usecase validation error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := NewMockcreateUsecase(ctrl)
		valErr := &trip.ValidationError{Field: "price", Message: "must be non-negative"}
		mockUC.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil, valErr)

		h := create.NewHandler(mockUC)
		rec := httptest.NewRecorder()
		h.CreateTrip(rec, newReq(validBody), api.CreateTripParams{})

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var prob api.Problem
		_ = json.NewDecoder(rec.Body).Decode(&prob)
		assert.Equal(t, "invalid_request", prob.Code)
	})

	t.Run("usecase internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := NewMockcreateUsecase(ctrl)
		mockUC.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil, errors.New("database down"))

		h := create.NewHandler(mockUC)
		rec := httptest.NewRecorder()
		h.CreateTrip(rec, newReq(validBody), api.CreateTripParams{})

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		var prob api.Problem
		_ = json.NewDecoder(rec.Body).Decode(&prob)
		assert.Equal(t, "internal_error", prob.Code)
	})
}
