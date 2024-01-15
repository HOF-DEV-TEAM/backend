package http_helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
)

const (
	PasswordMinLength = 6
)

func ParseJWTError(err error) (error, bool) {
	if e, ok := err.(*jwt.ValidationError); ok {
		switch e.Errors {
		case jwt.ValidationErrorMalformed, jwt.ValidationErrorExpired, jwt.ValidationErrorNotValidYet, jwt.ValidationErrorClaimsInvalid:
			//Token is malformed, Token is expired, Token is not active yet
			if e.Inner != nil {
				return e.Inner, true
			}
		default:
			return ErrUnauthorized, true

		}
	}
	return nil, false
}

var (
	ErrUnauthorized          = errors.New("unauthorized")
	ErrTokenInvalid          = errors.New("token contains an invalid number of segments")
	ErrNoTokenFound          = errors.New("no token found")
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
	fmt.Println(err)
	if _, ok := ParseJWTError(err); ok {
		return http.StatusUnauthorized
	}

	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrUserExists:
		return http.StatusConflict
	case ErrNameRequired, ErrEmailRequired, ErrInvalidRequest,
		ErrPasswordRequired, ErrRemovePassword, ErrPasswordLength, ErrEmptyLoginCredentials:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrUserPwd, ErrTokenInvalid:
		return http.StatusUnauthorized
	default:
		fmt.Println(err)
		return http.StatusInternalServerError
	}
}

type errorResponse struct {
	Err string `json:"error"`
} //	@name	errorResponse

type DefaultResponse struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Body    interface{} `json:"body,omitempty"`
} //	@name	DefaultResponse

func EncodeResult(w http.ResponseWriter, response interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	//TODO: will add this later when mobile is ready to change the response structure.
	// data := struct {
	// 	Data interface{} `json:"data"`
	// }{
	// 	Data: response,
	// }

	err := json.NewEncoder(w).Encode(response)
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
	err = json.NewEncoder(w).Encode(errorResponse{Err: err.Error()})
	if err != nil {
		return
	}
	return
}
