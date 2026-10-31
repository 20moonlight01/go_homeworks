package container

import (
	"net/http"

	api "github.com/20moonlight01/go_homeworks/internal/generated"
	httpService "github.com/20moonlight01/go_homeworks/internal/http/service"
	httpCreate "github.com/20moonlight01/go_homeworks/internal/http/trip/create"
	httpFinish "github.com/20moonlight01/go_homeworks/internal/http/trip/finish"
	httpGet "github.com/20moonlight01/go_homeworks/internal/http/trip/get"
)

type apiAdapter struct {
	api.Unimplemented
	create  *httpCreate.Handler
	get     *httpGet.Handler
	finish  *httpFinish.Handler
	service *httpService.Handler
}

func (a *apiAdapter) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	a.create.CreateTrip(w, r, params)
}
func (a *apiAdapter) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	a.get.GetTrip(w, r, tripId)
}
func (a *apiAdapter) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	a.finish.FinishTrip(w, r, tripId)
}
func (a *apiAdapter) Health(w http.ResponseWriter, r *http.Request) {
	a.service.Health(w, r)
}
func (a *apiAdapter) Ready(w http.ResponseWriter, r *http.Request) {
	a.service.Ready(w, r)
}
