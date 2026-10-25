package repository

import (
	"context"
	"database/sql"
	"errors"
	"stockflow/internal/entity"
	"stockflow/internal/model"

	"github.com/gofiber/fiber/v2"
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

func (p *ProductRepository) Create(ctx context.Context, request *model.CreateProductRequest) error {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	createQuery := `INSERT INTO products (sku, name, price, stock) VALUES ($1, $2, $3, $4) RETURNING id`
	var productID int64
	err = tx.QueryRowContext(ctx, createQuery, request.SKU, request.Name, request.Price, request.Stock).Scan(&productID)
	if err != nil {
		return err
	}

	if request.Stock > 0 {
		stockMovementQuery := `INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, $2, $3, $4)`
		stockMovement := entity.StockMovement{
			ProductID: productID,
			Type:      "in",
			Qty:       request.Stock,
			Note:      request.Note,
		}
		_, err = tx.ExecContext(ctx, stockMovementQuery, stockMovement.ProductID, stockMovement.Type, stockMovement.Qty, stockMovement.Note)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *ProductRepository) ExistsBySKU(ctx context.Context, SKU string) (bool, error) {
	getSKUQuery := `SELECT EXISTS (SELECT 1 FROM products WHERE sku = $1)`
	var exist bool
	if err := p.DB.QueryRowContext(ctx, getSKUQuery, SKU).Scan(&exist); err != nil {
		return false, err
	}
	return exist, nil
}

func (p *ProductRepository) UpdateProduct(ctx context.Context, request *model.UpdateProductRequest) (entity.Product, error) {
	updateQuery := `UPDATE products SET name = $1, price = $2, is_active = $3 WHERE id = $4 RETURNING id, sku, name, price, stock, is_active, created_at`
	var product entity.Product
	err := p.DB.QueryRowContext(ctx, updateQuery, request.Name, request.Price, *request.IsActive, request.ID).
		Scan(&product.ID, &product.SKU, &product.Name, &product.Price, &product.Stock, &product.IsActive, &product.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Product{}, fiber.ErrNotFound
		}
		return entity.Product{}, err
	}
	return product, nil
}
