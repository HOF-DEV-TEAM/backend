package audio_message

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/urlqueryhelper"
	"github.com/go-chi/chi"
)

type AudioMessageJSON struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	ImageUrl    string `json:"image_url"`
	AudioUrl    string `json:"audio_url,omitempty"`
	SeriesID    string `json:"series_id"`
	Description string `json:"description"`
	DateAdded   string `json:"date_added,omitempty"`
	LastUpdated string `json:"last_updated,omitempty"`
} // @name AudioMessageJSON

type AudioSeriesJSON struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	ImageUrl    string `json:"image_url"`
	Description string `json:"description"`
	DateAdded   string `json:"date_added,omitempty"`
	LastUpdated string `json:"last_updated,omitempty"`
} // @name AudioSeriesJSON

type PageResponse struct {
	TotalResults int32 `json:"totalResults"`
} // @name PageResponse

type GetAudiosSeriesResponse struct {
	AudioSeries []*AudioSeriesJSON `json:"audio_series"`
	Pagination  PageResponse       `json:"pagination"`
} // @name GetAudiosSeriesResponse

type GetAudiosMessagesResponse struct {
	AudioMessages []*AudioMessageJSON `json:"audio_messages"`
	Pagination    PageResponse        `json:"pagination"`
} // @name GetAudiosMessagesResponse

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

	return result
}

func (audioSeries *AudioSeriesJSON) ToAudioSeries() *AudioSeries {
	result := &AudioSeries{
		Title:       audioSeries.Title,
		Author:      audioSeries.Author,
		ImageUrl:    audioSeries.ImageUrl,
		Description: audioSeries.Description,
	}
	return result
}

func NewJSONAudioMessage(audioMessage *AudioMessage) *AudioMessageJSON {
	return &AudioMessageJSON{
		ID:          audioMessage.ID,
		Title:       audioMessage.Title,
		Author:      audioMessage.Author,
		ImageUrl:    audioMessage.ImageUrl,
		AudioUrl:    audioMessage.AudioUrl,
		SeriesID:    audioMessage.SeriesID.String,
		Description: audioMessage.Description,
	}
}

func NewJSONAudioSeries(audioSeries *AudioSeries) *AudioSeriesJSON {
	return &AudioSeriesJSON{
		ID:          audioSeries.ID,
		Title:       audioSeries.Title,
		Author:      audioSeries.Author,
		ImageUrl:    audioSeries.ImageUrl,
		Description: audioSeries.Description,
	}
}


// CreateAudioMessageHandler godoc
// @Summary Admin can create new audio messages
// @Description Admin will be able to create/insert new audio messages with the input payload
// @Tags CreateAudioMessages
// @Accept  json
// @Produce  json
// @Param AudioMessageJSON body AudioMessageJSON true "Create audio messages"
// @Success 200 {object} AudioMessageJSON
// @Router /audio_message/ [post]
func CreateAudioMessageHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
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
// @Summary Admin can create new audio series
// @Description Admin will be able to create/insert new audio series with the input payload
// @Tags CreateAudioSeries
// @Accept  json
// @Produce  json
// @Param AudioSeriesJSON body AudioSeriesJSON true "Create audio series"
// @Success 200 {object} AudioSeriesJSON
// @Router /audio_series/ [post]
func CreateAudioSeriesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
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
// @Summary Get an audio message
// @Description Users can now retrieve and see an audio message
// @Tags GetAudioMessages
// @Accept  json
// @Produce  json
// @Success 200 {object} GetAudiosMessagesResponse
// @Router /audio_messages/ [get]
func GetAudioMessagesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {	
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
// @Summary Get an audio series
// @Description Users can now retrieve and see an audio series
// @Tags GetAudioSeries
// @Accept  json
// @Produce  json
// @Success 200 {object} GetAudiosSeriesResponse
// @Router /audio_series/ [get]
func GetAudioSeriesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {

	result, err := svc.(Service).GetAudioSeries(r.Context())

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
		return
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}

// GetAudioMessageByIDHandler godoc
// @Summary Get an audio message
// @Description Users can now retrieve and see an audio message
// @Tags GetAnAudioMessage
// @Accept  json
// @Produce  json
// @Success 200 {object} AudioMessageJSON
// @Router /id/{id} [get]
func GetAudioMessageByIDHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	messageIdParam := chi.URLParam(r, "id")
	result, err := svc.(Service).GetAudioMessageByID(r.Context(), messageIdParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}

// GetAudioSeriesByIDHandler godoc
// @Summary Get an audio series
// @Description Users can now retrieve and see an audio series
// @Tags GetAnAudioSeries
// @Accept  json
// @Produce  json
// @Success 200 {object} AudioSeriesJSON
// @Router /series_id/{id} [get]
func GetAudioSeriesByIDHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {
	seriesIdParam := chi.URLParam(r, "id")
	result, err := svc.(Service).GetAudioSeriesByID(r.Context(), seriesIdParam)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, w)
	}
	http_helper.EncodeResult(w, result, http.StatusOK)
}
