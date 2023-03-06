package favourite

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"database/sql"
	"encoding/json"
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

// CreateFavouriteHandler godoc
//
//	@Summary		Create Audio Message
//	@Description	The endpoint takes an AudioMessageJSON requests and creates a new audio message
//	@Tags			Favourites
//	@Accept			json
//	@Produce		json
//	@Param			FavouriteJSON	body		FavouriteJSON	true	"create favourites request body"
//	@Success		200	{object}	http_helper.DefaultResponse{body=FavouriteJSON}
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
	fav.DateAdded = time.Now().Format(time.RFC3339)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	result, err := s.(Service).CreateFavourite(r.Context(), fav.ToFavourite())

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	payload := NewJSONFavourite(result)

	http_helper.EncodeResult(w, http_helper.DefaultResponse{Code: http.StatusOK, Success: true, Body: payload}, http.StatusOK)

}
