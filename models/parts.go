package models

import "time"

type Part struct {
	ID           int64      `json:"id"`
	SKU          string     `json:"sku"`
	Name         string     `json:"name"`
	Stock        int        `json:"stock"`
	ReorderLevel int        `json:"reorder_level"`
	Location     string     `json:"location"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}
