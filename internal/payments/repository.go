package payments

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Repository — SQL-доступ к журналу платежей. Строки только добавляются:
// методы Update/Delete намеренно отсутствуют.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий платежей.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// base — выборка платежей пользователя с вычисленным (актуальным) статусом.
const base = `
	SELECT pay.id, pay.contract_id, p.address, t.name, pay.amount,
	       to_char(pay.payment_date, 'YYYY-MM-DD') AS payment_date,
	       CASE WHEN pay.status = 'pending' AND pay.payment_date < CURRENT_DATE
	            THEN 'overdue' ELSE pay.status END AS effective_status,
	       pay.type, pay.description, pay.parent_id, pay.created_at
	FROM payments pay
	JOIN contracts c ON c.id = pay.contract_id
	JOIN properties p ON p.id = c.property_id
	JOIN tenants t ON t.id = c.tenant_id
	WHERE p.user_id = $1 AND t.user_id = $1`

// notSettled скрывает запись, у которой уже есть исполнение: в списке видна
// только фактическая оплата, а исходный план остаётся в журнале базы.
const notSettled = `
	AND NOT EXISTS (SELECT 1 FROM payments child WHERE child.parent_id = pay.id)`

func scan(row pgx.Row) (Payment, error) {
	var pay Payment
	err := row.Scan(&pay.ID, &pay.ContractID, &pay.PropertyAddress, &pay.TenantName, &pay.Amount,
		&pay.PaymentDate, &pay.Status, &pay.Type, &pay.Description, &pay.ParentID, &pay.CreatedAt)
	return pay, err
}

// Filter — условия выборки платежей.
type Filter struct {
	ContractID string
	PropertyID string
	Type       string
	Status     string
	From       string
	To         string
	Limit      int
	Offset     int
}

// List возвращает платежи пользователя по фильтру, новые сверху.
func (r *Repository) List(ctx context.Context, userID string, f Filter) ([]Payment, error) {
	var where strings.Builder
	args := []any{userID}
	add := func(clause string, value any) {
		args = append(args, value)
		fmt.Fprintf(&where, " AND %s = $%d", clause, len(args))
	}
	if f.ContractID != "" {
		add("pay.contract_id", f.ContractID)
	}
	if f.PropertyID != "" {
		add("c.property_id", f.PropertyID)
	}
	if f.Type != "" {
		add("pay.type", f.Type)
	}
	if f.From != "" {
		args = append(args, f.From)
		fmt.Fprintf(&where, " AND pay.payment_date >= $%d::date", len(args))
	}
	if f.To != "" {
		args = append(args, f.To)
		fmt.Fprintf(&where, " AND pay.payment_date <= $%d::date", len(args))
	}

	q := `SELECT * FROM (` + base + notSettled + where.String() + `) s`
	if f.Status != "" {
		args = append(args, f.Status)
		q += fmt.Sprintf(` WHERE s.effective_status = $%d`, len(args))
	}
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args = append(args, limit, max(f.Offset, 0))
	q += fmt.Sprintf(` ORDER BY s.payment_date DESC, s.created_at DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("payments.List: %w", err)
	}
	defer rows.Close()

	items := make([]Payment, 0, 16)
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("payments.List scan: %w", err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

// Get возвращает платёж пользователя.
func (r *Repository) Get(ctx context.Context, userID, id string) (Payment, error) {
	q := `SELECT * FROM (` + base + ` AND pay.id = $2) s`
	p, err := scan(r.pool.QueryRow(ctx, q, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, apperr.NotFound("платёж")
	}
	if err != nil {
		return Payment{}, fmt.Errorf("payments.Get: %w", err)
	}
	return p, nil
}

// Create добавляет запись журнала; договор должен принадлежать пользователю.
func (r *Repository) Create(ctx context.Context, userID string, in Input, parentID *string) (Payment, error) {
	const q = `
		INSERT INTO payments (contract_id, amount, payment_date, type, status, description, parent_id)
		SELECT c.id, $2, $3::date, $4, $5, $6, $7
		FROM contracts c
		JOIN properties p ON p.id = c.property_id
		JOIN tenants t ON t.id = c.tenant_id
		WHERE c.id = $1 AND p.user_id = $8 AND t.user_id = $8
		RETURNING id`
	var id string
	err := r.pool.QueryRow(ctx, q, in.ContractID, in.Amount, in.PaymentDate, in.Type,
		in.Status, in.Description, parentID, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, apperr.Validation("договор не найден")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Payment{}, apperr.Conflict("этот платёж уже исполнен")
	}
	if err != nil {
		return Payment{}, fmt.Errorf("payments.Create: %w", err)
	}
	return r.Get(ctx, userID, id)
}

// PendingSoon — планируемые платежи со сроком в ближайшие days дней.
func (r *Repository) PendingSoon(ctx context.Context, userID string, days int) ([]Payment, error) {
	q := `SELECT * FROM (` + base + notSettled + `
		AND pay.status = 'pending'
		AND pay.payment_date BETWEEN CURRENT_DATE AND CURRENT_DATE + $2::int) s
		ORDER BY s.payment_date`
	rows, err := r.pool.Query(ctx, q, userID, days)
	if err != nil {
		return nil, fmt.Errorf("payments.PendingSoon: %w", err)
	}
	defer rows.Close()

	items := make([]Payment, 0, 8)
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("payments.PendingSoon scan: %w", err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}
