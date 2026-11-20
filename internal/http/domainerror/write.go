package domainerror

import (
	"encoding/json"
	"net/http"

	"github.com/20moonlight01/go_homeworks/internal/http/ctxlog"
)

func Write(w http.ResponseWriter, r *http.Request, err error) {
	if entry := ctxlog.FromContext(r.Context()); entry != nil {
		entry.Err = err
	}

	status, prob := Problem(err)
	instance := r.URL.Path
	prob.Instance = &instance

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(prob)
}
