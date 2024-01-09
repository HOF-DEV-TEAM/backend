package auth

import (
	"bitbucket.org/hofng/hofApp/pkg/globalParameters/entity"
	"encoding/json"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"bitbucket.org/hofng/hofApp/pkg/user"
)

type AuthenticateRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginWithDeviceRequest struct {
	LoginRequest
	DeviceIdentifier string `json:"device_identifier" validate:"required"`
}

type UserSession struct {
	User            *user.UserJSON                 `json:"user"`
	Subscription    *subscription.SubscriptionJSON `json:"subscription"`
	GlobalVariables *entity.GlobalParameters       `json:"global_variables,omitempty"`
	Token           string                         `json:"token"`
	RefreshToken    string                         `json:"refresh_token"`
} //	@name	UserSession

// SignInHandler godoc
//
//	@Summary		Create a new session
//	@Description	Authenticates a user and returns a session
//	@Tags			Sessions
//	@Accept			json
//	@Produce		json
//	@Param			LoginWithDeviceRequest	body		LoginWithDeviceRequest	true	"Sign in user"
//	@Success		200						{object}	UserSession
//	@Router			/session/sign_in [post]
func SignInHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginRequest LoginWithDeviceRequest
		err := json.NewDecoder(r.Body).Decode(&loginRequest)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		result, err := svc.Login(r.Context(), &loginRequest)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}
}
func AdminSignInHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginRequest LoginRequest
		err := json.NewDecoder(r.Body).Decode(&loginRequest)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		result, err := svc.AdminLogin(r.Context(), &loginRequest)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}
}

func AuthenticateHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := security.VerifyRequest(r, security.TokenFromHeader)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		var req AuthenticateRequest

		err = json.NewDecoder(r.Body).Decode(&req)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		result, err := svc.Authenticate(r.Context(), tokenString, req.RefreshToken)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}
}
