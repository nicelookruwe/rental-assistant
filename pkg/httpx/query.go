package httpx

import (
	"net/http"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// DateLayout — формат дат в API.
const DateLayout = "2006-01-02"

// QueryDate читает необязательный параметр-дату формата ГГГГ-ММ-ДД.
func QueryDate(r *http.Request, name string) (string, error) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return "", nil
	}
	if _, err := time.Parse(DateLayout, v); err != nil {
		return "", apperr.Validation("параметр %s должен быть датой ГГГГ-ММ-ДД", name)
	}
	return v, nil
}
