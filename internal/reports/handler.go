package reports

import (
	"net/http"

	"github.com/ailina/rental-assistant/internal/auth"
	"github.com/ailina/rental-assistant/pkg/httpx"
)

// Handler — HTTP-слой модуля reports.
type Handler struct {
	svc *Service
}

// NewHandler создаёт обработчики /api/reports.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Summary — GET /api/reports/summary.
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	summary, err := h.svc.Summary(r.Context(), userID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, summary)
}
