package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appContent "bitbucket.org/hofng/hofApp/internal/application/content"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
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

// CreateMessage handles POST requests to create a new audio message.
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

// ListMessages handles GET requests to list audio messages with optional filters.
func (h *ContentHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := appContent.MessageListFilter{
		Search:   q.Get("search"),
		SeriesID: q.Get("series_id"),
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

// GetMessage handles GET requests to retrieve a single audio message by ID.
func (h *ContentHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "message_id"))
	if err != nil {
		response.BadRequest(w, "invalid message_id")
		return
	}

	m, err := h.svc.GetMessage(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, m)
}

// UpdateMessage handles PUT/PATCH requests to update an existing audio message.
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

// DeleteMessage handles DELETE requests to remove an audio message by ID.
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

// CreateSeries handles POST requests to create a new audio series.
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

// ListSeries handles GET requests to list all audio series.
func (h *ContentHandler) ListSeries(w http.ResponseWriter, r *http.Request) {
	series, total, err := h.svc.ListSeries(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSONList(w, http.StatusOK, series, total)
}

// GetSeries handles GET requests to retrieve a single audio series by ID.
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

// UpdateSeries handles PUT/PATCH requests to update an existing audio series.
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

// DeleteSeries handles DELETE requests to remove an audio series by ID.
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

// CreateMeditation handles POST requests to create a new meditation.
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

// ListMeditations handles GET requests to list meditations.
func (h *ContentHandler) ListMeditations(w http.ResponseWriter, r *http.Request) {
	admin := r.URL.Query().Get("admin") == "true"
	meditations, err := h.svc.ListMeditations(r.Context(), admin)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSONList(w, http.StatusOK, meditations, int64(len(meditations)))
}

// GetMeditation handles GET requests to retrieve a single meditation by ID.
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

// UpdateMeditation handles PUT/PATCH requests to update an existing meditation.
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

// DeleteMeditation handles DELETE requests to remove a meditation by ID.
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

// GetHomepage handles GET requests to retrieve the homepage content aggregation.
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
