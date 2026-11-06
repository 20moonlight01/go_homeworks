package trip_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
)

func TestNewTrip(t *testing.T) {
	validUserID := uuid.New()
	validDriverID := uuid.New()

	tests := []struct {
		name         string
		price        int64
		startLat     float64
		startLon     float64
		endLat       float64
		endLon       float64
		wantErr      bool
		wantErrField string
	}{
		{
			name:     "success",
			price:    100,
			startLat: 55.0,
			startLon: 37.0,
			endLat:   56.0,
			endLon:   38.0,
			wantErr:  false,
		},
		{
			name:         "negative price",
			price:        -1,
			startLat:     55.0,
			startLon:     37.0,
			endLat:       56.0,
			endLon:       38.0,
			wantErr:      true,
			wantErrField: "price",
		},
		{
			name:         "invalid start latitude",
			price:        100,
			startLat:     91.0,
			startLon:     37.0,
			endLat:       56.0,
			endLon:       38.0,
			wantErr:      true,
			wantErrField: "start_point.latitude",
		},
		{
			name:         "invalid start longitude",
			price:        100,
			startLat:     55.0,
			startLon:     181.0,
			endLat:       56.0,
			endLon:       38.0,
			wantErr:      true,
			wantErrField: "start_point.longitude",
		},
		{
			name:         "invalid end latitude",
			price:        100,
			startLat:     55.0,
			startLon:     37.0,
			endLat:       91.0,
			endLon:       38.0,
			wantErr:      true,
			wantErrField: "end_point.latitude",
		},
		{
			name:         "invalid end longitude",
			price:        100,
			startLat:     55.0,
			startLon:     37.0,
			endLat:       56.0,
			endLon:       181.0,
			wantErr:      true,
			wantErrField: "end_point.longitude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := trip.NewTrip(
				validUserID, validDriverID,
				tt.startLat, tt.startLon,
				tt.endLat, tt.endLon,
				tt.price,
			)

			if tt.wantErr {
				require.Error(t, err)
				var valErr *trip.ValidationError
				require.True(t, errors.As(err, &valErr))
				assert.Equal(t, tt.wantErrField, valErr.Field)
				assert.Nil(t, tr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tr)
				assert.Equal(t, trip.StatusActive, tr.Status)
				assert.NotEqual(t, uuid.Nil, tr.Id)
				assert.Nil(t, tr.FinishedAt)
			}
		})
	}
}

func TestTrip_Finish(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name       string
		initial    trip.Status
		wantErr    error
		wantStatus trip.Status
	}{
		{
			name:       "active to completed",
			initial:    trip.StatusActive,
			wantErr:    nil,
			wantStatus: trip.StatusCompleted,
		},
		{
			name:       "completed to completed returns error",
			initial:    trip.StatusCompleted,
			wantErr:    trip.ErrTripCompleted,
			wantStatus: trip.StatusCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &trip.Trip{Status: tt.initial}
			err := tr.Finish(now)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, tr.FinishedAt)
				assert.True(t, tr.FinishedAt.Equal(now))
			}
			assert.Equal(t, tt.wantStatus, tr.Status)
		})
	}
}
