package models

import "time"

type StockMovement struct {
	ID        int64     `json:"id"`
	PartID    int64     `json:"part_id"`
	Quantity  int       `json:"quantity"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}
