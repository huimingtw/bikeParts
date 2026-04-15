package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huimingtw/bikeparts/handler"
	"github.com/huimingtw/bikeparts/middleware"
)

func NewRouter(
	h *handler.Handler,
	ic *middleware.IdempotencyCache,
	logger *slog.Logger,
	assets http.FileSystem,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))

	router.StaticFS("/static", assets)
	router.GET("/", func(c *gin.Context) {
		c.FileFromFS("index.html", assets)
	})

	api := router.Group("/api")
	api.GET("/parts", h.GetParts)
	api.GET("/parts/:id", h.GetPartByID)
	api.POST("/parts", h.CreatePart)
	api.PUT("/parts/:id", h.UpdatePart)
	api.DELETE("/parts/:id", h.DeletePart)
	api.GET("/notifications", h.GetNotifications)

	api.POST("/parts/:id/increase", ic.Middleware(), h.IncreasePartStock)
	api.POST("/parts/:id/decrease", ic.Middleware(), h.DecreasePartStock)

	api.GET("/settings", h.GetSettings)
	api.PUT("/settings", h.SaveSettings)
	api.POST("/settings/test-email", h.TestEmail)

	return router
}
