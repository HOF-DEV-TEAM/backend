package paystack

import (
	"context"

	"bitbucket.org/hofng/hofApp/pkg/subscription"
)

type paystackService struct {
	payStackClient *PayStackClientHttp
}

func NewPaystackService(payStackClient *PayStackClientHttp) subscription.SubscriptionService {
	return &paystackService{payStackClient: payStackClient}
}

func (pc *paystackService) CreateSubscription(ctx context.Context) {
	pc.payStackClient.CreateSubscription(ctx)
}

func (pc *paystackService) CreateSubscriptionPlan(ctx context.Context, subscritpionPlan *subscription.SubscriptionPlanRequest) (*subscription.SubscriptionPlan, error) {
	result, err := pc.payStackClient.CreateSubscriptionPlan(ctx, subscritpionPlan)

	if err != nil {
		return nil, err
	}
	
	return result.ToSubscriptionPlan(), err
}

func (pc *paystackService) CancelSubscription(ctx context.Context) {

}

func (pc *paystackService) ChangeSubscription(ctx context.Context) {
}

func (pc *paystackService) MakePayment(ctx context.Context) {
}
