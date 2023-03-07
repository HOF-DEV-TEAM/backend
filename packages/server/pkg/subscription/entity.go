package subscription

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
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
	switch strings.ToLower(string(from)) {
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

// MarshalText interface implementation FreqEnum into text.
func (e FreqEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())	
}


// UnMarshalText interface implementation FreqEnum into text.
func (e *FreqEnum) UnmarshalJSON(from []byte) error {	
	switch strings.ToLower(string(from)) {
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

// sql/driver.Valuer interface implementation for TypeEnum
func (e FreqEnum) Value() (driver.Value, error) {
	switch e {
	case Hourly:
		return 0, nil
	case Daily:
		return 1, nil
	case Weekly:
		return 2, nil
	case Monthly:		
		return 3, nil
	case Quarterly:
		return 4, nil
	case Yearly:
		return 5, nil

	default:
		return 3, nil
	}
}

// UnMarshalText interface implementation FreqEnum into text.
// func (e *FreqEnum) Scan(src interface{}) error {
// 	if src == nil {
// 		*e = Monthly
// 		return nil
// 	}
	
// 	buf, ok := src.([]byte)

// 	if !ok {
// 		return errors.New("invalid FreqEnum")		
// 	}
// 	fmt.Println(string(buf), "buffer")
// 	return e.UnMarshalText(buf)
// }


type TypeEnum uint16

const (
	Premium = TypeEnum(1)
	Regular = TypeEnum(0)
)

func (e TypeEnum) String() string {
	switch e {
	case Premium:
		return "Premium"
	case Regular:
		return "Regular"
	default:
		return "invalid"
	}
}

// MarshalText interface implementation TypeEnum into text.
func (e TypeEnum) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

// UnMarshalText interface implementation TypeEnum into text.
func (e *TypeEnum) UnMarshalText(from []byte) error {
	switch strings.ToLower(string(from)) {
	case "premium":
		*e = Premium
	case "regular":
		*e = Regular
	default:
		return errors.New("invalid TypeEnum")
	}
	return nil
}

// MarshalText interface implementation TypeEnum into text.
func (e TypeEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())	
}

// UnMarshalText interface implementation TypeEnum into text.
func (e *TypeEnum) UnmarshalJSON(from []byte) error {
	s := strings.ToLower(string(from))	
	switch s {
	case "premium":		
		*e = Premium
	case "regular":
		*e = Regular
	default:
		return errors.New("invalid TypeEnum")
	}
	return nil
}


// sql/driver.Valuer interface implementation for TypeEnum
func (e TypeEnum) Value() (driver.Value, error) {
	switch e {
	case Premium:
		return 1, nil
	case Regular:
		return 0, nil
	default:
		return 0, nil
	}
}

// UnMarshalText interface implementation TypeEnum into text.
// func (e *TypeEnum) Scan(src interface{}) error {
// 	if src == nil {
// 		*e = Regular
// 		return nil
// 	}
	
// 	buf, ok := src.([]byte)

// 	if !ok {
// 		return errors.New("invalid TypeEnum")		
// 	}
// 	return e.UnMarshalText(buf)
// }

type SubscriptionPlan struct {
	ID                     string         `sql:"id"`
	Name                   string         `sql:"name" validate:"required"`
	Type                   TypeEnum  	  `sql:"int"`
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
	//SubscriptionPlan
	Type   TypeEnum  `sql:"int"`
	Freq   FreqEnum       `sql:"freq"`
	Fee    float64        `sql:"float64"`
	PlanCode string `sql:"code"`

	//SubscriptionOffering
	Name string `sql:"name" validate:"required"`

	ID                     string         `sql:"id"`
	Status                 int            `sql:"freq"`
	DateAdded              sql.NullString `sql:"date_added"`
	LastUpdated            sql.NullString `sql:"last_updated"`
	SubscriptionPlanID     sql.NullString `sql:"subscription_plan_id"`
	SubscriptionOfferingID sql.NullString `sql:"subscription_offering_id"`
	DeletedAt              sql.NullString `sql:"deleted_at"`
}

type Subscription struct {
	ID                 string         `sql:"id"`
	Status             int            `sql:"status"`
	UserID             string         `sql:"user_id" validate:"required"`
	SubscriptionPlanID string         `sql:"subscription_plan_id" validate:"required"`
	NextPaymentDate    sql.NullString `sql:"next_payment_date"`
	DateAdded          sql.NullString `sql:"date_added"`
	LastUpdated        sql.NullString `sql:"last_updated"`
	DeletedAt          sql.NullString `sql:"deleted_at"`
}
