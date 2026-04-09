package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type NotificationResponse struct {
	ID           int64     `json:"id"`
	PartID       int64     `json:"part_id"`
	SKU          string    `json:"sku"`
	Name         string    `json:"name"`
	ReorderLevel int       `json:"reorder_level"`
	CreatedAt    time.Time `json:"created_at"`
}

func (h *Handler) GetNotifications(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.db.QueryContext(ctx, `
	SELECT n.id, n.part_id, p.sku, p.name, p.reorder_level, n.created_at
		FROM low_stock_notifications n
		JOIN parts p ON n.part_id = p.id
		WHERE n.deleted_at IS NULL
	`)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to query notifications", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer rows.Close()

	notifications := []NotificationResponse{}
	for rows.Next() {
		var n NotificationResponse
		if err := rows.Scan(&n.ID, &n.PartID, &n.SKU, &n.Name, &n.ReorderLevel, &n.CreatedAt); err != nil {
			h.logger.ErrorContext(ctx, "failed to scan notification", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		h.logger.ErrorContext(ctx, "row iteration error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}
