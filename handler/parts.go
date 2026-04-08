package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huimingtw/bikeparts/models"
)

func (h *Handler) GetParts(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.db.QueryContext(ctx, `
		SELECT id, sku, name, stock, reorder_level, created_at, updated_at
		FROM parts
		WHERE deleted_at IS NULL
	`)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to query parts", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer rows.Close()

	parts := []models.Part{}
	for rows.Next() {
		var p models.Part
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Stock, &p.ReorderLevel, &p.CreatedAt, &p.UpdatedAt); err != nil {
			h.logger.ErrorContext(ctx, "failed to scan part", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		parts = append(parts, p)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, parts)
}

func (h *Handler) GetPartByID(c *gin.Context) {
	ctx := c.Request.Context()
	var p models.Part
	err := h.db.QueryRowContext(ctx, `
		SELECT id, sku, name, stock, reorder_level, created_at, updated_at
		FROM parts
		WHERE id = ? AND deleted_at IS NULL
	`, c.Param("id"),
	).Scan(&p.ID, &p.SKU, &p.Name, &p.Stock, &p.ReorderLevel, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "part not found"})
		return
	}
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to scan part by id", "id", c.Param("id"), "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) CreatePart(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handler) UpdatePart(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handler) DeletePart(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
