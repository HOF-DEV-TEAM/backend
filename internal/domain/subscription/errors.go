package subscription

import "bitbucket.org/hofng/hofApp/internal/domain/shared"

var (
	// ErrPlanNotFound is returned when a plan cannot be found.
	ErrPlanNotFound = shared.ErrNotFound{Resource: "subscription plan", ID: ""}
	// ErrSubscriptionNotFound is returned when a subscription cannot be found.
	ErrSubscriptionNotFound = shared.ErrNotFound{Resource: "subscription", ID: ""}
	// ErrOfferingNotFound is returned when an offering cannot be found.
	ErrOfferingNotFound = shared.ErrNotFound{Resource: "subscription offering", ID: ""}
	// ErrAlreadySubscribed is returned when a user already has an active subscription.
	ErrAlreadySubscribed = shared.ErrConflict{Message: "user already has an active subscription"}
	// ErrPaymentFailed is returned when payment verification fails.
	ErrPaymentFailed = shared.ErrInvalidInput{Message: "payment verification failed"}
	// ErrSubscriptionExpired is returned when a subscription has expired.
	ErrSubscriptionExpired = shared.ErrForbidden{Message: "subscription has expired"}
)
