package tenants

import (
	"net/http"

	"github.com/ailina/rental-assistant/internal/auth"
	"github.com/ailina/rental-assistant/pkg/httpx"
)

// Handler — HTTP-слой модуля tenants.
type Handler struct {
	svc *Service
}

// NewHandler создаёт обработчики /api/tenants.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List — GET /api/tenants.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// Get — GET /api/tenants/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
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
	t, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, t)
}

// Create — POST /api/tenants.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in Input
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	t, err := h.svc.Create(r.Context(), userID, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, t)
}

// Update — PUT /api/tenants/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
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
	var in Input
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	t, err := h.svc.Update(r.Context(), userID, id, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, t)
}

// Delete — DELETE /api/tenants/{id}.
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
