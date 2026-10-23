package model

import (
	"time"
)

type ProductResponse struct {
	ID        int64     `json:"id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Price     int64     `json:"price"`
	Stock     int64     `json:"stock"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchProductRequest struct {
	SKU  string `json:"sku" validate:"max=20"`
	Name string `json:"name" validate:"max=200"`
	Page int    `json:"page" validate:"min=1"`
	Size int    `json:"limit" validate:"min=1,max=100"`
}
