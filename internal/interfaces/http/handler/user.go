package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appUser "bitbucket.org/hofng/hofApp/internal/application/user"
	domainUser "bitbucket.org/hofng/hofApp/internal/domain/user"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/middleware"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
)

// UserHandler groups all user-management HTTP endpoints.
type UserHandler struct {
	svc          appUser.Service
	serverURL    string
	templatePath string
	// parseToken extracts a user UUID from a signed JWT string.
	// Injected so the handler stays free of the security package import.
	parseToken func(token string) (uuid.UUID, error)
}

// NewUserHandler creates a UserHandler.
// parseToken should call jwtSvc.Parse and return the user UUID from the claims.
func NewUserHandler(
	svc appUser.Service,
	serverURL string,
	templatePath string,
	parseToken func(token string) (uuid.UUID, error),
) *UserHandler {
	return &UserHandler{
		svc:          svc,
		serverURL:    serverURL,
		templatePath: templatePath,
		parseToken:   parseToken,
	}
}

// SignUp godoc
// @Summary      Create a new user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body appUser.SignUpRequest true "Sign up payload"
// @Success      201 {object} domainUser.User
// @Router       /session/sign_up [post]
func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req appUser.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	u, err := h.svc.SignUp(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, appUser.ToUserResponse(u))
}

// UpdateProfile godoc
// @Summary      Update the authenticated user's profile
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appUser.UpdateProfileRequest true "Profile fields"
// @Success      200
// @Router       /user/update [post]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req appUser.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.UpdateProfile(r.Context(), userID, req); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "profile updated"})
}

// ForgotPassword godoc
// @Summary      Initiate password reset
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body appUser.ForgotPasswordRequest true "Email"
// @Success      200
// @Router       /session/forgot_password [post]
func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req appUser.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.ForgotPassword(r.Context(), req); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "if that email exists, a reset code has been sent"})
}

// VerifyOTP godoc
// @Summary      Verify a password reset OTP
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body appUser.VerifyOTPRequest true "OTP"
// @Success      200 {object} domainUser.User
// @Router       /session/verify_token [put]
func (h *UserHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req appUser.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	u, err := h.svc.VerifyOTP(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, u)
}

// ResetPassword godoc
// @Summary      Set a new password after OTP verification
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body appUser.ResetPasswordRequest true "New password"
// @Success      200
// @Router       /user/reset_password [post]
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req appUser.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.ResetPassword(r.Context(), req); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}

// ChangePassword godoc
// @Summary      Change password for the authenticated user
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appUser.ChangePasswordRequest true "Passwords"
// @Success      200
// @Router       /user/change_password [post]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req appUser.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.ChangePassword(r.Context(), userID, req); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "password changed"})
}

// ── Roles ─────────────────────────────────────────────────────────────────────

// AssignRoles godoc
// @Summary      Assign roles to a user (admin)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appUser.AssignRolesRequest true "Roles"
// @Success      200
// @Router       /user/roles [post]
func (h *UserHandler) AssignRoles(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req appUser.AssignRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	// Populate from JWT so callers don't need to send user_id in the body.
	req.UserID = userID.String()

	if err := h.svc.AssignRoles(r.Context(), req); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "roles assigned"})
}

// GetRoles godoc
// @Summary      Get roles for the authenticated user
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} domainUser.Role
// @Router       /user/roles [get]
func (h *UserHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	roles, err := h.svc.GetRoles(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, roles)
}

// ── Favourites ────────────────────────────────────────────────────────────────

// AddFavourite godoc
// @Summary      Bookmark an audio message
// @Tags         favourites
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appUser.AddFavouriteRequest true "Message to bookmark"
// @Success      201
// @Router       /user/favourite [post]
func (h *UserHandler) AddFavourite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req appUser.AddFavouriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.AddFavourite(r.Context(), userID, req); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "added to favourites"})
}

// GetFavourites godoc
// @Summary      List the authenticated user's favourites
// @Tags         favourites
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} domainUser.FavouriteMessage
// @Router       /user/favourite/favs [get]
func (h *UserHandler) GetFavourites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	favs, err := h.svc.GetFavourites(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSONList(w, http.StatusOK, favs, int64(len(favs)))
}

// DeleteFavourite godoc
// @Summary      Remove an audio message from favourites
// @Tags         favourites
// @Security     BearerAuth
// @Produce      json
// @Param        message_id path string true "Message UUID"
// @Success      200
// @Router       /user/favourite/delete/{message_id} [delete]
func (h *UserHandler) DeleteFavourite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	messageID, err := uuid.Parse(chi.URLParam(r, "message_id"))
	if err != nil {
		response.BadRequest(w, "invalid message_id")
		return
	}

	if err := h.svc.DeleteFavourite(r.Context(), userID, messageID); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "removed from favourites"})
}

// ── Devices ───────────────────────────────────────────────────────────────────

// RegisterDevice godoc
// @Summary      Register a device for the user identified by email
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        email path string true "User email"
// @Param        body body appUser.DeviceInput true "Device info"
// @Success      201
// @Router       /session/device/{email} [post]
func (h *UserHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var input appUser.DeviceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	rec, err := h.svc.RegisterDevice(r.Context(), userID, input)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, rec)
}

// GetDevices godoc
// @Summary      List all registered devices for the authenticated user
// @Tags         devices
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} domainUser.DeviceRecord
// @Router       /user/devices/all [get]
func (h *UserHandler) GetDevices(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	rec, err := h.svc.GetDevices(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, rec)
}

// DeleteDevice godoc
// @Summary      Remove a registered device
// @Tags         devices
// @Security     BearerAuth
// @Produce      json
// @Param        identifier path string true "Device identifier"
// @Success      200
// @Router       /user/devices/delete/{identifier} [delete]
func (h *UserHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	identifier := chi.URLParam(r, "identifier")
	if err := h.svc.DeleteDevice(r.Context(), userID, identifier); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "device removed"})
}

// UpdateDeviceStatus godoc
// @Summary      Change a device's active/inactive status
// @Tags         devices
// @Security     BearerAuth
// @Produce      json
// @Param        identifier path string true "Device identifier"
// @Param        status     path string true "Status (ACTIVE|INACTIVE)"
// @Success      200
// @Router       /user/devices/update/{identifier}/{status} [put]
func (h *UserHandler) UpdateDeviceStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	identifier := chi.URLParam(r, "identifier")
	status := domainUser.DeviceStatus(chi.URLParam(r, "status"))

	if err := h.svc.UpdateDeviceStatus(r.Context(), userID, identifier, status); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "device status updated"})
}

// ── App version ───────────────────────────────────────────────────────────────

// GetAppVersion godoc
// @Summary      Get the current app version record
// @Tags         app-version
// @Security     BearerAuth
// @Produce      json
// @Param        version_id path string true "Version UUID"
// @Success      200 {object} domainUser.AppVersion
// @Router       /user/app_version/version/{version_id} [get]
func (h *UserHandler) GetAppVersion(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "version_id"))
	if err != nil {
		response.BadRequest(w, "invalid version_id")
		return
	}

	v, err := h.svc.GetAppVersion(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, v)
}

// UpdateAppVersion godoc
// @Summary      Update the promoted app version (admin)
// @Tags         app-version
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appUser.AppVersionUpdateRequest true "Version"
// @Success      200
// @Router       /user/app_version/admin/update [put]
func (h *UserHandler) UpdateAppVersion(w http.ResponseWriter, r *http.Request) {
	var req appUser.AppVersionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.UpdateAppVersion(r.Context(), req); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "app version updated"})
}

// ── Email verification ────────────────────────────────────────────────────────

// SendEmailVerification godoc
// @Summary      Send an email verification link
// @Description  Public endpoint — call this immediately after sign-up to trigger the verification email.
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        body body appUser.SendEmailVerificationRequest true "User email"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /session/send_verify_email [post]
func (h *UserHandler) SendEmailVerification(w http.ResponseWriter, r *http.Request) {
	var req appUser.SendEmailVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.SendEmailVerification(r.Context(), req.Email, h.serverURL); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "verification email sent"})
}

// VerifyEmail godoc
// @Summary      Complete email verification via link token
// @Description  Browser-facing endpoint — opened from the link in the verification email.
//
//	On success or failure it renders a branded HTML page, not JSON.
//
// @Tags         session
// @Produce      html
// @Param        token path string true "JWT embedded in the verification link"
// @Success      200
// @Failure      401
// @Router       /verify_email/{token} [get]
func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	tokenStr := chi.URLParam(r, "token")

	userID, err := h.parseToken(tokenStr)
	if err != nil {
		h.renderHTMLPage(w, http.StatusUnauthorized, "verify_email_error.page.tmpl")
		return
	}

	if err := h.svc.VerifyEmail(r.Context(), userID); err != nil {
		h.renderHTMLPage(w, http.StatusInternalServerError, "verify_email_error.page.tmpl")
		return
	}

	h.renderHTMLPage(w, http.StatusOK, "verify_email_success.page.tmpl")
}

// renderHTMLPage serves a static HTML file from the configured template directory.
func (h *UserHandler) renderHTMLPage(w http.ResponseWriter, status int, filename string) {
	content, err := os.ReadFile(filepath.Join(h.templatePath, filename))
	if err != nil {
		http.Error(w, "page unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(content)
}

