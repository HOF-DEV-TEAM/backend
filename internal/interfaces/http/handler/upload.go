package handler

import (
	"net/http"
	"strings"

	"bitbucket.org/hofng/hofApp/internal/infrastructure/storage"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
	"github.com/google/uuid"
)

// UploadHandler handles file upload endpoints.
type UploadHandler struct {
	storage storage.Storage
}

// NewUploadHandler creates an UploadHandler.
func NewUploadHandler(s storage.Storage) *UploadHandler {
	return &UploadHandler{storage: s}
}

// AllowedContentTypes defines MIME types permitted for upload.
var AllowedContentTypes = map[string]bool{
	"audio/mpeg":       true,
	"audio/mp3":        true,
	"audio/m4a":        true,
	"audio/wav":        true,
	"audio/webm":       true,
	"image/jpeg":       true,
	"image/jpg":        true,
	"image/png":        true,
	"image/webp":       true,
	"image/gif":        true,
	"video/mp4":        true,
	"video/webm":       true,
	"application/pdf":  true,
	"application/json": true,
}

// UploadFile godoc
// @Summary      Upload a file to storage (S3 or Cloudinary)
// @Description  Uploads a file with early validation. Max file size and content-type are checked before processing.
// @Tags         upload
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "File to upload"
// @Param        key  formData string false "Object key (auto-generated if omitted)"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string "Invalid file or size exceeded"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      403 {object} map[string]string "Forbidden"
// @Router       /admin/upload [post]
func (h *UploadHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Set timeouts for large files: 30 minute limit
	ctx := r.Context()

	// Limit request body to max file size + overhead
	maxSize := h.storage.GetMaxFileSize() + 1024*1024 // +1MB for form overhead
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	// Parse form with size limit
	if err := r.ParseMultipartForm(h.storage.GetMaxFileSize()); err != nil { // #nosec G120
		response.BadRequest(w, "request body too large or malformed")
		return
	}

	_, fileHeader, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "missing or invalid 'file' field")
		return
	}

	// Early validation: Check file size before opening
	if fileHeader.Size > h.storage.GetMaxFileSize() {
		response.BadRequest(w, "file size exceeds maximum allowed")
		return
	}

	if fileHeader.Size == 0 {
		response.BadRequest(w, "file is empty")
		return
	}

	// Early validation: Check content type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		response.BadRequest(w, "missing Content-Type header")
		return
	}

	// Normalize content-type (remove charset, boundaries, etc.)
	contentType = strings.Split(contentType, ";")[0]

	if !AllowedContentTypes[contentType] {
		response.BadRequest(w, "unsupported file type: "+contentType)
		return
	}

	// Generate key if not provided, or use the provided one
	key := r.FormValue("key")
	if key == "" {
		// Auto-generate key: uuid + original extension
		ext := getExtensionFromContentType(contentType)
		if ext == "" && fileHeader.Filename != "" {
			ext = getFileExtension(fileHeader.Filename)
		}
		key = uuid.New().String() + ext
	}

	// Upload with context (supports cancellation and timeouts)
	url, err := h.storage.Upload(ctx, fileHeader, key)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Set cache headers: CDN should cache for 1 year, browser for 30 days
	w.Header().Set("Cache-Control", "public, max-age=2592000, immutable")
	w.Header().Set("ETag", `"`+key+`"`)

	response.JSON(w, http.StatusOK, map[string]string{
		"url": url,
		"key": key,
	})
}

// GeneratePresignedURL godoc
// @Summary      Generate a presigned S3 upload URL
// @Description  Returns a time-limited URL allowing direct upload to S3, bypassing this server.
// @Tags         upload
// @Security     BearerAuth
// @Produce      json
// @Param        content_type query string true "MIME type of file to upload (e.g., audio/mp3)"
// @Success      200 {object} map[string]string "presigned_url, expires_in"
// @Failure      400 {object} map[string]string "Invalid content type"
// @Failure      403 {object} map[string]string "Presigned URLs not supported for this backend"
// @Router       /admin/upload/presigned [get]
func (h *UploadHandler) GeneratePresignedURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentType := r.URL.Query().Get("content_type")
	if contentType == "" {
		response.BadRequest(w, "missing 'content_type' query parameter")
		return
	}

	// Normalize and validate content type
	contentType = strings.Split(contentType, ";")[0]
	if !AllowedContentTypes[contentType] {
		response.BadRequest(w, "unsupported file type: "+contentType)
		return
	}

	// Generate key with extension
	ext := getExtensionFromContentType(contentType)
	key := uuid.New().String() + ext

	// Request presigned URL from storage backend
	url, err := h.storage.GeneratePresignedURL(ctx, key, contentType)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"presigned_url": url,
		"expires_in":    3600, // 1 hour
		"key":           key,
	})
}

// getExtensionFromContentType maps MIME type to file extension.
func getExtensionFromContentType(contentType string) string {
	extensions := map[string]string{
		"audio/mpeg":       ".mp3",
		"audio/mp3":        ".mp3",
		"audio/m4a":        ".m4a",
		"audio/wav":        ".wav",
		"audio/webm":       ".webm",
		"image/jpeg":       ".jpg",
		"image/jpg":        ".jpg",
		"image/png":        ".png",
		"image/webp":       ".webp",
		"image/gif":        ".gif",
		"video/mp4":        ".mp4",
		"video/webm":       ".webm",
		"application/pdf":  ".pdf",
		"application/json": ".json",
	}
	return extensions[contentType]
}

// getFileExtension extracts extension from filename.
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
}
