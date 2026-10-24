package test

import (
	"errors"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
)

const (
	searchQuery = `
	SELECT p.id, p.sku, p.name, p.price, p.stock, p.is_active, p.created_at,
	(SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = p.id) AS last_movement_at
	FROM products p
	WHERE ($1::text = '' OR p.sku ILIKE '%' || $1::text || '%' OR p.name ILIKE '%' || $1::text || '%')
	ORDER BY p.id DESC
	LIMIT $2 OFFSET $3`
	countQuery = `
	SELECT COUNT(*) FROM products p
	WHERE ($1::text = '' OR p.sku ILIKE '%' || $1::text || '%' OR p.name ILIKE '%' || $1::text || '%')`
	existsSKUQuery      = `SELECT EXISTS (SELECT 1 FROM products WHERE sku = $1)`
	createProductQuery  = `INSERT INTO products (sku, name, price, stock) VALUES ($1, $2, $3, $4) RETURNING id`
	insertMovementQuery = `INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, $2, $3, $4)`
)

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

func existsRows(v bool) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"exists"}).AddRow(v)
}

func idRows(id int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id"}).AddRow(id)
}

func statusCode(err error) int {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Code
	}
	return 0
}
