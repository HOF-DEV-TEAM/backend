package user

import (
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"net/http"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
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
	IsVerified  IsVerifiedEnum `json:"is_verified"`
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

type FavouriteJSON struct {
	ID     uuid.UUID     `json:"id,omitempty"`
	UserID uuid.UUID     `json:"user_id"`
	Fav    []FavBodyJSON `json:"fav"`
} // @name FavouriteJSON

type FavBodyJSON struct {
	MessageID uuid.UUID `json:"message_id"`
	SeriesID  string    `json:"series_id"`
	Fav       bool      `json:"fav"`
	DateAdded string    `json:"date_added"`
	DeletedAt string    `json:"deleted_at"`
} // @name FavBodyJSON

type FavMessageJSON struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Fav         bool      `json:"fav"`
	MessageID   uuid.UUID `json:"message_id"`
	SeriesID    uuid.UUID `json:"series_id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	ImageUrl    string    `json:"image_url"`
	AudioUrl    string    `json:"audio_url"`
	Description string    `json:"description"`
} // @name FavMessageJSON

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
	}

	return result
}

func (u *User) ToJSON() *UserJSON {
	return &UserJSON{
		ID:         u.ID,
		Email:      u.Email,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Address:    u.Address,
		Mobile:     u.Mobile.String,
		Gender:     u.Gender,
		Username:   u.UserName,
		IsVerified: u.IsVerified,
	}
}

// GetUserHandler godoc
//
//	@Summary		Sign up a new user
//	@Description	Creates a new user with the input payload
//	@Tags			Sessions
//	@Accept			json
//	@Produce		json
//	@Param			SignUpUserRequestJSON	body		SignUpUserRequestJSON	true	"Create user"
//	@Success		200						{object}	UserJSON
//	@Router			/session/sign_up [post]
func GetUserHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getUserHandler, svc)
}

func getUserHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
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
	payload := result.ToJSON()

	http_helper.EncodeResult(w, payload, http.StatusOK)
}

// ForgotPasswordHandler godoc
//
//	@Summary		User forgets their password
//	@Description	User can request for a password change with the input payload
//	@Tags			Password
//	@Accept			json
//	@Produce		json
//	@Param			ForgotPasswordPayload	body		ForgotPasswordPayload	true	"Forgot password"
//	@Success		200						{object}	OTPResponse
//	@Router			/session/forgot_password [post]
func ForgotPasswordHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request ForgotPasswordPayload
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		otpResponse, err := svc.ForgotPassword(request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}

		http_helper.EncodeResult(w, otpResponse, http.StatusOK)
	}
}

// VerifyPasswordResetOTPHandler godoc
//
//	@Summary		Verify password reset OTP
//	@Description	The endpoint verifies the OTP input from the user
//	@Tags			Password
//	@Accept			json
//	@Produce		json
//	@Param			VerifyOTP	body		VerifyOTP	true	"Verify OTP"
//	@Success		200						{object}	UserSession
//	@Router			/session/verify_token [post]
func VerifyPasswordResetOTPHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request VerifyOTP
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		result, err := svc.VerifyPasswordResetOTP(r.Context(), &request)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		payload := UserSession{
			Token: result.Token,
		}

		http_helper.EncodeResult(w, payload, http.StatusOK)
	}
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
func ResetPasswordHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resetPasswordRequest ResetPasswordPayload
		err := json.NewDecoder(r.Body).Decode(&resetPasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		_, err = svc.ResetPassword(r.Context(), resetPasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return

		}

		http_helper.EncodeResult(w, http_helper.DefaultResponse{Code: http.StatusOK, Success: true}, http.StatusOK)
	}
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
func ChangePasswordHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var changePasswordRequest ChangePasswordPayload
		err := json.NewDecoder(r.Body).Decode(&changePasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		_, err = svc.ChangePassword(r.Context(), changePasswordRequest)
		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return

		}

		http_helper.EncodeResult(w, http_helper.DefaultResponse{Code: http.StatusOK, Success: true}, http.StatusOK)
	}
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
		ID:          fav.ID,
		Fav:         fav.Fav,
		UserID:      fav.UserID,
		MessageID:   fav.MessageID,
		SeriesID:    fav.SeriesID,
		Title:       fav.Title,
		Author:      fav.Author,
		ImageUrl:    fav.ImageUrl,
		AudioUrl:    fav.AudioUrl,
		Description: fav.Description,
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
//	@Success		200	{object}	FavouriteJSON
//
// @Failure 400 {object} http_helper.errorResponse
//
//	@Router			/fav [post]
func CreateFavouriteHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createFavouriteHandler, svc)
}
func createFavouriteHandler(w http.ResponseWriter, r *http.Request, s interface{}) {
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

	result, err := s.(Service).CreateFavourite(r.Context(), fav.ToFavourite())
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	payload := NewJSONFavourite(result)

	http_helper.EncodeResult(w, http_helper.DefaultResponse{Code: http.StatusOK, Success: true, Body: payload}, http.StatusOK)
}

// GetFavouritesHandler godoc
//
//	@Summary		Get Favourites
//	@Description	Retrieves all favourite/liked messages
//	@Tags			Favourites
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	GetFavouritesResponse
//	@Failure		400			{object}	http_helper.errorResponse
//	@Router			/favourites [get]
func GetFavouritesHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getFavouritesHandler, svc)
}
func getFavouritesHandler(w http.ResponseWriter, r *http.Request, s interface{}) {

	result, err := s.(Service).GetFavourites(r.Context())
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	http_helper.EncodeResult(w, result, http.StatusOK)
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
func DeleteFavouritesHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(deleteFavouritesHandler, svc)
}
func deleteFavouritesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	messageIdParam := chi.URLParam(r, "message_id")
	result, err := svc.(Service).DeleteFavourite(r.Context(), messageIdParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}

func UpdateUserProfileHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(updateUserProfileHandler, svc)
}
func updateUserProfileHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	var userJSON UserJSON
	err := json.NewDecoder(r.Body).Decode(&userJSON)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	user := userJSON.ToUser()
	result, err := svc.(Service).UpdateUserProfile(r.Context(), user)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	http_helper.EncodeResult(w, result, http.StatusOK)
}
