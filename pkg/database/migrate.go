package database

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`

// Migrate применяет все ещё не применённые файлы вида NNNN_name.up.sql из fsys
// по возрастанию имени. Каждая миграция выполняется в отдельной транзакции,
// поэтому прерванный запуск не оставит половину изменений.
func Migrate(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) error {
	if _, err := pool.Exec(ctx, migrationsTable); err != nil {
		return fmt.Errorf("таблица schema_migrations: %w", err)
	}

	names, err := fs.Glob(fsys, "*.up.sql")
	if err != nil {
		return fmt.Errorf("список миграций: %w", err)
	}
	sort.Strings(names)

	for _, name := range names {
		version := strings.TrimSuffix(name, ".up.sql")

		var applied bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
		).Scan(&applied)
		if err != nil {
			return fmt.Errorf("проверка миграции %s: %w", version, err)
		}
		if applied {
			continue
		}

		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("чтение миграции %s: %w", name, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("транзакция миграции %s: %w", version, err)
		}
		// Simple protocol позволяет выполнить файл с несколькими выражениями.
		if _, err := tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("миграция %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, version,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("отметка миграции %s: %w", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("коммит миграции %s: %w", version, err)
		}
	}
	return nil
}
