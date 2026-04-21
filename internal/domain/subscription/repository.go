package subscription

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines every persistence operation needed for subscriptions.
type Repository interface {
	// Plans
	CreatePlan(ctx context.Context, p *Plan) error
	GetPlans(ctx context.Context) ([]Plan, int64, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*Plan, error)
	DeletePlan(ctx context.Context, id uuid.UUID) error

	// Offerings
	CreateOffering(ctx context.Context, o *Offering) error
	GetOfferings(ctx context.Context) ([]Offering, int64, error)
	DeleteOffering(ctx context.Context, id uuid.UUID) error

	// Plan offerings
	CreatePlanOffering(ctx context.Context, po *PlanOffering) error
	GetPlanOfferings(ctx context.Context) ([]PlanOffering, int64, error)

	// Subscriptions
	CreateSubscription(ctx context.Context, s *Subscription) error
	GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	GetSubscriptionByUserAndPlan(ctx context.Context, userID, planID uuid.UUID) (*Subscription, error)
	GetAllSubscriptions(ctx context.Context) ([]Subscription, int64, error)
	UpdateSubscriptionStatus(ctx context.Context, id uuid.UUID, status Status) error
	UpdateSubscriptionByCode(ctx context.Context, code string, status Status, nextPaymentDate *string) error

	// Global parameters
	GetGlobalParameters(ctx context.Context) (*GlobalParameters, error)
	UpdateGlobalParameters(ctx context.Context, params *GlobalParameters) error
}

// PaymentProvider abstracts a third-party payment processor (e.g. Paystack).
type PaymentProvider interface {
	InitializeTransaction(ctx context.Context, req InitTransactionRequest) (*TransactionResponse, error)
	VerifyTransaction(ctx context.Context, reference string) (*VerifyTransactionResponse, error)
	DisableSubscription(ctx context.Context, code, token string) (*DisableResponse, error)
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*CustomerResponse, error)
}

// InitTransactionRequest is the input for beginning a payment transaction.
type InitTransactionRequest struct {
	Email     string
	Amount    int64
	PlanCode  string
	Reference string
	Metadata  map[string]interface{}
}

// TransactionResponse is the payment provider's response to a transaction start.
type TransactionResponse struct {
	AuthorizationURL string
	AccessCode       string
	Reference        string
}

// VerifyTransactionResponse contains subscription details after verification.
type VerifyTransactionResponse struct {
	Status          bool
	CustomerCode    string
	CustomerID      string
	SubscriptionCode string
	NextPaymentDate string
	PlanCode        string
}

// DisableResponse is returned when a subscription is cancelled at the provider.
type DisableResponse struct {
	Message string
}

// CreateCustomerRequest carries the data needed to register a customer.
type CreateCustomerRequest struct {
	Email     string
	FirstName string
	LastName  string
	Phone     string
}

// CustomerResponse is the provider's response after creating a customer record.
type CustomerResponse struct {
	CustomerCode string
	CustomerID   string
}
