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

func (r *PayStackClientHttp) CreateSubscription(ctx context.Context, subRequest *subscription.SubscriptionRequest) (*PaystackCustomerSubscriptionResponse, error) {
	resp, err := r.doPostSubscription(ctx, subRequest)

	if err != nil {
		r.logger.Error("msg", zap.String("paystack subscription", err.Error()))
	}
	defer r.close(ctx, resp)

	bytes, errRead := io.ReadAll(resp.Body)

	if errRead != nil {
		return nil, errRead
	}

	var response PaystackCustomerSubscriptionResponse

	json.Unmarshal(bytes, &response)
	r.logger.Info("msg", zap.String(response.Message, ""))

	if !response.Status {
		return nil, errors.New(response.Message)
	}
	return &response, nil
}

func (r *PayStackClientHttp) GetSubscription(ctx context.Context, idOrCode string) (*PaystackCustomerSubscriptionResponse, error) {
	resp, err := r.goGetSubscription(ctx, idOrCode)

	if err != nil {
		r.logger.Error("msg", zap.String("paystack get suscription", err.Error()))
		return nil, err
	}

	defer r.close(ctx, resp)

	bytes, errRead := io.ReadAll(resp.Body)

	if errRead != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	var response PaystackCustomerSubscriptionResponse

	json.Unmarshal(bytes, &response)

	r.logger.Info("msg", zap.String(response.Message, ""))

	if !response.Status {
		return nil, errors.New(response.Message)
	}

	return &response, nil
}

func (r *PayStackClientHttp) goGetSubscription(ctx context.Context, idOrCode string) (*http.Response, error) {
	url := fmt.Sprintf(
		"%s/subscription/%s",
		r.config.Addr,
		idOrCode,
	)

	headerValues, err := r.getHeaders(ctx)

	if err != nil {
		return nil, err
	}

	return r.httpCaller.DoGet(ctx, headerValues, url)
}

func (r *PayStackClientHttp) CreateSubscriptionPlan(ctx context.Context, planInfo *subscription.SubscriptionPlanRequest) (*PaystackPlanResponse, error) {
	//Paystack fee 100 or greater
	planRequest := *planInfo
	planRequest.Fee = planInfo.Fee * 100

	resp, err := r.doPostSubscriptionPlan(ctx, &planRequest)

	if err != nil {
		r.logger.Error("msg", zap.String("paystack subscription", err.Error()))
		return nil, err
	}

	defer r.close(ctx, resp)

	bytes, errRead := io.ReadAll(resp.Body)

	if errRead != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	var response PaystackPlanResponse

	json.Unmarshal(bytes, &response)

	r.logger.Info("msg", zap.String(response.Message, ""))

	if !response.Status {
		return nil, errors.New(response.Message)
	}

	return &response, nil
}

func (r *PayStackClientHttp) doPostSubscription(ctx context.Context, subRequest *subscription.SubscriptionRequest) (*http.Response, error) {
	url := fmt.Sprintf(
		"%s/subscription",
		r.config.Addr,
	)

	body, err := json.Marshal(subRequest)

	if err != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	headerValues, err := r.getHeaders(ctx)

	if err != nil {
		return nil, err
	}

	return r.httpCaller.DoPost(ctx, headerValues, url, body)
}

func (r *PayStackClientHttp) VerifySubscription(ctx context.Context, subRef string) (*PaystackVerifySubscriptionResponse, error) {
	resp, err := r.doVerifySubscription(ctx, subRef)

	if err != nil {
		r.logger.Error("msg", zap.String("paystack subscription verification", err.Error()))
		return nil, err
	}

	defer r.close(ctx, resp)

	bytes, errRead := io.ReadAll(resp.Body)

	if errRead != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	var response PaystackVerifySubscriptionResponse

	err = json.Unmarshal(bytes, &response)
	if err != nil {
		r.logger.Error("msg", zap.String("paystack subscription verification", err.Error()))
		return nil, err
	}

	r.logger.Info("VerifySubscription", zap.String(response.Message, ""))
	r.logger.Info("VerifySubscription", zap.Any("response", response))

	if !response.Status {
		return nil, errors.New(response.Message)
	}

	return &response, nil
}

func (r *PayStackClientHttp) doVerifySubscription(ctx context.Context, subRef string) (*http.Response, error) {
	url := fmt.Sprintf(
		"%s/transaction/verify/%s",
		r.config.Addr,
		subRef,
	)

	r.logger.Info("msg", zap.String("calling verify subscription", url))
	headerValues, err := r.getHeaders(ctx)

	if err != nil {
		return nil, err
	}

	return r.httpCaller.DoGet(ctx, headerValues, url)
}

func (r *PayStackClientHttp) InitializeTransaction(ctx context.Context, req *subscription.InitializePaystackTransaction) (*subscription.TransactionInitializationResponse, error) {
	req.Channels = []string{"card"}
	resp, err := r.doInitializeTransaction(ctx, req)
	if err != nil {
		r.logger.Error("msg", zap.String("paystack initialising payment", err.Error()))
		return nil, err
	}
	defer r.close(ctx, resp)

	bytes, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	var response subscription.TransactionInitializationResponse

	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return nil, err
	}

	r.logger.Info("msg", zap.String(response.Message, ""))

	if !response.Status {
		return nil, errors.New(response.Message)
	}

	return &response, nil
}

func (r *PayStackClientHttp) doInitializeTransaction(ctx context.Context, req *subscription.InitializePaystackTransaction) (*http.Response, error) {
	url := fmt.Sprintf(
		"%s/transaction/initialize",
		r.config.Addr,
	)

	r.logger.Info("msg", zap.String("calling initialize transaction", url))
	body, err := json.Marshal(req)

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

func (r *PayStackClientHttp) CreateCustomer(ctx context.Context, customer *PaystackCustomer) (*PaystackCustomerResponse, error) {
	resp, err := r.doPostCustomer(ctx, customer)

	if err != nil {
		r.logger.Error("msg", zap.String("paystack customer creation", err.Error()))
		return nil, err
	}

	defer r.close(ctx, resp)

	bytes, errRead := io.ReadAll(resp.Body)

	if errRead != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	var response PaystackCustomerResponse

	json.Unmarshal(bytes, &response)

	r.logger.Info("msg", zap.String(response.Message, ""))

	if !response.Status {
		return nil, errors.New(response.Message)
	}

	return &response, nil
}

func (r *PayStackClientHttp) doPostCustomer(ctx context.Context, customer *PaystackCustomer) (*http.Response, error) {
	url := fmt.Sprintf(
		"%s/customer",
		r.config.Addr,
	)

	body, err := json.Marshal(customer)

	if err != nil {
		return nil, http_helper.ErrInvalidRequest
	}

	headerValues, err := r.getHeaders(ctx)

	if err != nil {
		return nil, err
	}

	return r.httpCaller.DoPost(ctx, headerValues, url, body)
}
