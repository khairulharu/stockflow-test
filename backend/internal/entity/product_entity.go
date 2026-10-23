package entity

import "time"

type Product struct {
	ID        int64
	SKU       string
	Name      string
	Price     int64
	Stock     int64
	IsActive  bool
	CreatedAt time.Time
}
