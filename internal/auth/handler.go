package auth

import (
	"net/http"

	"github.com/ailina/rental-assistant/pkg/httpx"
)

// Handler — HTTP-слой модуля аутентификации.
type Handler struct {
	svc *Service
}

// NewHandler создаёт обработчики /api/auth/*.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register — POST /api/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	res, err := h.svc.Register(r.Context(), in.Email, in.Password)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, res)
}

// Login — POST /api/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	res, err := h.svc.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// Me — GET /api/auth/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := UserID(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	user, err := h.svc.Profile(r.Context(), userID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}
