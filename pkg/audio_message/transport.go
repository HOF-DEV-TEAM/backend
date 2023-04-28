package audio_message

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/urlqueryhelper"
	"github.com/go-chi/chi/v5"
)

type AudioMessageJSON struct {
	ID           string `json:"id,omitempty"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	ImageUrl     string `json:"image_url"`
	AudioUrl     string `json:"audio_url,omitempty"`
	SeriesID     string `json:"series_id"`
	Description  string `json:"description"`
	DateAdded    string `json:"date_added,omitempty"`
	LastUpdated  string `json:"last_updated,omitempty"`
	DateReleased string `json:"date_released"`
} //	@name	AudioMessageJSON

type AudioSeriesJSON struct {
	ID           string `json:"id,omitempty"` //	@Param	series_id
	Title        string `json:"title"`
	Author       string `json:"author"`
	ImageUrl     string `json:"image_url"`
	Description  string `json:"description"`
	DateAdded    string `json:"date_added,omitempty"`
	LastUpdated  string `json:"last_updated,omitempty"`
	DateReleased string `json:"date_released"`
	OfTheMonth   *bool  `json:"of_the_month,omitempty"`
} //	@name	AudioSeriesJSON

type MeditationJSON struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Status    string `json:"status"`
	DateAdded string `json:"date_added"`
	DeletedAt string `json:"deleted_at"`
}

type HomepageJSON struct {
	AudioSeries []*AudioSeriesJSON `json:"audio_series"`
	Meditation  []*MeditationJSON  `json:"meditation"`
}

type PageResponse struct {
	TotalResults int32 `json:"totalResults"`
} //	@name	PageResponse

type GetAudiosSeriesResponse struct {
	AudioSeries []*AudioSeriesJSON `json:"audio_series"`
	Pagination  PageResponse       `json:"pagination"`
} //	@name	GetAudiosSeriesResponse

type GetAudiosMessagesResponse struct {
	AudioMessages []*AudioMessageJSON `json:"audio_messages"`
	Pagination    PageResponse        `json:"pagination"`
} //	@name	GetAudiosMessagesResponse

func (am *AudioMessageJSON) ToAudioMessage() *AudioMessage {
	result := &AudioMessage{
		Title:       am.Title,
		Author:      am.Author,
		ImageUrl:    am.ImageUrl,
		AudioUrl:    am.AudioUrl,
		Description: am.Description,
	}

	if am.SeriesID != "" {
		result.SeriesID = sql.NullString{Valid: true, String: am.SeriesID}
	}
	if am.DateReleased != "" {
		result.DateReleased = sql.NullString{Valid: true, String: am.DateReleased}
	}

	return result
}

func (audioSeries *AudioSeriesJSON) ToAudioSeries() *AudioSeries {
	result := &AudioSeries{
		Title:       audioSeries.Title,
		Author:      audioSeries.Author,
		ImageUrl:    audioSeries.ImageUrl,
		Description: audioSeries.Description,
		OfTheMonth:  audioSeries.OfTheMonth,
	}
	if audioSeries.DateReleased != "" {
		result.DateReleased = sql.NullString{Valid: true, String: audioSeries.DateReleased}
	}

	return result
}

func ToMeditations(meditations []*MeditationJSON) []*Meditation {
	var results []*Meditation
	for _, meditation := range meditations {
		result := Meditation{
			Name:   meditation.Name,
			Image:  meditation.Image,
			Status: meditation.Status,
		}
		if meditation.DateAdded != "" {
			result.DateAdded = sql.NullString{
				Valid:  true,
				String: meditation.DateAdded,
			}
		}

		results = append(results, &result)
	}

	return results
}

func (meditation *MeditationJSON) ToMeditation() *Meditation {
	result := &Meditation{
		Name:   meditation.Name,
		Image:  meditation.Image,
		Status: meditation.Status,
	}
	return result
}

func (homepage *HomepageJSON) ToHomePage() *Homepage {
	var (
		audioSeries    AudioSeries
		meditation     Meditation
		audioSeriesAll []*AudioSeries
		meditations    []*Meditation
	)

	for _, series := range homepage.AudioSeries {
		audioSeries = AudioSeries{
			ID:          series.ID,
			Title:       series.Title,
			Author:      series.Author,
			ImageUrl:    series.ImageUrl,
			Description: series.Description,
			OfTheMonth:  series.OfTheMonth,
		}

		if series.DateAdded != "" {
			audioSeries.DateAdded = sql.NullString{
				Valid:  true,
				String: series.DateAdded,
			}
		}
		if series.DateReleased != "" {
			audioSeries.DateReleased = sql.NullString{
				Valid:  true,
				String: series.DateReleased,
			}
		}

		audioSeriesAll = append(audioSeriesAll, &audioSeries)
	}

	for _, m := range homepage.Meditation {
		meditation = Meditation{
			ID:     m.ID,
			Name:   m.Name,
			Image:  m.Image,
			Status: m.Status,
		}
		if m.DateAdded != "" {
			meditation.DateAdded = sql.NullString{
				Valid:  true,
				String: m.DateAdded,
			}
		}

		meditations = append(meditations, &meditation)
	}

	result := Homepage{
		AudioSeries: audioSeriesAll,
		Meditation:  meditations,
	}

	return &result
}

func NewJSONAudioMessage(audioMessage *AudioMessage) *AudioMessageJSON {
	return &AudioMessageJSON{
		ID:           audioMessage.ID,
		Title:        audioMessage.Title,
		Author:       audioMessage.Author,
		ImageUrl:     audioMessage.ImageUrl,
		AudioUrl:     audioMessage.AudioUrl,
		SeriesID:     audioMessage.SeriesID.String,
		Description:  audioMessage.Description,
		DateReleased: audioMessage.DateReleased.String,
	}
}

func NewJSONAudioSeries(audioSeries *AudioSeries) *AudioSeriesJSON {
	return &AudioSeriesJSON{
		ID:           audioSeries.ID,
		Title:        audioSeries.Title,
		Author:       audioSeries.Author,
		ImageUrl:     audioSeries.ImageUrl,
		Description:  audioSeries.Description,
		DateReleased: audioSeries.DateReleased.String,
		OfTheMonth:   audioSeries.OfTheMonth,
	}
}

func NewJSONMeditation(m *Meditation) *MeditationJSON {
	result := MeditationJSON{
		ID:        m.ID,
		Name:      m.Name,
		Image:     m.Image,
		Status:    m.Status,
		DateAdded: m.DateAdded.String,
	}

	return &result
}

func NewJSONHomePage(homepage *Homepage) *HomepageJSON {
	var (
		audioSeriesJSON []*AudioSeriesJSON
		meditationJSON  []*MeditationJSON
	)

	for _, series := range homepage.AudioSeries {
		var audioSeries AudioSeriesJSON
		audioSeries = AudioSeriesJSON{
			ID:           series.ID,
			Title:        series.Title,
			Author:       series.Author,
			ImageUrl:     series.ImageUrl,
			Description:  series.Description,
			DateReleased: series.DateReleased.String,
			DateAdded:    series.DateAdded.String,
			OfTheMonth:   series.OfTheMonth,
		}

		audioSeriesJSON = append(audioSeriesJSON, &audioSeries)
	}

	for _, m := range homepage.Meditation {
		var meditation MeditationJSON

		meditation = MeditationJSON{
			ID:        m.ID,
			Name:      m.Name,
			Image:     m.Image,
			Status:    m.Status,
			DateAdded: m.DateAdded.String,
		}

		meditationJSON = append(meditationJSON, &meditation)
	}

	return &HomepageJSON{
		AudioSeries: audioSeriesJSON,
		Meditation:  meditationJSON,
	}
}

// CreateAudioMessageHandler godoc
//
//	@Summary		Create Audio Message
//	@Description	The endpoint takes an AudioMessageJSON requests and creates a new audio message
//	@Tags			Audio Message
//	@Accept			json
//	@Produce		json
//	@Param			AudioMessageJSON	body		AudioMessageJSON	true	"create audio message request body"
//	@Success		200					{object}	AudioMessageJSON
//
//	@Failure		400					{object}	http_helper.errorResponse
//
//	@Router			/audio_message [post]
func CreateAudioMessageHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createAudioMessageHandler, svc)
}

func createAudioMessageHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	var audioMessage AudioMessageJSON
	err := json.NewDecoder(r.Body).Decode(&audioMessage)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	result, err := svc.(Service).CreateAudioMessage(r.Context(), audioMessage.ToAudioMessage())

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	payload := NewJSONAudioMessage(result)

	http_helper.EncodeResult(w, payload, http.StatusOK)
}

// CreateAudioSeriesHandler godoc
//
//	@Summary		Admin can create new audio series
//	@Description	Admin will be able to create/insert new audio series with the input payload
//	@Tags			Audio Series
//	@Accept			json
//	@Produce		json
//	@Param			AudioSeriesJSON	body		AudioSeriesJSON	true	"Create audio series request body"
//	@Success		200				{object}	AudioSeriesJSON
//
//	@Failure		400				{object}	http_helper.errorResponse
//
//	@Router			/audio_series [post]

func CreateAudioSeriesHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createAudioSeriesHandler, svc)
}

func createAudioSeriesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	var audioSeries AudioSeriesJSON
	err := json.NewDecoder(r.Body).Decode(&audioSeries)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	result, err := svc.(Service).CreateAudioSeries(r.Context(), audioSeries.ToAudioSeries())

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	payload := NewJSONAudioSeries(result)
	http_helper.EncodeResult(w, payload, http.StatusOK)
}

// GetAudioMessagesHandler godoc
//
//	@Summary		Get Audio Message
//	@Description	Retrieves an audio message
//	@Tags			Audio Message
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	GetAudiosMessagesResponse
//
//	@Failure		400			{object}	http_helper.errorResponse
//
//	@Param			series_id	path		string	false	"search message by series id => returns all messages if value is * i.e series_id=* or omitted, returns non-series messages if value is ? i.e series_id=?"
//	@Router			/audio_message [get]

func GetAudioMessagesHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getAudioMessagesHandler, svc)
}

func getAudioMessagesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	var search Filter
	urlqueryhelper.Bind(&search, r)

	result, err := svc.(Service).GetAudioMessages(r.Context(), &search)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	http_helper.EncodeResult(w, result, http.StatusOK)
}

// GetAudioSeriesHandler godoc
//
//	@Summary		Get an audio series
//	@Description	Retrieve an audio series
//	@Tags			Audio Series
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	GetAudiosSeriesResponse
//
//	@Failure		400	{object}	http_helper.errorResponse
//
//	@Router			/audio_series [get]

func GetAudioSeriesHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getAudioSeriesHandler, svc)
}

func getAudioSeriesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {

	result, err := svc.(Service).GetAudioSeries(r.Context())

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}

// GetAudioMessageByIDHandler godoc
//
//	@Summary		Get Audio Message
//	@Description	Get Audio Message By ID
//	@Tags			Audio Message
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	AudioMessageJSON
//	@Success		200			{object}	AudioMessageJSON
//
//	@Failure		400			{object}	http_helper.errorResponse
//	@Param			message_id	path		string	true	"audio message id"
//
//	@Router			/audio_message/id/message/{message_id} [get]

func GetAudioMessageByIDHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getAudioMessageByIDHandler, svc)
}

func getAudioMessageByIDHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	messageIdParam := chi.URLParam(r, "message_id")
	result, err := svc.(Service).GetAudioMessageByID(r.Context(), messageIdParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}

// GetAudioSeriesByIDHandler godoc
//
//	@Summary		Get an audio series
//	@Description	Get Audio Series By ID
//	@Tags			Audio Series
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	AudioSeriesJSON
//
//	@Failure		400			{object}	http_helper.errorResponse
//	@Param			series_id	path		string	true	"audio series id"
//
//	@Router			/audio_series/id/series/{series_id} [get]

func GetAudioSeriesByIDHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getAudioSeriesByIDHandler, svc)
}

func getAudioSeriesByIDHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	seriesIdParam := chi.URLParam(r, "series_id")
	result, err := svc.(Service).GetAudioSeriesByID(r.Context(), seriesIdParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}

// UpdateAudioMessagesByIDHandler godoc
//
//	@Summary		Update Audio Message by ID
//	@Description	The endpoint takes an AudioMessageJSON requests and update the audio message
//	@Tags			Audio Message
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	uuid.UUID
//
//	@Failure		400			{object}	http_helper.errorResponse
//	@Param			message_id	path		string	true	"audio message id"
//
//	@Router			/update/{message_id} [put]
func UpdateAudioMessagesByIDHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(updateAudioMessagesByIDHandler, svc)
}

func updateAudioMessagesByIDHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	messageIdParam := chi.URLParam(r, "message_id")
	var messageJSON AudioMessageJSON
	err := json.NewDecoder(r.Body).Decode(&messageJSON)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	message := messageJSON.ToAudioMessage()
	result, err := svc.(Service).UpdateAudioMessagesByID(r.Context(), *message, messageIdParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	http_helper.EncodeResult(w, result, http.StatusOK)
}

// UpdateAudioSeriesByIDHandler godoc
//
//	@Summary		Update Audio Series by ID
//	@Description	The endpoint takes an AudioSeriesJSON requests and update the audio series
//	@Tags			Audio Series
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	uuid.UUID
//
//	@Failure		400			{object}	http_helper.errorResponse
//	@Param			series_id	path		string	true	"audio series id"
//
//	@Router			/update/{series_id} [put]
func UpdateAudioSeriesByIDHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(updateAudioSeriesByIDHandler, svc)
}

func updateAudioSeriesByIDHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	seriesIdParam := chi.URLParam(r, "series_id")
	var seriesJSON AudioSeriesJSON
	err := json.NewDecoder(r.Body).Decode(&seriesJSON)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	series := seriesJSON.ToAudioSeries()
	result, err := svc.(Service).UpdateAudioSeriesByID(r.Context(), *series, seriesIdParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	http_helper.EncodeResult(w, result, http.StatusOK)
}

// DeleteAudioMessagesByIDHandler godoc
//
//	@Summary		Delete Audio Message by ID
//	@Description	The endpoint takes nothing as the request body and update the audio message by ID
//	@Tags			Audio Message
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	uuid.UUID
//
//	@Failure		400			{object}	http_helper.errorResponse
//	@Param			message_id	path		string	true	"audio message id"
//
//	@Router			/delete/{message_id} [delete]
func DeleteAudioMessagesByIDHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(deleteAudioMessagesByIDHandler, svc)
}

func deleteAudioMessagesByIDHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	messageIdParam := chi.URLParam(r, "message_id")

	result, err := svc.(Service).DeleteAudioMessagesByID(r.Context(), messageIdParam)
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
func DeleteAudioSeriesByIDHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(deleteAudioSeriesByIDHandler, svc)
}

func deleteAudioSeriesByIDHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	seriesIdParam := chi.URLParam(r, "series_id")

	result, err := svc.(Service).DeleteAudioSeriesByID(r.Context(), seriesIdParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}

func HomePageDirectoryHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(homePageDirectoryHandler, svc)
}

func homePageDirectoryHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	result, err := svc.(Service).HomePageDirectory(r.Context())

	home := NewJSONHomePage(result)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, home, http.StatusOK)

}

func CreateMeditationHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createMeditationHandler, svc)
}

func createMeditationHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	var meditation []*MeditationJSON
	err := json.NewDecoder(r.Body).Decode(&meditation)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	result, err := svc.(Service).CreateMeditation(r.Context(), ToMeditations(meditation))

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	http_helper.EncodeResult(w, result, http.StatusOK)
}

func UpdateMeditationByIDHandler(svc Service) http.HandlerFunc {
	return http_helper.NewHTTPHandler(updateMeditationByIDHandler, svc)
}

func updateMeditationByIDHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	medParam := chi.URLParam(r, "meditation_id")
	var meditationJSON MeditationJSON
	err := json.NewDecoder(r.Body).Decode(&meditationJSON)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	med := meditationJSON.ToMeditation()
	result, err := svc.(Service).UpdateMeditationByID(r.Context(), med.Status, medParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}

	http_helper.EncodeResult(w, result, http.StatusOK)
}
