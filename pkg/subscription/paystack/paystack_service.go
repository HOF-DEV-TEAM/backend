package paystack

import (
	"context"
	"database/sql"
	"fmt"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"bitbucket.org/hofng/hofApp/pkg/user"
	"github.com/gofrs/uuid"
)

type paystackService struct {
	payStackClient *PayStackClientHttp
	userRepo       user.Repository
	subRepo        subscription.Repository
	config         *security.SecurityConfig
}

func NewPaystackService(payStackClient *PayStackClientHttp, userRepo user.Repository, subRepo subscription.Repository, config *security.SecurityConfig) subscription.SubscriptionService {
	return &paystackService{payStackClient: payStackClient, userRepo: userRepo, config: config}
}

func (pc *paystackService) updatePaystack(ctx context.Context, userId, paystackCustomerId, payStackcustomerCode string) (uuid.UUID, error) {
	return pc.userRepo.UpdatePaystack(ctx, &user.User{
		ID:                   userId,
		PaystackCustomerId:   sql.NullString{String: paystackCustomerId, Valid: true},
		PaystackCustomerCode: sql.NullString{String: payStackcustomerCode, Valid: true},
		IsVerified:           user.IsVerifiedEnum(1),
	})
}

func (pc *paystackService) CreateSubscriptionPlan(ctx context.Context, subscritpionPlan *subscription.SubscriptionPlanRequest) (*subscription.SubscriptionPlan, error) {
	result, err := pc.payStackClient.CreateSubscriptionPlan(ctx, subscritpionPlan)

	if err != nil {
		return nil, err
	}

	sub := result.ToSubscriptionPlan()
	return &sub, err
}

func (pc *paystackService) VerifySubscription(ctx context.Context, subReq subscription.VerifySubRequest) (*subscription.Subscription, error) {
	subResponse, err := pc.payStackClient.VerifySubscription(ctx, subReq.RefId)

	if err != nil {
		return nil, err
	}

	if subResponse.Data.Status == "success" {
		//update customer code if not set
		claims, ok := ctx.Value(pc.config.JWTClaimsContextKey).(*security.JWTClaim)

		if !ok {
			return nil, http_helper.ErrUnauthorized
		}

		userId := claims.JWTClaimsMain.LoggedInUserId
		customerId := subResponse.Data.Customer.ID
		customerCode := subResponse.Data.Customer.CustomerCode

		//create subscription
		// update user customer code

		_, err = pc.updatePaystack(ctx, userId, fmt.Sprintf("%d", customerId), customerCode)

		if err != nil {
			return nil, err
		}

		sub := subResponse.Data.ToSubscription()
		
		return &sub, nil
	}

	return nil, err
}

func (pc *paystackService) CancelSubscription(ctx context.Context) {

}

func (pc *paystackService) ChangeSubscription(ctx context.Context) {
}

func (pc *paystackService) MakePayment(ctx context.Context) {
}

func (pc *paystackService) GetSubscriptionPlanOfferings(ctx context.Context) ([]*subscription.SubscriptionPlanOffering, int, error) {
	return []*subscription.SubscriptionPlanOffering{}, 0, nil
}

func (pc *paystackService) CreateSubscriptionPlanOffering(ctx context.Context, _ *subscription.SubscriptionPlanOfferingRequest) (string, error) {
	return "", nil
}

func (pc *paystackService) GetSubscription(ctx context.Context, _ string) (*subscription.Subscription, error) {
	return nil, nil
}

func (pc *paystackService) CreateSubscription(ctx context.Context, _ *subscription.Subscription) (*subscription.Subscription, error) {
	return nil, nil
}