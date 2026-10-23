package test

import (
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	countAllQuery = "SELECT COUNT(*) FROM products"
	selectBase    = "SELECT id, sku, name, price, stock, is_active, created_at FROM products"
	orderBy       = "ORDER BY created_at DESC, id DESC"
)

var createdAt = time.Date(2026, 10, 22, 18, 8, 50, 0, time.UTC)

func productRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "sku", "name", "price", "stock", "is_active", "created_at"})
}

func countRows(n int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(n)
}
