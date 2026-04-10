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

func TestUpdatePart(t *testing.T) {
	truncateTables()

	// seed one part to update
	testDB.Exec(
		"INSERT INTO parts (sku, name, stock, reorder_level, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
		"BRK-001", "Brake Pad", 10, 5,
	)

	t.Run("bad request", func(t *testing.T) {
		t.Run("missing required fields", func(t *testing.T) {
			body := handler.CreatePartRequest{
				Stock:        10,
				ReorderLevel: 5,
				// missing SKU and Name
			}
			b, _ := json.Marshal(body)
			resp := makeRequest("PUT", "/api/parts/1", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})

		t.Run("invalid stock", func(t *testing.T) {
			body := handler.CreatePartRequest{
				SKU:          "BRK-001",
				Name:         "Brake Pad",
				Stock:        -1,
				ReorderLevel: 5,
			}
			b, _ := json.Marshal(body)
			resp := makeRequest("PUT", "/api/parts/1", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})

		t.Run("invalid reorder level", func(t *testing.T) {
			body := handler.CreatePartRequest{
				SKU:          "BRK-001",
				Name:         "Brake Pad",
				Stock:        10,
				ReorderLevel: -1,
			}
			b, _ := json.Marshal(body)
			resp := makeRequest("PUT", "/api/parts/1", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
	})

	t.Run("non-existing part", func(t *testing.T) {
		body := handler.CreatePartRequest{
			SKU:          "BRK-001",
			Name:         "Brake Pad",
			Stock:        10,
			ReorderLevel: 5,
		}
		b, _ := json.Marshal(body)
		resp := makeRequest("PUT", "/api/parts/999", bytes.NewReader(b))

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("should update part", func(t *testing.T) {
		body := handler.CreatePartRequest{
			SKU:          "BRK-001-UPDATED",
			Name:         "Brake Pad Updated",
			Stock:        20,
			ReorderLevel: 8,
		}
		b, _ := json.Marshal(body)
		resp := makeRequest("PUT", "/api/parts/1", bytes.NewReader(b))

		assert.Equal(t, http.StatusOK, resp.Code)

		var part models.Part
		testDB.QueryRow("SELECT sku, name, stock, reorder_level FROM parts WHERE id = 1").
			Scan(&part.SKU, &part.Name, &part.Stock, &part.ReorderLevel)

		assert.Equal(t, "BRK-001-UPDATED", part.SKU)
		assert.Equal(t, "Brake Pad Updated", part.Name)
		assert.Equal(t, 20, part.Stock)
		assert.Equal(t, 8, part.ReorderLevel)
	})
}
