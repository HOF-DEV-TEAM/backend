package favourite

import (
	"database/sql"
	"github.com/gofrs/uuid"
)

type Favourite struct {
	ID        uuid.UUID      `sql:"id"`
	UserID    uuid.UUID      `sql:"user_id" validate:"required"`
	MessageID uuid.UUID      `sql:"message_id" validate:"required"`
	SeriesID  sql.NullString `sql:"series_id"`
	Fav       bool           `sql:"fav"`
	DateAdded sql.NullString `sql:"date_added"`
	DeletedAt sql.NullString `sql:"deleted_at"`
}

type FavMessage struct {
	ID          uuid.UUID `sql:"id"`
	UserID      uuid.UUID `sql:"user_id" validate:"required"`
	Fav         bool      `sql:"fav"`
	MessageID   uuid.UUID `sql:"message_id" validate:"required"`
	SeriesID    uuid.UUID `sql:"series_id"`
	Title       string    `sql:"title" validate:"required"`
	Author      string    `sql:"author" validate:"required"`
	ImageUrl    string    `sql:"image_url"`
	AudioUrl    string    `sql:"audio_url" validate:"required"`
	Description string    `sql:"description"`
}
