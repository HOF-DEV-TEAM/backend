package interfaces

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/hofng/hofApp/pkg/audio_message"
)


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


func GetAudioSeriesHandler(w http.ResponseWriter, r *http.Request, svc interface{}) {

	result, err := svc.(audio_message.Service).GetAudioSeries(r.Context())

	if err != nil {
		EncodeJSONError(r.Context(), err, w)
		return
	}
	encodeResult(w, result, http.StatusOK)
}