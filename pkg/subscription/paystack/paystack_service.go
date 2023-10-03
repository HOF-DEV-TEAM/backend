package paystack

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"log"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"bitbucket.org/hofng/hofApp/pkg/user"
	"github.com/gofrs/uuid"
)

type PaystackService struct {
	payStackClient *PayStackClientHttp
	userRepo       user.Repository
	subRepo        subscription.Repository
	config         *security.SecurityConfig
}

func NewPaystackService(payStackClient *PayStackClientHttp, subRepo subscription.Repository, userRepo user.Repository, config *security.SecurityConfig) *PaystackService {
	return &PaystackService{payStackClient: payStackClient, subRepo: subRepo, userRepo: userRepo, config: config}
}

func (pc *PaystackService) updatePaystack(ctx context.Context, userId, paystackCustomerId, payStackcustomerCode string) (uuid.UUID, error) {
	return pc.userRepo.UpdatePaystack(ctx, &user.User{
		ID:                   userId,
		PaystackCustomerId:   sql.NullString{String: paystackCustomerId, Valid: true},
		PaystackCustomerCode: sql.NullString{String: payStackcustomerCode, Valid: true},
		IsVerified:           user.IsVerifiedEnum(1),
	})
}

func (pc *PaystackService) CreateSubscriptionPlan(ctx context.Context, subscritpionPlan *subscription.SubscriptionPlanRequest) (*subscription.SubscriptionPlan, error) {
	result, err := pc.payStackClient.CreateSubscriptionPlan(ctx, subscritpionPlan)

	if err != nil {
		return nil, err
	}

	sub := result.ToSubscriptionPlan()
	return &sub, err
}

func (pc *PaystackService) VerifySubscription(ctx context.Context, subReq subscription.VerifySubRequest) (*subscription.Subscription, error) {
	subResponse, err := pc.payStackClient.VerifySubscription(ctx, subReq.RefId)
	if err != nil || subResponse == nil {
		return nil, err
	}

	pc.payStackClient.logger.Info("subResponse: ", zap.Any("data", subResponse))
	if subResponse.Data.Status == "success" {
		//update customer code if not set
		claims, ok := ctx.Value(pc.config.JWTClaimsContextKey).(*security.JWTClaim[any])

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

func (p *PaystackService) HandleInvoiceUpdate(ctx context.Context, eventResponse *EventResponse) error {
	subData := eventResponse.Data.Subscription
	log.Println("Invoice: ", subData)
	//todo run both functions concurrently in a goroutine
	user, err := p.userRepo.GetByCustomerCode(ctx, eventResponse.Data.Customer.CustomerCode)

	if err != nil || user == nil {
		return err
	}

	//get additional information from paystack
	existingSub, err := p.payStackClient.GetSubscription(ctx, subData.SubscriptionCode)

	if err != nil || existingSub == nil {
		return err
	}

	//check if subscription exists locally
	sub := &subscription.Subscription{UserID: user.ID}

	subResult, err := p.subRepo.GetSubscription(ctx, sub)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if subResult != nil {
		now := sql.NullString{
			String: time.Now().Format(time.RFC3339),
			Valid:  true,
		}

		newSub := &subscription.Subscription{
			SubCode:         subData.SubscriptionCode,
			NextPaymentDate: parseDateTime(subData.NextPaymentDate),
			LastUpdated:     now,
			Status:          1,
		}

		//subscription already exists; update next payment date and subscription
		_, err := p.subRepo.UpdateSubscription(ctx, user.ID, newSub)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PaystackService) HandleSubscriptionCreate(ctx context.Context, eventResponse *EventResponse) error {
	p.payStackClient.logger.Info("SubsscriptionCreateEvent", zap.Any("sub response", eventResponse.Data.Subscription))
	p.payStackClient.logger.Info("SubsscriptionCreateEvent", zap.Any("all response", *eventResponse))

	subData := eventResponse.Data
	//todo run both functions concurrently in a goroutine
	user, err := p.userRepo.GetByCustomerCode(ctx, eventResponse.Data.Customer.CustomerCode)
	if err != nil || user == nil {
		return err
	}

	subPlan, err := p.subRepo.GetPlan(ctx, subData.Plan.PlanCode)
	//subplan exists at this point
	if err != nil {
		return err
	}
	//check if subscription exists locally
	sub := &subscription.Subscription{UserID: user.ID, SubCode: subData.SubscriptionCode}

	subResult, err := p.subRepo.GetSubscription(ctx, sub)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	now := sql.NullString{
		String: time.Now().Format(time.RFC3339),
		Valid:  true,
	}

	newSub := &subscription.Subscription{
		SubCode:         subData.SubscriptionCode,
		NextPaymentDate: parseDateTime(subData.NextPaymentDate),
		LastUpdated:     now,
		Status:          1,
	}

	if subResult != nil {
		//subscription already exists; update next payment date and subscription
		_, err := p.subRepo.UpdateSubscription(ctx, user.ID, newSub)
		if err != nil {
			return err
		}
		return nil
	}

	//create new sub
	newSub.DateAdded = now
	newSub.UserID = user.ID
	newSub.SubscriptionPlanID = subPlan.ID
	newSub.NextPaymentDate = sql.NullString{
		String: subData.PaystackSubscription.NextPaymentDate,
		Valid:  true,
	}
	_, err = p.subRepo.CreateSubscription(ctx, newSub)

	if err != nil {
		return err
	}
	return nil
}

func (p *PaystackService) HandleCancelSubscription(ctx context.Context, eventResponse *EventResponse) error {
	subData := eventResponse.Data
	//todo run both functions concurrently in a goroutine
	user, err := p.userRepo.GetByCustomerCode(ctx, eventResponse.Data.Customer.CustomerCode)

	if err != nil || user == nil {
		return err
	}

	sub := &subscription.Subscription{UserID: user.ID, SubCode: subData.SubscriptionCode}

	subResult, err := p.subRepo.GetSubscription(ctx, sub)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	newSub := &subscription.Subscription{
		NextPaymentDate: sql.NullString{Valid: true},
		LastUpdated: sql.NullString{
			String: time.Now().Format(time.RFC3339),
			Valid:  true,
		},
		Status: 2,
	}

	if subResult != nil {
		//subscription already exists; update next payment date and subscription
		_, err := p.subRepo.UpdateSubscription(ctx, user.ID, newSub)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (pc *PaystackService) InitializeTransaction(ctx context.Context, req subscription.InitializePaystackTransaction) (*subscription.TransactionInitializationResponse, error) {
	result, err := pc.payStackClient.InitializeTransaction(ctx, &req)

	if err != nil {
		return nil, err
	}

	return result, err
}
