package create

import "github.com/20moonlight01/go_homeworks/internal/domain/trip"

const msgIsRequired = "is required"

func validateTripPositionDataRequired(raw tripPositionDataRaw) error {
	switch {
	case raw.Coordinates == nil:
		return &trip.ValidationError{Field: "coordinates", Message: msgIsRequired}
	case raw.RecordedAt == nil:
		return &trip.ValidationError{Field: "recorded_at", Message: msgIsRequired}
	case raw.Coordinates.Latitude == nil:
		return &trip.ValidationError{Field: "coordinates_latitude", Message: msgIsRequired}
	case raw.Coordinates.Longitude == nil:
		return &trip.ValidationError{Field: "coordinates_longitude", Message: msgIsRequired}
	}

	return nil
}
