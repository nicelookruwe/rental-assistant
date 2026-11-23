package documents

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ailina/rental-assistant/pkg/apperr"
)

// Repository — SQL-доступ к documents. Изоляция — через properties.user_id.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий документов.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const base = `
	SELECT d.id, d.property_id, p.address, d.file_url, d.file_name, d.document_type,
	       d.size_bytes, d.created_at
	FROM documents d
	JOIN properties p ON p.id = d.property_id
	WHERE p.user_id = $1`

// List возвращает документы пользователя (можно ограничить квартирой).
func (r *Repository) List(ctx context.Context, userID, propertyID string) ([]Document, error) {
	q := base
	args := []any{userID}
	if propertyID != "" {
		args = append(args, propertyID)
		q += " AND d.property_id = $2"
	}
	q += " ORDER BY d.created_at DESC"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("documents.List: %w", err)
	}
	defer rows.Close()

	items := make([]Document, 0, 8)
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.PropertyID, &d.PropertyAddress, &d.FileURL, &d.FileName,
			&d.DocumentType, &d.SizeBytes, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("documents.List scan: %w", err)
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

// Get возвращает документ пользователя.
func (r *Repository) Get(ctx context.Context, userID, id string) (Document, error) {
	var d Document
	err := r.pool.QueryRow(ctx, base+" AND d.id = $2", userID, id).
		Scan(&d.ID, &d.PropertyID, &d.PropertyAddress, &d.FileURL, &d.FileName,
			&d.DocumentType, &d.SizeBytes, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Document{}, apperr.NotFound("документ")
	}
	if err != nil {
		return Document{}, fmt.Errorf("documents.Get: %w", err)
	}
	return d, nil
}

// Create вставляет метаданные, только если квартира принадлежит пользователю.
func (r *Repository) Create(ctx context.Context, userID string, in Input) (Document, error) {
	const q = `
		INSERT INTO documents (property_id, file_url, file_name, document_type, size_bytes)
		SELECT p.id, $3, $4, $5, $6
		FROM properties p
		WHERE p.id = $2 AND p.user_id = $1
		RETURNING id`
	var id string
	err := r.pool.QueryRow(ctx, q, userID, in.PropertyID, in.FileURL, in.FileName,
		in.DocumentType, in.SizeBytes).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Document{}, apperr.Validation("квартира не найдена")
	}
	if err != nil {
		return Document{}, fmt.Errorf("documents.Create: %w", err)
	}
	return r.Get(ctx, userID, id)
}

// Delete удаляет запись о документе; файл удаляет сервис.
func (r *Repository) Delete(ctx context.Context, userID, id string) error {
	const q = `
		DELETE FROM documents d
		USING properties p
		WHERE d.id = $2 AND p.id = d.property_id AND p.user_id = $1`
	tag, err := r.pool.Exec(ctx, q, userID, id)
	if err != nil {
		return fmt.Errorf("documents.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("документ")
	}
	return nil
}
