package subscription

import (
	"encoding/json"	
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
)

type SubscriptionRequest struct {
	Customer string `json:"customer,omitempty"`
	Plan  	 string `json:"plan,omitempty"`
}

type SubscriptionPlanRequest struct {
	Name string `json:"name,omitempty"`
	Freq string `json:"interval,omitempty"`
	Fee  int    `json:"amount,omitempty"`
}


type SubscriptionOfferingRequest struct {
	Name string `json:"name"`
}

type SubscriptionOfferingResponse struct {
	OfferingID string `json:"offering_id"`
}


type CustomerInfo struct {}

func CreateSubscriptionHandler(svc SubscriptionService) http.HandlerFunc {
	return http_helper.NewHTTPHandler(createSubscriptionHandler, svc)
}

func createSubscriptionHandler(wr http.ResponseWriter, r *http.Request, svc interface{}) {
	var subReq  SubscriptionRequest
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
