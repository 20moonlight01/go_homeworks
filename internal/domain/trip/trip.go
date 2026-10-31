package trip

import (
	"time"
	"uuid"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
)

type Point struct {
	Latitude  float64
	Longitude float64
}

type Trip struct {
	Id         uuid.UUID
	UserId     uuid.UUID
	DriverId   uuid.UUID
	StartPoint Point
	EndPoint   Point
	Price      int64
	Status     Status
	StartedAt  time.Time
	FinishedAt *time.Time
}

func NewTrip(
	userId uuid.UUID,
	driverId uuid.UUID,
	startLatitude float64,
	startLongitude float64,
	endLatitude float64,
	endLongitude float64,
	price int64) (*Trip, error) {
	if price < 0 {
		return nil, &ValidationError{Field: "price", Message: "must be non-negative"}
	}
	if startLatitude < -90 || startLatitude > 90 {
		return nil, &ValidationError{Field: "start_point.latitude", Message: "out of range"}
	}
	if startLongitude < -180 || startLongitude > 180 {
		return nil, &ValidationError{Field: "start_point.longitude", Message: "out of range"}
	}
	if endLatitude < -90 || endLatitude > 90 {
		return nil, &ValidationError{Field: "end_point.latitude", Message: "out of range"}
	}
	if endLongitude < -180 || endLongitude > 180 {
		return nil, &ValidationError{Field: "end_point.longitude", Message: "out of range"}
	}

	return &Trip{
		UserId:   userId,
		DriverId: driverId,
		StartPoint: Point{
			Latitude:  startLatitude,
			Longitude: startLongitude,
		},
		EndPoint: Point{
			Latitude:  endLatitude,
			Longitude: endLongitude,
		},
		Price:     price,
		Status:    StatusActive,
		StartedAt: time.Now().UTC(),
	}, nil
}

func (t *Trip) Finish(at time.Time) error {
	if t.Status != StatusActive {
		return ErrTripCompleted
	}

	t.Status = StatusCompleted
	t.FinishedAt = &at

	return nil
}
