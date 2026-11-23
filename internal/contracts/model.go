// Package contracts — договоры аренды.
package contracts

import (
	"strings"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
	"github.com/ailina/rental-assistant/pkg/ids"
)

// Статусы договора.
const (
	StatusDraft      = "draft"
	StatusActive     = "active"
	StatusTerminated = "terminated"
	StatusExpired    = "expired"
)

// DateLayout — формат дат в API.
const DateLayout = "2006-01-02"

// Contract — договор аренды с подписями объекта и арендатора для интерфейса.
type Contract struct {
	ID              string    `json:"id"`
	PropertyID      string    `json:"property_id"`
	TenantID        string    `json:"tenant_id"`
	PropertyAddress string    `json:"property_address"`
	TenantName      string    `json:"tenant_name"`
	StartDate       string    `json:"start_date"`
	EndDate         string    `json:"end_date"`
	RentAmount      float64   `json:"rent_amount"`
	Deposit         float64   `json:"deposit"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Input — данные создания и обновления договора.
type Input struct {
	PropertyID string  `json:"property_id"`
	TenantID   string  `json:"tenant_id"`
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	RentAmount float64 `json:"rent_amount"`
	Deposit    float64 `json:"deposit"`
	Status     string  `json:"status"`
}

// Normalize приводит строки к каноническому виду и подставляет статус по умолчанию.
func (in *Input) Normalize() {
	in.PropertyID = strings.TrimSpace(in.PropertyID)
	in.TenantID = strings.TrimSpace(in.TenantID)
	in.StartDate = strings.TrimSpace(in.StartDate)
	in.EndDate = strings.TrimSpace(in.EndDate)
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	if in.Status == "" {
		in.Status = StatusActive
	}
}

// Validate проверяет обязательные поля, даты, суммы и статус.
func (in Input) Validate() error {
	if !ids.Valid(in.PropertyID) {
		return apperr.Validation("выберите квартиру для договора")
	}
	if !ids.Valid(in.TenantID) {
		return apperr.Validation("выберите арендатора для договора")
	}
	start, err := time.Parse(DateLayout, in.StartDate)
	if err != nil {
		return apperr.Validation("дата начала должна быть в формате ГГГГ-ММ-ДД")
	}
	end, err := time.Parse(DateLayout, in.EndDate)
	if err != nil {
		return apperr.Validation("дата окончания должна быть в формате ГГГГ-ММ-ДД")
	}
	if !end.After(start) {
		return apperr.Validation("дата окончания должна быть позже даты начала")
	}
	if end.Sub(start) > 50*365*24*time.Hour {
		return apperr.Validation("срок договора не может превышать 50 лет")
	}
	if in.RentAmount < 0 {
		return apperr.Validation("арендная плата не может быть отрицательной")
	}
	if in.Deposit < 0 {
		return apperr.Validation("залог не может быть отрицательным")
	}
	switch in.Status {
	case StatusDraft, StatusActive, StatusTerminated, StatusExpired:
	default:
		return apperr.Validation("неизвестный статус договора: %s", in.Status)
	}
	return nil
}
