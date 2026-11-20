package create

import (
	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
	uc "github.com/20moonlight01/go_homeworks/internal/usecase/trip/position/save"
)

func toUsecaseRequest(d api.PositionData, tripId api.TripId) uc.Request {
	return uc.Request{
		TripId:     tripId,
		Latitude:   d.Coordinates.Latitude,
		Longitude:  d.Coordinates.Longitude,
		RecordedAt: d.RecordedAt,
	}
}

func toAPIResponse(p *trip.Position) api.TripPosition {
	return api.TripPosition{
		Id:     p.Id,
		TripId: p.TripId,
		Coordinates: api.Coordinates{
			Latitude:  p.Latitude,
			Longitude: p.Longitude,
		},
		RecordedAt: p.RecordedAt,
	}
}
