package payments

import (
	"context"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

type store interface {
	List(ctx context.Context, userID string, f Filter) ([]Payment, error)
	Get(ctx context.Context, userID, id string) (Payment, error)
	Create(ctx context.Context, userID string, in Input, parentID *string) (Payment, error)
	PendingSoon(ctx context.Context, userID string, days int) ([]Payment, error)
}

// Service — бизнес-логика платежей.
type Service struct {
	repo store
	now  func() time.Time
}

// NewService собирает сервис платежей.
func NewService(repo store) *Service {
	return &Service{repo: repo, now: time.Now}
}

// List возвращает платежи по фильтру.
func (s *Service) List(ctx context.Context, userID string, f Filter) ([]Payment, error) {
	return s.repo.List(ctx, userID, f)
}

// Get возвращает платёж.
func (s *Service) Get(ctx context.Context, userID, id string) (Payment, error) {
	return s.repo.Get(ctx, userID, id)
}

// Create добавляет запись в журнал.
func (s *Service) Create(ctx context.Context, userID string, in Input) (Payment, error) {
	in.Normalize()
	if err := in.Validate(); err != nil {
		return Payment{}, err
	}
	return s.repo.Create(ctx, userID, in, nil)
}

// PendingSoon возвращает планируемые платежи со сроком в ближайшие days дней.
func (s *Service) PendingSoon(ctx context.Context, userID string, days int) ([]Payment, error) {
	if days <= 0 {
		days = 7
	}
	if days > 365 {
		days = 365
	}
	return s.repo.PendingSoon(ctx, userID, days)
}

// Settlement — данные для фиксации оплаты планируемого платежа.
type Settlement struct {
	PaymentDate string   `json:"payment_date"`
	Amount      *float64 `json:"amount"`
	Description string   `json:"description"`
}

// Settle фиксирует оплату, создавая новую запись журнала и не меняя исходную:
// планируемый платёж остаётся историей, а факт оплаты — отдельная строка.
func (s *Service) Settle(ctx context.Context, userID, id string, in Settlement) (Payment, error) {
	parent, err := s.repo.Get(ctx, userID, id)
	if err != nil {
		return Payment{}, err
	}
	if parent.ParentID != nil {
		return Payment{}, apperr.Conflict("платёж уже является исполнением другого платежа")
	}
	if parent.Status == StatusPaid {
		return Payment{}, apperr.Conflict("платёж уже оплачен")
	}
	if parent.Status == StatusCanceled {
		return Payment{}, apperr.Conflict("отменённый платёж нельзя оплатить")
	}

	date := in.PaymentDate
	if date == "" {
		date = s.now().Format(DateLayout)
	}
	amount := parent.Amount
	if in.Amount != nil {
		amount = *in.Amount
	}
	description := in.Description
	if description == "" {
		description = "Оплата по платежу от " + parent.PaymentDate
	}

	settlement := Input{
		ContractID:  parent.ContractID,
		Amount:      amount,
		PaymentDate: date,
		Type:        parent.Type,
		Status:      StatusPaid,
		Description: description,
	}
	settlement.Normalize()
	if err := settlement.Validate(); err != nil {
		return Payment{}, err
	}
	return s.repo.Create(ctx, userID, settlement, &parent.ID)
}
