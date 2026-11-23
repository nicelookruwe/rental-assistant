package contracts

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Repository — SQL-доступ к contracts. Принадлежность проверяется через
// properties.user_id и tenants.user_id, поэтому чужие договоры недостижимы.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий договоров.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const selectColumns = `
	c.id, c.property_id, c.tenant_id,
	p.address, t.name,
	to_char(c.start_date, 'YYYY-MM-DD'), to_char(c.end_date, 'YYYY-MM-DD'),
	c.rent_amount, c.deposit, c.status, c.created_at, c.updated_at`

const fromJoin = `
	FROM contracts c
	JOIN properties p ON p.id = c.property_id
	JOIN tenants t ON t.id = c.tenant_id`

func scanContract(row pgx.Row) (Contract, error) {
	var c Contract
	err := row.Scan(&c.ID, &c.PropertyID, &c.TenantID, &c.PropertyAddress, &c.TenantName,
		&c.StartDate, &c.EndDate, &c.RentAmount, &c.Deposit, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

// Filter — условия списка договоров.
type Filter struct {
	Status     string
	PropertyID string
}

// List возвращает договоры пользователя с необязательными фильтрами.
func (r *Repository) List(ctx context.Context, userID string, f Filter) ([]Contract, error) {
	q := `SELECT ` + selectColumns + fromJoin + ` WHERE p.user_id = $1 AND t.user_id = $1`
	args := []any{userID}
	if f.Status != "" {
		args = append(args, f.Status)
		q += fmt.Sprintf(" AND c.status = $%d", len(args))
	}
	if f.PropertyID != "" {
		args = append(args, f.PropertyID)
		q += fmt.Sprintf(" AND c.property_id = $%d", len(args))
	}
	q += " ORDER BY c.start_date DESC"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("contracts.List: %w", err)
	}
	defer rows.Close()

	items := make([]Contract, 0, 8)
	for rows.Next() {
		c, err := scanContract(rows)
		if err != nil {
			return nil, fmt.Errorf("contracts.List scan: %w", err)
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// Get возвращает договор пользователя.
func (r *Repository) Get(ctx context.Context, userID, id string) (Contract, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+selectColumns+fromJoin+` WHERE c.id = $1 AND p.user_id = $2 AND t.user_id = $2`, id, userID)
	c, err := scanContract(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Contract{}, apperr.NotFound("договор")
	}
	if err != nil {
		return Contract{}, fmt.Errorf("contracts.Get: %w", err)
	}
	return c, nil
}

// Create вставляет договор, только если и квартира, и арендатор принадлежат
// пользователю: INSERT ... SELECT возвращает 0 строк для чужих объектов.
func (r *Repository) Create(ctx context.Context, userID string, in Input) (Contract, error) {
	const q = `
		INSERT INTO contracts (property_id, tenant_id, start_date, end_date, rent_amount, deposit, status)
		SELECT p.id, t.id, $3::date, $4::date, $5, $6, $7
		FROM properties p, tenants t
		WHERE p.id = $1 AND p.user_id = $8 AND t.id = $2 AND t.user_id = $8
		RETURNING id`
	var id string
	err := r.pool.QueryRow(ctx, q, in.PropertyID, in.TenantID, in.StartDate, in.EndDate,
		in.RentAmount, in.Deposit, in.Status, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Contract{}, apperr.Validation("квартира или арендатор не найдены")
	}
	if err != nil {
		return Contract{}, fmt.Errorf("contracts.Create: %w", err)
	}
	return r.Get(ctx, userID, id)
}

// Update обновляет договор пользователя.
func (r *Repository) Update(ctx context.Context, userID, id string, in Input) (Contract, error) {
	const q = `
		UPDATE contracts c
		SET property_id = $3, tenant_id = $4, start_date = $5::date, end_date = $6::date,
		    rent_amount = $7, deposit = $8, status = $9
		WHERE c.id = $1
		  AND EXISTS (SELECT 1 FROM properties p WHERE p.id = c.property_id AND p.user_id = $2)
		  AND EXISTS (SELECT 1 FROM properties p WHERE p.id = $3 AND p.user_id = $2)
		  AND EXISTS (SELECT 1 FROM tenants t WHERE t.id = $4 AND t.user_id = $2)
		RETURNING c.id`
	var updated string
	err := r.pool.QueryRow(ctx, q, id, userID, in.PropertyID, in.TenantID, in.StartDate,
		in.EndDate, in.RentAmount, in.Deposit, in.Status).Scan(&updated)
	if errors.Is(err, pgx.ErrNoRows) {
		return Contract{}, apperr.NotFound("договор")
	}
	if err != nil {
		return Contract{}, fmt.Errorf("contracts.Update: %w", err)
	}
	return r.Get(ctx, userID, updated)
}

// EndingSoon возвращает действующие договоры, срок которых истекает в ближайшие
// days дней (включая сегодня). Используется модулем уведомлений.
func (r *Repository) EndingSoon(ctx context.Context, userID string, days int) ([]Contract, error) {
	q := `SELECT ` + selectColumns + fromJoin + `
		WHERE p.user_id = $1 AND t.user_id = $1
		  AND c.status IN ('active', 'draft')
		  AND c.end_date BETWEEN CURRENT_DATE AND CURRENT_DATE + $2::int
		ORDER BY c.end_date`
	rows, err := r.pool.Query(ctx, q, userID, days)
	if err != nil {
		return nil, fmt.Errorf("contracts.EndingSoon: %w", err)
	}
	defer rows.Close()

	items := make([]Contract, 0, 4)
	for rows.Next() {
		c, err := scanContract(rows)
		if err != nil {
			return nil, fmt.Errorf("contracts.EndingSoon scan: %w", err)
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// Delete удаляет договор, если по нему нет платежей.
func (r *Repository) Delete(ctx context.Context, userID, id string) error {
	const q = `DELETE FROM contracts c
		WHERE c.id = $1
		  AND EXISTS (SELECT 1 FROM properties p WHERE p.id = c.property_id AND p.user_id = $2)`
	tag, err := r.pool.Exec(ctx, q, id, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return apperr.Conflict("нельзя удалить договор: по нему есть платежи")
		}
		return fmt.Errorf("contracts.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("договор")
	}
	return nil
}
