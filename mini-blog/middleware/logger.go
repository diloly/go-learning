package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// Logger — 请求日志中间件
// 记录每个请求的方法、路径、状态码、耗时
// ─────────────────────────────────────────────────

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// 放行，执行后续 handler
		c.Next()

		// handler 执行完毕，记录日志
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		// 生产环境可以用 logrus/zap 替代 fmt
		fmt.Printf("[LOG] %s | %3d | %13v | %-6s | %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			status,
			latency,
			method,
			path,
		)
	}
}
