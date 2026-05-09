package handler

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	appSub "bitbucket.org/hofng/hofApp/internal/application/subscription"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/middleware"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
)

// SubscriptionHandler groups all subscription-related HTTP endpoints.
type SubscriptionHandler struct {
	svc            appSub.Service
	paystackSecret string
	log            *zap.Logger
}

// NewSubscriptionHandler creates a SubscriptionHandler.
func NewSubscriptionHandler(svc appSub.Service, paystackSecret string, log *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc, paystackSecret: paystackSecret, log: log}
}

// ── Plans ─────────────────────────────────────────────────────────────────────

// CreatePlan godoc
// @Summary      Create a subscription plan (admin only)
// @Tags         subscription
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appSub.CreatePlanRequest true "Plan payload"
// @Success      201
// @Router       /admin/subscription/plan [post]
// CreatePlan handles POST requests to create a new subscription plan.
func (h *SubscriptionHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req appSub.CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	plan, err := h.svc.CreatePlan(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, plan)
}

// ListPlans godoc
// @Summary      List subscription plans
// @Tags         subscription
// @Security     BearerAuth
// @Produce      json
// @Success      200
// @Router       /subscription/plan [get]
// ListPlans handles GET requests to list all subscription plans.
func (h *SubscriptionHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, total, err := h.svc.ListPlans(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSONList(w, http.StatusOK, plans, total)
}

// GetPlan godoc
// @Summary      Get a subscription plan by ID
// @Tags         subscription
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Plan ID"
// @Success      200
// @Router       /subscription/plan/{id} [get]
// GetPlan handles GET requests to retrieve a subscription plan by ID.
func (h *SubscriptionHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid plan id")
		return
	}

	plan, err := h.svc.GetPlan(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, plan)
}

// DeletePlan godoc
// @Summary      Delete a subscription plan (admin only)
// @Tags         subscription
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Plan ID"
// @Success      200
// @Router       /admin/subscription/plan/{id} [delete]
// DeletePlan handles DELETE requests to remove a subscription plan by ID.
func (h *SubscriptionHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid plan id")
		return
	}

	if err := h.svc.DeletePlan(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "plan deleted"})
}

// ── Offerings ─────────────────────────────────────────────────────────────────

// CreateOffering godoc
// @Summary      Create a subscription offering (admin only)
// @Tags         subscription
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appSub.CreateOfferingRequest true "Offering payload"
// @Success      201
// @Router       /admin/subscription/offering [post]
// CreateOffering handles POST requests to create a new subscription offering.
func (h *SubscriptionHandler) CreateOffering(w http.ResponseWriter, r *http.Request) {
	var req appSub.CreateOfferingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	o, err := h.svc.CreateOffering(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, o)
}

// ListOfferings godoc
// @Summary      List subscription offerings
// @Tags         subscription
// @Security     BearerAuth
// @Produce      json
// @Success      200
// @Router       /subscription/offering [get]
// ListOfferings handles GET requests to list all subscription offerings.
func (h *SubscriptionHandler) ListOfferings(w http.ResponseWriter, r *http.Request) {
	offerings, total, err := h.svc.ListOfferings(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSONList(w, http.StatusOK, offerings, total)
}

// DeleteOffering godoc
// @Summary      Delete a subscription offering (admin only)
// @Tags         subscription
// @Security     BearerAuth
// @Produce      json
// @Param        offering_id path string true "Offering ID"
// @Success      200
// @Router       /admin/subscription/offering/delete/{offering_id} [delete]
// DeleteOffering handles DELETE requests to remove a subscription offering by ID.
func (h *SubscriptionHandler) DeleteOffering(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "offering_id"))
	if err != nil {
		response.BadRequest(w, "invalid offering_id")
		return
	}

	if err := h.svc.DeleteOffering(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "offering deleted"})
}

// ── Plan offerings ────────────────────────────────────────────────────────────

// CreatePlanOffering godoc
// @Summary      Create a plan-offering mapping (admin only)
// @Tags         subscription
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appSub.CreatePlanOfferingRequest true "Plan offering payload"
// @Success      201
// @Router       /admin/subscription/plan/offering [post]
// CreatePlanOffering handles POST requests to link an offering to a subscription plan.
func (h *SubscriptionHandler) CreatePlanOffering(w http.ResponseWriter, r *http.Request) {
	var req appSub.CreatePlanOfferingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	po, err := h.svc.CreatePlanOffering(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, po)
}

// ListPlanOfferings godoc
// @Summary      List plan-offering mappings
// @Tags         subscription
// @Security     BearerAuth
// @Produce      json
// @Success      200
// @Router       /subscription/plan/offering [get]
// ListPlanOfferings handles GET requests to list all plan-offering associations.
func (h *SubscriptionHandler) ListPlanOfferings(w http.ResponseWriter, r *http.Request) {
	pos, total, err := h.svc.ListPlanOfferings(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSONList(w, http.StatusOK, pos, total)
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

// ListSubscriptions godoc
// @Summary      List subscriptions
// @Tags         subscription
// @Security     BearerAuth
// @Produce      json
// @Success      200
// @Router       /subscription [get]
// ListSubscriptions handles GET requests to list all user subscriptions.
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs, total, err := h.svc.ListSubscriptions(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSONList(w, http.StatusOK, subs, total)
}

// VerifySubscription godoc
// @Summary      Verify a subscription payment
// @Tags         subscription
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appSub.VerifySubscriptionRequest true "Verification payload"
// @Success      200
// @Router       /subscription/verify [post]
// VerifySubscription handles POST requests to verify and activate a user subscription.
func (h *SubscriptionHandler) VerifySubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req appSub.VerifySubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	sub, err := h.svc.VerifySubscription(r.Context(), userID, req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, sub)
}

// InitializeTransaction godoc
// @Summary      Initialize a Paystack transaction
// @Tags         subscription
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appSub.InitTransactionRequest true "Transaction payload"
// @Success      200
// @Router       /subscription/transaction [post]
// InitializeTransaction handles POST requests to initiate a Paystack payment transaction.
func (h *SubscriptionHandler) InitializeTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req appSub.InitTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	txResp, err := h.svc.InitializeTransaction(r.Context(), userID, req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, txResp)
}

// DisableSubscription godoc
// @Summary      Disable a subscription
// @Tags         subscription
// @Security     BearerAuth
// @Produce      json
// @Param        code path string true "Subscription code"
// @Param        token query string true "Paystack token"
// @Success      200
// @Router       /subscription/disable/{code} [delete]
// DisableSubscription handles POST requests to cancel a user's active subscription.
func (h *SubscriptionHandler) DisableSubscription(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	token := r.URL.Query().Get("token")

	resp, err := h.svc.DisableSubscription(r.Context(), appSub.DisableSubscriptionRequest{
		Code: code, Token: token,
	})
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// PaystackWebhook receives and processes Paystack webhook events.
// @Summary      Handle Paystack webhook events
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Param        body body appSub.WebhookEvent true "Webhook payload"
// @Success      200
// @Router       /subscription/webhook [post]
// Always responds 200 to prevent Paystack retries for app-level errors.
func (h *SubscriptionHandler) PaystackWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify HMAC-SHA512 signature when secret is configured.
	if h.paystackSecret != "" {
		sig := r.Header.Get("X-Paystack-Signature")
		mac := hmac.New(sha512.New, []byte(h.paystackSecret))
		_, _ = mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(sig), []byte(expected)) {
			h.log.Warn("paystack webhook: invalid signature")
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	var event appSub.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		h.log.Warn("paystack webhook: unmarshal error", zap.Error(err))
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.svc.HandleWebhookEvent(r.Context(), &event); err != nil {
		h.log.Error("paystack webhook: handler error", zap.Error(err))
	}

	w.WriteHeader(http.StatusOK)
}
