package repository

import (
	"context"
	"database/sql"
	"fmt"
	"stockflow/internal/entity"
	"stockflow/internal/model"
	"strings"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		DB: db,
	}
}

func (p *ProductRepository) Search(ctx context.Context, request *model.SearchProductRequest) ([]entity.Product, int64, error) {
	where, args := p.FilterProducts(request)

	var total int64
	countQuery := "SELECT COUNT(*) FROM products" + where
	if err := p.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []entity.Product{}, 0, nil
	}

	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	query := fmt.Sprintf(`
		SELECT id, sku, name, price, stock, is_active, created_at
		FROM products%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d`, where, limitPos, offsetPos)

	args = append(args, request.Size, (request.Page-1)*request.Size)

	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]entity.Product, 0, request.Size)
	for rows.Next() {
		var pr entity.Product
		if err := rows.Scan(
			&pr.ID,
			&pr.SKU,
			&pr.Name,
			&pr.Price,
			&pr.Stock,
			&pr.IsActive,
			&pr.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		products = append(products, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (p *ProductRepository) FilterProducts(request *model.SearchProductRequest) (string, []any) {
	var conditions []string
	var args []any

	if sku := request.SKU; sku != "" {
		args = append(args, "%"+escapeLike(sku)+"%")
		conditions = append(conditions, fmt.Sprintf("sku ILIKE $%d", len(args)))
	}

	if name := request.Name; name != "" {
		args = append(args, "%"+escapeLike(name)+"%")
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func escapeLike(s string) string {
	return likeEscaper.Replace(s)
}
