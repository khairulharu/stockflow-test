package model

type SearchProductRequest struct {
	SKU  string `json:"sku" validate:"max=20"`
	Name string `json:"name" validate:"max=200"`
	Page int    `json:"page" validate:"min=1"`
	Size int    `json:"limit" validate:"min=1,max=100"`
}
