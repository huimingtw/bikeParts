package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huimingtw/bikeparts/models"
)

func (h *Handler) MailTest(c *gin.Context) {
	part := models.Part{
		ID:           1,
		Name:         "Test Part",
		SKU:          "TEST123",
		Stock:        5,
		ReorderLevel: 10,
	}
	err := h.mailer.SendLowStockEmail(part)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
