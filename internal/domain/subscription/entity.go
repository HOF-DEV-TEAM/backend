// Package subscription defines the subscription domain entities and value objects.
package subscription

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle state of a subscription.
type Status int

const (
	// StatusInactive marks a subscription that is not active.
	StatusInactive Status = 0
	// StatusActive marks a subscription that is currently active.
	StatusActive Status = 1
	// StatusCanceled marks a subscription that was canceled.
	StatusCanceled Status = 2
	// StatusPastDue marks a subscription that missed payment.
	StatusPastDue Status = 3
)

func (s Status) String() string {
	switch s {
	case StatusActive:
		return "active"
	case StatusCanceled:
		return "canceled"
	case StatusPastDue:
		return "past_due"
	default:
		return "inactive"
	}
}

// BillingFrequency describes how often a plan is billed.
type BillingFrequency int

const (
	// FreqHourly bills once per hour.
	FreqHourly BillingFrequency = 0
	// FreqDaily bills once per day.
	FreqDaily BillingFrequency = 1
	// FreqWeekly bills once per week.
	FreqWeekly BillingFrequency = 2
	// FreqMonthly bills once per month.
	FreqMonthly BillingFrequency = 3
	// FreqQuarterly bills once per quarter.
	FreqQuarterly BillingFrequency = 4
	// FreqYearly bills once per year.
	FreqYearly BillingFrequency = 5
)

// PlanType categorizes the tier of a subscription plan.
type PlanType int

const (
	// PlanTypeRegular is the default subscription tier.
	PlanTypeRegular PlanType = 0
	// PlanTypePremium is the premium subscription tier.
	PlanTypePremium PlanType = 1
)

// Provider is a registered payment provider (e.g. Paystack).
type Provider struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string     `gorm:"type:varchar(200);not null"`
	CreatedAt time.Time  `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:last_updated;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

// TableName returns the database table for providers.
func (Provider) TableName() string { return "subscription_provider" }

// Plan describes a purchasable subscription tier.
type Plan struct {
	ID             uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name           string           `gorm:"type:varchar(200);not null"`
	Type           PlanType         `gorm:"type:int;default:0"`
	Frequency      BillingFrequency `gorm:"column:freq;type:int;default:3"`
	Fee            float64          `gorm:"type:decimal(10,2)"`
	Status         Status           `gorm:"type:int;default:1"`
	Currency       string           `gorm:"type:varchar(10);default:'NGN'"`
	Code           string           `gorm:"type:varchar(200)"`
	ProviderPlanID *uuid.UUID       `gorm:"column:plan_id;type:uuid"`
	ProviderID     *uuid.UUID       `gorm:"column:subscription_provider_id;type:uuid"`
	CreatedAt      time.Time        `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt      time.Time        `gorm:"column:last_updated;autoUpdateTime"`
	DeletedAt      *time.Time       `gorm:"column:deleted_at"`
}

// TableName returns the database table for plans.
func (Plan) TableName() string { return "subscription_plans" }

// Offering is a named feature that can be attached to subscription plans.
type Offering struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name       string     `gorm:"type:varchar(200);not null"`
	Status     Status     `gorm:"type:int;default:1"`
	ProviderID *uuid.UUID `gorm:"column:subscription_provider_id;type:uuid"`
	CreatedAt  time.Time  `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"column:last_updated;autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

// TableName returns the database table for offerings.
func (Offering) TableName() string { return "subscription_offerings" }

// PlanOffering is the join between a plan and an offering.
type PlanOffering struct {
	ID         uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PlanID     uuid.UUID        `gorm:"column:subscription_plan_id;type:uuid;not null"`
	OfferingID uuid.UUID        `gorm:"column:subscription_offering_id;type:uuid;not null"`
	Frequency  BillingFrequency `gorm:"column:freq;type:int"`
	Type       PlanType         `gorm:"type:int"`
	Fee        float64          `gorm:"type:decimal(10,2)"`
	Code       string           `gorm:"type:varchar(200)"`
	Currency   string           `gorm:"type:varchar(10)"`
	Name       string           `gorm:"type:varchar(200);not null"`
	Status     Status           `gorm:"type:int;default:1"`
	CreatedAt  time.Time        `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt  time.Time        `gorm:"column:last_updated;autoUpdateTime"`
	DeletedAt  *time.Time       `gorm:"column:deleted_at"`
}

// TableName returns the database table for plan offerings.
func (PlanOffering) TableName() string { return "subscription_plan_offerings" }

// Subscription is a user's active entitlement record.
type Subscription struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	PlanID          uuid.UUID  `gorm:"column:subscription_plan_id;type:uuid;not null"`
	SubCode         string     `gorm:"column:sub_code;type:varchar(200)"`
	Status          Status     `gorm:"type:int;default:1"`
	NextPaymentDate *time.Time `gorm:"column:next_payment_date"`
	CreatedAt       time.Time  `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"column:last_updated;autoUpdateTime"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`

	Plan *Plan `gorm:"foreignKey:PlanID"`
}

// TableName returns the database table for subscriptions.
func (Subscription) TableName() string { return "subscriptions" }

// IsExpired reports whether the subscription's next payment date has passed.
func (s *Subscription) IsExpired() bool {
	if s.NextPaymentDate == nil {
		return false
	}
	return time.Now().After(*s.NextPaymentDate)
}

// GlobalParameters holds app-wide feature flags.
type GlobalParameters struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ActivateSubscription bool      `gorm:"column:activate_subscription;default:true"`
}

// TableName returns the database table for global parameters.
func (GlobalParameters) TableName() string { return "global_parameters" }
