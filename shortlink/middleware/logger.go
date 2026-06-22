package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		fmt.Printf("[%s] %3d | %13v | %-6s | %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Writer.Status(),
			latency,
			c.Request.Method,
			path,
		)
	}
}
