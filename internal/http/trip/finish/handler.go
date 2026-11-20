package finish

import (
	"encoding/json"
	"net/http"

	api "github.com/20moonlight01/go_homeworks/internal/generated"
	"github.com/20moonlight01/go_homeworks/internal/http/domainerror"
)

type Handler struct {
	uc finishUsecase
}

func NewHandler(uc finishUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	t, err := h.uc.Execute(r.Context(), tripId)
	if err != nil {
		domainerror.Write(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toAPIResponse(t))
}
