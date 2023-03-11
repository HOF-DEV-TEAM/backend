package favourite

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"net/http"
	"time"
)

type FavouriteJSON struct {
	ID     uuid.UUID     `json:"id,omitempty"`
	UserID uuid.UUID     `json:"user_id"`
	Fav    []FavBodyJSON `json:"fav"`
} // @name FavouriteJSON

//type FavsJSON struct {
//	Favourite []FavBodyJSON `json:"favourite"`
//} // @name FavsJSON
//

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
//	@Param			fav_id	path		string	true	"favourite id"
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
