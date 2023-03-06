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
