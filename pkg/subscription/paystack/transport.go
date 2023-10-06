package paystack

import (
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"database/sql"
	"strconv"
	"time"
)

// Paystack related types

// common paystack payload
type PaystackResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type PaystackCustomer struct {
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Email        string `json:"email"`
	Phone        string `json:"phone,omitempty"`
	CustomerCode string `json:"customer_code"`
	ID           int    `json:"id"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// Paystack plans
type PaystackPlan struct {
	Name         string                `json:"name"`
	Interval     subscription.FreqEnum `json:"interval,string,omit_empty"`
	Currency     string                `json:"currency"`
	PlanCode     string                `json:"plan_code"`
	Amount       float64               `json:"amount"`
	SendInvoices int                   `json:"send_invoices"`
	SendSms      bool                  `json:"send_sms"`
	IsArchived   bool                  `json:"is_archived"`
	ID           int                   `json:"id"`
	CreatedAt    string                `json:"createdAt"`
	UpdatedAt    string                `json:"updatedAt"`
}

type PaystackPlanResponse struct {
	PaystackResponse
	Data PaystackPlan `json:"data"`
}

// paystack subscription
type PaystackSubscription struct {
	NextPaymentDate  string `json:"next_payment_date"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
	SubscriptionCode string `json:"subscription_code"`
	//to verify transaction status //success or failure
	Status string `json:"status"`
}

type PaystackCustomerSubscription struct {
	Customer PaystackCustomer `json:"customer"`
	Plan     PaystackPlan     `json:"plan"`
	PaystackSubscription
}

type PaystackCustomerSubscriptionResponse struct {
	PaystackResponse
	Data PaystackCustomerSubscription `json:"data"`
}

type VerifyPlanResponse struct {
	Plan     PaystackPlan     `json:"plan_object"`
	Customer PaystackCustomer `json:"customer"`
	PaystackSubscription
}

type PaystackVerifySubscriptionResponse struct {
	PaystackResponse
	Data VerifyPlanResponse `json:"data"`
}

type PaystackCustomerResponse struct {
	PaystackResponse
	Data PaystackCustomer `json:"data"`
}

// utils to convert to generic subscription types
func parseDateTime(dateString string) sql.NullString {
	if _, err := time.Parse(time.RFC3339, dateString); err == nil {
		return sql.NullString{String: dateString, Valid: true}
	}
	return sql.NullString{}
}

func (paystackResponse *PaystackPlanResponse) ToSubscriptionPlan() subscription.SubscriptionPlan {
	data := paystackResponse.Data
	plan := subscription.SubscriptionPlan{
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

func (subResponse *PaystackCustomerSubscription) ToSubscription() subscription.Subscription {
	return subscription.Subscription{
		NextPaymentDate: parseDateTime(subResponse.NextPaymentDate),
		DateAdded:       parseDateTime(subResponse.CreatedAt),
		LastUpdated:     parseDateTime(subResponse.UpdatedAt),
	}
}

func (subResponse *VerifyPlanResponse) ToSubscription() subscription.Subscription {
	return subscription.Subscription{
		NextPaymentDate: parseDateTime(subResponse.NextPaymentDate),
		DateAdded:       parseDateTime(subResponse.CreatedAt),
		LastUpdated:     parseDateTime(subResponse.UpdatedAt),
	}
}
