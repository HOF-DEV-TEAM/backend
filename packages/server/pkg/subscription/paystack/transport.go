package paystack

import (
	"database/sql"
	"strconv"
	"time"

	"bitbucket.org/hofng/hofApp/pkg/subscription"
)

type PaystackResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type PlanResponseData struct {
	Name         string                `json:"name"`
	Interval     subscription.FreqEnum `json:"interval"`
	Currency     string                `json:"currency"`
	PlanCode     string                `json:"plan_code"`
	Amount       float64               `json:"amount"`
	SendInvoices bool                  `json:"send_invoices"`
	SendSms      bool                  `json:"send_sms"`
	IsArchived   bool                  `json:"is_archived"`
	ID           int                   `json:"id"`
	CreatedAt    string                `json:"createdAt"`
	UpdatedAt    string                `json:"updatedAt"`
}

type CustomerResponseData struct {
	Email        string `json:"email"`
	CustomerCode string `json:"customer_code"`
	ID           int    `json:"id"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type PlanResponse struct {
	PaystackResponse
	Data PlanResponseData `json:"data"`
}

type SubscriptionResponseData struct {
	Customer         CustomerResponseData `json:"customer"`
	Plan             PlanResponseData     `json:"plan"`	
	NextPaymentDate  string               `json:"next_payment_date"`
	CreatedAt        string               `json:"createdAt"`
	UpdatedAt        string               `json:"updatedAt"`
	SubscriptionCode string               `json:"subscription_code"`
}

type SubscriptionResponse struct {
	PaystackResponse
	Data SubscriptionResponseData `json:"data"`
}

type CustomerResponse struct {
	PaystackResponse
	Data CustomerResponseData `json:"data"`
}

func (paystackResponse *PlanResponse) ToSubscriptionPlan() *subscription.SubscriptionPlan {
	data := paystackResponse.Data
	plan := &subscription.SubscriptionPlan{
		Name:     data.Name,
		Code:     data.PlanCode,
		Freq:     data.Interval,
		Fee:      data.Amount,
		PlanId:   sql.NullString{String: strconv.Itoa(data.ID), Valid: true},
		Currency: data.Currency,
	}

	plan.DateAdded = parseDateTime(data.CreatedAt)
	plan.LastUpdated = parseDateTime(data.UpdatedAt)

	return plan
}

func (subResponse *SubscriptionResponse) ToSubscription() *subscription.Subscription {
	data := subResponse.Data
	sub := &subscription.Subscription{}

	sub.NextPaymentDate = parseDateTime(data.NextPaymentDate)
	sub.DateAdded = parseDateTime(data.CreatedAt)
	sub.LastUpdated = parseDateTime(data.UpdatedAt)
	return sub
}

func parseDateTime(dateString string) sql.NullString {
	if _, err := time.Parse(time.RFC3339, dateString); err == nil {
		return sql.NullString{String: dateString, Valid: true}
	}
	return sql.NullString{}
}
