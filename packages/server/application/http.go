package application

import (
	"net/http"

	"go.uber.org/zap"
)

type HTTPHandler struct {
	log *zap.Logger
}


func NewHTTPHandler(fn func (wr http.ResponseWriter, rd *http.Request, svc interface{}), svc interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, svc)
	}
}

