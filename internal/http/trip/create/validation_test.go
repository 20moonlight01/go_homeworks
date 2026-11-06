package create_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/http/trip/create"
)

const (
	fieldUserId   = "user_id"
	fieldDriverId = "driver_id"
)

func TestHandler_CreateTrip_Validation(t *testing.T) {
	getValidBody := func() map[string]interface{} {
		return map[string]interface{}{
			fieldUserId:   "5cb72c04-7650-45c9-a79b-bcdba0631e0c",
			fieldDriverId: "8860b315-ec86-42eb-a17c-7c163d721ff5",
			"start_point": map[string]interface{}{"latitude": 59.9398, "longitude": 30.3146},
			"end_point":   map[string]interface{}{"latitude": 59.929, "longitude": 30.3626},
			fieldPrice:    1450,
		}
	}

	tests := []struct {
		name      string
		mutate    func(m map[string]interface{})
		wantField string
	}{
		{
			name:      "missing user_id",
			mutate:    func(m map[string]interface{}) { delete(m, fieldUserId) },
			wantField: fieldUserId,
		},
		{
			name:      "missing driver_id",
			mutate:    func(m map[string]interface{}) { delete(m, fieldDriverId) },
			wantField: fieldDriverId,
		},
		{
			name:      "missing start_point",
			mutate:    func(m map[string]interface{}) { delete(m, "start_point") },
			wantField: "start_point",
		},
		{
			name:      "missing end_point",
			mutate:    func(m map[string]interface{}) { delete(m, "end_point") },
			wantField: "end_point",
		},
		{
			name:      "missing price",
			mutate:    func(m map[string]interface{}) { delete(m, fieldPrice) },
			wantField: fieldPrice,
		},
		{
			name: "missing start_point.latitude",
			mutate: func(m map[string]interface{}) {
				sp := m["start_point"].(map[string]interface{})
				delete(sp, "latitude")
			},
			wantField: "start_point_latitude",
		},
		{
			name: "missing start_point.longitude",
			mutate: func(m map[string]interface{}) {
				sp := m["start_point"].(map[string]interface{})
				delete(sp, "longitude")
			},
			wantField: "start_point_longitude",
		},
		{
			name: "missing end_point.latitude",
			mutate: func(m map[string]interface{}) {
				ep := m["end_point"].(map[string]interface{})
				delete(ep, "latitude")
			},
			wantField: "end_point_latitude",
		},
		{
			name: "missing end_point.longitude",
			mutate: func(m map[string]interface{}) {
				ep := m["end_point"].(map[string]interface{})
				delete(ep, "longitude")
			},
			wantField: "end_point_longitude",
		},
		{
			name: "empty user_id (zero UUID)",
			mutate: func(m map[string]interface{}) {
				m[fieldUserId] = "00000000-0000-0000-0000-000000000000"
			},
			wantField: fieldUserId,
		},
		{
			name: "empty driver_id (zero UUID)",
			mutate: func(m map[string]interface{}) {
				m[fieldDriverId] = "00000000-0000-0000-0000-000000000000"
			},
			wantField: fieldDriverId,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := NewMockcreateUsecase(ctrl)
			h := create.NewHandler(mockUC)

			bodyMap := getValidBody()
			tt.mutate(bodyMap)

			body, _ := json.Marshal(bodyMap)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			h.CreateTrip(rec, req, api.CreateTripParams{})

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var prob api.Problem
			_ = json.NewDecoder(rec.Body).Decode(&prob)
			assert.Equal(t, "invalid_request", prob.Code)
			assert.Contains(t, *prob.Detail, tt.wantField)
		})
	}
}
