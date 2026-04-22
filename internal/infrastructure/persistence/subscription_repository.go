package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainSub "bitbucket.org/hofng/hofApp/internal/domain/subscription"
	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type subscriptionRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewSubscriptionRepository returns a GORM-backed implementation of subscription.Repository.
func NewSubscriptionRepository(db *gorm.DB, log *zap.Logger) domainSub.Repository {
	return &subscriptionRepository{db: db, log: log}
}

// ── Plans ─────────────────────────────────────────────────────────────────────

func (r *subscriptionRepository) CreatePlan(ctx context.Context, p *domainSub.Plan) error {
	if result := r.db.WithContext(ctx).Create(p); result.Error != nil {
		return fmt.Errorf("creating subscription plan: %w", result.Error)
	}
	return nil
}

func (r *subscriptionRepository) GetPlans(ctx context.Context) ([]domainSub.Plan, int64, error) {
	var plans []domainSub.Plan
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("date_added DESC")

	var total int64
	if err := q.Model(&domainSub.Plan{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting plans: %w", err)
	}
	if result := q.Find(&plans); result.Error != nil {
		return nil, 0, fmt.Errorf("listing plans: %w", result.Error)
	}
	return plans, total, nil
}

func (r *subscriptionRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*domainSub.Plan, error) {
	var p domainSub.Plan
	result := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&p, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "subscription plan", ID: id.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting plan by id: %w", result.Error)
	}
	return &p, nil
}

func (r *subscriptionRepository) DeletePlan(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domainSub.Plan{}).
		Where("id = ?", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return fmt.Errorf("deleting plan: %w", result.Error)
	}
	return nil
}

// ── Offerings ─────────────────────────────────────────────────────────────────

func (r *subscriptionRepository) CreateOffering(ctx context.Context, o *domainSub.Offering) error {
	if result := r.db.WithContext(ctx).Create(o); result.Error != nil {
		return fmt.Errorf("creating offering: %w", result.Error)
	}
	return nil
}

func (r *subscriptionRepository) GetOfferings(ctx context.Context) ([]domainSub.Offering, int64, error) {
	var offerings []domainSub.Offering
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("date_added DESC")

	var total int64
	if err := q.Model(&domainSub.Offering{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting offerings: %w", err)
	}
	if result := q.Find(&offerings); result.Error != nil {
		return nil, 0, fmt.Errorf("listing offerings: %w", result.Error)
	}
	return offerings, total, nil
}

func (r *subscriptionRepository) DeleteOffering(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domainSub.Offering{}).
		Where("id = ?", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return fmt.Errorf("deleting offering: %w", result.Error)
	}
	return nil
}

// ── Plan offerings ────────────────────────────────────────────────────────────

func (r *subscriptionRepository) CreatePlanOffering(ctx context.Context, po *domainSub.PlanOffering) error {
	if result := r.db.WithContext(ctx).Create(po); result.Error != nil {
		return fmt.Errorf("creating plan offering: %w", result.Error)
	}
	return nil
}

func (r *subscriptionRepository) GetPlanOfferings(ctx context.Context) ([]domainSub.PlanOffering, int64, error) {
	var planOfferings []domainSub.PlanOffering
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("date_added DESC")

	var total int64
	if err := q.Model(&domainSub.PlanOffering{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting plan offerings: %w", err)
	}
	if result := q.Find(&planOfferings); result.Error != nil {
		return nil, 0, fmt.Errorf("listing plan offerings: %w", result.Error)
	}
	return planOfferings, total, nil
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

func (r *subscriptionRepository) CreateSubscription(ctx context.Context, s *domainSub.Subscription) error {
	if result := r.db.WithContext(ctx).Create(s); result.Error != nil {
		return fmt.Errorf("creating subscription: %w", result.Error)
	}
	return nil
}

func (r *subscriptionRepository) GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID) (*domainSub.Subscription, error) {
	var s domainSub.Subscription
	result := r.db.WithContext(ctx).
		Preload("Plan").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("date_added DESC").
		First(&s)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "subscription", ID: userID.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting subscription by user id: %w", result.Error)
	}
	return &s, nil
}

func (r *subscriptionRepository) GetSubscriptionByUserAndPlan(ctx context.Context, userID, planID uuid.UUID) (*domainSub.Subscription, error) {
	var s domainSub.Subscription
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND subscription_plan_id = ? AND deleted_at IS NULL", userID, planID).
		First(&s)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "subscription", ID: userID.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting subscription by user and plan: %w", result.Error)
	}
	return &s, nil
}

func (r *subscriptionRepository) GetSubscriptionByCode(ctx context.Context, subCode string) (*domainSub.Subscription, error) {
	var s domainSub.Subscription
	result := r.db.WithContext(ctx).
		Preload("Plan").
		Where("sub_code = ? AND deleted_at IS NULL", subCode).
		First(&s)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "subscription", ID: subCode}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting subscription by code: %w", result.Error)
	}
	return &s, nil
}

func (r *subscriptionRepository) GetAllSubscriptions(ctx context.Context) ([]domainSub.Subscription, int64, error) {
	var subs []domainSub.Subscription
	q := r.db.WithContext(ctx).Preload("Plan").Where("deleted_at IS NULL").Order("date_added DESC")

	var total int64
	if err := q.Model(&domainSub.Subscription{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting subscriptions: %w", err)
	}
	if result := q.Find(&subs); result.Error != nil {
		return nil, 0, fmt.Errorf("listing subscriptions: %w", result.Error)
	}
	return subs, total, nil
}

func (r *subscriptionRepository) UpdateSubscriptionStatus(ctx context.Context, id uuid.UUID, status domainSub.Status) error {
	result := r.db.WithContext(ctx).Model(&domainSub.Subscription{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("updating subscription status: %w", result.Error)
	}
	return nil
}

// UpsertSubscription creates or updates the subscription for the given user+plan pair.
func (r *subscriptionRepository) UpsertSubscription(ctx context.Context, s *domainSub.Subscription) error {
	existing, err := r.GetSubscriptionByUserAndPlan(ctx, s.UserID, s.PlanID)
	if err != nil && !shared.IsNotFound(err) {
		return fmt.Errorf("upsert subscription lookup: %w", err)
	}
	if existing != nil {
		updates := map[string]any{
			"status":      s.Status,
			"sub_code":    s.SubCode,
			"last_updated": time.Now(),
		}
		if s.NextPaymentDate != nil {
			updates["next_payment_date"] = s.NextPaymentDate
		}
		result := r.db.WithContext(ctx).Model(&domainSub.Subscription{}).
			Where("id = ?", existing.ID).Updates(updates)
		return result.Error
	}
	if result := r.db.WithContext(ctx).Create(s); result.Error != nil {
		return fmt.Errorf("creating subscription: %w", result.Error)
	}
	return nil
}

// UpdateSubscriptionByCode updates status and next payment date for a subscription
// identified by its Paystack subscription code.
func (r *subscriptionRepository) UpdateSubscriptionByCode(ctx context.Context, subCode string, status domainSub.Status, nextPaymentDate *time.Time) error {
	updates := map[string]any{"status": status, "last_updated": time.Now()}
	if nextPaymentDate != nil {
		updates["next_payment_date"] = nextPaymentDate
	}
	result := r.db.WithContext(ctx).Model(&domainSub.Subscription{}).
		Where("sub_code = ? AND deleted_at IS NULL", subCode).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("updating subscription by code: %w", result.Error)
	}
	return nil
}

// ── Global parameters ─────────────────────────────────────────────────────────

func (r *subscriptionRepository) GetGlobalParameters(ctx context.Context) (*domainSub.GlobalParameters, error) {
	var params domainSub.GlobalParameters
	result := r.db.WithContext(ctx).First(&params)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Seed default if none exists.
		params = domainSub.GlobalParameters{ActivateSubscription: true}
		if err := r.db.WithContext(ctx).Create(&params).Error; err != nil {
			return nil, fmt.Errorf("seeding global parameters: %w", err)
		}
		return &params, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting global parameters: %w", result.Error)
	}
	return &params, nil
}

func (r *subscriptionRepository) UpdateGlobalParameters(ctx context.Context, params *domainSub.GlobalParameters) error {
	result := r.db.WithContext(ctx).Save(params)
	if result.Error != nil {
		return fmt.Errorf("updating global parameters: %w", result.Error)
	}
	return nil
}
