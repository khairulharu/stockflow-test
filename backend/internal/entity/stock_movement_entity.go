package entity

import (
	"time"
)

type StockMovement struct {
	ID        int64
	ProductID int64
	Type      string
	Qty       int
	Note      string
	CreatedAt time.Time
}
