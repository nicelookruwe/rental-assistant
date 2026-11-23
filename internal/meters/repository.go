package meters

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Repository — SQL-доступ к meters. Update/Delete отсутствуют: история
// показаний неизменяема (в БД это закреплено триггером).
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий показаний.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Filter — условия выборки показаний.
type Filter struct {
	PropertyID string
	MeterType  string
	Limit      int
}

const base = `
	SELECT m.id, m.property_id, p.address, m.meter_type, m.value,
	       to_char(m.recorded_at, 'YYYY-MM-DD'), m.note, m.created_at
	FROM meters m
	JOIN properties p ON p.id = m.property_id
	WHERE p.user_id = $1`

// List возвращает историю показаний, свежие сверху.
func (r *Repository) List(ctx context.Context, userID string, f Filter) ([]Reading, error) {
	q := base
	args := []any{userID}
	if f.PropertyID != "" {
		args = append(args, f.PropertyID)
		q += fmt.Sprintf(" AND m.property_id = $%d", len(args))
	}
	if f.MeterType != "" {
		args = append(args, f.MeterType)
		q += fmt.Sprintf(" AND m.meter_type = $%d", len(args))
	}
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args = append(args, limit)
	q += fmt.Sprintf(" ORDER BY m.recorded_at DESC, m.created_at DESC LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("meters.List: %w", err)
	}
	defer rows.Close()

	items := make([]Reading, 0, 16)
	for rows.Next() {
		var m Reading
		if err := rows.Scan(&m.ID, &m.PropertyID, &m.PropertyAddress, &m.MeterType, &m.Value,
			&m.RecordedAt, &m.Note, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("meters.List scan: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

// Get возвращает одно показание пользователя.
func (r *Repository) Get(ctx context.Context, userID, id string) (Reading, error) {
	var m Reading
	err := r.pool.QueryRow(ctx, base+` AND m.id = $2`, userID, id).
		Scan(&m.ID, &m.PropertyID, &m.PropertyAddress, &m.MeterType, &m.Value, &m.RecordedAt, &m.Note, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reading{}, apperr.NotFound("показание")
	}
	if err != nil {
		return Reading{}, fmt.Errorf("meters.Get: %w", err)
	}
	return m, nil
}

// Latest возвращает последнее показание по каждому счётчику.
func (r *Repository) Latest(ctx context.Context, userID, propertyID string) ([]Latest, error) {
	q := `
		SELECT DISTINCT ON (m.property_id, m.meter_type)
		       m.property_id, p.address, m.meter_type, m.value, to_char(m.recorded_at, 'YYYY-MM-DD')
		FROM meters m
		JOIN properties p ON p.id = m.property_id
		WHERE p.user_id = $1`
	args := []any{userID}
	if propertyID != "" {
		args = append(args, propertyID)
		q += fmt.Sprintf(" AND m.property_id = $%d", len(args))
	}
	q += ` ORDER BY m.property_id, m.meter_type, m.recorded_at DESC, m.created_at DESC`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("meters.Latest: %w", err)
	}
	defer rows.Close()

	items := make([]Latest, 0, 8)
	for rows.Next() {
		var l Latest
		if err := rows.Scan(&l.PropertyID, &l.PropertyAddress, &l.MeterType, &l.Value, &l.RecordedAt); err != nil {
			return nil, fmt.Errorf("meters.Latest scan: %w", err)
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

// Create добавляет показание; квартира должна принадлежать пользователю.
func (r *Repository) Create(ctx context.Context, userID string, in Input) (Reading, error) {
	const q = `
		INSERT INTO meters (user_id, property_id, meter_type, value, recorded_at, note)
		SELECT $2, p.id, $3, $4, $5::date, $6
		FROM properties p
		WHERE p.id = $1 AND p.user_id = $2
		RETURNING id`
	var id string
	err := r.pool.QueryRow(ctx, q, in.PropertyID, userID, in.MeterType, in.Value, in.RecordedAt, in.Note).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reading{}, apperr.Validation("квартира не найдена")
	}
	if err != nil {
		return Reading{}, fmt.Errorf("meters.Create: %w", err)
	}
	return r.Get(ctx, userID, id)
}
