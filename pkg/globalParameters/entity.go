package globalParameters

import (
	"database/sql"
	"github.com/gofrs/uuid"
)

type GlobalParameters struct {
	ID                   uuid.UUID      `json:"id" sql:"id"`
	ActivateSubscription bool           `json:"activate_subscription" sql:"activate_subscription"`
	DateCreated          sql.NullString `json:"date_created" sql:"date_created"`
	LastUpdated          sql.NullString `json:"last_updated" sql:"last_updated"`
}

type UpdateGlobalParameters struct {
	ID                   string         `json:"id" sql:"id"`
	ActivateSubscription *bool          `json:"activate_subscription" sql:"activate_subscription"`
	LastUpdated          sql.NullString `json:"last_updated" sql:"last_updated"`
}
