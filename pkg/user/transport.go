package user

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/mailer"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"net/http"
	"time"
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
	ID               string         `json:"id"`
	Username         string         `json:"username"`
	Password         string         `json:"password,omitempty"`
	Email            string         `json:"email"`
	FirstName        string         `json:"first_name"`
	LastName         string         `json:"last_name"`
	Address          string         `json:"address,omitempty"`
	Mobile           string         `json:"mobile,omitempty"`
	Gender           string         `json:"gender,omitempty"`
	IsVerified       IsVerifiedEnum `json:"is_verified"`
	Devices          []Device       `json:"devices,omitempty"`
	LatestAppVersion VersionManager `json:"latest_app_version,omitempty"`
	NewJWTToken      string         `json:"newToken,omitempty"`
} //	@name	UserJSON

type SignUpUserRequestJSON struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	Devices   []Device `json:"devices"`
} //	@name	SignUpUserRequestJSON

type FavouriteJSON struct {
	ID     uuid.UUID     `json:"id,omitempty"`
	UserID uuid.UUID     `json:"user_id"`
	Fav    []FavBodyJSON `json:"fav"`
} //	@name	FavouriteJSON

type FavBodyJSON struct {
	MessageID uuid.UUID `json:"message_id"`
	SeriesID  string    `json:"series_id"`
	Fav       bool      `json:"fav"`
	DateAdded string    `json:"date_added"`
	DeletedAt string    `json:"deleted_at"`
} //	@name	FavBodyJSON

type FavMessageJSON struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Fav          bool      `json:"fav"`
	MessageID    uuid.UUID `json:"message_id"`
	SeriesID     uuid.UUID `json:"series_id"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	ImageUrl     string    `json:"image_url"`
	AudioUrl     string    `json:"audio_url"`
	Description  string    `json:"description"`
	DateReleased string    `sql:"date_released"`
	IsFree       bool      `sql:"is_free"`
} //	@name	FavMessageJSON

type PageResponse struct {
	TotalResults int32 `json:"totalResults"`
} //	@name	PageResponse

type GetFavouritesResponse struct {
	Favourites []*FavMessageJSON `json:"favourites"`
	Pagination PageResponse      `json:"pagination"`
} //	@name	GetAudiosSeriesResponse

func (u *UserJSON) ToUser() *User {
	result := &User{
		ID:         u.ID,
		Email:      u.Email,
		Password:   u.Password,
		UserName:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Address:    u.Address,
		Gender:     u.Gender,
		IsVerified: u.IsVerified,
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
		Devices:   u.Devices,
	}

	return result
}

func (u *User) ToJSON() *UserJSON {
	return &UserJSON{
		ID:               u.ID,
		Email:            u.Email,
		FirstName:        u.FirstName,
		LastName:         u.LastName,
		Address:          u.Address,
		Mobile:           u.Mobile.String,
		Gender:           u.Gender,
		Username:         u.UserName,
		IsVerified:       u.IsVerified,
		LatestAppVersion: u.LatestAppVersion,
		//Devices:    u.Devices,
	}
}

// SignupUserHandler godoc
//
//	@Summary		Signs up a new user
//	@Description	Creates a new user with the input payload
//	@Tags			Sessions
//	@Accept			json
//	@Produce		json
//	@Param			SignUpUserRequestJSON	body		SignUpUserRequestJSON	true	"Create user"
//	@Success		200						{object}	UserJSON
//	@Router			/session/sign_up [post]
func SignupUserHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var u SignUpUserRequestJSON
		err := json.NewDecoder(r.Body).Decode(&u)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		result, err := s.SignUp(r.Context(), u.ToSignUpUser(), u.Devices)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		payload := result.ToJSON()

		http_helper.EncodeResult(w, payload, http.StatusOK)
	}
	return
}

// ForgotPasswordHandler godoc
//
//	@Summary		User forgets their password
//	@Description	User can request for a password change with the input payload
//	@Tags			Password
//	@Accept			json
//	@Produce		json
//	@Param			ForgotPasswordPayload	body		ForgotPasswordPayload	true	"Forgot password"
//	@Success		200						{object}	http_helper.DefaultResponse
//	@Router			/session/forgot_password [post]
func ForgotPasswordHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var request ForgotPasswordPayload
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		err = s.ForgotPassword(request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, http_helper.DefaultResponse{
			Code:    http.StatusOK,
			Success: true,
			Body:    "OTP sent",
		}, http.StatusOK)
	}
	return
}

// VerifyPasswordResetOTPHandler godoc
//
//	@Summary		Verify password reset OTP
//	@Description	The endpoint verifies the OTP input from the user
//	@Tags			Password
//	@Accept			json
//	@Produce		json
//	@Param			VerifyOTP	body		VerifyOTP	true	"Verify OTP"
//	@Success		200			{object}	UserSession
//	@Router			/session/verify_token [post]
func VerifyPasswordResetOTPHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var request VerifyOTP
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		result, err := s.VerifyPasswordResetOTP(r.Context(), &request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		payload := UserSession{
			Token: result.Token,
		}

		http_helper.EncodeResult(w, payload, http.StatusOK)
	}
	return
}

// ResetPasswordHandler godoc
//
//	@Summary		Reset user password
//	@Description	Create new password with the input payload
//	@Tags			Password
//	@Accept			json
//	@Produce		json
//	@Param			ResetPasswordPayload	body		ResetPasswordPayload	true	"Reset password"
//	@Success		200						{object}	http_helper.DefaultResponse
//	@Router			/user/reset_password [post]
func ResetPasswordHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var resetPasswordRequest ResetPasswordPayload
		err := json.NewDecoder(r.Body).Decode(&resetPasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		_, err = s.ResetPassword(r.Context(), resetPasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return

		}

		http_helper.EncodeResult(w, http_helper.DefaultResponse{Code: http.StatusOK, Success: true}, http.StatusOK)
	}
	return
}

// ChangePasswordHandler godoc
//
//	@Summary		Change user password
//	@Description	Create new password with the input payload
//	@Tags			Password
//	@Accept			json
//	@Produce		json
//	@Param			ChangePasswordPayload	body		ChangePasswordPayload	true	"Change password"
//	@Success		200						{object}	http_helper.DefaultResponse
//	@Router			/user/change_password [post]
func ChangePasswordHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var changePasswordRequest ChangePasswordPayload
		err := json.NewDecoder(r.Body).Decode(&changePasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		_, err = s.ChangePassword(r.Context(), changePasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return

		}

		http_helper.EncodeResult(w, http_helper.DefaultResponse{Code: http.StatusOK, Success: true}, http.StatusOK)
	}
	return
}

func (fav *FavouriteJSON) ToFavourite() *Favourites {
	var favBody []FavBody

	for _, val := range fav.Fav {
		fav := FavBody{
			MessageID: val.MessageID,
			SeriesID:  val.SeriesID,
			Fav:       val.Fav,
			DateAdded: val.DateAdded,
			DeletedAt: val.DeletedAt,
		}
		favBody = append(favBody, fav)
	}

	//result := &Favourites{
	//	UserID: fav.UserID,
	//	Fav: Favs{
	//		favBody,
	//	},
	//}

	result := &Favourites{
		UserID: fav.UserID,
		Fav:    favBody,
	}

	return result
}

func NewJSONFavourite(fav *Favourites) *FavouriteJSON {
	var favBodyJson []FavBodyJSON

	//for _, val := range fav.Fav.Favourite {
	//	fav := FavBodyJSON{
	//		MessageID: val.MessageID,
	//		SeriesID:  val.SeriesID,
	//		Fav:       val.Fav,
	//		DateAdded: val.DateAdded,
	//		DeletedAt: val.DeletedAt,
	//	}
	//	favBodyJson = append(favBodyJson, fav)
	//}

	for _, val := range fav.Fav {
		fav := FavBodyJSON{
			MessageID: val.MessageID,
			SeriesID:  val.SeriesID,
			Fav:       val.Fav,
			DateAdded: val.DateAdded,
			DeletedAt: val.DeletedAt,
		}
		favBodyJson = append(favBodyJson, fav)
	}

	return &FavouriteJSON{
		ID:     fav.ID,
		UserID: fav.UserID,
		Fav:    favBodyJson,
	}
}

func NewJSONFavMessage(fav *FavMessage) *FavMessageJSON {
	return &FavMessageJSON{
		ID:           fav.ID,
		Fav:          fav.Fav,
		UserID:       fav.UserID,
		MessageID:    fav.MessageID,
		SeriesID:     fav.SeriesID,
		Title:        fav.Title,
		Author:       fav.Author,
		ImageUrl:     fav.ImageUrl,
		AudioUrl:     fav.AudioUrl,
		Description:  fav.Description,
		DateReleased: fav.DateReleased.String,
		IsFree:       fav.IsFree,
	}
}

// CreateFavouriteHandler godoc
//
//	@Summary		Create Favourites
//	@Description	The endpoint takes a FavouriteJSON requests and creates a new favourite
//	@Tags			Favourites
//	@Accept			json
//	@Produce		json
//	@Param			FavouriteJSON	body		FavouriteJSON	true	"create favourites request body"
//	@Success		200				{object}	FavouriteJSON
//
//	@Failure		400				{object}	http_helper.errorResponse
//
//	@Router			/fav [post]
func CreateFavouriteHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var (
			fav       FavouriteJSON
			Favourite []FavBodyJSON
		)
		err := json.NewDecoder(r.Body).Decode(&fav)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		for _, bodyJSON := range fav.Fav {
			bodyJSON.DateAdded = time.Now().Format(time.RFC3339)
			Favourite = append(Favourite, bodyJSON)
		}
		fav.Fav = Favourite

		result, err := s.CreateFavourite(r.Context(), fav.ToFavourite())
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		payload := NewJSONFavourite(result)

		http_helper.EncodeResult(w, http_helper.DefaultResponse{Code: http.StatusOK, Success: true, Body: payload}, http.StatusOK)
	}

	return
}

// GetFavouritesHandler godoc
//
//	@Summary		Get Favourites
//	@Description	Retrieves all favourite/liked messages
//	@Tags			Favourites
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	GetFavouritesResponse
//	@Failure		400	{object}	http_helper.errorResponse
//	@Router			/favourites [get]
func GetFavouritesHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		result, err := s.GetFavourites(r.Context())
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}

	return
}

// DeleteFavouritesHandler godoc
//
//	@Summary		Delete favourite by ID
//	@Description	The endpoint takes nothing as the request body and deletes the favourite by ID
//	@Tags			Favourites
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	uuid.UUID
//
//	@Param			message_id	path		string	true	"message id"
//	@Failure		400			{object}	http_helper.errorResponse
//
//	@Router			/delete/{series_id} [delete]
func DeleteFavouritesHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		messageIdParam := chi.URLParam(r, "message_id")
		result, err := s.DeleteFavourite(r.Context(), messageIdParam)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		http_helper.EncodeResult(w, result, http.StatusOK)
	}
	return
}

func UpdateUserProfileHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var userJSON UserJSON
		err := json.NewDecoder(r.Body).Decode(&userJSON)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		user := userJSON.ToUser()
		result, err := s.UpdateUserProfile(r.Context(), user)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}
	return
}

func BuildDeviceHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var devices DeviceManager
		email := chi.URLParam(r, "email")

		err := json.NewDecoder(r.Body).Decode(&devices)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		result, err := s.BuildDevice(r.Context(), &devices, email)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}
	return
}

func GetDevicesHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		result, err := s.GetDevices(r.Context())
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}
	return
}

func DeleteDeviceHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		identifier := chi.URLParam(r, "identifier")
		result, err := s.DeleteDevice(r.Context(), identifier)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}
	return
}

func UpdateDeviceHandler(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		identifier := chi.URLParam(r, "identifier")
		status := chi.URLParam(r, "status")
		result, err := s.UpdateDevice(r.Context(), status, identifier)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}
	return
}

func UpdateAppVersion(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var appVersion VersionManager
		err := json.NewDecoder(r.Body).Decode(&appVersion)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		result, err := s.UpdateAppVersion(r.Context(), appVersion)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}
	return
}

func GetAppVersion(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		versionIDParam := chi.URLParam(r, "version_id")
		result, err := s.GetAppVersion(r.Context(), versionIDParam)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, result, http.StatusOK)
	}

	return
}

func SendEmailVerificationLink(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		var user UserJSON

		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		err = s.SendEmailVerificationLink(r.Context(), user.Email)

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, http_helper.DefaultResponse{
			Code:    http.StatusOK,
			Success: true,
			Body:    "Verification link sent",
		}, http.StatusOK)
	}
	return
}

func VerifyEmail(s *UserService) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		t := mailer.Template{}
		data := mailer.Message{
			DataMap: map[string]string{
				"HofRoundLogo": fmt.Sprintf("%s/HoF_Logo_White.png", s.config.AwsConfiguration.BucketPath),
				"ThisIsHome1":  fmt.Sprintf("%s/home1.jpg", s.config.AwsConfiguration.BucketPath),
			},
		}

		err := s.VerifyEmail(r.Context())

		if err != nil {
			_ = t.Create(w, "verify_email_error.page.tmpl", data)
			return
		}

		_ = t.Create(w, "verify_email_success.page.tmpl", data)
	}
	return
}
