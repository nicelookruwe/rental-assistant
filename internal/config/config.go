// Package config читает настройки приложения из переменных окружения.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config — настройки запуска API.
type Config struct {
	Env            string
	Addr           string
	DatabaseURL    string
	JWTSecret      string
	JWTTTL         time.Duration
	FilesDir       string
	WebDir         string
	AllowedOrigins []string
	RunMigrations  bool
}

// Load собирает конфигурацию из окружения и проверяет обязательные поля.
func Load() (Config, error) {
	cfg := Config{
		Env:            env("APP_ENV", "development"),
		Addr:           env("HTTP_ADDR", ":8080"),
		DatabaseURL:    env("DATABASE_URL", ""),
		JWTSecret:      env("JWT_SECRET", ""),
		FilesDir:       env("FILES_DIR", "./data/files"),
		WebDir:         env("WEB_DIR", "./web"),
		AllowedOrigins: splitList(env("CORS_ORIGINS", "http://localhost:5173")),
		RunMigrations:  envBool("RUN_MIGRATIONS", true),
	}

	ttl, err := time.ParseDuration(env("JWT_TTL", "168h"))
	if err != nil || ttl <= 0 {
		return Config{}, fmt.Errorf("config: некорректный JWT_TTL")
	}
	cfg.JWTTTL = ttl

	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("config: не задан DATABASE_URL")
	}
	if len(cfg.JWTSecret) < 16 {
		return Config{}, fmt.Errorf("config: JWT_SECRET должен быть не короче 16 символов")
	}
	if cfg.Env == "production" && len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("config: в production JWT_SECRET должен быть не короче 32 символов")
	}
	return cfg, nil
}

// IsProduction сообщает, что приложение запущено в production.
func (c Config) IsProduction() bool { return c.Env == "production" }

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return def
}

func envBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return b
}

func splitList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
