package interfaces

import (
	"context"
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


func EncodeJSONError(_ context.Context, err error, w http.ResponseWriter) {
    if err == nil {
        panic("encodeError with nil error")
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    _ = json.NewEncoder(w).Encode(map[string]interface{}{
        "error": err.Error(),
    })
}