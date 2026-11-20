package trip

import (
	"time"

	"github.com/google/uuid"
)

type Position struct {
	Id         int64
	TripId     uuid.UUID
	Latitude   float64
	Longitude  float64
	RecordedAt time.Time
	CreatedAt  time.Time
}

func NewPosition(tripId uuid.UUID, latitude, longitude float64, recordedAt, now time.Time) (*Position, error) {
	if latitude < -90 || latitude > 90 {
		return nil, &ValidationError{Field: "coordinates.latitude", Message: msgOutOfRange}
	}
	if longitude < -180 || longitude > 180 {
		return nil, &ValidationError{Field: "coordinates.longitude", Message: msgOutOfRange}
	}
	if recordedAt.After(now.Add(time.Minute)) {
		return nil, &ValidationError{Field: "recorded_at", Message: "must not be in the future"}
	}
	return &Position{
		TripId:     tripId,
		Latitude:   latitude,
		Longitude:  longitude,
		RecordedAt: recordedAt.UTC(),
	}, nil
}
