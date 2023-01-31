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

func NewHTTPHandler(fn func(wr http.ResponseWriter, rd *http.Request, svc interface{}), svc interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, svc)
	}
}

type DefaultResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Code    int    `json:"code"`
} // @name DefaultResponse

func encodeResult(w http.ResponseWriter, response interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(&response)
	if err != nil {
		return
	}
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
