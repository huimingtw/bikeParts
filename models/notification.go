package models

import "time"

type LowStockNotification struct {
	ID        int64      `json:"id"`
	PartID    int64      `json:"part_id"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
