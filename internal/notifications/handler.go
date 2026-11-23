package notifications

import (
	"net/http"

	"github.com/ailina/rental-assistant/internal/auth"
	"github.com/ailina/rental-assistant/pkg/httpx"
)

// Handler — HTTP-слой модуля notifications.
type Handler struct {
	svc *Service
}

// NewHandler создаёт обработчики /api/notifications.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List — GET /api/notifications?status=pending.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	items, err := h.svc.List(r.Context(), userID, r.URL.Query().Get("status"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// Generate — POST /api/notifications/generate: создаёт напоминания по текущим данным.
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	created, err := h.svc.Generate(r.Context(), userID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": created, "created": len(created)})
}

// MarkRead — POST /api/notifications/{id}/read.
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
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
	n, err := h.svc.MarkRead(r.Context(), userID, id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, n)
}
