package interfaces

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"

	"bitbucket.org/hofng/hofApp/pkg/user"
)

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the input paylod
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body User true "Create user"
// @Success 200 {object} User
// @Router /user [post]

func CreateGetUserHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	var u user.SignUpUserRequestJSON
	err := json.NewDecoder(r.Body).Decode(&u)

	if err != nil {
		encodeResult(w, err, http.StatusInternalServerError)
		return
	}

	result, err := svc.(user.Service).SignUp(r.Context(), u.ToSignUpUser())

	if err != nil {
		EncodeJSONError(r.Context(), err, w)
		return
	}
	encodeResult(w, user.NewJSONUser(result), http.StatusOK)
}

func CreateSignInHandler(svc user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req user.LoginRequestJSON
		err := json.NewDecoder(r.Body).Decode(&req)

		if err != nil {
			encodeResult(w, err, http.StatusInternalServerError)
			return
		}

		result, err := svc.Login(r.Context(), req.Email, req.Password)

		if err != nil {
			EncodeJSONError(r.Context(), err, w)
			return
		}

		encodeResult(
			w, 
			user.UserSession{
				User: user.NewJSONUser(result.User), 
				Token: result.Token,
			}, 
			http.StatusOK)
		}
}

func ForgotPasswordHandler(svc user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request user.ForgotPasswordPayload
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			encodeResult(w, err, http.StatusInternalServerError)
			return
		}
		_, err = svc.ForgotPassword(request)
		if err != nil {
			encodeResult(w, err, http.StatusInternalServerError)
			return
		}
		encodeResult(w, request, http.StatusOK)
	}
}

func ResetPasswordHandler(svc user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resetPasswordRequest user.ResetPasswordPayload
		err := json.NewDecoder(r.Body).Decode(&resetPasswordRequest)
		if err != nil {
			encodeResult(w, err, http.StatusInternalServerError)
			return
		}
		passwordTokenParam := chi.URLParam(r, "token")

		_, err = svc.VerifyPasswordToken(resetPasswordRequest, passwordTokenParam)
		if err != nil {
			encodeResult(w, err, http.StatusBadRequest)
			return
		}

		_, err = svc.ResetPassword(resetPasswordRequest)
		if err != nil {
			encodeResult(w, err, http.StatusBadRequest)
			return

		}

		encodeResult(w, DefaultResponse{Message: "success", Code: http.StatusOK, Success: true}, http.StatusOK)
	}
}
