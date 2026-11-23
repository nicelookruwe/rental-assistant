package httpx

import (
	"net/http"

	"github.com/ailina/rental-assistant/pkg/apperr"
	"github.com/ailina/rental-assistant/pkg/ids"
)

// PathID достаёт и проверяет параметр пути {id}.
func PathID(r *http.Request) (string, error) {
	id := r.PathValue("id")
	if !ids.Valid(id) {
		return "", apperr.Validation("некорректный идентификатор")
	}
	return id, nil
}
