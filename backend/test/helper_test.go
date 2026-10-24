package test

import (
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const searchQuery = `
SELECT p.id, p.sku, p.name, p.price, p.stock, p.is_active, p.created_at,
(SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = p.id) AS last_movement_at
FROM products p
WHERE ($1::text = '' OR p.sku ILIKE '%' || $1::text || '%' OR p.name ILIKE '%' || $1::text || '%')
ORDER BY p.id DESC
LIMIT $2 OFFSET $3`

const countQuery = `
SELECT COUNT(*) FROM products p
WHERE ($1::text = '' OR p.sku ILIKE '%' || $1::text || '%' OR p.name ILIKE '%' || $1::text || '%')`

var createdAt = time.Date(2026, 10, 22, 18, 8, 50, 0, time.UTC)
var lastMovement = createdAt.Add(2 * time.Hour)

func productRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "sku", "name", "price", "stock", "is_active", "created_at", "last_movement_at",
	})
}

func countRows(n int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(n)
}
