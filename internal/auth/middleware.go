package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

type contextKey string

const userIDKey contextKey = "auth.user_id"

// Middleware требует заголовок Authorization: Bearer <JWT> и кладёт ID
// пользователя в контекст запроса.
func Middleware(tokens *TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || strings.TrimSpace(token) == "" {
				writeUnauthorized(w)
				return
			}
			userID, err := tokens.Parse(strings.TrimSpace(token))
			if err != nil {
				writeUnauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	http.Error(w, `{"error":"требуется авторизация"}`, http.StatusUnauthorized)
}

// UserID достаёт ID пользователя из контекста; без него возвращает 401.
func UserID(ctx context.Context) (string, error) {
	id, ok := ctx.Value(userIDKey).(string)
	if !ok || id == "" {
		return "", apperr.Unauthorized("требуется авторизация")
	}
	return id, nil
}
