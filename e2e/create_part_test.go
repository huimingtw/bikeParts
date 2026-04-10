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

func TestCreatePart(t *testing.T) {
	truncateTables()

	t.Run("bad request", func(t *testing.T) {
		t.Run("missing required fields", func(t *testing.T) {
			body := handler.CreatePartRequest{
				Stock:        10,
				ReorderLevel: 5,
				// missing SKU and Name
			}
			b, _ := json.Marshal(body)
			resp := makeRequest("POST", "/api/parts", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
		t.Run("invalid stock", func(t *testing.T) {
			body := handler.CreatePartRequest{
				SKU:          "INV-001",
				Name:         "Invalid Stock",
				Stock:        -5, // invalid negative stock
				ReorderLevel: 5,
			}
			b, _ := json.Marshal(body)
			resp := makeRequest("POST", "/api/parts", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
		t.Run("invalid reorder level", func(t *testing.T) {
			body := handler.CreatePartRequest{
				SKU:          "INV-002",
				Name:         "Invalid Reorder Level",
				Stock:        10,
				ReorderLevel: -3, // invalid negative reorder level
			}
			b, _ := json.Marshal(body)
			resp := makeRequest("POST", "/api/parts", bytes.NewReader(b))

			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
	})

	t.Run("should create a new part", func(t *testing.T) {
		body := handler.CreatePartRequest{
			SKU:          "NEW-001",
			Name:         "New Part",
			Stock:        50,
			ReorderLevel: 10,
		}
		b, _ := json.Marshal(body)
		resp := makeRequest("POST", "/api/parts", bytes.NewReader(b))

		part := models.Part{}
		err := testDB.QueryRow("SELECT id, sku, name, stock, reorder_level FROM parts WHERE sku = ?", body.SKU).
			Scan(&part.ID, &part.SKU, &part.Name, &part.Stock, &part.ReorderLevel)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusCreated, resp.Code)
		assert.Equal(t, "NEW-001", part.SKU)
		assert.Equal(t, "New Part", part.Name)
		assert.Equal(t, 50, part.Stock)
		assert.Equal(t, 10, part.ReorderLevel)
	})
}
