package handler

import (
	"net/http"

	"bitbucket.org/hofng/hofApp/internal/infrastructure/storage"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
)

// UploadHandler handles file upload endpoints.
type UploadHandler struct {
	storage *storage.S3Storage
}

// NewUploadHandler creates an UploadHandler.
func NewUploadHandler(s *storage.S3Storage) *UploadHandler {
	return &UploadHandler{storage: s}
}

// UploadFile godoc
// @Summary      Upload a file to S3
// @Tags         upload
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "File to upload"
// @Param        key  formData string true "S3 object key"
// @Success      200 {object} map[string]string
// @Router       /upload [post]
func (h *UploadHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.BadRequest(w, "file too large or malformed form")
		return
	}

	_, fileHeader, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "missing file field")
		return
	}

	key := r.FormValue("key")
	if key == "" {
		key = fileHeader.Filename
	}

	url, err := h.storage.Upload(r.Context(), fileHeader, key)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"url": url})
}
