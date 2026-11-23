package meters

import (
	"net/http"
	"strconv"

	"github.com/ailina/rental-assistant/internal/auth"
	"github.com/ailina/rental-assistant/pkg/apperr"
	"github.com/ailina/rental-assistant/pkg/httpx"
	"github.com/ailina/rental-assistant/pkg/ids"
)

// Handler — HTTP-слой модуля meters.
type Handler struct {
	svc *Service
}

// NewHandler создаёт обработчики /api/meters.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List — GET /api/meters?property_id=&meter_type=&limit=.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	q := r.URL.Query()
	f := Filter{PropertyID: q.Get("property_id"), MeterType: q.Get("meter_type"), Limit: 0}
	if f.PropertyID != "" && !ids.Valid(f.PropertyID) {
		httpx.Fail(w, r, apperr.Validation("некорректный property_id"))
		return
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}
	items, err := h.svc.List(r.Context(), userID, f)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// Latest — GET /api/meters/latest?property_id=.
func (h *Handler) Latest(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	propertyID := r.URL.Query().Get("property_id")
	if propertyID != "" && !ids.Valid(propertyID) {
		httpx.Fail(w, r, apperr.Validation("некорректный property_id"))
		return
	}
	items, err := h.svc.Latest(r.Context(), userID, propertyID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// Get — GET /api/meters/{id}.
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
	m, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, m)
}

// Create — POST /api/meters.
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
	m, err := h.svc.Create(r.Context(), userID, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, m)
}
