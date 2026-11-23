package notifications

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ailina/rental-assistant/internal/contracts"
	"github.com/ailina/rental-assistant/internal/payments"
)

// Окна генерации напоминаний.
const (
	PaymentWindowDays  = 7
	ContractWindowDays = 30
)

// DateLayout — формат дат, в котором модули отдают даты в API.
const DateLayout = "2006-01-02"

type store interface {
	Insert(ctx context.Context, userID, message string, scheduledAt time.Time, dedupeKey string) (Notification, bool, error)
	List(ctx context.Context, userID, status string) ([]Notification, error)
	MarkRead(ctx context.Context, userID, id string) (Notification, error)
}

type paymentSource interface {
	PendingSoon(ctx context.Context, userID string, days int) ([]payments.Payment, error)
}

type contractSource interface {
	EndingSoon(ctx context.Context, userID string, days int) ([]contracts.Contract, error)
}

// Service — бизнес-логика напоминаний.
type Service struct {
	repo      store
	payments  paymentSource
	contracts contractSource
	now       func() time.Time
}

// NewService собирает сервис напоминаний.
func NewService(repo store, pay paymentSource, con contractSource) *Service {
	return &Service{repo: repo, payments: pay, contracts: con, now: time.Now}
}

// List возвращает напоминания; status — фильтр (pending/read) или пусто.
func (s *Service) List(ctx context.Context, userID, status string) ([]Notification, error) {
	if status != "" && status != StatusPending && status != StatusRead {
		status = ""
	}
	return s.repo.List(ctx, userID, status)
}

// MarkRead помечает напоминание прочитанным.
func (s *Service) MarkRead(ctx context.Context, userID, id string) (Notification, error) {
	return s.repo.MarkRead(ctx, userID, id)
}

// Generate строит напоминания о ближайших платежах и истекающих договорах.
// Повторный вызов не создаёт дублей благодаря dedupe_key.
func (s *Service) Generate(ctx context.Context, userID string) ([]Notification, error) {
	created := make([]Notification, 0, 8)

	paymentsDue, err := s.payments.PendingSoon(ctx, userID, PaymentWindowDays)
	if err != nil {
		return nil, err
	}
	for _, p := range paymentsDue {
		n, ok, err := s.repo.Insert(ctx, userID, PaymentMessage(p), atDate(p.PaymentDate, s.now()), "payment:"+p.ID)
		if err != nil {
			return nil, err
		}
		if ok {
			created = append(created, n)
		}
	}

	ending, err := s.contracts.EndingSoon(ctx, userID, ContractWindowDays)
	if err != nil {
		return nil, err
	}
	for _, c := range ending {
		n, ok, err := s.repo.Insert(ctx, userID, ContractMessage(c), atDate(c.EndDate, s.now()), "contract:"+c.ID)
		if err != nil {
			return nil, err
		}
		if ok {
			created = append(created, n)
		}
	}

	return created, nil
}

// PaymentMessage — текст напоминания об оплате.
func PaymentMessage(p payments.Payment) string {
	kind := "Платёж"
	switch p.Type {
	case payments.TypeRent:
		kind = "Аренда"
	case payments.TypeUtility:
		kind = "Коммунальный платёж"
	case payments.TypeDeposit:
		kind = "Депозит"
	}
	who := p.PropertyAddress
	if p.TenantName != "" {
		who = who + ", " + p.TenantName
	}
	return fmt.Sprintf("%s %s ₽ по объекту «%s» ожидается %s", kind, money(p.Amount), who, p.PaymentDate)
}

// ContractMessage — текст напоминания об окончании договора.
func ContractMessage(c contracts.Contract) string {
	return fmt.Sprintf("Договор по объекту «%s» с арендатором %s заканчивается %s",
		c.PropertyAddress, c.TenantName, c.EndDate)
}

func money(v float64) string {
	neg := v < 0
	if neg {
		v = -v
	}

	whole := int64(v)
	kopecks := int64(math.Round((v - float64(whole)) * 100))
	if kopecks == 100 {
		whole++
		kopecks = 0
	}

	digits := strconv.FormatInt(whole, 10)
	var b strings.Builder
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(d)
	}

	out := b.String() + "," + fmt.Sprintf("%02d", kopecks)
	if neg {
		return "−" + out
	}
	return out
}

// atDate превращает дату ГГГГ-ММ-ДД в момент времени (полдень UTC),
// чтобы уведомление не «съезжало» на сутки из-за часового пояса.
func atDate(s string, fallback time.Time) time.Time {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return fallback
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
}
