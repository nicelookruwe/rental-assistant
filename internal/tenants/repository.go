package tenants

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Repository — SQL-доступ к tenants с фильтром по владельцу.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий арендаторов.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const columns = `id, user_id, name, phone, email, created_at, updated_at`

// List возвращает арендаторов пользователя.
func (r *Repository) List(ctx context.Context, userID string) ([]Tenant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+columns+` FROM tenants WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("tenants.List: %w", err)
	}
	defer rows.Close()

	items := make([]Tenant, 0, 8)
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Phone, &t.Email, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("tenants.List scan: %w", err)
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

// Get возвращает арендатора пользователя.
func (r *Repository) Get(ctx context.Context, userID, id string) (Tenant, error) {
	var t Tenant
	err := r.pool.QueryRow(ctx,
		`SELECT `+columns+` FROM tenants WHERE id = $1 AND user_id = $2`, id, userID).
		Scan(&t.ID, &t.UserID, &t.Name, &t.Phone, &t.Email, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, apperr.NotFound("арендатор")
	}
	if err != nil {
		return Tenant{}, fmt.Errorf("tenants.Get: %w", err)
	}
	return t, nil
}

// Create вставляет арендатора.
func (r *Repository) Create(ctx context.Context, userID string, in Input) (Tenant, error) {
	var t Tenant
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tenants (user_id, name, phone, email) VALUES ($1, $2, $3, $4)
		 RETURNING `+columns, userID, in.Name, in.Phone, in.Email).
		Scan(&t.ID, &t.UserID, &t.Name, &t.Phone, &t.Email, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return Tenant{}, fmt.Errorf("tenants.Create: %w", err)
	}
	return t, nil
}

// Update обновляет арендатора пользователя.
func (r *Repository) Update(ctx context.Context, userID, id string, in Input) (Tenant, error) {
	var t Tenant
	err := r.pool.QueryRow(ctx,
		`UPDATE tenants SET name = $3, phone = $4, email = $5
		 WHERE id = $1 AND user_id = $2
		 RETURNING `+columns, id, userID, in.Name, in.Phone, in.Email).
		Scan(&t.ID, &t.UserID, &t.Name, &t.Phone, &t.Email, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, apperr.NotFound("арендатор")
	}
	if err != nil {
		return Tenant{}, fmt.Errorf("tenants.Update: %w", err)
	}
	return t, nil
}

// Delete удаляет арендатора, если на него не ссылаются договоры.
func (r *Repository) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tenants WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return apperr.Conflict("нельзя удалить арендатора: с ним есть договоры")
		}
		return fmt.Errorf("tenants.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("арендатор")
	}
	return nil
}
