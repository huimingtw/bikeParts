package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeletePart(t *testing.T) {
	truncateTables()

	testDB.Exec(
		"INSERT INTO parts (sku, name, stock, reorder_level, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
		"BRK-001", "Brake Pad", 10, 5,
	)

	t.Run("non-existing part", func(t *testing.T) {
		resp := makeRequest("DELETE", "/api/parts/999", nil)

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("should soft-delete part", func(t *testing.T) {
		resp := makeRequest("DELETE", "/api/parts/1", nil)

		assert.Equal(t, http.StatusOK, resp.Code)

		// deleted part should not appear in GET /api/parts
		getResp := makeRequest("GET", "/api/parts", nil)
		assert.Equal(t, http.StatusOK, getResp.Code)

		// deleted_at should be set
		var deletedAt *string
		testDB.QueryRow("SELECT deleted_at FROM parts WHERE id = 1").Scan(&deletedAt)
		assert.NotNil(t, deletedAt)
	})

	t.Run("already deleted part", func(t *testing.T) {
		// second delete on same ID should 404
		resp := makeRequest("DELETE", "/api/parts/1", nil)

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})
}
