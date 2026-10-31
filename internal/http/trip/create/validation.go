package create

import (
	openapi "github.com/oapi-codegen/runtime/types"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
)

const (
	msgIsRequired         = "is required"
	msgMustBeNonEmptyUuid = "must be non-empty uuid"
)

func validateTripDataRequired(raw tripDataRaw) error {
	switch {
	case raw.UserId == nil:
		return &trip.ValidationError{Field: "user_id", Message: msgIsRequired}
	case raw.DriverId == nil:
		return &trip.ValidationError{Field: "driver_id", Message: msgIsRequired}
	case raw.StartPoint == nil:
		return &trip.ValidationError{Field: "start_point", Message: msgIsRequired}
	case raw.EndPoint == nil:
		return &trip.ValidationError{Field: "end_point", Message: msgIsRequired}
	case raw.Price == nil:
		return &trip.ValidationError{Field: "price", Message: msgIsRequired}
	case raw.StartPoint.Latitude == nil:
		return &trip.ValidationError{Field: "start_point_latitude", Message: msgIsRequired}
	case raw.StartPoint.Longitude == nil:
		return &trip.ValidationError{Field: "start_point_longitude", Message: msgIsRequired}
	case raw.EndPoint.Latitude == nil:
		return &trip.ValidationError{Field: "end_point_latitude", Message: msgIsRequired}
	case raw.EndPoint.Longitude == nil:
		return &trip.ValidationError{Field: "end_point_longitude", Message: msgIsRequired}
	}

	return nil
}

func validateTripData(req api.CreateTripJSONRequestBody) error {
	if req.UserId == (openapi.UUID{}) {
		return &trip.ValidationError{Field: "user_id", Message: msgMustBeNonEmptyUuid}
	}

	if req.DriverId == (openapi.UUID{}) {
		return &trip.ValidationError{Field: "driver_id", Message: msgMustBeNonEmptyUuid}
	}

	return nil
}
