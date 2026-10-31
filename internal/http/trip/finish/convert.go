package finish

import (
	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func toAPIResponse(t *trip.Trip) api.Trip {
	return api.Trip{
		Id:         openapi_types.UUID(t.Id),
		UserId:     openapi_types.UUID(t.UserId),
		DriverId:   openapi_types.UUID(t.DriverId),
		StartPoint: api.Coordinates{Latitude: t.StartPoint.Latitude, Longitude: t.StartPoint.Longitude},
		EndPoint:   api.Coordinates{Latitude: t.EndPoint.Latitude, Longitude: t.EndPoint.Longitude},
		Price:      t.Price,
		Status:     api.TripStatus(t.Status),
		StartedAt:  t.StartedAt,
		FinishedAt: t.FinishedAt,
	}
}
