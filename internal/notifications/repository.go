package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Repository — SQL-доступ к notifications.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий уведомлений.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const columns = `id, message, scheduled_at, status, created_at`

// Insert добавляет напоминание, если такого ещё нет.
// Возвращает created=false для уже существующего dedupe_key.
func (r *Repository) Insert(ctx context.Context, userID, message string, scheduledAt time.Time, dedupeKey string) (Notification, bool, error) {
	var n Notification
	err := r.pool.QueryRow(ctx,
		`INSERT INTO notifications (user_id, message, scheduled_at, dedupe_key)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, dedupe_key) DO NOTHING
		 RETURNING `+columns, userID, message, scheduledAt, dedupeKey).
		Scan(&n.ID, &n.Message, &n.ScheduledAt, &n.Status, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Notification{}, false, nil
	}
	if err != nil {
		return Notification{}, false, fmt.Errorf("notifications.Insert: %w", err)
	}
	return n, true, nil
}

// List возвращает уведомления пользователя, ближайшие сверху.
func (r *Repository) List(ctx context.Context, userID, status string) ([]Notification, error) {
	q := `SELECT ` + columns + ` FROM notifications WHERE user_id = $1`
	args := []any{userID}
	if status != "" {
		args = append(args, status)
		q += " AND status = $2"
	}
	q += " ORDER BY scheduled_at ASC, created_at DESC LIMIT 200"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("notifications.List: %w", err)
	}
	defer rows.Close()

	items := make([]Notification, 0, 8)
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.Message, &n.ScheduledAt, &n.Status, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("notifications.List scan: %w", err)
		}
		items = append(items, n)
	}
	return items, rows.Err()
}

// MarkRead помечает напоминание прочитанным.
func (r *Repository) MarkRead(ctx context.Context, userID, id string) (Notification, error) {
	var n Notification
	err := r.pool.QueryRow(ctx,
		`UPDATE notifications SET status = 'read'
		 WHERE id = $1 AND user_id = $2
		 RETURNING `+columns, id, userID).
		Scan(&n.ID, &n.Message, &n.ScheduledAt, &n.Status, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Notification{}, apperr.NotFound("уведомление")
	}
	if err != nil {
		return Notification{}, fmt.Errorf("notifications.MarkRead: %w", err)
	}
	return n, nil
}
