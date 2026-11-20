package list

import (
	"github.com/20moonlight01/go_homeworks/internal/domain/trip"
	api "github.com/20moonlight01/go_homeworks/internal/generated"
)

func toAPIResponse(positions []*trip.Position) api.TripPositionsResponse {
	out := make([]api.TripPosition, 0, len(positions))
	for _, p := range positions {
		out = append(out, api.TripPosition{
			Id:     p.Id,
			TripId: p.TripId,
			Coordinates: api.Coordinates{
				Latitude:  p.Latitude,
				Longitude: p.Longitude,
			},
			RecordedAt: p.RecordedAt,
		})
	}
	return api.TripPositionsResponse{Positions: out}
}
