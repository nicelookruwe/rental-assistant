package notifications

import (
	"context"
	"testing"
	"time"

	"github.com/ailina/rental-assistant/internal/contracts"
	"github.com/ailina/rental-assistant/internal/payments"
)

type fakeStore struct {
	seen map[string]bool
	list []Notification
}

func (f *fakeStore) Insert(ctx context.Context, userID, message string, scheduledAt time.Time, dedupeKey string) (Notification, bool, error) {
	if f.seen == nil {
		f.seen = map[string]bool{}
	}
	if f.seen[dedupeKey] {
		return Notification{}, false, nil
	}
	f.seen[dedupeKey] = true
	n := Notification{ID: dedupeKey, Message: message, ScheduledAt: scheduledAt, Status: StatusPending}
	f.list = append(f.list, n)
	return n, true, nil
}

func (f *fakeStore) List(ctx context.Context, userID, status string) ([]Notification, error) {
	return f.list, nil
}

func (f *fakeStore) MarkRead(ctx context.Context, userID, id string) (Notification, error) {
	for i := range f.list {
		if f.list[i].ID == id {
			f.list[i].Status = StatusRead
			return f.list[i], nil
		}
	}
	return Notification{}, nil
}

type fakePayments struct{ items []payments.Payment }

func (f fakePayments) PendingSoon(ctx context.Context, userID string, days int) ([]payments.Payment, error) {
	return f.items, nil
}

type fakeContracts struct{ items []contracts.Contract }

func (f fakeContracts) EndingSoon(ctx context.Context, userID string, days int) ([]contracts.Contract, error) {
	return f.items, nil
}

func newService(store *fakeStore) *Service {
	return NewService(store, fakePayments{items: []payments.Payment{{
		ID:              "pay-1",
		Type:            payments.TypeRent,
		Amount:          45000,
		PaymentDate:     "2026-11-25",
		PropertyAddress: "Тверская 12, кв. 5",
		TenantName:      "Петров П. П.",
	}}}, fakeContracts{items: []contracts.Contract{{
		ID:              "con-1",
		PropertyAddress: "Тверская 12, кв. 5",
		TenantName:      "Петров П. П.",
		EndDate:         "2026-12-01",
	}}})
}

func TestGenerateCreatesReminders(t *testing.T) {
	store := &fakeStore{}
	svc := newService(store)

	created, err := svc.Generate(context.Background(), "user")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("создано %d напоминаний, ожидалось 2", len(created))
	}
	if created[0].ScheduledAt.Hour() != 12 {
		t.Fatalf("время напоминания = %s, ожидался полдень", created[0].ScheduledAt)
	}
}

func TestGenerateDoesNotDuplicate(t *testing.T) {
	store := &fakeStore{}
	svc := newService(store)

	if _, err := svc.Generate(context.Background(), "user"); err != nil {
		t.Fatalf("первый Generate: %v", err)
	}
	again, err := svc.Generate(context.Background(), "user")
	if err != nil {
		t.Fatalf("второй Generate: %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("повторный вызов создал %d дублей", len(again))
	}
}

func TestPaymentMessageIsReadable(t *testing.T) {
	msg := PaymentMessage(payments.Payment{
		Type:            payments.TypeUtility,
		Amount:          15234.5,
		PaymentDate:     "2026-11-20",
		PropertyAddress: "Тверская 12",
	})
	want := "Коммунальный платёж 15 234,50 ₽ по объекту «Тверская 12» ожидается 2026-11-20"
	if msg != want {
		t.Fatalf("текст напоминания:\n получили: %s\n ожидали: %s", msg, want)
	}
}
