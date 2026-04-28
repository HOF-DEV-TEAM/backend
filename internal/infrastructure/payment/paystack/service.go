// Package paystack adapts the REST client to the subscription domain interface.
package paystack

import (
	"context"
	"fmt"

	domainSub "bitbucket.org/hofng/hofApp/internal/domain/subscription"
	"go.uber.org/zap"
)

// Service adapts the Paystack Client to the domain's PaymentProvider interface.
type Service struct {
	client *Client
	log    *zap.Logger
}

// NewService wraps a Paystack Client as a domain PaymentProvider.
func NewService(client *Client, log *zap.Logger) domainSub.PaymentProvider {
	return &Service{client: client, log: log}
}

// InitializeTransaction creates a Paystack payment session and returns the authorization URL.
func (s *Service) InitializeTransaction(ctx context.Context, req domainSub.InitTransactionRequest) (*domainSub.TransactionResponse, error) {
	authURL, accessCode, ref, err := s.client.InitializeTransaction(
		ctx, req.Email, req.Amount, req.PlanCode, req.Reference,
	)
	if err != nil {
		return nil, fmt.Errorf("initializing paystack transaction: %w", err)
	}
	return &domainSub.TransactionResponse{
		AuthorizationURL: authURL,
		AccessCode:       accessCode,
		Reference:        ref,
	}, nil
}

// VerifyTransaction confirms a completed payment by its reference string.
func (s *Service) VerifyTransaction(ctx context.Context, reference string) (*domainSub.VerifyTransactionResponse, error) {
	resp, err := s.client.VerifyTransaction(ctx, reference)
	if err != nil {
		return nil, fmt.Errorf("verifying paystack transaction: %w", err)
	}

	return &domainSub.VerifyTransactionResponse{
		Status:           resp.Data.Status == "success",
		CustomerCode:     resp.Data.Customer.Code,
		CustomerID:       fmt.Sprintf("%d", resp.Data.Customer.ID),
		SubscriptionCode: resp.Data.Subscription.SubscriptionCode,
		NextPaymentDate:  resp.Data.Subscription.NextPaymentDate,
		PlanCode:         resp.Data.Subscription.Plan.PlanCode,
	}, nil
}

// DisableSubscription cancels an active Paystack subscription by code and token.
func (s *Service) DisableSubscription(ctx context.Context, code, token string) (*domainSub.DisableResponse, error) {
	if err := s.client.DisableSubscription(ctx, code, token); err != nil {
		return nil, fmt.Errorf("disabling paystack subscription: %w", err)
	}
	return &domainSub.DisableResponse{Message: "subscription disabled"}, nil
}

// CreateCustomer registers a new customer record in Paystack.
func (s *Service) CreateCustomer(ctx context.Context, req domainSub.CreateCustomerRequest) (*domainSub.CustomerResponse, error) {
	code, id, err := s.client.CreateCustomer(ctx, req.Email, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("creating paystack customer: %w", err)
	}
	return &domainSub.CustomerResponse{
		CustomerCode: code,
		CustomerID:   id,
	}, nil
}
