package subscription

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type SubscriptionRequest struct {
	Customer string `json:"customer,omitempty"`
	Plan     string `json:"plan,omitempty"`
}

type VerifySubRequest struct {
	PlanId string `json:"plan_id"`
	RefId  string `json:"ref_id"`
}

type InitializePaystackTransaction struct {
	Email    string   `json:"email"`
	Amount   string   `json:"amount"`
	Plan     string   `json:"plan"`
	Channels []string `json:"channels"`
}

type TransactionInitializationRequest struct {
	PlanID string `json:"plan_id"`
}

type TransactionInitializationResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationUrl string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

type DisableSubscriptionRequest struct {
	Code  string `json:"code"`
	Token string `json:"token"`
}

type DisableSubscriptionPayload struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
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

type SubscriptionJSON struct {
	ID                 string   `json:"id"`
	Status             bool     `json:"status"`
	SubStatusCode      int      `json:"sub_status_code"`
	UserID             string   `json:"user_id"`
	SubscriptionPlanID string   `json:"subscription_plan_id"`
	NextPaymentDate    string   `json:"next_payment_date"`
	LastUpdated        string   `json:"last_updated"`
	DateAdded          string   `json:"date_added"`
	SubCode            string   `json:"sub_code"`
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
		SubStatusCode:      sub.Status,
		UserID:             sub.UserID,
		NextPaymentDate:    sub.NextPaymentDate.String,
		LastUpdated:        sub.LastUpdated.String,
		DateAdded:          sub.DateAdded.String,
		SubCode:            sub.SubCode,
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
	Type                   TypeEnum   `json:"type"`
	Freq                   FreqEnum   `json:"freq"`
	Fee                    float64    `json:"fee"`
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

func CreateSubscriptionPlanHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createSubscriptionPlanHandler, svc)

}

// Get subscriptions
func GetSubscriptionsHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getSubscriptionHandler, svc)

}

func getSubscriptionHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	payload, _, err := svc.(SubscriptionService).GetSubscriptions(r.Context())

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, http_helper.DefaultResponse{Body: payload, Code: 200, Success: true}, http.StatusOK)
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

func DeleteSubscriptionOfferingHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(deleteSubscriptionOfferingHandler, svc)
}

func deleteSubscriptionOfferingHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	subscriptionOfferingID := chi.URLParam(r, "offering_id")

	result, err := svc.(Service).DeleteSubscriptionOfferingByID(r.Context(), subscriptionOfferingID)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, result, http.StatusOK)
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

func InitializeTransactionPlanHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(initializeTransactionPlanHandler, svc)
}

func initializeTransactionPlanHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	var request TransactionInitializationRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}
	log.Println("initializeTransactionPlanHandler request: ", request)
	payload, err := svc.(SubscriptionService).InitializeTransaction(r.Context(), request)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, payload, http.StatusOK)
}

func DisableSubscriptionHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(disableSubscriptionHandler, svc)
}

func disableSubscriptionHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	subCode := chi.URLParam(r, "code")

	log.Println("disableSubscription request")
	payload, err := svc.(SubscriptionService).DisableSubscription(r.Context(), subCode)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, payload, http.StatusOK)
}

func GetSubscriptionPlansHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getSubscriptionPlansHandler, svc)
}

func getSubscriptionPlansHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	payload, _, err := svc.(Service).GetSubscriptionPlans(r.Context())

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, http_helper.DefaultResponse{Body: payload, Code: 200, Success: true}, http.StatusOK)
}

func GetSubscriptionPlanByIdHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(getSubscriptionPlanByIdHandler, svc)
}

func getSubscriptionPlanByIdHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	subPlanId := chi.URLParam(r, "id")

	payload, err := svc.(Service).GetSubscriptionPlanById(r.Context(), subPlanId)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, http_helper.DefaultResponse{Body: payload.ToJSON(), Code: 200, Success: true}, http.StatusOK)
}

func DeleteSubscriptionPlanHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(deleteSubscriptionPlanHandler, svc)
}

func deleteSubscriptionPlanHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	subPlanId := chi.URLParam(r, "id")

	result, err := svc.(Service).DeleteSubscriptionPlanById(r.Context(), subPlanId)
	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}
	http_helper.EncodeResult(wr, result, http.StatusOK)
}

func GetOfferingsHandler(s Service) (fn http.HandlerFunc) {
	fn = func(w http.ResponseWriter, r *http.Request) {
		result, _, err := s.GetOfferings(r.Context())

		if err != nil {
			http_helper.EncodeJSONError(r.Context(), err, w)
			return
		}
		http_helper.EncodeResult(w, result, http.StatusOK)
	}
	return
}
