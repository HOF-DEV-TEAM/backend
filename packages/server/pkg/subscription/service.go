package subscription

import "context"

type SubscriptionService interface {
	CreateSubscription(ctx context.Context)
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
	subProvider SubscriptionService //implements subsription service
}

func NewService(subProvider SubscriptionService, repo Repository) Service {
	return &subscriptionSvc{subProvider: subProvider, repo: repo}
}

func (ss *subscriptionSvc) CreateSubscription(ctx context.Context) {
	ss.subProvider.CreateSubscription(ctx)
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
