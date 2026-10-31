package domainerror

import (
	"encoding/json"
	"net/http"
)

func Write(w http.ResponseWriter, r *http.Request, err error) {
	status, prob := Problem(err)
	instance := r.URL.Path
	prob.Instance = &instance

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(prob)
}
