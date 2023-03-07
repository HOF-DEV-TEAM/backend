package subscription

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
)

type SubscriptionRequest struct {
	Customer string `json:"customer,omitempty"`
	Plan     string `json:"plan,omitempty"`
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
	ID   string   `json:"id"`
	Type TypeEnum   `json:"type"`
	Freq FreqEnum `json:"freq"`
	Fee  float64  `json:"fee"`
	Code string   `json:"code"`
}

type PlanOfferingResponse struct {
	SubscriptionPlanKey
	Offerings []string `json:"offerings"`
}

type SubscriptionPlanOfferingResponse struct {
	Offerings map[string][]*PlanOfferingResponse `json:"offerings"`
}

func CreateSubscriptionHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createSubscriptionHandler, svc)
}

func createSubscriptionHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	var subReq SubscriptionRequest
	err := json.NewDecoder(r.Body).Decode(&subReq)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	sub, err := svc.(SubscriptionService).CreateSubscription(r.Context(), &subReq)

	if err != nil {
		http_helper.EncodeJSONError(r.Context(), err, wr)
		return
	}

	http_helper.EncodeResult(wr, sub, http.StatusOK)

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

	http_helper.EncodeResult(wr, payload, http.StatusOK)

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
			ID:   plan.SubscriptionPlanID.String,
			Type: plan.Type,
			Code: plan.PlanCode,
			Freq: plan.Freq,
			Fee:  plan.Fee,
		}
		planOfferings[key] = append(planOfferings[key], plan.Name)
	}

	for key, offerings := range planOfferings {
		plans = append(plans, &PlanOfferingResponse{key, offerings})		
	}	

	for _, plan  := range plans {		
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
