package create

import (
	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func validateTripDataRequired(raw tripDataRaw) error {
	switch {
	case raw.UserId == nil:
		return &trip.ValidationError{Field: "user_id", Message: "is required"}
	case raw.DriverId == nil:
		return &trip.ValidationError{Field: "driver_id", Message: "is required"}
	case raw.StartPoint == nil:
		return &trip.ValidationError{Field: "start_point", Message: "is required"}
	case raw.EndPoint == nil:
		return &trip.ValidationError{Field: "end_point", Message: "is required"}
	case raw.Price == nil:
		return &trip.ValidationError{Field: "price", Message: "is required"}
	case raw.StartPoint.Latitude == nil:
		return &trip.ValidationError{Field: "start_point_latitude", Message: "is required"}
	case raw.StartPoint.Longitude == nil:
		return &trip.ValidationError{Field: "start_point_longitude", Message: "is required"}
	case raw.EndPoint.Latitude == nil:
		return &trip.ValidationError{Field: "end_point_latitude", Message: "is required"}
	case raw.EndPoint.Longitude == nil:
		return &trip.ValidationError{Field: "end_point_longitude", Message: "is required"}
	}

	return nil
}

func validateTripData(req api.CreateTripJSONRequestBody) error {
	if req.UserId == (openapi_types.UUID{}) {
		return &trip.ValidationError{Field: "user_id", Message: "must be non-empty uuid"}
	}

	if req.DriverId == (openapi_types.UUID{}) {
		return &trip.ValidationError{Field: "driver_id", Message: "must be non-empty uuid"}
	}

	return nil
}
