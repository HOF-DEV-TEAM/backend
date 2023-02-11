package http_helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"	
)

const (
	PasswordMinLength = 6
)

var (
	ErrUnauthorized 		 = errors.New("unauthorized")
	ErrNoTokenFound 		 = errors.New("no token found")
    ErrUnauthorizedRequest   = errors.New("unauthorized request. please check your credentials")
    ErrNotFound              = errors.New("not found")
    ErrQueryRepository       = errors.New("there was an error executing the query")
    ErrInvalidAccount        = errors.New("this is an invalid account")
    ErrInvalidRequest        = errors.New("invalid request")           // Always should be 400 Bad Request
    ErrUserPwd               = errors.New("invalid login credentials") // This causes 401 Unauthorized
    ErrEmptyLoginCredentials = errors.New("invalid login credentials") // This causes 400 Bad Request
    ErrEmailRequired         = errors.New("email is required")
    ErrNameRequired          = errors.New("name is required")
    ErrPasswordRequired      = errors.New("password is required")
    ErrPasswordLength        = fmt.Errorf("the password minimum length is %d", PasswordMinLength)
    ErrRemovePassword        = errors.New("invalid request")
    ErrUserExists            = errors.New("user with the same email address already exists")
)


func CodeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrUserExists:
		return http.StatusConflict
	case ErrNameRequired, ErrEmailRequired, ErrInvalidRequest, 
		ErrPasswordRequired, ErrRemovePassword, ErrPasswordLength, ErrEmptyLoginCredentials:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrUserPwd:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

type errorResponse struct {
	Err 	string 	`json:"error"`
}

type DefaultResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Code    int    `json:"code"`
}

func EncodeResult(w http.ResponseWriter, response interface{}, code int) {
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
	w.WriteHeader(CodeFrom(err))
	json.NewEncoder(w).Encode(errorResponse{Err: err.Error()})
}
