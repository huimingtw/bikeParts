package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/huimingtw/bikeparts/models"

	"github.com/stretchr/testify/assert"
)

func TestGetParts(t *testing.T) {
	truncateTables()
	now := time.Now()
	parts := []models.Part{
		{SKU: "BRK-001", Name: "Brake Pad", Stock: 10, ReorderLevel: 5},
		{SKU: "TIR-002", Name: "Tire", Stock: 20, ReorderLevel: 10},
		{SKU: "CHN-003", Name: "Chain", Stock: 15, ReorderLevel: 7},
		{SKU: "SAD-004", Name: "Saddle", Stock: 5, ReorderLevel: 3},
		{SKU: "HND-005", Name: "Handlebar", Stock: 8, ReorderLevel: 4},
		// deleted
		{SKU: "DEL-006", Name: "Deleted Part", Stock: 0, ReorderLevel: 0, DeletedAt: &now},
	}
	tx, _ := testDB.Begin()
	for _, p := range parts {
		tx.Exec(
			"INSERT INTO parts (sku, name, stock, reorder_level, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
			p.SKU, p.Name, p.Stock, p.ReorderLevel,
		)
	}
	tx.Commit()
	t.Run("should return all parts", func(t *testing.T) {

		resp := makeRequest("GET", "/api/parts", nil)
		body, _ := io.ReadAll(resp.Body)
		var responseParts []models.Part
		json.Unmarshal(body, &responseParts)

		assert.Equal(t, http.StatusOK, resp.Code)

		assert.Equal(t, len(parts), len(responseParts))
	})

	t.Run("should return part with specified ID", func(t *testing.T) {
		t.Run("existing part ID=1", func(t *testing.T) {
			resp := makeRequest("GET", "/api/parts/1", nil)
			body, _ := io.ReadAll(resp.Body)
			var p models.Part
			json.Unmarshal(body, &p)

			assert.Equal(t, http.StatusOK, resp.Code)

			assert.Equal(t, "BRK-001", p.SKU)
			assert.Equal(t, "Brake Pad", p.Name)
			assert.Equal(t, 10, p.Stock)
			assert.Equal(t, 5, p.ReorderLevel)
		})

		t.Run("existing part ID=3", func(t *testing.T) {
			resp := makeRequest("GET", "/api/parts/3", nil)
			body, _ := io.ReadAll(resp.Body)
			var p models.Part
			json.Unmarshal(body, &p)

			assert.Equal(t, http.StatusOK, resp.Code)

			assert.Equal(t, "CHN-003", p.SKU)
			assert.Equal(t, "Chain", p.Name)
			assert.Equal(t, 15, p.Stock)
			assert.Equal(t, 7, p.ReorderLevel)
		})

		t.Run("non-existing part", func(t *testing.T) {
			resp := makeRequest("GET", "/api/parts/999", nil)
			body, _ := io.ReadAll(resp.Body)
			var errorResp map[string]string
			json.Unmarshal(body, &errorResp)

			assert.Equal(t, http.StatusNotFound, resp.Code)
		})
	})
}
