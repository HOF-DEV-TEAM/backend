package subscription

import "errors"

var (
	ErrPlanNotFound         = errors.New("subscription plan not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrOfferingNotFound     = errors.New("subscription offering not found")
	ErrAlreadySubscribed    = errors.New("user already has an active subscription")
	ErrPaymentFailed        = errors.New("payment verification failed")
	ErrSubscriptionExpired  = errors.New("subscription has expired")
)
