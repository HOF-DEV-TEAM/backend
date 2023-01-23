package interfaces

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type HTTPHandler struct {
	log *zap.Logger
}

func New(log *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		log: log,

	}
}

func encodeResult(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	json.NewEncoder(w).Encode(&result)
}
