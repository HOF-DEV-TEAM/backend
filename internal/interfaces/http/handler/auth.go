package handler

import (
	"encoding/json"
	"net/http"

	appAuth "bitbucket.org/hofng/hofApp/internal/application/auth"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
)

// AuthHandler groups all authentication HTTP endpoints.
type AuthHandler struct {
	svc appAuth.Service
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(svc appAuth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// SignIn godoc
// @Summary      Sign in with email and password
// @Description  Include the optional `device` field to register/refresh the device on login.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body appAuth.LoginRequest true "Login credentials (device field optional)"
// @Success      200 {object} appAuth.SessionResponse
// @Router       /session/sign_in [post]
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req appAuth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	session, err := h.svc.Login(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, session)
}

// AdminSignIn godoc
// @Summary      Admin sign in
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body appAuth.AdminLoginRequest true "Admin credentials"
// @Success      200 {object} appAuth.SessionResponse
// @Router       /session/sign_in/admin [post]
func (h *AuthHandler) AdminSignIn(w http.ResponseWriter, r *http.Request) {
	var req appAuth.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	session, err := h.svc.AdminLogin(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, session)
}

// Authenticate godoc
// @Summary      Refresh a session using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body appAuth.AuthenticateRequest true "Tokens"
// @Success      200 {object} appAuth.SessionResponse
// @Router       /session/authenticate [post]
func (h *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req appAuth.AuthenticateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	session, err := h.svc.Authenticate(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, session)
}
