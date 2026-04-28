package subscription

import "errors"

var (
	// ErrPlanNotFound is returned when a plan cannot be found.
	ErrPlanNotFound = errors.New("subscription plan not found")
	// ErrSubscriptionNotFound is returned when a subscription cannot be found.
	ErrSubscriptionNotFound = errors.New("subscription not found")
	// ErrOfferingNotFound is returned when an offering cannot be found.
	ErrOfferingNotFound = errors.New("subscription offering not found")
	// ErrAlreadySubscribed is returned when a user already has an active subscription.
	ErrAlreadySubscribed = errors.New("user already has an active subscription")
	// ErrPaymentFailed is returned when payment verification fails.
	ErrPaymentFailed = errors.New("payment verification failed")
	// ErrSubscriptionExpired is returned when a subscription has expired.
	ErrSubscriptionExpired = errors.New("subscription has expired")
)
