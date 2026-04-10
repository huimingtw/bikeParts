package router

import (
	"github.com/gin-gonic/gin"
	"github.com/huimingtw/bikeparts/handler"
	"github.com/huimingtw/bikeparts/middleware"
)

func NewRouter(
	h *handler.Handler,
	ic *middleware.IdempotencyCache,
) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	api.GET("/parts", h.GetParts)
	api.GET("/parts/:id", h.GetPartByID)
	api.POST("/parts", h.CreatePart)
	api.PUT("/parts/:id", h.UpdatePart)
	api.DELETE("/parts/:id", h.DeletePart)
	api.GET("/notifications", h.GetNotifications)

	api.POST("/parts/:id/increase", ic.Middleware(), h.IncreasePartStock)
	api.POST("/parts/:id/decrease", ic.Middleware(), h.DecreasePartStock)

	return router
}
