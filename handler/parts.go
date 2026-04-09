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

type GetPartByIDRequest struct {
	ID int `uri:"id" binding:"required"`
}

func (h *Handler) GetPartByID(c *gin.Context) {
	req := GetPartByIDRequest{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()
	var p models.Part
	err = h.db.QueryRowContext(ctx, `
		SELECT id, sku, name, stock, reorder_level, created_at, updated_at
		FROM parts
		WHERE id = ? AND deleted_at IS NULL
	`, req.ID,
	).Scan(&p.ID, &p.SKU, &p.Name, &p.Stock, &p.ReorderLevel, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "part not found"})
		return
	}
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to scan part by id", "id", req.ID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, p)
}

type CreatePartRequest struct {
	SKU          string `json:"sku" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Stock        int    `json:"stock" binding:"min=0"`
	ReorderLevel int    `json:"reorder_level" binding:"min=0"`
}

func (h *Handler) CreatePart(c *gin.Context) {
	req := CreatePartRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO parts (sku, name, stock, reorder_level, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	`, req.SKU, req.Name, req.Stock, req.ReorderLevel)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to insert part", "sku", req.SKU, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func (h *Handler) UpdatePart(c *gin.Context) {
	req := GetPartByIDRequest{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	body := CreatePartRequest{}
	err = c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()
	result, err := h.db.ExecContext(ctx, `
		UPDATE parts
		SET sku = ?, name = ?, stock = ?, reorder_level = ?, updated_at = datetime('now')
		WHERE id = ? AND deleted_at IS NULL
	`, body.SKU, body.Name, body.Stock, body.ReorderLevel, req.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to update part", "id", req.ID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get rows affected for update", "id", req.ID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "part not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) DeletePart(c *gin.Context) {
	req := GetPartByIDRequest{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()
	result, err := h.db.ExecContext(ctx, `
		UPDATE parts
		SET deleted_at = datetime('now')
		WHERE id = ? AND deleted_at IS NULL
	`, req.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to delete part", "id", req.ID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get rows affected for delete", "id", req.ID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "part not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
