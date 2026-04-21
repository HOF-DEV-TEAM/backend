package subscription

import (
	"context"
	"fmt"

	domainSub "bitbucket.org/hofng/hofApp/internal/domain/subscription"
	domainUser "bitbucket.org/hofng/hofApp/internal/domain/user"
	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var validate = validator.New()

// Service exposes all subscription use cases.
type Service interface {
	// Plans
	CreatePlan(ctx context.Context, req CreatePlanRequest) (*domainSub.Plan, error)
	ListPlans(ctx context.Context) ([]domainSub.Plan, int64, error)
	GetPlan(ctx context.Context, id uuid.UUID) (*domainSub.Plan, error)
	DeletePlan(ctx context.Context, id uuid.UUID) error

	// Offerings
	CreateOffering(ctx context.Context, req CreateOfferingRequest) (*domainSub.Offering, error)
	ListOfferings(ctx context.Context) ([]domainSub.Offering, int64, error)
	DeleteOffering(ctx context.Context, id uuid.UUID) error

	// Plan offerings
	CreatePlanOffering(ctx context.Context, req CreatePlanOfferingRequest) (*domainSub.PlanOffering, error)
	ListPlanOfferings(ctx context.Context) ([]domainSub.PlanOffering, int64, error)

	// Subscriptions
	ListSubscriptions(ctx context.Context) ([]domainSub.Subscription, int64, error)
	GetUserSubscription(ctx context.Context, userID uuid.UUID) (*domainSub.Subscription, error)
	VerifySubscription(ctx context.Context, userID uuid.UUID, req VerifySubscriptionRequest) (*domainSub.Subscription, error)
	InitializeTransaction(ctx context.Context, userID uuid.UUID, req InitTransactionRequest) (*domainSub.TransactionResponse, error)
	DisableSubscription(ctx context.Context, req DisableSubscriptionRequest) (*domainSub.DisableResponse, error)

	// Global parameters
	GetGlobalParameters(ctx context.Context) (*domainSub.GlobalParameters, error)
	UpdateGlobalParameters(ctx context.Context, req UpdateGlobalParamsRequest) (*domainSub.GlobalParameters, error)
}

type subscriptionService struct {
	repo     domainSub.Repository
	provider domainSub.PaymentProvider
	userRepo domainUser.Repository
	log      *zap.Logger
}

// NewService creates the subscription application service.
func NewService(
	repo domainSub.Repository,
	provider domainSub.PaymentProvider,
	userRepo domainUser.Repository,
	log *zap.Logger,
) Service {
	return &subscriptionService{
		repo:     repo,
		provider: provider,
		userRepo: userRepo,
		log:      log,
	}
}

// ── Plans ─────────────────────────────────────────────────────────────────────

func (s *subscriptionService) CreatePlan(ctx context.Context, req CreatePlanRequest) (*domainSub.Plan, error) {
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	p := &domainSub.Plan{
		Name:      req.Name,
		Type:      domainSub.PlanType(req.Type),
		Frequency: domainSub.BillingFrequency(req.Frequency),
		Fee:       req.Fee,
		Currency:  req.Currency,
		Code:      req.Code,
		Status:    domainSub.StatusActive,
	}

	if err := s.repo.CreatePlan(ctx, p); err != nil {
		return nil, fmt.Errorf("create plan: %w", err)
	}
	return p, nil
}

func (s *subscriptionService) ListPlans(ctx context.Context) ([]domainSub.Plan, int64, error) {
	return s.repo.GetPlans(ctx)
}

func (s *subscriptionService) GetPlan(ctx context.Context, id uuid.UUID) (*domainSub.Plan, error) {
	return s.repo.GetPlanByID(ctx, id)
}

func (s *subscriptionService) DeletePlan(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePlan(ctx, id)
}

// ── Offerings ─────────────────────────────────────────────────────────────────

func (s *subscriptionService) CreateOffering(ctx context.Context, req CreateOfferingRequest) (*domainSub.Offering, error) {
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	o := &domainSub.Offering{
		Name:   req.Name,
		Status: domainSub.StatusActive,
	}
	if err := s.repo.CreateOffering(ctx, o); err != nil {
		return nil, fmt.Errorf("create offering: %w", err)
	}
	return o, nil
}

func (s *subscriptionService) ListOfferings(ctx context.Context) ([]domainSub.Offering, int64, error) {
	return s.repo.GetOfferings(ctx)
}

func (s *subscriptionService) DeleteOffering(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteOffering(ctx, id)
}

// ── Plan offerings ────────────────────────────────────────────────────────────

func (s *subscriptionService) CreatePlanOffering(ctx context.Context, req CreatePlanOfferingRequest) (*domainSub.PlanOffering, error) {
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		return nil, shared.ErrInvalidInput{Field: "subscription_plan_id", Message: "invalid UUID"}
	}
	offeringID, err := uuid.Parse(req.OfferingID)
	if err != nil {
		return nil, shared.ErrInvalidInput{Field: "subscription_offering_id", Message: "invalid UUID"}
	}

	po := &domainSub.PlanOffering{
		PlanID:     planID,
		OfferingID: offeringID,
		Name:       req.Name,
		Frequency:  domainSub.BillingFrequency(req.Frequency),
		Type:       domainSub.PlanType(req.Type),
		Fee:        req.Fee,
		Code:       req.Code,
		Currency:   req.Currency,
		Status:     domainSub.StatusActive,
	}
	if err := s.repo.CreatePlanOffering(ctx, po); err != nil {
		return nil, fmt.Errorf("create plan offering: %w", err)
	}
	return po, nil
}

func (s *subscriptionService) ListPlanOfferings(ctx context.Context) ([]domainSub.PlanOffering, int64, error) {
	return s.repo.GetPlanOfferings(ctx)
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

func (s *subscriptionService) ListSubscriptions(ctx context.Context) ([]domainSub.Subscription, int64, error) {
	return s.repo.GetAllSubscriptions(ctx)
}

func (s *subscriptionService) GetUserSubscription(ctx context.Context, userID uuid.UUID) (*domainSub.Subscription, error) {
	return s.repo.GetSubscriptionByUserID(ctx, userID)
}

func (s *subscriptionService) VerifySubscription(ctx context.Context, userID uuid.UUID, req VerifySubscriptionRequest) (*domainSub.Subscription, error) {
	result, err := s.provider.VerifyTransaction(ctx, req.Reference)
	if err != nil {
		return nil, fmt.Errorf("verifying payment: %w", err)
	}
	if !result.Status {
		return nil, domainSub.ErrPaymentFailed
	}

	// Update or create the user's Paystack customer info.
	if err := s.userRepo.UpdatePaystackInfo(ctx, userID, result.CustomerCode, result.CustomerID); err != nil {
		s.log.Warn("failed to update paystack customer info", zap.Error(err))
	}

	// Find the plan by Paystack plan code.
	plans, _, err := s.repo.GetPlans(ctx)
	if err != nil {
		return nil, err
	}
	var planID uuid.UUID
	for _, p := range plans {
		if p.Code == result.PlanCode {
			planID = p.ID
			break
		}
	}
	if planID == uuid.Nil {
		return nil, shared.ErrNotFound{Resource: "plan", ID: result.PlanCode}
	}

	sub := &domainSub.Subscription{
		UserID: userID,
		PlanID: planID,
		Status: domainSub.StatusActive,
	}
	if result.NextPaymentDate != "" {
		nextDate := &result.NextPaymentDate
		_ = s.repo.UpdateSubscriptionByCode(ctx, result.SubscriptionCode, domainSub.StatusActive, nextDate)
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("recording subscription: %w", err)
	}

	return sub, nil
}

func (s *subscriptionService) InitializeTransaction(ctx context.Context, userID uuid.UUID, req InitTransactionRequest) (*domainSub.TransactionResponse, error) {
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		return nil, shared.ErrInvalidInput{Field: "plan_id", Message: "invalid UUID"}
	}

	plan, err := s.repo.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	txReq := domainSub.InitTransactionRequest{
		Email:    req.Email,
		Amount:   int64(plan.Fee * 100),
		PlanCode: plan.Code,
	}
	return s.provider.InitializeTransaction(ctx, txReq)
}

func (s *subscriptionService) DisableSubscription(ctx context.Context, req DisableSubscriptionRequest) (*domainSub.DisableResponse, error) {
	return s.provider.DisableSubscription(ctx, req.Code, req.Token)
}

// ── Global parameters ─────────────────────────────────────────────────────────

func (s *subscriptionService) GetGlobalParameters(ctx context.Context) (*domainSub.GlobalParameters, error) {
	return s.repo.GetGlobalParameters(ctx)
}

func (s *subscriptionService) UpdateGlobalParameters(ctx context.Context, req UpdateGlobalParamsRequest) (*domainSub.GlobalParameters, error) {
	params, err := s.repo.GetGlobalParameters(ctx)
	if err != nil {
		return nil, err
	}

	params.ActivateSubscription = req.ActivateSubscription
	if err := s.repo.UpdateGlobalParameters(ctx, params); err != nil {
		return nil, err
	}
	return params, nil
}
