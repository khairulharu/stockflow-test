package converter

import (
	"stockflow/internal/entity"
	"stockflow/internal/model"
)

func ProductToResponse(product *entity.Product) *model.ProductResponse {
	return &model.ProductResponse{
		ID:        product.ID,
		SKU:       product.SKU,
		Name:      product.Name,
		Price:     product.Price,
		Stock:     product.Stock,
		IsActive:  product.IsActive,
		CreatedAt: product.CreatedAt,
	}
}
