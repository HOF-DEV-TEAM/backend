package content

// CreateMessageRequest is the payload for adding a new audio message.
type CreateMessageRequest struct {
	Title        string  `json:"title"        validate:"required"`
	Author       string  `json:"author"       validate:"required"`
	AudioURL     string  `json:"audio_url"    validate:"required"`
	ImageURL     string  `json:"image_url"`
	Description  string  `json:"description"`
	SeriesID     string  `json:"series_id"`
	DateReleased string  `json:"date_released"`
	IsFree bool `json:"is_free"`
	// Access controls visibility: "leaders", "stewards", "members". Defaults to "members".
	Access string `json:"access"`
	// IsPrivate hides the message from all non-admin users regardless of access level.
	IsPrivate bool `json:"is_private"`
}

// UpdateMessageRequest carries the fields that may be changed on an existing message.
type UpdateMessageRequest struct {
	Title        string `json:"title"`
	Author       string `json:"author"`
	AudioURL     string `json:"audio_url"`
	ImageURL     string `json:"image_url"`
	Description  string `json:"description"`
	SeriesID     string `json:"series_id"`
	DateReleased string `json:"date_released"`
	IsFree *bool `json:"is_free"`
	// Optional access change. Valid values: "leaders", "stewards", "members".
	Access *string `json:"access"`
	// IsPrivate, when non-nil, updates the private visibility flag.
	IsPrivate *bool `json:"is_private"`
}

// CreateSeriesRequest is the payload for adding a new audio series.
type CreateSeriesRequest struct {
	Title        string `json:"title"      validate:"required"`
	Author       string `json:"author"`
	ImageURL     string `json:"image_url"  validate:"required"`
	Description  string `json:"description"`
	DateReleased string `json:"date_released"`
	OfTheMonth   bool   `json:"of_the_month"`
}

// UpdateSeriesRequest carries the fields that may be changed on an existing series.
type UpdateSeriesRequest struct {
	Title        string `json:"title"`
	Author       string `json:"author"`
	ImageURL     string `json:"image_url"`
	Description  string `json:"description"`
	DateReleased string `json:"date_released"`
	OfTheMonth   *bool  `json:"of_the_month"`
}

// CreateMeditationRequest adds a new meditation item.
type CreateMeditationRequest struct {
	Name   string `json:"name"   validate:"required"`
	Image  string `json:"image"`
	Text   string `json:"text"`
	Status string `json:"status"`
}

// UpdateMeditationRequest changes fields on an existing meditation.
type UpdateMeditationRequest struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	Text   string `json:"text"`
	Status string `json:"status"`
}

// MessageListFilter carries query parameters for listing messages.
type MessageListFilter struct {
	Search   string `json:"search"`
	SeriesID string `json:"series_id"`
	Access   string `json:"access"`
	IsFree   *bool  `json:"is_free"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	// IsAdmin, when true, allows private messages to be included in results.
	IsAdmin bool `json:"-"`
}
