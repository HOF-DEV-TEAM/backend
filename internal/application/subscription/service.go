package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	domainSub "bitbucket.org/hofng/hofApp/internal/domain/subscription"
	domainUser "bitbucket.org/hofng/hofApp/internal/domain/user"
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

	// Webhook
	HandleWebhookEvent(ctx context.Context, event *WebhookEvent) error

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

	// Update the user's Paystack customer info.
	if err := s.userRepo.UpdatePaystackInfo(ctx, userID, result.CustomerCode, result.CustomerID); err != nil {
		s.log.Warn("failed to update paystack customer info", zap.Error(err))
	}

	// Find the plan by Paystack plan code.
	plans, _, err := s.repo.GetPlans(ctx)
	if err != nil {
		return nil, err
	}
	var planID uuid.UUID
	for i := range plans {
		if plans[i].Code == result.PlanCode {
			planID = plans[i].ID
			break
		}
	}
	if planID == uuid.Nil {
		return nil, shared.ErrNotFound{Resource: "plan", ID: result.PlanCode}
	}

	sub := &domainSub.Subscription{
		UserID:  userID,
		PlanID:  planID,
		SubCode: result.SubscriptionCode,
		Status:  domainSub.StatusActive,
	}
	if result.NextPaymentDate != "" {
		if t, err := parsePaystackDate(result.NextPaymentDate); err == nil {
			sub.NextPaymentDate = &t
		}
	}

	if err := s.repo.UpsertSubscription(ctx, sub); err != nil {
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

// ── Webhook ───────────────────────────────────────────────────────────────────

func (s *subscriptionService) HandleWebhookEvent(ctx context.Context, event *WebhookEvent) error {
	s.log.Info("paystack webhook received", zap.String("event", string(event.Event)))
	switch event.Event {
	case EventChargeSuccess:
		return s.handleChargeSuccess(ctx, event)
	case EventInvoiceUpdate:
		return s.handleInvoiceUpdate(ctx, event)
	case EventSubscriptionCreate:
		return s.handleSubscriptionCreate(ctx, event)
	case EventSubscriptionNotRenew:
		return s.handleSubscriptionNotRenew(ctx, event)
	case EventInvoicePaymentFailed:
		return s.handleInvoicePaymentFailed(ctx, event)
	default:
		s.log.Info("unhandled paystack event", zap.String("event", string(event.Event)))
	}
	return nil
}

func (s *subscriptionService) handleChargeSuccess(ctx context.Context, event *WebhookEvent) error {
	user, err := s.userRepo.GetByCustomerCode(ctx, event.Data.Customer.CustomerCode)
	if err != nil {
		return fmt.Errorf("charge.success: lookup user by customer code: %w", err)
	}

	sub, err := s.repo.GetSubscriptionByUserID(ctx, user.ID)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("charge.success: get subscription: %w", err)
	}

	nextPayment, nextPaymentDate := parseDatePtr(event.Data.Subscription.NextPaymentDate)
	if !nextPayment {
		nextPaymentDate2, nextPaymentDate3 := parseDatePtr(event.Data.NextPaymentDate)
		nextPayment = nextPaymentDate2
		nextPaymentDate = nextPaymentDate3
	}
	_ = nextPayment

	subCode := event.Data.Subscription.SubscriptionCode
	if subCode == "" {
		subCode = event.Data.SubscriptionCode
	}
	return s.repo.UpdateSubscriptionByCode(ctx, sub.SubCode, domainSub.StatusActive, nextPaymentDate)
}

func (s *subscriptionService) handleInvoiceUpdate(ctx context.Context, event *WebhookEvent) error {
	subCode := event.Data.Subscription.SubscriptionCode
	if subCode == "" {
		subCode = event.Data.SubscriptionCode
	}
	if subCode == "" {
		return nil
	}

	_, nextPaymentDate := parseDatePtr(event.Data.Subscription.NextPaymentDate)
	if nextPaymentDate == nil {
		_, nextPaymentDate = parseDatePtr(event.Data.NextPaymentDate)
	}

	return s.repo.UpdateSubscriptionByCode(ctx, subCode, domainSub.StatusActive, nextPaymentDate)
}

func (s *subscriptionService) handleSubscriptionCreate(ctx context.Context, event *WebhookEvent) error {
	user, err := s.userRepo.GetByEmail(ctx, event.Data.Customer.Email)
	if err != nil {
		return fmt.Errorf("subscription.create: lookup user by email: %w", err)
	}

	plans, _, err := s.repo.GetPlans(ctx)
	if err != nil {
		return fmt.Errorf("subscription.create: list plans: %w", err)
	}
	var planID uuid.UUID
	for i := range plans {
		if plans[i].Code == event.Data.Plan.PlanCode {
			planID = plans[i].ID
			break
		}
	}
	if planID == uuid.Nil {
		s.log.Warn("subscription.create: plan not found locally", zap.String("plan_code", event.Data.Plan.PlanCode))
		return nil
	}

	subCode := event.Data.SubscriptionCode
	sub := &domainSub.Subscription{
		UserID:  user.ID,
		PlanID:  planID,
		SubCode: subCode,
		Status:  domainSub.StatusActive,
	}
	_, sub.NextPaymentDate = parseDatePtr(event.Data.NextPaymentDate)

	if err := s.repo.UpsertSubscription(ctx, sub); err != nil {
		return fmt.Errorf("subscription.create: upsert: %w", err)
	}
	return nil
}

func (s *subscriptionService) handleSubscriptionNotRenew(ctx context.Context, event *WebhookEvent) error {
	subCode := event.Data.SubscriptionCode
	if subCode == "" {
		return nil
	}
	return s.repo.UpdateSubscriptionByCode(ctx, subCode, domainSub.StatusPastDue, nil)
}

func (s *subscriptionService) handleInvoicePaymentFailed(ctx context.Context, event *WebhookEvent) error {
	subCode := event.Data.Subscription.SubscriptionCode
	if subCode == "" {
		subCode = event.Data.SubscriptionCode
	}
	if subCode == "" {
		return nil
	}
	return s.repo.UpdateSubscriptionByCode(ctx, subCode, domainSub.StatusCanceled, nil)
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

// ── helpers ───────────────────────────────────────────────────────────────────

// parsePaystackDate parses the date formats Paystack may return.
func parsePaystackDate(s string) (time.Time, error) {
	formats := []string{time.RFC3339, "2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05Z"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable date: %s", s)
}

// parseDatePtr parses a date string and returns a pointer; nil if empty or invalid.
func parseDatePtr(s string) (bool, *time.Time) {
	if s == "" {
		return false, nil
	}
	t, err := parsePaystackDate(s)
	if err != nil {
		return false, nil
	}
	return true, &t
}
