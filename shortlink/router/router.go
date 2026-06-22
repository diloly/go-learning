package router

import (
	"shortlink/handler"
	"shortlink/middleware"

	"github.com/gin-gonic/gin"
)

// Setup 注册路由
func Setup(h *handler.ShortLinkHandler, mode string) *gin.Engine {
	gin.SetMode(mode)

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	// 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// ── 短链接核心接口 ──
	// 注意: GET /:code 必须放在最后，否则会覆盖上面的路由
	api := r.Group("/api/v1")
	{
		api.POST("/shorten", h.Shorten)
		api.GET("/:code/stats", h.Stats)
	}

	// 重定向（放在根路径，短码直接跟在域名后）
	r.GET("/:code", h.Redirect)

	return r
}
