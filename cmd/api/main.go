// Команда api запускает REST API Rental Assistant.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ailina/rental-assistant/internal/api"
	"github.com/ailina/rental-assistant/internal/auth"
	"github.com/ailina/rental-assistant/internal/config"
	"github.com/ailina/rental-assistant/internal/contracts"
	"github.com/ailina/rental-assistant/internal/documents"
	"github.com/ailina/rental-assistant/internal/meters"
	"github.com/ailina/rental-assistant/internal/notifications"
	"github.com/ailina/rental-assistant/internal/payments"
	"github.com/ailina/rental-assistant/internal/properties"
	"github.com/ailina/rental-assistant/internal/reports"
	"github.com/ailina/rental-assistant/internal/tenants"
	"github.com/ailina/rental-assistant/internal/users"
	"github.com/ailina/rental-assistant/migrations"
	"github.com/ailina/rental-assistant/pkg/database"
	"github.com/ailina/rental-assistant/pkg/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("запуск не удался", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logLevel := slog.LevelInfo
	if cfg.Env == "development" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if cfg.RunMigrations {
		if err := database.Migrate(ctx, pool, migrations.FS); err != nil {
			return err
		}
		slog.Info("миграции применены")
	}

	files, err := storage.NewLocal(cfg.FilesDir, "/api/documents")
	if err != nil {
		return err
	}

	tokens, err := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL)
	if err != nil {
		return err
	}

	usersRepo := users.NewRepository(pool)
	propertiesRepo := properties.NewRepository(pool)
	tenantsRepo := tenants.NewRepository(pool)
	contractsRepo := contracts.NewRepository(pool)
	paymentsRepo := payments.NewRepository(pool)
	metersRepo := meters.NewRepository(pool)
	documentsRepo := documents.NewRepository(pool)
	notificationsRepo := notifications.NewRepository(pool)
	reportsRepo := reports.NewRepository(pool)

	contractsSvc := contracts.NewService(contractsRepo)
	paymentsSvc := payments.NewService(paymentsRepo)
	notificationsSvc := notifications.NewService(notificationsRepo, paymentsSvc, contractsSvc)

	handlers := api.Handlers{
		Auth:          auth.NewHandler(auth.NewService(usersRepo, tokens)),
		Properties:    properties.NewHandler(properties.NewService(propertiesRepo)),
		Tenants:       tenants.NewHandler(tenants.NewService(tenantsRepo)),
		Contracts:     contracts.NewHandler(contractsSvc),
		Payments:      payments.NewHandler(paymentsSvc),
		Meters:        meters.NewHandler(meters.NewService(metersRepo)),
		Documents:     documents.NewHandler(documents.NewService(documentsRepo, files)),
		Notifications: notifications.NewHandler(notificationsSvc),
		Reports:       reports.NewHandler(reports.NewService(reportsRepo, contractsSvc)),
	}

	router := api.NewRouter(api.Options{
		Tokens:         tokens,
		Handlers:       handlers,
		AllowedOrigins: cfg.AllowedOrigins,
		WebDir:         cfg.WebDir,
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           withRequestLog(router),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("API слушает", "addr", cfg.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("останавливаюсь по сигналу")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// withRequestLog пишет строку в лог по каждому запросу и ловит паники.
func withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if p := recover(); p != nil {
				slog.Error("паника в обработчике", "path", r.URL.Path, "panic", p)
				if !rec.wrote {
					http.Error(rec, `{"error":"внутренняя ошибка сервера"}`, http.StatusInternalServerError)
				}
			}
			slog.Info("запрос",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds())
		}()
		next.ServeHTTP(rec, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if !s.wrote {
		s.status = code
		s.wrote = true
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	s.wrote = true
	return s.ResponseWriter.Write(b)
}
