package payments

import (
	"net/http"
	"strconv"

	"github.com/ailina/rental-assistant/internal/auth"
	"github.com/ailina/rental-assistant/pkg/apperr"
	"github.com/ailina/rental-assistant/pkg/httpx"
	"github.com/ailina/rental-assistant/pkg/ids"
)

// Handler — HTTP-слой модуля payments.
type Handler struct {
	svc *Service
}

// NewHandler создаёт обработчики /api/payments.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List — GET /api/payments?contract_id=&property_id=&type=&status=&from=&to=&limit=&offset=.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	q := r.URL.Query()
	f := Filter{
		ContractID: q.Get("contract_id"),
		PropertyID: q.Get("property_id"),
		Type:       q.Get("type"),
		Status:     q.Get("status"),
		Limit:      atoiDefault(q.Get("limit"), 100),
		Offset:     atoiDefault(q.Get("offset"), 0),
	}
	if f.ContractID != "" && !ids.Valid(f.ContractID) {
		httpx.Fail(w, r, apperr.Validation("некорректный contract_id"))
		return
	}
	if f.PropertyID != "" && !ids.Valid(f.PropertyID) {
		httpx.Fail(w, r, apperr.Validation("некорректный property_id"))
		return
	}
	if f.From, err = httpx.QueryDate(r, "from"); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if f.To, err = httpx.QueryDate(r, "to"); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	items, err := h.svc.List(r.Context(), userID, f)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// Get — GET /api/payments/{id}.
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
	p, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, p)
}

// Create — POST /api/payments.
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
	p, err := h.svc.Create(r.Context(), userID, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, p)
}

// Settle — POST /api/payments/{id}/settle: фиксирует оплату новой записью журнала.
func (h *Handler) Settle(w http.ResponseWriter, r *http.Request) {
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
	var in Settlement
	if r.ContentLength > 0 {
		if err := httpx.Decode(w, r, &in); err != nil {
			httpx.Fail(w, r, err)
			return
		}
	}
	p, err := h.svc.Settle(r.Context(), userID, id, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, p)
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
