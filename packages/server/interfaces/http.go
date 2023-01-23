package interfaces

import (
	"bitbucket.org/hofng/hofApp/domain/repository"
	"encoding/json"
	"net/http"
)

type HTTPHandler struct {
	repo repository.Repositories
}

func New(repo repository.Repositories) *HTTPHandler {
	return &HTTPHandler{
		repo: repo,
	}
}

func encodeResult(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	json.NewEncoder(w).Encode(&result)
}
