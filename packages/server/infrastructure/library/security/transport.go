package security

import (
	"encoding/json"
	"net/http"
)

type authResponse struct {
	Err string `json:"message"`
}
func EncodeJSONError(err error, w http.ResponseWriter) {
	code := http.StatusUnauthorized
	msg := err.Error()
	w.WriteHeader(code)

	err = json.NewEncoder(w).Encode(authResponse{Err: msg})
	if err != nil {
		panic(err)
	}
}