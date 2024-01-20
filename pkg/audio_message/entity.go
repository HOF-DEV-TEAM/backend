package audio_message

import "database/sql"

type AudioMessage struct {
	ID           string         `sql:"id"`
	Title        string         `sql:"title" validate:"required"`
	Author       string         `sql:"author" validate:"required"`
	DateAdded    sql.NullString `sql:"date_added"`
	LastUpdated  sql.NullString `sql:"last_updated"`
	ImageUrl     string         `sql:"image_url"`
	AudioUrl     string         `sql:"audio_url" validate:"required"`
	SeriesID     sql.NullString `sql:"series_id"`
	Description  string         `sql:"description"`
	DeletedAt    sql.NullString `sql:"deleted_at"`
	DateReleased sql.NullString `sql:"date_released"`
	IsFree       *bool          `sql:"is_free"`
}

type AudioSeries struct {
	ID           string         `sql:"id"`
	Title        string         `sql:"title" validate:"required"`
	Author       string         `sql:"author"`
	ImageUrl     string         `sql:"image_url" validate:"required"`
	DateAdded    sql.NullString `sql:"date_added"`
	LastUpdated  sql.NullString `sql:"last_updated"`
	Description  string         `sql:"description"`
	DeletedAt    sql.NullString `sql:"deleted_at"`
	DateReleased sql.NullString `sql:"date_released"`
	OfTheMonth   *bool          `sql:"of_the_month"`
}

type Meditation struct {
	ID        string         `sql:"id" json:"id"`
	Name      string         `sql:"name" json:"name"`
	Image     string         `sql:"image" json:"image"`
	Status    string         `sql:"status" json:"status"`
	DateAdded sql.NullString `sql:"date_added" json:"-"`
	DeletedAt sql.NullString `sql:"deleted_at" json:"-"`
}

type MeditationResponse struct {
	AffectedRows int64 `json:"affected_row" sql:"affected_row"`
}

type Homepage struct {
	AudioSeries []*AudioSeries `sql:"audio_series"`
	Meditation  []*Meditation  `sql:"meditation"`
}

type DefaultResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
} // @name DefaultResponse

//<tr>
//<img src="https://s3.amazonaws.com/goninja/hof/thisIsHome.jpeg"
//" alt="welcome_to_the_faith_factory"
//style="max-width:100%;">
//</tr>
