package audio_message

import "database/sql"

type AudioMessage struct {
	ID          int            `sql:"id"`
	Title       string         `sql:"title" validate:"required"`
	Author      string         `sql:"author" validate:"required"`
	DateAdded   sql.NullString `sql:"date_added"`
	LastUpdated sql.NullString `sql:"last_updated"`
	ImageUrl    string         `sql:"image_url"`
	AudioUrl    string         `sql:"audio_url" validate:"required"`
	SeriesID    int            `sql:"id"`
	Description string         `sql:"description"`
}

type AudioSeries struct {
	ID          int            `sql:"id"`
	Title       string         `sql:"title" validate:"required"`
	Author      string         `sql:"image_url"`
	ImageUrl    string         `sql:"image_url" validate:"required"`
	DateAdded   sql.NullString `sql:"date_added"`
	LastUpdated sql.NullString `sql:"last_updated"`
	Description string         `sql:"description"`
}
