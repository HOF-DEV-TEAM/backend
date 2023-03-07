package paystack

import (
	"context"
	"database/sql"
	"fmt"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"bitbucket.org/hofng/hofApp/pkg/user"
)

type paystackService struct {
	payStackClient *PayStackClientHttp
	userRepo user.Repository
	config *security.SecurityConfig
}

func NewPaystackService(payStackClient *PayStackClientHttp, userRepo user.Repository, config *security.SecurityConfig) subscription.SubscriptionService {
	return &paystackService{payStackClient: payStackClient, userRepo: userRepo, config: config}
}

func (pc *paystackService) CreateSubscription(ctx context.Context, subReq *subscription.SubscriptionRequest) (*subscription.Subscription, error) {
	claims, ok := ctx.Value(pc.config.JWTClaimsContextKey).(*security.JWTClaim)

	if !ok {
		return nil, http_helper.ErrUnauthorized
	}
	//check if user is subscribed to the same plan - return  plan if true
	user, err := pc.userRepo.GetById(ctx, claims.JWTClaimsMain.LoggedInUserId)
		
	if err != nil {
		return nil, err
	}

	var paystackCustomerCode string

	if !user.PaystackCustomerCode.Valid {
		customer := PaystackCustomer{
			Email: user.Email,
			FirstName: user.FirstName,
			LastName: user.LastName,
			Phone: user.Mobile.String,
		}

		paystackUser, err := pc.payStackClient.CreateCustomer(ctx, &customer)
		if err != nil {
			return nil, err
		}
		paystackCustomerCode = paystackUser.Data.CustomerCode

		user.PaystackCustomerCode = sql.NullString{String: paystackCustomerCode, Valid: true}
		user.PaystackCustomerId = sql.NullString{String: fmt.Sprintf("%d", paystackUser.Data.ID), Valid: true}
		

		_, err = pc.userRepo.UpdatePaystack(ctx, user)

		if err != nil {
			return nil, err
		}		
	}
	
	subReq.Customer = user.PaystackCustomerCode.String
	//get active plan - if there's none create one?
	sub, err := pc.payStackClient.CreateSubscription(ctx, subReq)
	if err  != nil {
		return nil, err
	}

	userSub := sub.ToSubscription()

	userSub.UserID = user.ID
	return userSub, err
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

func (pc *paystackService) GetSubscriptionPlanOfferings(ctx context.Context) ([]*subscription.SubscriptionPlanOffering, int, error) {
	return []*subscription.SubscriptionPlanOffering{}, 0, nil
}

func (pc *paystackService) CreateSubscriptionPlanOffering(ctx context.Context, _ *subscription.SubscriptionPlanOfferingRequest) (string, error) {
	return "", nil
}