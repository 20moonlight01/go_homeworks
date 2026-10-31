package finish

import (
	openapi "github.com/oapi-codegen/runtime/types"

	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
)

func toAPIResponse(t *trip.Trip) api.Trip {
	return api.Trip{
		Id:         openapi.UUID(t.Id),
		UserId:     openapi.UUID(t.UserId),
		DriverId:   openapi.UUID(t.DriverId),
		StartPoint: api.Coordinates{Latitude: t.StartPoint.Latitude, Longitude: t.StartPoint.Longitude},
		EndPoint:   api.Coordinates{Latitude: t.EndPoint.Latitude, Longitude: t.EndPoint.Longitude},
		Price:      t.Price,
		Status:     api.TripStatus(t.Status),
		StartedAt:  t.StartedAt,
		FinishedAt: t.FinishedAt,
	}
}
