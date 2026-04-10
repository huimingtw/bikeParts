package e2e

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/huimingtw/bikeparts/models"

	"github.com/stretchr/testify/assert"
)

func TestGetParts(t *testing.T) {
	t.Run("should return all parts", func(t *testing.T) {
		truncateTables()
		parts := []models.Part{
			{SKU: "BRK-001", Name: "Brake Pad", Stock: 10, ReorderLevel: 5},
			{SKU: "TIR-002", Name: "Tire", Stock: 20, ReorderLevel: 10},
			{SKU: "CHN-003", Name: "Chain", Stock: 15, ReorderLevel: 7},
			{SKU: "SAD-004", Name: "Saddle", Stock: 5, ReorderLevel: 3},
			{SKU: "HND-005", Name: "Handlebar", Stock: 8, ReorderLevel: 4},
		}
		tx, _ := testDB.Begin()
		for _, p := range parts {
			tx.Exec(
				"INSERT INTO parts (sku, name, stock, reorder_level, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
				p.SKU, p.Name, p.Stock, p.ReorderLevel,
			)
		}
		tx.Commit()

		resp := makeRequest("GET", "/api/parts", nil)
		body, _ := io.ReadAll(resp.Body)
		var responseParts []models.Part
		json.Unmarshal(body, &responseParts)

		assert.Equal(t, 200, resp.Code)
		assert.Equal(t, "application/json; charset=utf-8", resp.Header().Get("Content-Type"))
		assert.Equal(t, len(parts), len(responseParts))
	})
}
