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

// ── Webhook event types ───────────────────────────────────────────────────────

// EventType identifies a Paystack webhook event.
type EventType string

const (
	EventChargeSuccess         EventType = "charge.success"
	EventInvoiceUpdate         EventType = "invoice.update"
	EventSubscriptionCreate    EventType = "subscription.create"
	EventSubscriptionNotRenew  EventType = "subscription.not_renew"
	EventInvoicePaymentFailed  EventType = "invoice.payment_failed"
)

// WebhookEvent is the parsed Paystack webhook payload.
type WebhookEvent struct {
	Event EventType        `json:"event"`
	Data  WebhookEventData `json:"data"`
}

// WebhookEventData carries the fields common across all Paystack event types.
type WebhookEventData struct {
	SubscriptionCode string              `json:"subscription_code"`
	NextPaymentDate  string              `json:"next_payment_date"`
	Customer         WebhookCustomer     `json:"customer"`
	Plan             WebhookPlan         `json:"plan"`
	Subscription     WebhookSubscription `json:"subscription"`
}

// WebhookCustomer is the customer block inside a Paystack event.
type WebhookCustomer struct {
	Email        string `json:"email"`
	CustomerCode string `json:"customer_code"`
}

// WebhookPlan is the plan block inside a Paystack event.
type WebhookPlan struct {
	PlanCode string `json:"plan_code"`
}

// WebhookSubscription is the nested subscription block (invoice events).
type WebhookSubscription struct {
	SubscriptionCode string `json:"subscription_code"`
	NextPaymentDate  string `json:"next_payment_date"`
}
