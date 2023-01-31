package audio_message

type AudioMessageJSON struct {
	ID          int    `json:"id,omitempty"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	ImageUrl    string `json:"image_url"`
	AudioUrl    string `json:"audio_url,omitempty"`
	SeriesID    int    `json:"series_id"`
	Description string `json:"description"`
	DateAdded   string `json:"date_added,omitempty"`
	LastUpdated string `json:"last_updated,omitempty"`
} // @name AudioMessageJSON

type AudioSeriesJSON struct {
	ID          int    `json:"id,omitempty"`
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

func (audioMessage *AudioMessageJSON) ToAudioMessage() *AudioMessage {
	result := &AudioMessage{
		Title:       audioMessage.Title,
		Author:      audioMessage.Author,
		ImageUrl:    audioMessage.ImageUrl,
		AudioUrl:    audioMessage.AudioUrl,
		SeriesID:    audioMessage.SeriesID,
		Description: audioMessage.Description,
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
		SeriesID:    audioMessage.SeriesID,
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
