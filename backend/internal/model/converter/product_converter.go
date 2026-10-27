package converter

import (
	"stockflow/internal/entity"
	"stockflow/internal/model"
	"time"
)

func ProductWithMovementToResponse(product *entity.ProductWithMovement) *model.ProductResponseWithMovement {
	var lastMovementAt *time.Time
	if product.LastMovementAt.Valid {
		t := product.LastMovementAt.Time
		lastMovementAt = &t
	}
	return &model.ProductResponseWithMovement{
		ID:             product.ID,
		SKU:            product.SKU,
		Name:           product.Name,
		Price:          product.Price,
		Stock:          product.Stock,
		IsActive:       product.IsActive,
		CreatedAt:      product.CreatedAt,
		LastMovementAt: lastMovementAt,
	}
}

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
