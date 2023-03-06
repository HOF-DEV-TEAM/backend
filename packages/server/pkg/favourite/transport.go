package favourite

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"net/http"
	"time"
)

func NullSQL(s sql.NullString) string {
	if !s.Valid {
		return ""
	}

	return s.String
}

type FavouriteJSON struct {
	ID        uuid.UUID `json:"id,omitempty"`
	UserID    uuid.UUID `json:"user_id"`
	MessageID uuid.UUID `json:"message_id"`
	SeriesID  string    `json:"series_id"`
	Fav       bool      `json:"fav"`
	DateAdded string    `json:"date_added"`
	DeletedAt string    `json:"deleted_at"`
} // @name FavouriteJSON

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

func (fav *FavouriteJSON) ToFavourite() *Favourite {

	result := &Favourite{
		UserID:    fav.UserID,
		MessageID: fav.MessageID,
	}
	if fav.Fav {
		result.Fav = fav.Fav
	}
	if fav.SeriesID != "" {
		result.SeriesID = sql.NullString{Valid: true, String: fav.SeriesID}
	}
	if fav.DateAdded != "" {
		result.DateAdded = sql.NullString{Valid: true, String: fav.DateAdded}
	}
	if fav.DeletedAt != "" {
		result.DateAdded = sql.NullString{Valid: true, String: fav.DeletedAt}
	}
	return result
}

func NewJSONFavourite(fav *Favourite) *FavouriteJSON {
	return &FavouriteJSON{
		ID:        fav.ID,
		UserID:    fav.UserID,
		MessageID: fav.MessageID,
		SeriesID:  fav.SeriesID.String,
		Fav:       fav.Fav,
		DateAdded: fav.DateAdded.String,
		DeletedAt: fav.DeletedAt.String,
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
	var fav FavouriteJSON

	err := json.NewDecoder(r.Body).Decode(&fav)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	fav.DateAdded = time.Now().Format(time.RFC3339)
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

// DeleteAudioSeriesByIDHandler godoc
//
//	@Summary		Delete Audio Series by ID
//	@Description	The endpoint takes nothing as the request body and update the audio series by ID
//	@Tags			Audio Series
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	uuid.UUID
//
//	@Param			series_id	path		string	true	"audio series id"
//	@Failure		400			{object}	http_helper.errorResponse
//
//	@Router			/delete/{series_id} [delete]
func DeleteFavouritesHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(deleteFavouritesHandler, svc)
}

func deleteFavouritesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	favIdParam := chi.URLParam(r, "fav_id")
	result, err := svc.(Service).DeleteFavourite(r.Context(), favIdParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}
