package handler

import (
	"encoding/json"
	"net/http"

	appSub "bitbucket.org/hofng/hofApp/internal/application/subscription"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
)

// AdminHandler groups admin-only HTTP endpoints.
type AdminHandler struct {
	subSvc appSub.Service
}

// NewAdminHandler creates an AdminHandler.
func NewAdminHandler(subSvc appSub.Service) *AdminHandler {
	return &AdminHandler{subSvc: subSvc}
}

// GetGlobalParameters godoc
// @Summary      Get global application parameters
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Success      200
// @Router       /admin/global [get]
func (h *AdminHandler) GetGlobalParameters(w http.ResponseWriter, r *http.Request) {
	params, err := h.subSvc.GetGlobalParameters(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, params)
}

// UpdateGlobalParameters godoc
// @Summary      Update global application parameters
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appSub.UpdateGlobalParamsRequest true "Parameters"
// @Success      200
// @Router       /admin/global [put]
func (h *AdminHandler) UpdateGlobalParameters(w http.ResponseWriter, r *http.Request) {
	var req appSub.UpdateGlobalParamsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	params, err := h.subSvc.UpdateGlobalParameters(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, params)
}
