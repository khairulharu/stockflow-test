package model

import (
	"strings"
	"time"
)

type ProductResponse struct {
	ID             int64      `json:"id"`
	SKU            string     `json:"sku"`
	Name           string     `json:"name"`
	Price          int64      `json:"price"`
	Stock          int64      `json:"stock"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	LastMovementAt *time.Time `json:"last_movement_at"`
}

type SearchProductRequest struct {
	Search string `json:"search"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

type CreateProductRequest struct {
	SKU   string `json:"sku" validate:"required,max=50"`
	Name  string `json:"name" validate:"required,max=255"`
	Price int64  `json:"price" validate:"gte=0"`
	Stock int    `json:"stock" validate:"gte=0"`
	Note  string `json:"note" validate:"max=255"`
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (s *SearchProductRequest) EscapedSearch() string {
	return likeEscaper.Replace(s.Search)
}
