package paystack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"go.uber.org/zap"
)

type CustomerInfo struct {
	Customer string `json:"customer"`
	Plan     string `plan:"customer"`
}

type PayStackClientHttp struct {
	logger     *zap.Logger
	config     *config.PaystackConfig
	httpCaller http_helper.HttpClient	
}

func NewPayStackHttpClient(config *config.PaystackConfig, httpCaller http_helper.HttpClient, logger *zap.Logger) *PayStackClientHttp {
	return &PayStackClientHttp{
		config:     config,
		logger:     logger,
		httpCaller: httpCaller,
	}
}

func (r *PayStackClientHttp) close(_ context.Context, resp *http.Response) {
	if resp.Body == nil {
		return
	}

	if err := resp.Body.Close(); err != nil {
		r.logger.Error("msg", zap.String("Error closing reponse body", err.Error()))
	}
}

func (r *PayStackClientHttp) getHeaders(_ context.Context) (http_helper.HttpHeader, error) {
	headerValues := http_helper.HttpHeader{}
	headerValues["Content-Type"] = "application/json"
	headerValues["Authorization"] = fmt.Sprintf(
		"Bearer %s",
		r.config.PaystackSecret,
	)
	return headerValues, nil
}

func (r *PayStackClientHttp) CreateSubscription(ctx context.Context) {
	resp, err := r.doPostSubscription(ctx, CustomerInfo{})

	if err != nil {
		r.logger.Error("msg", zap.String("paystack subscription", err.Error()))
	}
	defer r.close(ctx, resp)

	bytes, errRead := io.ReadAll(resp.Body)

	if errRead != nil {
		return
	}

	var response CustomerInfo

	json.Unmarshal(bytes, &response)
	r.logger.Info("msg", zap.String(response.Customer, ""))
}

func (r *PayStackClientHttp) CreateSubscriptionPlan(ctx context.Context, planInfo *subscription.SubscriptionPlanRequest) (*PaystackSubscriptionResponse, error) {
	resp, err := r.doPostSubscriptionPlan(ctx, planInfo)

	if err != nil {
		r.logger.Error("msg", zap.String("paystack subscription", err.Error()))
		return nil, err
	}

	defer r.close(ctx, resp)

	bytes, errRead := io.ReadAll(resp.Body)

	if errRead != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	var response PaystackSubscriptionResponse

	json.Unmarshal(bytes, &response)

	r.logger.Info("msg", zap.String(response.Message, ""))
	
	if !response.Status {
		return nil, errors.New(response.Message)
	}

	

	return &response, nil
}

func (r *PayStackClientHttp) doPostSubscription(ctx context.Context, customerInfo CustomerInfo) (*http.Response, error) {	
	url := fmt.Sprintf(
		"%s/subscription",
		r.config.Addr,
	)
	
	body, err := json.Marshal(customerInfo)

	if err != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	headerValues, err := r.getHeaders(ctx)

	if err != nil {
		return nil, err
	}

	return r.httpCaller.DoPost(ctx, headerValues, url, body)
}

func (r *PayStackClientHttp) doPostSubscriptionPlan(ctx context.Context, planInfo *subscription.SubscriptionPlanRequest) (*http.Response, error) {
	
	url := fmt.Sprintf(
		"%s/plan",
		r.config.Addr,
	)

	body, err := json.Marshal(planInfo)

	if err != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	headerValues, err := r.getHeaders(ctx)
	
	if err != nil {
		return nil, err
	}

	return r.httpCaller.DoPost(ctx, headerValues, url, body)
}
