// Package migrations содержит SQL-файлы схемы, встроенные в бинарник.
package migrations

import "embed"

// FS — встроенные файлы миграций.
//
//go:embed *.sql
var FS embed.FS
