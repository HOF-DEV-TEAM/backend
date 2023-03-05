package paystack

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"bitbucket.org/hofng/hofApp/pkg/subscription"
)

type PaystackSubscriptionResponseData struct {
	Name         string                `json:"name"`
	Interval     string                `json:"interval"`
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

type PaystackSubscriptionResponse struct {
	Status  bool                             `json:"status"`
	Message string                           `json:"message"`
	Data    PaystackSubscriptionResponseData `json:"data"`
}

func (paystackResponse *PaystackSubscriptionResponse) ToSubscriptionPlan() *subscription.SubscriptionPlan {
	data := paystackResponse.Data
	fmt.Println(data, "data")
	plan := &subscription.SubscriptionPlan{		
		Name:     data.Name,
		Code:     data.PlanCode,
		// Freq:     data.Interval,
		Fee:      data.Amount,
		PlanId:   sql.NullString{String: strconv.Itoa(data.ID), Valid: true},
		Currency: data.Currency,
	}

	if _, err := time.Parse(time.RFC3339, data.CreatedAt); err == nil {

		plan.DateAdded = sql.NullString{String: data.CreatedAt, Valid: true}

	}

	if _, err := time.Parse(time.RFC3339, data.UpdatedAt); err == nil {
		plan.LastUpdated = sql.NullString{String: data.UpdatedAt, Valid: true}

	}
	return plan
}
