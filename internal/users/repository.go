// Package users — доступ к учётным записям.
package users

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

// Repository — SQL-доступ к таблице users.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий пользователей.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Create вставляет пользователя; email приводится к нижнему регистру.
func (r *Repository) Create(ctx context.Context, email, passwordHash string) (User, error) {
	const q = `INSERT INTO users (email, password_hash) VALUES (lower($1), $2)
	           RETURNING id, email, password_hash, created_at`
	var u User
	err := r.pool.QueryRow(ctx, q, strings.TrimSpace(email), passwordHash).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, apperr.Conflict("пользователь с такой почтой уже зарегистрирован")
		}
		return User{}, fmt.Errorf("users.Create: %w", err)
	}
	return u, nil
}

// ByEmail ищет пользователя по почте без учёта регистра.
func (r *Repository) ByEmail(ctx context.Context, email string) (User, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE lower(email) = lower($1)`
	var u User
	err := r.pool.QueryRow(ctx, q, strings.TrimSpace(email)).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, apperr.NotFound("пользователь")
	}
	if err != nil {
		return User{}, fmt.Errorf("users.ByEmail: %w", err)
	}
	return u, nil
}

// ByID возвращает пользователя по идентификатору.
func (r *Repository) ByID(ctx context.Context, id string) (User, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`
	var u User
	err := r.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, apperr.NotFound("пользователь")
	}
	if err != nil {
		return User{}, fmt.Errorf("users.ByID: %w", err)
	}
	return u, nil
}
