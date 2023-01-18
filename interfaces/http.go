package interfaces

import (
	"bitbucket.org/hofng/hofApp/application"
	"encoding/json"
	"net/http"
)

type HTTPHandler struct {
	app application.Applications
}

func New(app application.Applications) *HTTPHandler {
	return &HTTPHandler{
		app: app,
	}
}

func encodeResult(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	json.NewEncoder(w).Encode(&result)
}
