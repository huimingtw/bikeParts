package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/huimingtw/bikeparts/handler"
	"github.com/stretchr/testify/assert"
)

func TestGetNotifications(t *testing.T) {
	truncateTables()

	// seed a part with stock below reorder level
	testDB.Exec(
		"INSERT INTO parts (sku, name, stock, reorder_level, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
		"BRK-001", "Brake Pad", 3, 5,
	)
	// manually insert a notification (skipping email side-effect)
	testDB.Exec(
		"INSERT INTO low_stock_notifications (part_id, created_at) VALUES (?, datetime('now'))",
		1,
	)

	t.Run("should return active notifications", func(t *testing.T) {
		resp := makeRequest("GET", "/api/notifications", nil)
		body, _ := io.ReadAll(resp.Body)

		var notifications []handler.NotificationResponse
		json.Unmarshal(body, &notifications)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Len(t, notifications, 1)
		assert.Equal(t, int64(1), notifications[0].PartID)
		assert.Equal(t, "BRK-001", notifications[0].SKU)
		assert.Equal(t, "Brake Pad", notifications[0].Name)
		assert.Equal(t, 5, notifications[0].ReorderLevel)
	})

	t.Run("should not return soft-deleted notifications", func(t *testing.T) {
		testDB.Exec("UPDATE low_stock_notifications SET deleted_at = datetime('now') WHERE part_id = 1")

		resp := makeRequest("GET", "/api/notifications", nil)
		body, _ := io.ReadAll(resp.Body)

		var notifications []handler.NotificationResponse
		json.Unmarshal(body, &notifications)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Len(t, notifications, 0)
	})
}
