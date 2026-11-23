package properties

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Repository — SQL-доступ к properties. Все запросы фильтруются по user_id,
// поэтому чужой объект неотличим от несуществующего.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий квартир.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const columns = `id, user_id, address, area, description, created_at, updated_at`

// List возвращает квартиры пользователя, новые сверху.
func (r *Repository) List(ctx context.Context, userID string) ([]Property, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+columns+` FROM properties WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("properties.List: %w", err)
	}
	defer rows.Close()

	items := make([]Property, 0, 8)
	for rows.Next() {
		var p Property
		if err := rows.Scan(&p.ID, &p.UserID, &p.Address, &p.Area, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("properties.List scan: %w", err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

// Get возвращает квартиру пользователя по ID.
func (r *Repository) Get(ctx context.Context, userID, id string) (Property, error) {
	var p Property
	err := r.pool.QueryRow(ctx,
		`SELECT `+columns+` FROM properties WHERE id = $1 AND user_id = $2`, id, userID).
		Scan(&p.ID, &p.UserID, &p.Address, &p.Area, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Property{}, apperr.NotFound("квартира")
	}
	if err != nil {
		return Property{}, fmt.Errorf("properties.Get: %w", err)
	}
	return p, nil
}

// Create вставляет квартиру.
func (r *Repository) Create(ctx context.Context, userID string, in Input) (Property, error) {
	var p Property
	err := r.pool.QueryRow(ctx,
		`INSERT INTO properties (user_id, address, area, description) VALUES ($1, $2, $3, $4)
		 RETURNING `+columns, userID, in.Address, in.Area, in.Description).
		Scan(&p.ID, &p.UserID, &p.Address, &p.Area, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return Property{}, fmt.Errorf("properties.Create: %w", err)
	}
	return p, nil
}

// Update полностью заменяет поля квартиры.
func (r *Repository) Update(ctx context.Context, userID, id string, in Input) (Property, error) {
	var p Property
	err := r.pool.QueryRow(ctx,
		`UPDATE properties SET address = $3, area = $4, description = $5
		 WHERE id = $1 AND user_id = $2
		 RETURNING `+columns, id, userID, in.Address, in.Area, in.Description).
		Scan(&p.ID, &p.UserID, &p.Address, &p.Area, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Property{}, apperr.NotFound("квартира")
	}
	if err != nil {
		return Property{}, fmt.Errorf("properties.Update: %w", err)
	}
	return p, nil
}

// Delete удаляет квартиру пользователя.
func (r *Repository) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM properties WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return apperr.Conflict("нельзя удалить квартиру: к ней привязаны договоры")
		}
		return fmt.Errorf("properties.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("квартира")
	}
	return nil
}
