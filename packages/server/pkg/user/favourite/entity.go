package favourite

import (
	"github.com/gofrs/uuid"
)

type Favourites struct {
	ID     uuid.UUID `sql:"id"`
	UserID uuid.UUID `sql:"user_id" validate:"required"`
	Fav    []FavBody `sql:"fav" json:"fav"`
}

type FavBody struct {
	MessageID uuid.UUID `sql:"message_id" validate:"required" json:"message_id"`
	SeriesID  string    `sql:"series_id" json:"series_id"`
	Fav       bool      `sql:"fav" json:"fav"`
	DateAdded string    `sql:"date_added" json:"date_added"`
	DeletedAt string    `sql:"deleted_at" json:"deleted_at"`
}

type FavMessage struct {
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
}
