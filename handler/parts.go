package handler

import (
	"context"
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

type IDPathParam struct {
	ID int `uri:"id" binding:"required"`
}

func (h *Handler) GetPartByID(c *gin.Context) {
	req := IDPathParam{}
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
	req := IDPathParam{}
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
	req := IDPathParam{}
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

var ErrInsufficientStock = errors.New("insufficient stock")

type AdjustStockRequest struct {
	Amount int    `json:"amount" binding:"required,min=1"`
	Note   string `json:"note"`
}

func (h *Handler) adjustPartStock(ctx context.Context, partID int, amount int, note string) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentStock, reorderLevel int
	var sku, name string
	err = tx.QueryRowContext(ctx,
		"SELECT stock, reorder_level, sku, name FROM parts WHERE id = ? AND deleted_at IS NULL", partID).
		Scan(&currentStock, &reorderLevel, &sku, &name)
	if err != nil {
		return err
	}

	newStock := currentStock + amount
	if newStock < 0 {
		return ErrInsufficientStock
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE parts SET stock = ?, updated_at = datetime('now') WHERE id = ?", newStock, partID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO stock_movements (part_id, quantity, note) VALUES (?, ?, ?)", partID, amount, note)
	if err != nil {
		return err
	}

	part := models.Part{ID: int64(partID), SKU: sku, Name: name, Stock: newStock, ReorderLevel: reorderLevel}
	if amount > 0 {
		if err := h.notifier.ClearNotification(ctx, tx, part.ID); err != nil {
			return err
		}
	} else {
		if err := h.notifier.CheckAndNotify(ctx, tx, part); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (h *Handler) IncreasePartStock(c *gin.Context) {
	req := IDPathParam{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	body := AdjustStockRequest{}
	err = c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.adjustPartStock(c.Request.Context(), req.ID, body.Amount, body.Note); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "part not found"})
			return
		}
		h.logger.ErrorContext(c.Request.Context(), "failed to increase part stock", "id", req.ID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) DecreasePartStock(c *gin.Context) {
	req := IDPathParam{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	body := AdjustStockRequest{}
	err = c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.adjustPartStock(c.Request.Context(), req.ID, -body.Amount, body.Note); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "part not found"})
			return
		}
		if errors.Is(err, ErrInsufficientStock) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient stock"})
			return
		}
		h.logger.ErrorContext(c.Request.Context(), "failed to decrease part stock", "id", req.ID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
