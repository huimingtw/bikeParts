package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/huimingtw/bikeparts/handler"
	"github.com/huimingtw/bikeparts/models"
	"github.com/stretchr/testify/assert"
)

func TestIncreasePartStock(t *testing.T) {
	truncateTables()

	testDB.Exec(
		"INSERT INTO parts (sku, name, stock, reorder_level, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
		"BRK-001", "Brake Pad", 10, 5,
	)

	t.Run("bad request", func(t *testing.T) {
		t.Run("missing amount", func(t *testing.T) {
			b, _ := json.Marshal(handler.AdjustStockRequest{})
			resp := makeRequest("POST", "/api/parts/1/increase", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})

		t.Run("zero amount", func(t *testing.T) {
			b, _ := json.Marshal(map[string]any{"amount": 0})
			resp := makeRequest("POST", "/api/parts/1/increase", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
	})

	t.Run("non-existing part", func(t *testing.T) {
		b, _ := json.Marshal(handler.AdjustStockRequest{Amount: 5})
		resp := makeRequest("POST", "/api/parts/999/increase", bytes.NewReader(b))

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("should increase stock", func(t *testing.T) {
		b, _ := json.Marshal(handler.AdjustStockRequest{Amount: 5, Note: "restock"})
		resp := makeRequest("POST", "/api/parts/1/increase", bytes.NewReader(b))

		assert.Equal(t, http.StatusOK, resp.Code)

		var part models.Part
		testDB.QueryRow("SELECT stock FROM parts WHERE id = 1").Scan(&part.Stock)
		assert.Equal(t, 15, part.Stock)

		// stock movement should be recorded
		var qty int
		testDB.QueryRow("SELECT quantity FROM stock_movements WHERE part_id = 1").Scan(&qty)
		assert.Equal(t, 5, qty)
	})
}

func TestDecreasePartStock(t *testing.T) {
	truncateTables()

	testDB.Exec(
		"INSERT INTO parts (sku, name, stock, reorder_level, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
		"BRK-001", "Brake Pad", 10, 5,
	)

	t.Run("bad request", func(t *testing.T) {
		t.Run("missing amount", func(t *testing.T) {
			b, _ := json.Marshal(handler.AdjustStockRequest{})
			resp := makeRequest("POST", "/api/parts/1/decrease", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})

		t.Run("zero amount", func(t *testing.T) {
			b, _ := json.Marshal(map[string]any{"amount": 0})
			resp := makeRequest("POST", "/api/parts/1/decrease", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
	})

	t.Run("non-existing part", func(t *testing.T) {
		b, _ := json.Marshal(handler.AdjustStockRequest{Amount: 5})
		resp := makeRequest("POST", "/api/parts/999/decrease", bytes.NewReader(b))

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("insufficient stock", func(t *testing.T) {
		b, _ := json.Marshal(handler.AdjustStockRequest{Amount: 999})
		resp := makeRequest("POST", "/api/parts/1/decrease", bytes.NewReader(b))

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("should decrease stock", func(t *testing.T) {
		b, _ := json.Marshal(handler.AdjustStockRequest{Amount: 3, Note: "sold"})
		resp := makeRequest("POST", "/api/parts/1/decrease", bytes.NewReader(b))

		assert.Equal(t, http.StatusOK, resp.Code)

		var part models.Part
		testDB.QueryRow("SELECT stock FROM parts WHERE id = 1").Scan(&part.Stock)
		assert.Equal(t, 7, part.Stock)

		// stock movement should be recorded with negative quantity
		var qty int
		testDB.QueryRow("SELECT quantity FROM stock_movements WHERE part_id = 1").Scan(&qty)
		assert.Equal(t, -3, qty)
	})

	t.Run("should create low stock notification when stock <= reorder_level", func(t *testing.T) {
		// stock is currently 7, reorder_level is 5 — decrease to 5 to trigger notification
		b, _ := json.Marshal(handler.AdjustStockRequest{Amount: 2, Note: "trigger low stock"})
		resp := makeRequest("POST", "/api/parts/1/decrease", bytes.NewReader(b))

		assert.Equal(t, http.StatusOK, resp.Code)

		var count int
		testDB.QueryRow("SELECT COUNT(*) FROM low_stock_notifications WHERE part_id = 1 AND deleted_at IS NULL").Scan(&count)
		assert.Equal(t, 1, count)
	})
}
