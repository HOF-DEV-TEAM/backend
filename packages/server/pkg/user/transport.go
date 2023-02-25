package user

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"github.com/go-chi/chi/v5"
)

type UserAndToken struct {
	User  *User
	Token string
}

type UserSession struct {
	User  *UserJSON `json:"user"`
	Token string    `json:"token"`
} //	@name	UserSession

type AddressJSON string

type UserJSON struct {
	ID          string         `json:"id"`
	Username    string         `json:"username"`
	Password    string         `json:"password,omitempty"`
	Email       string         `json:"email"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	Address     string         `json:"address,omitempty"`
	Mobile      string         `json:"mobile,omitempty"`
	Gender      string         `json:"gender,omitempty"`
	IsVerified  IsVerifiedEnum `json:"is_verified,omitempty"`
	NewJWTToken string         `json:"newToken,omitempty"`
} //	@name	UserJSON

type LoginRequestJSON struct {
	Email    string `json:"email"`
	Password string `json:"password"`
} //	@name	LoginRequestJSON

type SignUpUserRequestJSON struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
} //	@name	SignUpUserRequestJSON

func (u *UserJSON) ToUser() *User {
	result := &User{
		ID:        u.ID,
		Email:     u.Email,
		Password:  u.Password,
		UserName:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Address:   u.Address,
		Gender:    u.Gender,
	}

	if u.Mobile != "" {
		result.Mobile = sql.NullString{Valid: true, String: u.Mobile}
	}
	return result
}

func (u *SignUpUserRequestJSON) ToSignUpUser() *SignUpUser {
	result := &SignUpUser{
		Email:     u.Email,
		Password:  u.Password,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}

	return result
}

func NewJSONUser(u *User) *UserJSON {
	return &UserJSON{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Address:   u.Address,
		Mobile:    u.Mobile.String,
		Gender:    u.Gender,
		Username:  u.UserName,
	}
}

// CreateGetUserHandler godoc
//
//	@Summary		Sign up a new user
//	@Description	Creates a new user with the input payload
//	@Tags			Sessions
//	@Accept			json
//	@Produce		json
//	@Param			SignUpUserRequestJSON	body		SignUpUserRequestJSON	true	"Create user"
//	@Success		200						{object}	UserJSON
//	@Router			/session/sign_up [post]
func CreateGetUserHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	var u SignUpUserRequestJSON
	err := json.NewDecoder(r.Body).Decode(&u)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	result, err := svc.(Service).SignUp(r.Context(), u.ToSignUpUser())

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	payload := NewJSONUser(result)

	http_helper.EncodeResult(w, payload, http.StatusOK)
}

// CreateSignInHandler godoc
//
//	@Summary		Create a new session
//	@Description	Authenticates a user and returns a session
//	@Tags			Sessions
//	@Accept			json
//	@Produce		json
//	@Param			LoginRequestJSON	body		LoginRequestJSON	true	"Sign in user"
//	@Success		200					{object}	UserSession
//	@Router			/session/sign_in [post]
func CreateSignInHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequestJSON
		err := json.NewDecoder(r.Body).Decode(&req)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		result, err := svc.Login(r.Context(), req.Email, req.Password)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		payload := UserSession{
			User:  NewJSONUser(result.User),
			Token: result.Token,
		}

		http_helper.EncodeResult(w, payload, http.StatusOK)
	}
}

// ForgotPasswordHandler godoc
//
//	@Summary		User forgets their password
//	@Description	User can request for a password change with the input payload
//	@Tags			Sessions
//	@Accept			json
//	@Produce		json
//	@Param			ForgotPasswordPayload	body		ForgotPasswordPayload	true	"Forgot password"
//	@Success		200						{object}	ForgotPasswordResponse
//	@Router			/session/forgot_password [post]
func ForgotPasswordHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ForgotPasswordPayload
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		url, err := svc.ForgotPassword(request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, url, http.StatusOK)
	}
}

// ResetPasswordHandler godoc
//
//	@Summary		Rest user password
//	@Description	Creat new password with the input payload
//	@Tags			Sessions
//	@Accept			json
//	@Produce		json
//	@Param			ResetPasswordPayload	body		ResetPasswordPayload	true	"Reset password"
//	@Success		200						{object}	http_helper.DefaultResponse
//	@Router			/session/reset_password/{password_token} [post]
func ResetPasswordHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resetPasswordRequest ResetPasswordPayload
		err := json.NewDecoder(r.Body).Decode(&resetPasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		passwordTokenParam := chi.URLParam(r, "token")
		_, err = svc.VerifyPasswordToken(resetPasswordRequest, passwordTokenParam)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		_, err = svc.ResetPassword(resetPasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return

		}

		http_helper.EncodeResult(w, http_helper.DefaultResponse{Code: http.StatusOK, Success: true}, http.StatusOK)
	}
}
