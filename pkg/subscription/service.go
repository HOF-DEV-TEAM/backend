package subscription

import (
	"context"
	"database/sql"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"bitbucket.org/hofng/hofApp/pkg/user"
)

type SubscriptionService interface {
	GetSubscriptions(ctx context.Context) ([]*SubscriptionJSON, int, error)
	GetSubscriptionPlans(ctx context.Context) ([]*SubscriptionPlan, int, error)
	CreateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error)
	DeleteSubscriptionPlanById(ctx context.Context, id string) (string, error)
	GetSubscription(ctx context.Context, userId string) (*Subscription, error)
	CreateSubscriptionPlan(ctx context.Context, subscriptionPlan *SubscriptionPlanRequest) (*SubscriptionPlan, error)
	GetSubscriptionPlanById(ctx context.Context, subPlanId string) (*SubscriptionPlan, error)
	GetSubscriptionPlanOfferings(ctx context.Context) ([]*SubscriptionPlanOffering, int, error)
	CreateSubscriptionPlanOffering(ctx context.Context, sub *SubscriptionPlanOfferingRequest) (string, error)
	VerifySubscription(ctx context.Context, subReq VerifySubRequest) (*Subscription, error)
}

type SubscriptionProviderService interface {
	CreateSubscriptionPlan(ctx context.Context, subscriptionPlan *SubscriptionPlanRequest) (*SubscriptionPlan, error)
	VerifySubscription(ctx context.Context, subReq VerifySubRequest) (*Subscription, error)
}

type Service interface {
	GetSession(ctx context.Context) (*user.UserSession, error)
	CreateSubscriptionOffering(ctx context.Context, offering *SubscriptionOfferingRequest) (string, error)
	SubscriptionService
}

type subscriptionSvc struct {
	repo        Repository
	userRepo    user.Repository
	config      *security.SecurityConfig
	subProvider SubscriptionProviderService //implements subscription provider service
}

func NewService(subProvider SubscriptionProviderService, repo Repository, config *security.SecurityConfig, userRepo user.Repository) Service {
	return &subscriptionSvc{subProvider: subProvider, repo: repo, userRepo: userRepo, config: config}
}

func (ss *subscriptionSvc) CreateSubscription(ctx context.Context, subReq *Subscription) (*Subscription, error) {
	sub, err := ss.repo.GetSubscriptionByUserAndPlanId(ctx, subReq.UserID, subReq.SubscriptionPlanID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if sub != nil {
		return sub, nil
	}
	return ss.repo.CreateSubscription(ctx, subReq)
}

func (ss *subscriptionSvc) CreateSubscriptionPlan(ctx context.Context, subscriptionPlan *SubscriptionPlanRequest) (*SubscriptionPlan, error) {
	//validate input
	planType := subscriptionPlan.Type
	var dummyType TypeEnum
	subscriptionPlan.Type = dummyType

	plan, err := ss.subProvider.CreateSubscriptionPlan(ctx, subscriptionPlan)

	if err != nil {
		return nil, err
	}

	plan.Type = planType
	return ss.repo.CreateSubscriptionPlan(ctx, plan)
}

func (ss *subscriptionSvc) CreateSubscriptionOffering(ctx context.Context, offering *SubscriptionOfferingRequest) (string, error) {
	return ss.repo.CreateSubscriptionOffering(ctx, offering)
}

func (ss *subscriptionSvc) GetSubscriptionPlanOfferings(ctx context.Context) ([]*SubscriptionPlanOffering, int, error) {
	return ss.repo.GetSubscriptionPlanOfferings(ctx)
}

func (ss *subscriptionSvc) VerifySubscription(ctx context.Context, subReq VerifySubRequest) (*Subscription, error) {
	claims, ok := ctx.Value(security.JWTClaimsContextKey).(*security.JWTClaim[any])
	if !ok {
		return nil, nil
	}

	sub, err := ss.subProvider.VerifySubscription(ctx, subReq)

	if err != nil {
		return nil, err
	}

	sub.UserID = claims.JWTClaimsMain.LoggedInUserId

	existingSub, err := ss.repo.GetSubscription(ctx, sub)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if existingSub != nil {
		_, err := ss.repo.UpdateSubscription(ctx, sub.UserID, sub)

		if err != nil {
			return nil, err
		}
	}

	sub.SubscriptionPlanID = subReq.PlanId
	sub.Status = 1
	return ss.CreateSubscription(ctx, sub)
}

func (ss *subscriptionSvc) CreateSubscriptionPlanOffering(ctx context.Context, subReq *SubscriptionPlanOfferingRequest) (string, error) {
	sub := &SubscriptionPlanOffering{
		SubscriptionPlanID:     sql.NullString{String: subReq.SubscriptionPlanId, Valid: true},
		SubscriptionOfferingID: sql.NullString{String: subReq.SubscriptionOfferingId, Valid: true},
		DateAdded: sql.NullString{
			String: time.Now().Format(time.RFC3339),
			Valid:  true,
		},
		LastUpdated: sql.NullString{
			String: time.Now().Format(time.RFC3339),
			Valid:  true,
		},
	}

	return ss.repo.CreateSubscriptionPlanOffering(ctx, sub)
}

func (ss *subscriptionSvc) GetSession(ctx context.Context) (*user.UserSession, error) {
	claims, ok := ctx.Value(security.JWTClaimsContextKey).(*security.JWTClaim[any])
	userId := claims.JWTClaimsMain.LoggedInUserId

	if !ok {
		return nil, nil
	}

	u, err := ss.userRepo.GetById(ctx, userId)
	if err != nil {

		return nil, err
	}

	updatedJWTToken, err := claims.PutUserIDAndSign(ss.config, claims.JWTClaimsMain.LoggedInUserId)

	if err != nil {
		return nil, err
	}

	return &user.UserSession{User: u.ToJSON(), Token: updatedJWTToken}, nil
}

func (ss *subscriptionSvc) GetSubscription(ctx context.Context, userId string) (*Subscription, error) {
	sub := &Subscription{
		UserID: userId,
	}

	_, err := ss.repo.GetSubscription(ctx, sub)

	if err != nil {
		return nil, err
	}

	return sub, err
}

func (ss *subscriptionSvc) GetSubscriptions(ctx context.Context) ([]*SubscriptionJSON, int, error) {
	return ss.repo.GetSubscriptions(ctx)
}

func (ss *subscriptionSvc) GetSubscriptionPlans(ctx context.Context) ([]*SubscriptionPlan, int, error) {
	return ss.repo.GetSubscriptionPlans(ctx)
}

func (ss *subscriptionSvc) DeleteSubscriptionPlanById(ctx context.Context, id string) (string, error) {
	return ss.repo.DeleteSubscriptionPlanById(ctx, id)
}

func (ss *subscriptionSvc) GetSubscriptionPlanById(ctx context.Context, subPlanId string) (*SubscriptionPlan, error) {
	return ss.repo.GetSubscriptionPlanById(ctx, subPlanId)
}
