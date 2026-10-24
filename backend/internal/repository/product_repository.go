package repository

import (
	"context"
	"database/sql"
	"stockflow/internal/entity"
	"stockflow/internal/model"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		DB: db,
	}
}

func (p *ProductRepository) Search(ctx context.Context, request *model.SearchProductRequest) ([]entity.ProductWithMovement, int64, error) {

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

	var total int64
	if err := p.DB.QueryRowContext(ctx, countQuery, request.Search).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []entity.ProductWithMovement{}, 0, nil
	}

	offset := (request.Page - 1) * request.Limit

	rows, err := p.DB.QueryContext(ctx, searchQuery, request.Search, request.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]entity.ProductWithMovement, 0)
	for rows.Next() {
		var item entity.ProductWithMovement

		if err := rows.Scan(&item.ID, &item.SKU, &item.Name, &item.Price, &item.Stock, &item.IsActive, &item.CreatedAt, &item.LastMovementAt); err != nil {
			return nil, 0, err
		}

		products = append(products, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
