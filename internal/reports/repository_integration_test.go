package reports

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryExcludesSettledPlansFromDebtAndUpcoming(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	defer pool.Close()

	stamp := time.Now().UnixNano()
	var userID, propertyID, tenantID, contractID string
	err = pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, 'hash') RETURNING id`,
		fmt.Sprintf("reports-settlement-%d@example.test", stamp),
	).Scan(&userID)
	if err != nil {
		t.Fatalf("create user (apply migrations first): %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM payments WHERE contract_id = $1`, contractID)
		_, _ = pool.Exec(ctx, `DELETE FROM contracts WHERE id = $1`, contractID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})
	if err := pool.QueryRow(ctx,
		`INSERT INTO properties (user_id, address, area) VALUES ($1, 'Test address', 1) RETURNING id`, userID,
	).Scan(&propertyID); err != nil {
		t.Fatalf("create property: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO tenants (user_id, name) VALUES ($1, 'Test tenant') RETURNING id`, userID,
	).Scan(&tenantID); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO contracts (property_id, tenant_id, start_date, end_date, rent_amount)
		VALUES ($1, $2, CURRENT_DATE - 30, CURRENT_DATE + 30, 100) RETURNING id`, propertyID, tenantID,
	).Scan(&contractID); err != nil {
		t.Fatalf("create contract: %v", err)
	}

	var settledPastID, unpaidPastID, settledFutureID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO payments (contract_id, amount, payment_date, type, status)
		VALUES ($1, 200, CURRENT_DATE - 1, 'rent', 'pending') RETURNING id`, contractID,
	).Scan(&settledPastID); err != nil {
		t.Fatalf("create settled past plan: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO payments (contract_id, amount, payment_date, type, status, parent_id)
		VALUES ($1, 200, CURRENT_DATE, 'rent', 'paid', $2)`, contractID, settledPastID,
	); err != nil {
		t.Fatalf("create past settlement: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO payments (contract_id, amount, payment_date, type, status)
		VALUES ($1, 100, CURRENT_DATE - 1, 'rent', 'pending') RETURNING id`, contractID,
	).Scan(&unpaidPastID); err != nil {
		t.Fatalf("create unpaid past plan: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO payments (contract_id, amount, payment_date, type, status)
		VALUES ($1, 300, CURRENT_DATE + 1, 'rent', 'pending') RETURNING id`, contractID,
	).Scan(&settledFutureID); err != nil {
		t.Fatalf("create settled future plan: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO payments (contract_id, amount, payment_date, type, status, parent_id)
		VALUES ($1, 300, CURRENT_DATE, 'rent', 'paid', $2)`, contractID, settledFutureID,
	); err != nil {
		t.Fatalf("create future settlement: %v", err)
	}

	repo := NewRepository(pool)
	counts, err := repo.Counts(ctx, userID)
	if err != nil {
		t.Fatalf("Counts: %v", err)
	}
	if counts.OverdueCount != 1 {
		t.Fatalf("overdue count = %d, want 1", counts.OverdueCount)
	}
	amount, count, err := repo.OverdueTotal(ctx, userID)
	if err != nil {
		t.Fatalf("OverdueTotal: %v", err)
	}
	if amount != 100 || count != 1 {
		t.Fatalf("overdue total = %.2f / %d, want 100.00 / 1", amount, count)
	}
	overdue, err := repo.Overdue(ctx, userID, 10)
	if err != nil {
		t.Fatalf("Overdue: %v", err)
	}
	if len(overdue) != 1 || overdue[0].ID != unpaidPastID {
		t.Fatalf("overdue = %#v, want only unpaid plan %s", overdue, unpaidPastID)
	}
	upcoming, err := repo.Upcoming(ctx, userID, 7, 10)
	if err != nil {
		t.Fatalf("Upcoming: %v", err)
	}
	if len(upcoming) != 0 {
		t.Fatalf("upcoming contains settled plan: %#v", upcoming)
	}
}
