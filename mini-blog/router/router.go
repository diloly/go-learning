package router

import (
	"mini-blog/handler"
	"mini-blog/middleware"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// Setup — 路由注册
// 定义所有 API 端点及其对应的 handler
// ─────────────────────────────────────────────────

func Setup(
	userHandler *handler.UserHandler,
	articleHandler *handler.ArticleHandler,
	categoryHandler *handler.CategoryHandler,
	mode string,
) *gin.Engine {
	gin.SetMode(mode)

	// 使用 gin.New() 而非 gin.Default()，手动挂载中间件
	r := gin.New()

	// ── 全局中间件（对所有请求生效） ──
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	// ── 健康检查 ──
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// ── API v1 路由组 ──
	api := r.Group("/api/v1")
	{
		// 用户模块 — 无需认证
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)

		// 分类模块 — 公开查询
		api.GET("/categories", categoryHandler.List)

		// 文章模块 — 公开查询
		api.GET("/articles", articleHandler.List)
		api.GET("/articles/:id", articleHandler.GetByID)

		// ── 需认证的路由组 ──
		auth := api.Group("")
		auth.Use(middleware.Auth())
		{
			// 分类 — 需认证
			auth.POST("/categories", categoryHandler.Create)

			// 文章 CRUD — 需认证
			auth.POST("/articles", articleHandler.Create)
			auth.PUT("/articles/:id", articleHandler.Update)
			auth.DELETE("/articles/:id", articleHandler.Delete)
		}
	}

	return r
}
