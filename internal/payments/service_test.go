package payments

import (
	"context"
	"testing"
	"time"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

const (
	testContract = "0b7f2c1e-8a51-4c2b-9f3d-2f7a6c1d4e10"
	testPayment  = "5f3e9a7c-1b2d-4e6f-8a90-7c3b1d2e4f56"
)

// fakeStore — журнал в памяти: хватает для проверки бизнес-логики без БД.
type fakeStore struct {
	items      []Payment
	nextID     string
	createCall int
}

func (f *fakeStore) List(ctx context.Context, userID string, filt Filter) ([]Payment, error) {
	return f.items, nil
}

func (f *fakeStore) Get(ctx context.Context, userID, id string) (Payment, error) {
	for _, p := range f.items {
		if p.ID == id {
			return p, nil
		}
	}
	return Payment{}, apperr.NotFound("платёж не найден")
}

func (f *fakeStore) Create(ctx context.Context, userID string, in Input, parentID *string) (Payment, error) {
	f.createCall++
	created := Payment{
		ID:          f.nextID,
		ContractID:  in.ContractID,
		Amount:      in.Amount,
		PaymentDate: in.PaymentDate,
		Type:        in.Type,
		Status:      in.Status,
		Description: in.Description,
		ParentID:    parentID,
	}
	f.items = append(f.items, created)
	return created, nil
}

func (f *fakeStore) PendingSoon(ctx context.Context, userID string, days int) ([]Payment, error) {
	return nil, nil
}

func TestCreateValidatesInput(t *testing.T) {
	store := &fakeStore{nextID: "new"}
	svc := NewService(store)

	_, err := svc.Create(context.Background(), "user", Input{
		ContractID:  testContract,
		Amount:      0,
		PaymentDate: "2026-11-01",
		Type:        TypeRent,
	})
	if err == nil {
		t.Fatal("платёж с нулевой суммой принят")
	}
	if store.createCall != 0 {
		t.Fatal("запись создана несмотря на ошибку валидации")
	}
}

func TestCreateFillsDefaults(t *testing.T) {
	store := &fakeStore{nextID: "new"}
	svc := NewService(store)

	created, err := svc.Create(context.Background(), "user", Input{
		ContractID:  testContract,
		Amount:      45000,
		PaymentDate: "2026-11-05",
		Type:        TypeRent,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Status != StatusPaid {
		t.Fatalf("статус = %q, ожидался %q", created.Status, StatusPaid)
	}
}

func TestSettleCreatesNewRecordAndKeepsHistory(t *testing.T) {
	store := &fakeStore{
		nextID: "settled",
		items: []Payment{{
			ID:          testPayment,
			ContractID:  testContract,
			Amount:      45000,
			PaymentDate: "2026-11-01",
			Type:        TypeRent,
			Status:      StatusPending,
		}},
	}
	svc := NewService(store)
	svc.now = func() time.Time { return time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC) }

	settled, err := svc.Settle(context.Background(), "user", testPayment, Settlement{})
	if err != nil {
		t.Fatalf("Settle: %v", err)
	}
	if settled.Status != StatusPaid {
		t.Fatalf("статус оплаты = %q, ожидался %q", settled.Status, StatusPaid)
	}
	if settled.ParentID == nil || *settled.ParentID != testPayment {
		t.Fatal("новая запись не связана с планируемым платежом")
	}
	if settled.PaymentDate != "2026-11-19" {
		t.Fatalf("дата оплаты = %q, ожидалась сегодняшняя", settled.PaymentDate)
	}
	if len(store.items) != 2 {
		t.Fatalf("в журнале %d записей, ожидалось 2", len(store.items))
	}
	original, _ := store.Get(context.Background(), "user", testPayment)
	if original.Status != StatusPending || original.Amount != 45000 {
		t.Fatal("исходная запись перезаписана — история платежей нарушена")
	}
}

func TestSettleRejectsAlreadyPaid(t *testing.T) {
	store := &fakeStore{items: []Payment{{
		ID:          testPayment,
		ContractID:  testContract,
		Amount:      1000,
		PaymentDate: "2026-11-01",
		Type:        TypeRent,
		Status:      StatusPaid,
	}}}
	svc := NewService(store)

	if _, err := svc.Settle(context.Background(), "user", testPayment, Settlement{}); err == nil {
		t.Fatal("оплаченный платёж удалось оплатить повторно")
	}
}

func TestSettleRejectsCanceled(t *testing.T) {
	store := &fakeStore{items: []Payment{{
		ID:          testPayment,
		ContractID:  testContract,
		Amount:      1000,
		PaymentDate: "2026-11-01",
		Type:        TypeRent,
		Status:      StatusCanceled,
	}}}
	svc := NewService(store)

	if _, err := svc.Settle(context.Background(), "user", testPayment, Settlement{}); err == nil {
		t.Fatal("отменённый платёж удалось оплатить")
	}
}
