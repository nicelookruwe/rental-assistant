// Package api собирает HTTP-маршруты приложения.
package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ailina/rental-assistant/internal/auth"
	"github.com/ailina/rental-assistant/internal/contracts"
	"github.com/ailina/rental-assistant/internal/documents"
	"github.com/ailina/rental-assistant/internal/meters"
	"github.com/ailina/rental-assistant/internal/notifications"
	"github.com/ailina/rental-assistant/internal/payments"
	"github.com/ailina/rental-assistant/internal/properties"
	"github.com/ailina/rental-assistant/internal/reports"
	"github.com/ailina/rental-assistant/internal/tenants"
	"github.com/ailina/rental-assistant/pkg/httpx"
)

// Handlers — все HTTP-обработчики приложения.
type Handlers struct {
	Auth          *auth.Handler
	Properties    *properties.Handler
	Tenants       *tenants.Handler
	Contracts     *contracts.Handler
	Payments      *payments.Handler
	Meters        *meters.Handler
	Documents     *documents.Handler
	Notifications *notifications.Handler
	Reports       *reports.Handler
}

// Options — настройки роутера.
type Options struct {
	Tokens         *auth.TokenManager
	Handlers       Handlers
	AllowedOrigins []string
	WebDir         string
}

// NewRouter собирает маршруты REST API и (если есть сборка фронтенда) SPA.
func NewRouter(opts Options) http.Handler {
	mux := http.NewServeMux()
	h := opts.Handlers

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Публичные маршруты.
	mux.HandleFunc("POST /api/auth/register", h.Auth.Register)
	mux.HandleFunc("POST /api/auth/login", h.Auth.Login)

	// Маршруты, требующие JWT.
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/auth/me", h.Auth.Me)

	protected.HandleFunc("GET /api/properties", h.Properties.List)
	protected.HandleFunc("POST /api/properties", h.Properties.Create)
	protected.HandleFunc("GET /api/properties/{id}", h.Properties.Get)
	protected.HandleFunc("PUT /api/properties/{id}", h.Properties.Update)
	protected.HandleFunc("DELETE /api/properties/{id}", h.Properties.Delete)

	protected.HandleFunc("GET /api/tenants", h.Tenants.List)
	protected.HandleFunc("POST /api/tenants", h.Tenants.Create)
	protected.HandleFunc("GET /api/tenants/{id}", h.Tenants.Get)
	protected.HandleFunc("PUT /api/tenants/{id}", h.Tenants.Update)
	protected.HandleFunc("DELETE /api/tenants/{id}", h.Tenants.Delete)

	protected.HandleFunc("GET /api/contracts", h.Contracts.List)
	protected.HandleFunc("POST /api/contracts", h.Contracts.Create)
	protected.HandleFunc("GET /api/contracts/{id}", h.Contracts.Get)
	protected.HandleFunc("PUT /api/contracts/{id}", h.Contracts.Update)
	protected.HandleFunc("DELETE /api/contracts/{id}", h.Contracts.Delete)

	protected.HandleFunc("GET /api/payments", h.Payments.List)
	protected.HandleFunc("POST /api/payments", h.Payments.Create)
	protected.HandleFunc("GET /api/payments/{id}", h.Payments.Get)
	protected.HandleFunc("POST /api/payments/{id}/settle", h.Payments.Settle)

	protected.HandleFunc("GET /api/meters", h.Meters.List)
	protected.HandleFunc("POST /api/meters", h.Meters.Create)
	protected.HandleFunc("GET /api/meters/latest", h.Meters.Latest)
	protected.HandleFunc("GET /api/meters/{id}", h.Meters.Get)

	protected.HandleFunc("GET /api/documents", h.Documents.List)
	protected.HandleFunc("POST /api/documents", h.Documents.Upload)
	protected.HandleFunc("GET /api/documents/{id}/file", h.Documents.Download)
	protected.HandleFunc("DELETE /api/documents/{id}", h.Documents.Delete)

	protected.HandleFunc("GET /api/notifications", h.Notifications.List)
	protected.HandleFunc("POST /api/notifications/generate", h.Notifications.Generate)
	protected.HandleFunc("POST /api/notifications/{id}/read", h.Notifications.MarkRead)

	protected.HandleFunc("GET /api/reports/summary", h.Reports.Summary)

	mux.Handle("/api/", auth.Middleware(opts.Tokens)(protected))

	// Собранный фронтенд: статика и SPA-fallback на index.html.
	if opts.WebDir != "" {
		if info, err := os.Stat(filepath.Join(opts.WebDir, "index.html")); err == nil && !info.IsDir() {
			mux.Handle("/", spaHandler(opts.WebDir))
		}
	}

	return cors(opts.AllowedOrigins)(mux)
}

func spaHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	index := filepath.Join(dir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if p := filepath.Join(dir, filepath.Clean("/"+r.URL.Path)); fileExists(p) {
			files.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, index)
	})
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// cors разрешает запросы с указанных источников (для dev-фронтенда на Vite).
func cors(origins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(origins))
	for _, o := range origins {
		allowed[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (allowed[origin] || allowed["*"]) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "600")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
