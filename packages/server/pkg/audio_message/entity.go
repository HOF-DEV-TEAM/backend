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
}
