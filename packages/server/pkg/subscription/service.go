package subscription

import (
	"context"
	"database/sql"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"bitbucket.org/hofng/hofApp/pkg/user"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, subReq *SubscriptionRequest) (*Subscription, error)
	CancelSubscription(ctx context.Context)
	ChangeSubscription(ctx context.Context)
	CreateSubscriptionPlan(ctx context.Context, subscriptionPlan *SubscriptionPlanRequest) (*SubscriptionPlan, error)
	GetSubscriptionPlanOfferings(ctx context.Context) ([]*SubscriptionPlanOffering, int, error)
	CreateSubscriptionPlanOffering(ctx context.Context, sub *SubscriptionPlanOfferingRequest) (string, error)
	VerifySubscription(ctx context.Context, subRef string) (*Subscription, error)
	MakePayment(ctx context.Context)
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
	subProvider SubscriptionService //implements subsription service
}

func NewService(subProvider SubscriptionService, repo Repository, config *security.SecurityConfig, userRepo user.Repository) Service {
	return &subscriptionSvc{subProvider: subProvider, repo: repo, userRepo: userRepo, config: config}
}

func (ss *subscriptionSvc) CreateSubscription(ctx context.Context, subReq *SubscriptionRequest) (*Subscription, error) {
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

	if err == sql.ErrNoRows {
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

func (ss *subscriptionSvc) MakePayment(ctx context.Context) {
	ss.subProvider.MakePayment(ctx)
}

func (ss *subscriptionSvc) CreateSubscriptionOffering(ctx context.Context, offering *SubscriptionOfferingRequest) (string, error) {
	return ss.repo.CreateSubscriptionOffering(ctx, offering)
}

func (ss *subscriptionSvc) GetSubscriptionPlanOfferings(ctx context.Context) ([]*SubscriptionPlanOffering, int, error) {
	return ss.repo.GetSubscriptionPlanOfferings(ctx)
}

func (ss *subscriptionSvc) VerifySubscription(ctx context.Context, subRef string) (*Subscription, error) {
	return ss.subProvider.VerifySubscription(ctx, subRef)
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
	claims, ok := ctx.Value(security.JWTClaimsContextKey).(*security.JWTClaim)
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

	return &user.UserSession{User: user.NewJSONUser(u), Token: updatedJWTToken}, nil
}
