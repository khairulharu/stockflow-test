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

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (s *SearchProductRequest) EscapedSearch() string {
	return likeEscaper.Replace(s.Search)
}
