// Package payments — журнал арендных и коммунальных платежей.
package payments

import (
	"strings"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
	"github.com/ailina/rental-assistant/pkg/ids"
)

// Типы платежей.
const (
	TypeRent    = "rent"
	TypeUtility = "utility"
	TypeDeposit = "deposit"
	TypeRepair  = "repair"
	TypeOther   = "other"
)

// Статусы платежей.
const (
	StatusPending  = "pending"
	StatusPaid     = "paid"
	StatusOverdue  = "overdue"
	StatusCanceled = "canceled"
)

// DateLayout — формат дат в API.
const DateLayout = "2006-01-02"

// Payment — запись журнала платежей.
// Поле Status приходит из БД уже «вычисленным»: планируемый платёж с прошедшей
// датой отдаётся как overdue, но сама строка не перезаписывается.
type Payment struct {
	ID              string    `json:"id"`
	ContractID      string    `json:"contract_id"`
	PropertyAddress string    `json:"property_address"`
	TenantName      string    `json:"tenant_name"`
	Amount          float64   `json:"amount"`
	PaymentDate     string    `json:"payment_date"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	Description     string    `json:"description"`
	ParentID        *string   `json:"parent_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// Input — новая запись журнала.
type Input struct {
	ContractID  string  `json:"contract_id"`
	Amount      float64 `json:"amount"`
	PaymentDate string  `json:"payment_date"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
}

// Normalize чистит строки и подставляет значения по умолчанию.
func (in *Input) Normalize() {
	in.ContractID = strings.TrimSpace(in.ContractID)
	in.PaymentDate = strings.TrimSpace(in.PaymentDate)
	in.Type = strings.ToLower(strings.TrimSpace(in.Type))
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	in.Description = strings.TrimSpace(in.Description)
	if in.Status == "" {
		in.Status = StatusPaid
	}
}

// Validate проверяет суммы, дату, тип и статус.
func (in Input) Validate() error {
	if !ids.Valid(in.ContractID) {
		return apperr.Validation("выберите договор для платежа")
	}
	if in.Amount <= 0 {
		return apperr.Validation("сумма платежа должна быть больше нуля")
	}
	if in.Amount > 100_000_000 {
		return apperr.Validation("сумма платежа выглядит нереалистично")
	}
	if _, err := time.Parse(DateLayout, in.PaymentDate); err != nil {
		return apperr.Validation("дата платежа должна быть в формате ГГГГ-ММ-ДД")
	}
	switch in.Type {
	case TypeRent, TypeUtility, TypeDeposit, TypeRepair, TypeOther:
	default:
		return apperr.Validation("неизвестный тип платежа: %s", in.Type)
	}
	switch in.Status {
	case StatusPending, StatusPaid, StatusCanceled:
	case StatusOverdue:
		return apperr.Validation("статус «просрочен» вычисляется автоматически")
	default:
		return apperr.Validation("неизвестный статус платежа: %s", in.Status)
	}
	if len([]rune(in.Description)) > 500 {
		return apperr.Validation("описание длиннее 500 символов")
	}
	return nil
}
