package handler

import (
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func validateTripDataRequired(raw tripDataRaw) string {
	switch {
	case raw.UserId == nil:
		return "user_id is required"
	case raw.DriverId == nil:
		return "driver_id is required"
	case raw.StartPoint == nil:
		return "start point is required"
	case raw.EndPoint == nil:
		return "end point is required"
	case raw.Price == nil:
		return "price is required"
	}

	if raw.StartPoint.Latitude == nil || raw.StartPoint.Longitude == nil {
		return "start point latitude and start point longitude are required"
	}
	if raw.EndPoint.Latitude == nil || raw.EndPoint.Longitude == nil {
		return "end point latitude and end point longitude are required"
	}

	return ""
}

func validateTripData(req api.CreateTripJSONRequestBody) string {
	if req.UserId == (openapi_types.UUID{}) {
		return "user_id must be non-empty uuid"
	}

	if req.DriverId == (openapi_types.UUID{}) {
		return "driver_id must be non-empty uuid"
	}

	if !(-90 <= req.StartPoint.Latitude && req.StartPoint.Latitude <= 90) ||
		!(-180 <= req.StartPoint.Longitude && req.StartPoint.Longitude <= 180) {
		return "invalid start point coords"
	}

	if !(-90 <= req.EndPoint.Latitude && req.EndPoint.Latitude <= 90) ||
		!(-180 <= req.EndPoint.Longitude && req.EndPoint.Longitude <= 180) {
		return "invalid end point coords"
	}

	if req.Price < 0 {
		return "price must be non-negative"
	}

	return ""
}
