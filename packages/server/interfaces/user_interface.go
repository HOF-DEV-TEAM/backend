package interfaces

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"

	"bitbucket.org/hofng/hofApp/pkg/user"
)

// CreateGetUserHandler godoc
// @Summary Create a new user
// @Description Create a new user with the input payload
// @Tags SignUp
// @Accept  json
// @Produce  json
// @Param SignUpUserRequestJSON body SignUpUserRequestJSON true "Create user"
// @Success 200 {object} UserJSON
// @Router /session/sign_up [post]
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

// CreateSignInHandler godoc
// @Summary Create a new sign in session for a user
// @Description Create a new sign in session with the input payload
// @Tags Login
// @Accept  json
// @Produce  json
// @Param LoginRequestJSON body LoginRequestJSON true "Sign in user"
// @Success 200 {object} UserSession
// @Router /session/sign_in [post]
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
				User:  user.NewJSONUser(result.User),
				Token: result.Token,
			},
			http.StatusOK)
	}
}

// ForgotPasswordHandler godoc
// @Summary User forgets their password
// @Description User can request for a password change with the input payload
// @Tags ForgotPassword
// @Accept  json
// @Produce  json
// @Param ForgotPasswordPayload body ForgotPasswordPayload true "Forgot password"
// @Success 200 {object} ForgotPasswordResponse
// @Router /session/forgot_password [post]
func ForgotPasswordHandler(svc user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request user.ForgotPasswordPayload
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			encodeResult(w, err, http.StatusInternalServerError)
			return
		}
		url, err := svc.ForgotPassword(request)
		if err != nil {
			encodeResult(w, err, http.StatusInternalServerError)
			return
		}
		encodeResult(w, url, http.StatusOK)
	}
}

// ResetPasswordHandler godoc
// @Summary User can reset their password
// @Description User can insert new passwords for a password change with the input payload
// @Tags ResetPassword
// @Accept  json
// @Produce  json
// @Param ResetPasswordPayload body ResetPasswordPayload true "Reset password"
// @Success 200 {object} DefaultResponse
// @Router /session/reset_password/{url-param} [post]
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
