package handler

import (
	"encoding/json"
	"net/http"

	api "github.com/20moonlight01/go_homeworks/internal/generated"
)

const problemTypeBase = "https://tripgo.example/problems/"

func writeProblemJSON(w http.ResponseWriter, r *http.Request, status int, code, title, errType, detail string) {
	instance := r.URL.Path
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(
		api.Problem{
			Type:     problemTypeBase + errType,
			Title:    title,
			Status:   int32(status),
			Detail:   &detail,
			Code:     code,
			Instance: &instance,
		},
	)
}
