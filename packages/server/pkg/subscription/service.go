package subscription

import (
	"context"
	"database/sql"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, subReq *SubscriptionRequest) (*Subscription, error)
	CancelSubscription(ctx context.Context)
	ChangeSubscription(ctx context.Context)
	CreateSubscriptionPlan(ctx context.Context, subscriptionPlan *SubscriptionPlanRequest) (*SubscriptionPlan, error)
	MakePayment(ctx context.Context)
}

type Service interface {
	CreateSubscriptionOffering(ctx context.Context, offering *SubscriptionOfferingRequest) (string, error)
	SubscriptionService
}

type subscriptionSvc struct {
	repo        Repository
	config *security.SecurityConfig
	subProvider SubscriptionService //implements subsription service
}

func NewService(subProvider SubscriptionService, repo Repository, config *security.SecurityConfig) Service {
	return &subscriptionSvc{subProvider: subProvider, repo: repo, config: config}
}

func (ss *subscriptionSvc) CreateSubscription(ctx context.Context, subReq *SubscriptionRequest)  (*Subscription, error) {
	claims, ok := ctx.Value(ss.config.JWTClaimsContextKey).(*security.JWTClaim)
	
	if !ok {
		return nil, http_helper.ErrInvalidAccount
	}

	//check if user is subscribed to the same plan - return  plan if true
	plan, err := ss.repo.GetPlan(ctx, subReq.Plan)	

	if err != nil {
		return nil, err
	}
	
	sub, err := ss.repo.GetSubscription(ctx, claims.JWTClaimsMain.LoggedInUserId, plan.ID)

	//user is already subscribed to plan
	if sub != nil {
		return sub, nil
	}

	if err == sql.ErrNoRows  {
		payStackSub, err := ss.subProvider.CreateSubscription(ctx, subReq)
		
		if err != nil {
			return nil, err
		}
		payStackSub.SubscriptionPlanID = plan.ID
		return ss.repo.CreateSubscription(ctx, payStackSub)		
	}

	return nil, err	
}

func (ss *subscriptionSvc) CancelSubscription(ctx context.Context) {
	ss.subProvider.ChangeSubscription(ctx)
}

func (ss *subscriptionSvc) ChangeSubscription(ctx context.Context) {
	ss.subProvider.CancelSubscription(ctx)
}

func (ss *subscriptionSvc) CreateSubscriptionPlan(ctx context.Context, subscriptionPlan *SubscriptionPlanRequest) (*SubscriptionPlan, error) {
	//validate input
	plan, err := ss.subProvider.CreateSubscriptionPlan(ctx, subscriptionPlan)

	if err != nil {
		return nil, err
	}
	
	return ss.repo.CreateSubscriptionPlan(ctx, plan)
	
}

func (ss *subscriptionSvc) MakePayment(ctx context.Context) {
	ss.subProvider.MakePayment(ctx)
}

func (ss *subscriptionSvc) CreateSubscriptionOffering(ctx context.Context, offering *SubscriptionOfferingRequest) (string, error) {
	return ss.repo.CreateSubscriptionOffering(ctx, offering)
}
