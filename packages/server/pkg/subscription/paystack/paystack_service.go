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
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Mobile.String,
		}

		paystackUser, err := pc.payStackClient.CreateCustomer(ctx, &customer)
		if err != nil {
			return nil, err
		}

		_, err = pc.updatePaystack(ctx, user.ID, fmt.Sprintf("%d", paystackUser.Data.ID), paystackCustomerCode)

		if err != nil {
			return nil, err
		}
	}

	subReq.Customer = user.PaystackCustomerCode.String
	//get active plan - if there's none create one?
	sub, err := pc.payStackClient.CreateSubscription(ctx, subReq)
	if err != nil {
		return nil, err
	}

	userSub := sub.Data.ToSubscription()

	userSub.UserID = user.ID
	return &userSub, err
}

func (pc *paystackService) CreateSubscriptionPlan(ctx context.Context, subscritpionPlan *subscription.SubscriptionPlanRequest) (*subscription.SubscriptionPlan, error) {
	result, err := pc.payStackClient.CreateSubscriptionPlan(ctx, subscritpionPlan)

	if err != nil {
		return nil, err
	}

	sub := result.ToSubscriptionPlan()
	return &sub, err
}

func (pc *paystackService) VerifySubscription(ctx context.Context, subRef string) (*subscription.Subscription, error) {
	subResponse, err := pc.payStackClient.VerifySubscription(ctx, subRef)

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
