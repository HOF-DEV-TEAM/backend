package subscription

// CreatePlanRequest adds a new subscription plan.
type CreatePlanRequest struct {
	Name      string  `json:"name"      validate:"required"`
	Type      int     `json:"type"`
	Frequency int     `json:"freq"`
	Fee       float64 `json:"fee"`
	Currency  string  `json:"currency"`
	Code      string  `json:"code"`
}

// CreateOfferingRequest adds a new subscription offering.
type CreateOfferingRequest struct {
	Name string `json:"name" validate:"required"`
}

// CreatePlanOfferingRequest links a plan to an offering with pricing details.
type CreatePlanOfferingRequest struct {
	PlanID     string  `json:"subscription_plan_id"     validate:"required,uuid"`
	OfferingID string  `json:"subscription_offering_id" validate:"required,uuid"`
	Name       string  `json:"name"                     validate:"required"`
	Frequency  int     `json:"freq"`
	Type       int     `json:"type"`
	Fee        float64 `json:"fee"`
	Code       string  `json:"code"`
	Currency   string  `json:"currency"`
}

// VerifySubscriptionRequest initiates post-payment subscription verification.
type VerifySubscriptionRequest struct {
	Reference string `json:"reference" validate:"required"`
}

// InitTransactionRequest begins a Paystack payment session.
type InitTransactionRequest struct {
	PlanID string `json:"plan_id" validate:"required,uuid"`
	Email  string `json:"email"   validate:"required,email"`
}

// DisableSubscriptionRequest cancels a subscription via Paystack.
type DisableSubscriptionRequest struct {
	Code  string `json:"code"  validate:"required"`
	Token string `json:"token" validate:"required"`
}

// UpdateGlobalParamsRequest updates app-wide feature flags.
type UpdateGlobalParamsRequest struct {
	ActivateSubscription bool `json:"activate_subscription"`
}
