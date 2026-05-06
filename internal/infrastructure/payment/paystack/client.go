// Package paystack provides a minimal Paystack REST API client.
package paystack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"go.uber.org/zap"
)

// Client is a minimal Paystack REST API client.
type Client struct {
	baseURL    string
	secretKey  string
	httpClient *http.Client
	log        *zap.Logger
}

// NewClient returns a Paystack Client.
func NewClient(cfg config.PaystackConfig, log *zap.Logger) *Client {
	return &Client{
		baseURL:   cfg.Addr,
		secretKey: cfg.Secret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: log,
	}
}

// ── API types ─────────────────────────────────────────────────────────────────

type initTransactionRequest struct {
	Email     string `json:"email"`
	Amount    int64  `json:"amount"`
	Plan      string `json:"plan,omitempty"`
	Reference string `json:"reference,omitempty"`
}

type initTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// VerifyTransactionResponse represents a verified Paystack transaction payload.
type VerifyTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status   string `json:"status"`
		Customer struct {
			Code string `json:"customer_code"`
			ID   int64  `json:"id"`
		} `json:"customer"`
		Subscription struct {
			SubscriptionCode string `json:"subscription_code"`
			NextPaymentDate  string `json:"next_payment_date"`
			Plan             struct {
				PlanCode string `json:"plan_code"`
			} `json:"plan"`
		} `json:"subscription"`
	} `json:"data"`
}

type disableSubscriptionRequest struct {
	Code  string `json:"code"`
	Token string `json:"token"`
}

type disableSubscriptionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type createCustomerRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type createCustomerResponse struct {
	Status bool `json:"status"`
	Data   struct {
		CustomerCode string `json:"customer_code"`
		ID           int64  `json:"id"`
	} `json:"data"`
}

// ── API methods ───────────────────────────────────────────────────────────────

// InitializeTransaction creates a Paystack payment session.
func (c *Client) InitializeTransaction(ctx context.Context, email string, amount int64, planCode, reference string) (authURL, accessCode, ref string, err error) {
	body := initTransactionRequest{
		Email:     email,
		Amount:    amount,
		Plan:      planCode,
		Reference: reference,
	}
	var resp initTransactionResponse
	if err := c.post(ctx, "/transaction/initialize", body, &resp); err != nil {
		return "", "", "", err
	}
	if !resp.Status {
		return "", "", "", shared.ErrInvalidInput{Message: resp.Message}
	}
	return resp.Data.AuthorizationURL, resp.Data.AccessCode, resp.Data.Reference, nil
}

// VerifyTransaction confirms a completed payment by reference.
func (c *Client) VerifyTransaction(ctx context.Context, reference string) (*VerifyTransactionResponse, error) {
	var resp VerifyTransactionResponse
	if err := c.get(ctx, "/transaction/verify/"+reference, &resp); err != nil {
		return nil, err
	}
	if !resp.Status {
		return nil, shared.ErrInvalidInput{Message: "payment verification failed: transaction not found or unsuccessful"}
	}
	return &resp, nil
}

// DisableSubscription cancels a subscription at Paystack.
func (c *Client) DisableSubscription(ctx context.Context, code, token string) error {
	body := disableSubscriptionRequest{Code: code, Token: token}
	var resp disableSubscriptionResponse
	if err := c.post(ctx, "/subscription/disable", body, &resp); err != nil {
		return err
	}
	if !resp.Status {
		return shared.ErrInvalidInput{Message: resp.Message}
	}
	return nil
}

// CreateCustomer registers a new customer in Paystack.
func (c *Client) CreateCustomer(ctx context.Context, email, firstName, lastName, phone string) (customerCode, customerID string, err error) {
	body := createCustomerRequest{
		Email: email, FirstName: firstName, LastName: lastName, Phone: phone,
	}
	var resp createCustomerResponse
	if err := c.post(ctx, "/customer", body, &resp); err != nil {
		return "", "", err
	}
	if !resp.Status {
		return "", "", shared.ErrInvalidInput{Message: "paystack create customer failed"}
	}
	return resp.Data.CustomerCode, fmt.Sprintf("%d", resp.Data.ID), nil
}

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func (c *Client) post(ctx context.Context, path string, body, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling paystack request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("building paystack request: %w", err)
	}
	c.setHeaders(req)

	return c.do(req, out)
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("building paystack request: %w", err)
	}
	c.setHeaders(req)
	return c.do(req, out)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req) //nolint:gosec // Paystack requests use a configured base URL, not user input.
	if err != nil {
		return fmt.Errorf("paystack HTTP call: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading paystack response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("paystack returned status %d: %s", resp.StatusCode, string(b))
	}

	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("decoding paystack response: %w", err)
	}
	return nil
}
