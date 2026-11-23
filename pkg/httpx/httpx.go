// Package httpx — небольшие помощники HTTP-слоя: JSON-ответы, разбор тела
// запроса, преобразование доменных ошибок в HTTP-коды.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

const maxJSONBody = 1 << 20 // 1 MiB

// ErrorResponse — единый формат ошибки API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// JSON отправляет payload со статусом и заголовком Content-Type.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("не удалось записать ответ", "error", err)
	}
}

// NoContent — 204 без тела.
func NoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

// Decode разбирает JSON-тело запроса: ограничивает размер, запрещает неизвестные
// поля и лишние значения после объекта.
func Decode(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return apperr.Validation("тело запроса пустое")
		}
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return apperr.Validation("тело запроса слишком большое")
		}
		return apperr.Validation("некорректный JSON: %v", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return apperr.Validation("в теле запроса должен быть один JSON-объект")
	}
	return nil
}

// Fail записывает ошибку в ответ, подбирая HTTP-код по категории.
func Fail(w http.ResponseWriter, r *http.Request, err error) {
	e, ok := apperr.As(err)
	if !ok {
		slog.Error("необработанная ошибка", "path", r.URL.Path, "error", err)
		JSON(w, http.StatusInternalServerError, ErrorResponse{Error: "внутренняя ошибка сервера"})
		return
	}
	status := http.StatusInternalServerError
	switch e.Kind {
	case apperr.KindValidation:
		status = http.StatusUnprocessableEntity
	case apperr.KindUnauthorized:
		status = http.StatusUnauthorized
	case apperr.KindForbidden:
		status = http.StatusForbidden
	case apperr.KindNotFound:
		status = http.StatusNotFound
	case apperr.KindConflict:
		status = http.StatusConflict
	}
	if e.Kind == apperr.KindInternal {
		slog.Error("внутренняя ошибка", "path", r.URL.Path, "error", e.Err)
	}
	JSON(w, status, ErrorResponse{Error: e.Message})
}
