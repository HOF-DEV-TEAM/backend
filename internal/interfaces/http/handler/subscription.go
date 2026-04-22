package handler

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	appSub "bitbucket.org/hofng/hofApp/internal/application/subscription"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/middleware"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SubscriptionHandler groups all subscription-related HTTP endpoints.
type SubscriptionHandler struct {
	svc             appSub.Service
	paystackSecret  string
	log             *zap.Logger
}

// NewSubscriptionHandler creates a SubscriptionHandler.
func NewSubscriptionHandler(svc appSub.Service, paystackSecret string, log *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc, paystackSecret: paystackSecret, log: log}
}

// ── Plans ─────────────────────────────────────────────────────────────────────

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

func (h *SubscriptionHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, total, err := h.svc.ListPlans(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSONList(w, http.StatusOK, plans, total)
}

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

func (h *SubscriptionHandler) ListOfferings(w http.ResponseWriter, r *http.Request) {
	offerings, total, err := h.svc.ListOfferings(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSONList(w, http.StatusOK, offerings, total)
}

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

func (h *SubscriptionHandler) ListPlanOfferings(w http.ResponseWriter, r *http.Request) {
	pos, total, err := h.svc.ListPlanOfferings(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSONList(w, http.StatusOK, pos, total)
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs, total, err := h.svc.ListSubscriptions(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSONList(w, http.StatusOK, subs, total)
}

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
		mac.Write(body)
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
