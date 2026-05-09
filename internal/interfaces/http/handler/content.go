package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appContent "bitbucket.org/hofng/hofApp/internal/application/content"
	domainContent "bitbucket.org/hofng/hofApp/internal/domain/content"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
)

// Ensure domain types are available for swagger documentation.
var (
	_ domainContent.AudioMessage
	_ domainContent.AudioSeries
	_ domainContent.Meditation
	_ domainContent.Homepage
)

// ContentHandler groups audio message and series HTTP endpoints.
type ContentHandler struct {
	svc appContent.Service
}

// NewContentHandler creates a ContentHandler.
func NewContentHandler(svc appContent.Service) *ContentHandler {
	return &ContentHandler{svc: svc}
}

// ── Audio messages ────────────────────────────────────────────────────────────

// CreateMessage godoc
// @Summary      Create a new audio message
// @Description  Create a new audio message with optional access control. Access levels: "leaders" (leaders only), "stewards" (stewards+leaders), "members" (all roles). Defaults to "members". URLs are trimmed and checked for uniqueness.
// @Tags         content
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appContent.CreateMessageRequest true "Message payload with optional 'access' field"
// @Success      201 {object} domainContent.AudioMessage
// @Failure      409 {object} map[string]string "Audio URL already exists"
// @Failure      400 {object} map[string]string "Invalid access value or other validation error"
// @Router       /admin/audio_message/ [post]
func (h *ContentHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req appContent.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	m, err := h.svc.CreateMessage(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, m)
}

// ListMessages godoc
// @Summary      List audio messages with optional filters
// @Description  List audio messages filtered by viewer role (access control). Access parameter controls which messages are returned based on role hierarchy: "leaders" sees all, "stewards" sees stewards+members, "members" sees members only.
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Param        search query string false "Search term in title or author"
// @Param        series_id query string false "Filter by series ID"
// @Param        is_free query boolean false "Filter by free status"
// @Param        access query string false "Viewer role for access control (leaders, stewards, members)"
// @Param        page query integer false "Page number" default(1)
// @Param        page_size query integer false "Page size" default(20)
// @Success      200 {array} domainContent.AudioMessage
// @Failure      400 {object} map[string]string "Invalid access parameter"
// @Router       /audio_message/ [get]
func (h *ContentHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// Validate access query param if provided
	if a := q.Get("access"); a != "" {
		switch strings.ToLower(a) {
		case "leaders", "stewards", "members":
			// ok
		default:
			response.BadRequest(w, "invalid access parameter: must be leaders, stewards, or members")
			return
		}
	}
	filter := appContent.MessageListFilter{
		Search:   q.Get("search"),
		SeriesID: q.Get("series_id"),
		Access:   q.Get("access"),
		Page:     intParam(q.Get("page"), 1),
		PageSize: intParam(q.Get("page_size"), 20),
	}
	if q.Get("is_free") == "true" {
		t := true
		filter.IsFree = &t
	} else if q.Get("is_free") == "false" {
		f := false
		filter.IsFree = &f
	}

	messages, total, err := h.svc.ListMessages(r.Context(), filter)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSONList(w, http.StatusOK, messages, total)
}

// GetMessage godoc
// @Summary      Get a single audio message by ID
// @Description  Retrieve a single audio message with optional viewer role for access control. Returns 403 if viewer role lacks permission for the message's access level.
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Param        message_id path string true "Message ID"
// @Param        access query string false "Viewer role for access control (leaders, stewards, members)"
// @Success      200 {object} domainContent.AudioMessage
// @Failure      403 {object} map[string]string "Access denied for viewer role"
// @Failure      400 {object} map[string]string "Invalid access parameter"
// @Failure      404 {object} map[string]string "Message not found"
// @Router       /audio_message/id/message/{message_id} [get]
func (h *ContentHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "message_id"))
	if err != nil {
		response.BadRequest(w, "invalid message_id")
		return
	}
	viewerRole := r.URL.Query().Get("access")
	if viewerRole != "" {
		switch strings.ToLower(viewerRole) {
		case "leaders", "stewards", "members":
			// ok
		default:
			response.BadRequest(w, "invalid access parameter: must be leaders, stewards, or members")
			return
		}
	}

	m, err := h.svc.GetMessage(r.Context(), id, viewerRole)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, m)
}

// UpdateMessage godoc
// @Summary      Update an existing audio message
// @Description  Update an audio message. Can change access level via 'access' field (leaders, stewards, members). URLs are trimmed and checked for uniqueness against other messages.
// @Tags         content
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        message_id path string true "Message ID"
// @Param        body body appContent.UpdateMessageRequest true "Updated message fields (all optional)"
// @Success      200 {object} domainContent.AudioMessage
// @Failure      409 {object} map[string]string "Audio URL already used by another message"
// @Failure      400 {object} map[string]string "Invalid access value or other validation error"
// @Router       /admin/audio_message/update/{message_id} [put]
func (h *ContentHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "message_id"))
	if err != nil {
		response.BadRequest(w, "invalid message_id")
		return
	}

	var req appContent.UpdateMessageRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	m, err := h.svc.UpdateMessage(r.Context(), id, &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, m)
}

// DeleteMessage godoc
// @Summary      Delete an audio message
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Param        message_id path string true "Message ID"
// @Success      200 {object} map[string]string
// @Router       /admin/audio_message/delete/{message_id} [delete]
func (h *ContentHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "message_id"))
	if err != nil {
		response.BadRequest(w, "invalid message_id")
		return
	}

	if err := h.svc.DeleteMessage(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ── Audio series ──────────────────────────────────────────────────────────────

// CreateSeries godoc
// @Summary      Create a new audio series
// @Tags         content
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appContent.CreateSeriesRequest true "Series payload"
// @Success      201 {object} domainContent.AudioSeries
// @Router       /admin/audio_series/ [post]
func (h *ContentHandler) CreateSeries(w http.ResponseWriter, r *http.Request) {
	var req appContent.CreateSeriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	s, err := h.svc.CreateSeries(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, s)
}

// ListSeries godoc
// @Summary      List all audio series
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} domainContent.AudioSeries
// @Router       /audio_series/ [get]
func (h *ContentHandler) ListSeries(w http.ResponseWriter, r *http.Request) {
	series, total, err := h.svc.ListSeries(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSONList(w, http.StatusOK, series, total)
}

// GetSeries godoc
// @Summary      Get a single audio series by ID
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Param        series_id path string true "Series ID"
// @Success      200 {object} domainContent.AudioSeries
// @Router       /audio_series/id/series/{series_id} [get]
func (h *ContentHandler) GetSeries(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "series_id"))
	if err != nil {
		response.BadRequest(w, "invalid series_id")
		return
	}

	s, err := h.svc.GetSeries(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, s)
}

// UpdateSeries godoc
// @Summary      Update an existing audio series
// @Tags         content
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        series_id path string true "Series ID"
// @Param        body body appContent.UpdateSeriesRequest true "Updated series fields"
// @Success      200 {object} domainContent.AudioSeries
// @Router       /admin/audio_series/update/{series_id} [put]
func (h *ContentHandler) UpdateSeries(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "series_id"))
	if err != nil {
		response.BadRequest(w, "invalid series_id")
		return
	}

	var req appContent.UpdateSeriesRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	s, err := h.svc.UpdateSeries(r.Context(), id, &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, s)
}

// DeleteSeries godoc
// @Summary      Delete an audio series
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Param        series_id path string true "Series ID"
// @Success      200 {object} map[string]string
// @Router       /admin/audio_series/delete/{series_id} [delete]
func (h *ContentHandler) DeleteSeries(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "series_id"))
	if err != nil {
		response.BadRequest(w, "invalid series_id")
		return
	}

	if err := h.svc.DeleteSeries(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ── Meditations ───────────────────────────────────────────────────────────────

// CreateMeditation godoc
// @Summary      Create a new meditation
// @Tags         content
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body appContent.CreateMeditationRequest true "Meditation payload"
// @Success      201 {object} domainContent.Meditation
// @Router       /admin/audio_message/meditation [post]
func (h *ContentHandler) CreateMeditation(w http.ResponseWriter, r *http.Request) {
	var req appContent.CreateMeditationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	m, err := h.svc.CreateMeditation(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, m)
}

// ListMeditations godoc
// @Summary      List meditations
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Param        admin query boolean false "Admin view"
// @Success      200 {array} domainContent.Meditation
// @Router       /audio_message/meditations [get]
func (h *ContentHandler) ListMeditations(w http.ResponseWriter, r *http.Request) {
	admin := r.URL.Query().Get("admin") == "true"
	meditations, err := h.svc.ListMeditations(r.Context(), admin)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSONList(w, http.StatusOK, meditations, int64(len(meditations)))
}

// GetMeditation godoc
// @Summary      Get a single meditation by ID
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Param        meditation_id path string true "Meditation ID"
// @Success      200 {object} domainContent.Meditation
// @Router       /audio_message/meditation/{meditation_id} [get]
func (h *ContentHandler) GetMeditation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "meditation_id"))
	if err != nil {
		response.BadRequest(w, "invalid meditation_id")
		return
	}

	m, err := h.svc.GetMeditation(r.Context(), id, false)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, m)
}

// UpdateMeditation godoc
// @Summary      Update an existing meditation
// @Tags         content
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        meditation_id path string true "Meditation ID"
// @Param        body body appContent.UpdateMeditationRequest true "Updated meditation fields"
// @Success      200 {object} domainContent.Meditation
// @Router       /admin/audio_message/meditation/{meditation_id} [put]
func (h *ContentHandler) UpdateMeditation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "meditation_id"))
	if err != nil {
		response.BadRequest(w, "invalid meditation_id")
		return
	}

	var req appContent.UpdateMeditationRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	m, err := h.svc.UpdateMeditation(r.Context(), id, &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, m)
}

// DeleteMeditation godoc
// @Summary      Delete a meditation
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Param        meditation_id path string true "Meditation ID"
// @Success      200 {object} map[string]string
// @Router       /admin/audio_message/meditation/delete/{meditation_id} [delete]
func (h *ContentHandler) DeleteMeditation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "meditation_id"))
	if err != nil {
		response.BadRequest(w, "invalid meditation_id")
		return
	}

	if err := h.svc.DeleteMeditation(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ── Homepage ──────────────────────────────────────────────────────────────────

// GetHomepage godoc
// @Summary      Get the homepage content aggregation
// @Tags         content
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} domainContent.Homepage
// @Router       /audio_series/home [get]
func (h *ContentHandler) GetHomepage(w http.ResponseWriter, r *http.Request) {
	homepage, err := h.svc.GetHomepage(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, homepage)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func intParam(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return def
	}
	return v
}
