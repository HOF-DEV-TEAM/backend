package audio_message

import "database/sql"

type SubscriptionProvider struct {
	ID          string         `sql:"id"`
	Name        string         `sql:"name" validate:"required"`
	IsDefault   int            `sql:"author"`
	DateAdded   sql.NullString `sql:"date_added"`
	LastUpdated sql.NullString `sql:"last_updated"`
	DeletedAt   sql.NullString `sql:"deleted_at"`
}

type SubscriptionPlan struct {
	ID                     string         `sql:"id"`
	Name                   string         `sql:"name" validate:"required"`
	Type                   int            `sql:"int"`
	Freq                   float64        `sql:"freq"`
	Fee                    int            `sql:"int"`
	Status                 int            `sql:"freq"`
	DateAdded              sql.NullString `sql:"date_added"`
	LastUpdated            sql.NullString `sql:"last_updated"`
	SubscritpionProviderID sql.NullString `sql:"subscription_provider_id"`
	DeletedAt              sql.NullString `sql:"deleted_at"`
}

type Subscription struct {
	ID                 string         `sql:"id"`
	Status             string         `sql:"status"`
	UserID             string         `sql:"user_id" validate:"required"`
	SubscriptionPlanID string         `sql:"subscription_plan_id" validate:"required"`
	DateAdded          sql.NullString `sql:"date_added"`
	LastUpdated        sql.NullString `sql:"last_updated"`
	DeletedAt          sql.NullString `sql:"deleted_at"`
}
