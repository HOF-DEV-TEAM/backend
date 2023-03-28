package subscription

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
)

type SubscriptionRequest struct {
	Customer string `json:"customer,omitempty"`
	Plan     string `json:"plan,omitempty"`
}

type VerifySubRequest struct {
	PlanId string `json:"plan_id"`
	RefId  string `json:"ref_id"`
}
type SubscriptionPlanRequest struct {
	Type TypeEnum `json:"type,string,omitempty"`
	Name string   `json:"name,omitempty"`
	Freq FreqEnum `json:"interval,string,omitempty"`
	Fee  int      `json:"amount,omitempty"`
}

type SubscriptionOfferingRequest struct {
	Name string `json:"name"`
}

type SubscriptionOfferingResponse struct {
	OfferingID string `json:"offering_id"`
}

type SubscriptionPlanOfferingRequest struct {
	SubscriptionPlanId     string `json:"subscription_plan_id"`
	SubscriptionOfferingId string `json:"subscription_offering_id"`
}

type SubscriptionPlanKey struct {
	ID       string   `json:"id"`
	Type     TypeEnum `json:"type"`
	Freq     FreqEnum `json:"freq"`
	Fee      float64  `json:"fee"`
	Code     string   `json:"code"`
	Currency string   `json:"currency"`
}

type PlanOfferingResponse struct {
	SubscriptionPlanKey
	Offerings []string `json:"offerings"`
}

type SubscriptionPlanOfferingResponse struct {
	Offerings map[string][]*PlanOfferingResponse `json:"offerings"`
}

type PaystackResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type PlanResponseData struct {
	Name         string   `json:"name"`
	Interval     FreqEnum `json:"interval,string,omit_empty"`
	Currency     string   `json:"currency"`
	PlanCode     string   `json:"code"`
	Amount       float64  `json:"amount"`
	SendInvoices bool     `json:"send_invoices"`
	SendSms      bool     `json:"send_sms"`
	IsArchived   bool     `json:"is_archived"`
	ID           int      `json:"id"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
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
	//to verify transaction status //success or failure
	Status string `json:"status"`
}

type SubscriptionJSON struct {
	ID                 string   `json:"id"`
	Status             bool     `json:"status"`
	UserID             string   `json:"user_id"`
	SubscriptionPlanID string   `json:"subscription_plan_id"`
	NextPaymentDate    string   `json:"next_payment_date"`
	Type               TypeEnum `json:"type"`
	Freq               FreqEnum `json:"freq"`
	Fee                float64  `json:"fee"`
	Currency           string   `json:"currency"`
	PlanCode           string   `json:"plan_code"`
}

func (sub *Subscription) ToJSON() *SubscriptionJSON {
	return &SubscriptionJSON{
		ID:                 sub.ID,
		Status:             sub.Status == 1,
		UserID:             sub.UserID,
		NextPaymentDate:    sub.NextPaymentDate.String,
		SubscriptionPlanID: sub.SubscriptionPlanID,
		Type:               sub.Type,
		Freq:               sub.Freq,
		Fee:                sub.Fee,
		PlanCode:           sub.PlanCode,
		Currency:           sub.Currency,
	}
}

type SubscriptionPlanJSON struct {
	ID                     string     `json:"id"`
	Name                   string     `json:"name"`
	Type                   TypeEnum   `json:"int"`
	Freq                   FreqEnum   `json:"freq"`
	Fee                    float64    `json:"float64"`
	Status                 StatusEnum `json:"status"`
	Currency               string     `json:"currency"`
	Code                   string     `json:"code"`
	DateAdded              string     `json:"date_added"`
	LastUpdated            string     `json:"last_updated"`
	PlanId                 string     `json:"plan_id"`
	SubscritpionProviderID string     `json:"subscription_provider_id"`
}

func (sub *SubscriptionPlan) ToJSON() *SubscriptionPlanJSON {
	return &SubscriptionPlanJSON{
		ID:                     sub.ID,
		Name:                   sub.Name,
		Type:                   sub.Type,
		Freq:                   sub.Freq,
		Fee:                    sub.Fee,
		Status:                 sub.Status,
		Currency:               sub.Currency,
		Code:                   sub.Code,
		PlanId:                 sub.PlanId.String,
		DateAdded:              sub.DateAdded.String,
		LastUpdated:            sub.LastUpdated.String,
		SubscritpionProviderID: sub.SubscritpionProviderID.String,
	}
}

type SubscriptionResponse struct {
	PaystackResponse
	Data SubscriptionResponseData `json:"data"`
}

type CustomerResponse struct {
	PaystackResponse
	Data CustomerResponseData `json:"data"`
}

func (paystackResponse *PlanResponse) ToSubscriptionPlan() SubscriptionPlan {
	data := paystackResponse.Data
	plan := SubscriptionPlan{
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

func (subResponse *SubscriptionResponseData) ToSubscription() Subscription {
	sub := Subscription{}

	sub.NextPaymentDate = parseDateTime(subResponse.NextPaymentDate)
	sub.DateAdded = parseDateTime(subResponse.CreatedAt)
	sub.LastUpdated = parseDateTime(subResponse.UpdatedAt)
	return sub
}

func parseDateTime(dateString string) sql.NullString {
	if _, err := time.Parse(time.RFC3339, dateString); err == nil {
		return sql.NullString{String: dateString, Valid: true}
	}
	return sql.NullString{}
}

func CreateSubscriptionPlanHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createSubscriptionPlanHandler, svc)

}

func createSubscriptionPlanHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	var subscriptionPlan SubscriptionPlanRequest

	err := json.NewDecoder(r.Body).Decode(&subscriptionPlan)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	payload, err := svc.(SubscriptionService).CreateSubscriptionPlan(r.Context(), &subscriptionPlan)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, payload.ToJSON(), http.StatusOK)
}

func CreateSubscriptionOfferingHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createSubscriptionOfferingHandler, svc)
}

func createSubscriptionOfferingHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	var offering SubscriptionOfferingRequest

	err := json.NewDecoder(r.Body).Decode(&offering)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	offeringId, err := svc.(Service).CreateSubscriptionOffering(r.Context(), &offering)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, SubscriptionOfferingResponse{OfferingID: offeringId}, http.StatusOK)

}

func GetSubscriptionPlanOfferingsHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getSubscriptionPlanOfferingsHandler, svc)

}

func getSubscriptionPlanOfferingsHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	result, _, err := svc.(Service).GetSubscriptionPlanOfferings(r.Context())

	planOfferings := make(map[SubscriptionPlanKey][]string)

	plans := make([]*PlanOfferingResponse, 0)
	monthlyPlans := make(map[string][]*PlanOfferingResponse)

	for _, plan := range result {
		key := SubscriptionPlanKey{
			ID:       plan.SubscriptionPlanID.String,
			Type:     plan.Type,
			Code:     plan.PlanCode,
			Freq:     plan.Freq,
			Fee:      plan.Fee,
			Currency: plan.Currency,
		}
		planOfferings[key] = append(planOfferings[key], plan.Name)
	}

	for key, offerings := range planOfferings {
		plans = append(plans, &PlanOfferingResponse{key, offerings})
	}

	for _, plan := range plans {
		key := plan.Freq.String()
		monthlyPlans[key] = append(monthlyPlans[key], plan)
	}

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, SubscriptionPlanOfferingResponse{monthlyPlans}, http.StatusOK)

}

func CreateSubscriptionPlanOfferingHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createSubscriptionPlanOfferingHandler, svc)

}

func createSubscriptionPlanOfferingHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	var offering SubscriptionPlanOfferingRequest

	err := json.NewDecoder(r.Body).Decode(&offering)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	planOfferingId, err := svc.(Service).CreateSubscriptionPlanOffering(r.Context(), &offering)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, planOfferingId, http.StatusOK)
}

func VerifySubscriptionHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(verifySubscriptionHandler, svc)

}

func verifySubscriptionHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	var subReq VerifySubRequest
	err := json.NewDecoder(r.Body).Decode(&subReq)

	s := svc.(Service)
	ctx := r.Context()
	_, err = s.VerifySubscription(ctx, subReq)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	payload := struct {
		Msg string `json:"msg"`
	}{
		Msg: "Verification Succesful",
	}

	http_helper.EncodeResult(wr, http_helper.DefaultResponse{Body: payload, Code: 200, Success: true}, http.StatusOK)
}

func CreateSubscriptionHookHandler(event Event) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createSubscriptionHookHandler, event)

}

func createSubscriptionHookHandler(wr http.ResponseWriter, r *http.Request, evt interface{}) {

	bytes, errRead := io.ReadAll(r.Body)

	if errRead != nil {
		http_helper.EncodeJSONError(r.Context(), errRead, wr)
		return
	}

	var event SubscriptionEvent

	json.Unmarshal(bytes, &event)

	subEvent := evt.(Event)

	subEvent.HandleEvent(r.Context(), &event)
}
