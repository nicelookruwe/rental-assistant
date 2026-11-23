package documents

import (
	"mime"
	"net/http"

	"github.com/ailina/rental-assistant/internal/auth"
	"github.com/ailina/rental-assistant/pkg/apperr"
	"github.com/ailina/rental-assistant/pkg/httpx"
)

// Handler — HTTP-слой модуля documents.
type Handler struct {
	svc *Service
}

// NewHandler создаёт обработчики /api/documents.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List — GET /api/documents?property_id=.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	items, err := h.svc.List(r.Context(), userID, r.URL.Query().Get("property_id"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// Upload — POST /api/documents (multipart/form-data: property_id, document_type, file).
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize+(1<<20))
	if err := r.ParseMultipartForm(4 << 20); err != nil {
		httpx.Fail(w, r, apperr.Validation("не удалось прочитать файл (максимум 20 МиБ)"))
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Fail(w, r, apperr.Validation("приложите файл в поле file"))
		return
	}
	defer file.Close()

	if header.Size > MaxUploadSize {
		httpx.Fail(w, r, apperr.Validation("файл больше 20 МиБ"))
		return
	}
	if _, err := ValidateFile(header.Filename, file); err != nil {
		httpx.Fail(w, r, err)
		return
	}

	doc, err := h.svc.Upload(r.Context(), userID,
		r.FormValue("property_id"), r.FormValue("document_type"),
		header.Filename, file, header.Size)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, doc)
}

// Download — GET /api/documents/{id}/file: отдаёт файл после проверки владельца.
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	id, err := httpx.PathID(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	doc, file, err := h.svc.Open(r.Context(), userID, id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": doc.FileName}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, doc.FileName, doc.CreatedAt, file)
}

// Delete — DELETE /api/documents/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	id, err := httpx.PathID(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.NoContent(w)
}
