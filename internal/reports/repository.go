package reports

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ailina/rental-assistant/internal/payments"
)

// Repository — агрегаты по данным пользователя.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий отчётов.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// unsettledPayment исключает исходный плановый платёж, когда для него уже
// создана дочерняя запись оплаты. Сам план остаётся в журнале, но не должен
// повторно попадать в задолженность и ожидаемые платежи.
const unsettledPayment = `
	AND NOT EXISTS (SELECT 1 FROM payments settlement WHERE settlement.parent_id = pm.id)`

// Counts считает квартиры, арендаторов, действующие договоры и просрочки.
func (r *Repository) Counts(ctx context.Context, userID string) (Counters, error) {
	const q = `
		SELECT
		  (SELECT count(*) FROM properties WHERE user_id = $1),
		  (SELECT count(*) FROM tenants WHERE user_id = $1),
		  (SELECT count(*) FROM contracts c
		     JOIN properties p ON p.id = c.property_id
		     WHERE p.user_id = $1 AND c.status = 'active'),
		  (SELECT count(*) FROM payments pm
		     JOIN contracts c ON c.id = pm.contract_id
		     JOIN properties p ON p.id = c.property_id
		     WHERE p.user_id = $1 AND pm.status IN ('pending', 'overdue')
		       AND pm.payment_date < CURRENT_DATE` + unsettledPayment + `)`
	var c Counters
	if err := r.pool.QueryRow(ctx, q, userID).Scan(
		&c.PropertiesCount, &c.TenantsCount, &c.ActiveContracts, &c.OverdueCount); err != nil {
		return Counters{}, fmt.Errorf("reports.Counts: %w", err)
	}
	return c, nil
}

// MonthMoney возвращает доходы и расходы текущего месяца.
func (r *Repository) MonthMoney(ctx context.Context, userID string) (Money, error) {
	const q = `
		SELECT
		  COALESCE(SUM(pm.amount) FILTER (WHERE pm.type IN ('rent', 'deposit', 'other')), 0),
		  COALESCE(SUM(pm.amount) FILTER (WHERE pm.type IN ('utility', 'repair')), 0)
		FROM payments pm
		JOIN contracts c ON c.id = pm.contract_id
		JOIN properties p ON p.id = c.property_id
		WHERE p.user_id = $1
		  AND pm.status = 'paid'
		  AND pm.payment_date >= date_trunc('month', CURRENT_DATE)
		  AND pm.payment_date < date_trunc('month', CURRENT_DATE) + interval '1 month'`
	var m Money
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&m.MonthIncome, &m.MonthExpenses); err != nil {
		return Money{}, fmt.Errorf("reports.MonthMoney: %w", err)
	}
	m.MonthProfit = m.MonthIncome - m.MonthExpenses
	return m, nil
}

// OverdueTotal суммирует неоплаченные платежи с прошедшей датой.
func (r *Repository) OverdueTotal(ctx context.Context, userID string) (amount float64, count int, err error) {
	const q = `
		SELECT COALESCE(SUM(pm.amount), 0), count(*)
		FROM payments pm
		JOIN contracts c ON c.id = pm.contract_id
		JOIN properties p ON p.id = c.property_id
		WHERE p.user_id = $1
		  AND pm.status IN ('pending', 'overdue')
		  AND pm.payment_date < CURRENT_DATE` + unsettledPayment
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&amount, &count); err != nil {
		return 0, 0, fmt.Errorf("reports.OverdueTotal: %w", err)
	}
	return amount, count, nil
}

const paymentColumns = `pm.id, pm.contract_id, pr.address, t.name, pm.amount,
	to_char(pm.payment_date, 'YYYY-MM-DD'), pm.type, pm.status, pm.description, pm.parent_id, pm.created_at`

const paymentJoin = `
	FROM payments pm
	JOIN contracts c ON c.id = pm.contract_id
	JOIN properties pr ON pr.id = c.property_id
	JOIN tenants t ON t.id = c.tenant_id
	WHERE pr.user_id = $1`

// Upcoming возвращает ожидаемые платежи в ближайшие days дней.
func (r *Repository) Upcoming(ctx context.Context, userID string, days, limit int) ([]payments.Payment, error) {
	q := `SELECT ` + paymentColumns + paymentJoin + `
		AND pm.status = 'pending'
		AND pm.payment_date BETWEEN CURRENT_DATE AND CURRENT_DATE + $2::int` + unsettledPayment + `
		ORDER BY pm.payment_date
		LIMIT $3`
	return r.paymentList(ctx, q, userID, days, limit)
}

// Overdue возвращает просроченные платежи.
func (r *Repository) Overdue(ctx context.Context, userID string, limit int) ([]payments.Payment, error) {
	q := `SELECT ` + paymentColumns + paymentJoin + `
		AND pm.status IN ('pending', 'overdue')
		AND pm.payment_date < CURRENT_DATE` + unsettledPayment + `
		ORDER BY pm.payment_date
		LIMIT $2`
	return r.paymentList(ctx, q, userID, limit)
}

// Recent возвращает последние операции журнала.
func (r *Repository) Recent(ctx context.Context, userID string, limit int) ([]payments.Payment, error) {
	q := `SELECT ` + paymentColumns + paymentJoin + `
		ORDER BY pm.created_at DESC, pm.payment_date DESC
		LIMIT $2`
	return r.paymentList(ctx, q, userID, limit)
}

// Monthly возвращает доходы и расходы за последние months месяцев.
func (r *Repository) Monthly(ctx context.Context, userID string, months int) ([]MonthPoint, error) {
	const q = `
		WITH span AS (
		  SELECT generate_series(
		           date_trunc('month', CURRENT_DATE) - make_interval(months => $2::int - 1),
		           date_trunc('month', CURRENT_DATE),
		           interval '1 month') AS month
		)
		SELECT to_char(s.month, 'YYYY-MM'),
		       COALESCE(SUM(pm.amount) FILTER (WHERE pm.type IN ('rent', 'deposit', 'other')), 0),
		       COALESCE(SUM(pm.amount) FILTER (WHERE pm.type IN ('utility', 'repair')), 0)
		FROM span s
		LEFT JOIN payments pm ON pm.status = 'paid'
		     AND pm.payment_date >= s.month
		     AND pm.payment_date < s.month + interval '1 month'
		     AND pm.contract_id IN (
		         SELECT c.id FROM contracts c
		         JOIN properties p ON p.id = c.property_id
		         WHERE p.user_id = $1)
		GROUP BY s.month
		ORDER BY s.month`

	rows, err := r.pool.Query(ctx, q, userID, months)
	if err != nil {
		return nil, fmt.Errorf("reports.Monthly: %w", err)
	}
	defer rows.Close()

	points := make([]MonthPoint, 0, months)
	for rows.Next() {
		var p MonthPoint
		if err := rows.Scan(&p.Month, &p.Income, &p.Expenses); err != nil {
			return nil, fmt.Errorf("reports.Monthly scan: %w", err)
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (r *Repository) paymentList(ctx context.Context, q, userID string, args ...any) ([]payments.Payment, error) {
	rows, err := r.pool.Query(ctx, q, append([]any{userID}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("reports.payments: %w", err)
	}
	defer rows.Close()

	items := make([]payments.Payment, 0, 8)
	for rows.Next() {
		var p payments.Payment
		if err := rows.Scan(&p.ID, &p.ContractID, &p.PropertyAddress, &p.TenantName, &p.Amount,
			&p.PaymentDate, &p.Type, &p.Status, &p.Description, &p.ParentID, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("reports.payments scan: %w", err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}
