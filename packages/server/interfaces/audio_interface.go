package interfaces

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/hofng/hofApp/pkg/audio_message"
)

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
	var audioMessage audio_message.AudioMessageJSON
	err := json.NewDecoder(r.Body).Decode(&audioMessage)

	if err != nil {
		encodeResult(w, err, http.StatusInternalServerError)
		return
	}

	result, err := svc.(audio_message.Service).CreateAudioMessage(r.Context(), audioMessage.ToAudioMessage())

	if err != nil {
		EncodeJSONError(r.Context(), err, w)
		return
	}
	encodeResult(w, audio_message.NewJSONAudioMessage(result), http.StatusOK)
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
	var audioSeries audio_message.AudioSeriesJSON
	err := json.NewDecoder(r.Body).Decode(&audioSeries)

	if err != nil {
		encodeResult(w, err, http.StatusInternalServerError)
		return
	}

	result, err := svc.(audio_message.Service).CreateAudioSeries(r.Context(), audioSeries.ToAudioSeries())

	if err != nil {
		EncodeJSONError(r.Context(), err, w)
		return
	}
	encodeResult(w, audio_message.NewJSONAudioSeries(result), http.StatusOK)
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

	//TODO: Make this a robust search type
	seriesId := r.URL.Query().Get("series_id")

	result, err := svc.(audio_message.Service).GetAudioMessages(r.Context(), seriesId)

	if err != nil {
		EncodeJSONError(r.Context(), err, w)
		return
	}
	encodeResult(w, result, http.StatusOK)
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

	result, err := svc.(audio_message.Service).GetAudioSeries(r.Context())

	if err != nil {
		EncodeJSONError(r.Context(), err, w)
		return
	}
	encodeResult(w, result, http.StatusOK)
}
