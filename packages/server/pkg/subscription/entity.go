package subscription

import (
	"database/sql"
	"errors"	
)

type SubscriptionProvider struct {
	ID          string         `sql:"id"`
	Name        string         `sql:"name" validate:"required"`
	IsDefault   int            `sql:"author"`
	DateAdded   sql.NullString `sql:"date_added"`
	LastUpdated sql.NullString `sql:"last_updated"`
	DeletedAt   sql.NullString `sql:"deleted_at"`
}

type StatusEnum uint16

const (
	Inactive = StatusEnum(0)
	Active   = StatusEnum(1)
	Archived = StatusEnum(2)
)

func (e StatusEnum) String() string {
	switch e {
	case Inactive:
		return "0"
	case Active:
		return "1"
	case Archived:
		return "2"
	default:
		return "invalid"
	}
}

// MarshalText interface implementation StatusEnum into text.
func (e StatusEnum) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

// UnMarshalText interface implementation StatusEnum into text.
func (e *StatusEnum) UnMarshalText(from []byte) error {
	switch string(from) {
	case "0":
		*e = Inactive
	case "1":
		*e = Active
	case "2":
		*e = Archived
	default:
		return errors.New("invalid StatusEnum")
	}
	return nil
}

type FreqEnum uint16

const (
	Hourly    = FreqEnum(0)
	Daily     = FreqEnum(1)
	Weekly    = FreqEnum(2)
	Monthly   = FreqEnum(3)
	Quarterly = FreqEnum(4)
	Yearly    = FreqEnum(5)
)

func (e FreqEnum) String() string {
	switch e {
	case Hourly:
		return "hourly"
	case Daily:
		return "daily"
	case Weekly:
		return "weekly"
	case Monthly:
		return "monthly"
	case Quarterly:
		return "quarterly"
	case Yearly:
		return "yearly"
		
	default:
		return "invalid"
	}
}

// MarshalText interface implementation FreqEnum into text.
func (e FreqEnum) MarshalText() ([]byte, error) {	
	return []byte(e.String()), nil
}

// UnMarshalText interface implementation FreqEnum into text.
func (e *FreqEnum) UnMarshalText(from []byte) error {	
	switch string(from) {
	case "hourly":
		*e = Hourly
	case "daily":
		*e = Daily
	case "weekly":
		*e = Weekly
	case "monthly":		
		*e = Monthly
	case "quarterly":
		*e = Quarterly
	case "yearly", "annually":
		*e = Yearly
	default:
		return errors.New("invalid FreqEnum")
	}
	return nil
}

type SubscriptionPlan struct {
	ID                     string         `sql:"id"`
	Name                   string         `sql:"name" validate:"required"`
	Type                   int            `sql:"int"`
	Freq                   FreqEnum       `sql:"freq"`
	Fee                    float64        `sql:"float64"`
	Status                 StatusEnum     `sql:"status"`
	Currency               string         `sql:"currency"`
	Code                   string         `sql:"code"`
	DateAdded              sql.NullString `sql:"date_added"`
	LastUpdated            sql.NullString `sql:"last_updated"`
	PlanId                 sql.NullString `sql:"plan_id"`
	SubscritpionProviderID sql.NullString `sql:"subscription_provider_id"`
	DeletedAt              sql.NullString `sql:"deleted_at"`
}

type SubscriptionOffering struct {
	ID                     string         `sql:"id"`
	Name                   string         `sql:"name" validate:"required"`
	Status                 int            `sql:"status"`
	DateAdded              sql.NullString `sql:"date_added"`
	LastUpdated            sql.NullString `sql:"last_updated"`
	SubscritpionProviderID sql.NullString `sql:"subscription_provider_id"`
	DeletedAt              sql.NullString `sql:"deleted_at"`
}

type SubscriptionPlanOffering struct {
	ID                     string         `sql:"id"`
	Status                 int            `sql:"freq"`
	DateAdded              sql.NullString `sql:"date_added"`
	LastUpdated            sql.NullString `sql:"last_updated"`
	SubscritpionProviderID sql.NullString `sql:"subscription_provider_id"`
	SubscritpionOfferingID sql.NullString `sql:"subscription_offering_id"`
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
