package router

import (
	"io/fs"
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
	frontendFS fs.FS,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))

	router.StaticFS("/static", http.FS(frontendFS))
	router.GET("/", func(c *gin.Context) {
		// 直接讀取並回傳，避免 http.FileServer 對 /index.html 發出 ./ redirect 造成無限迴圈
		data, err := fs.ReadFile(frontendFS, "index.html")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
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
