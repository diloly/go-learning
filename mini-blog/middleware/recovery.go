package middleware

import (
	"net/http"
	"runtime/debug"

	"mini-blog/utils"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// Recovery — 异常捕获中间件
// 防止 panic 导致进程崩溃，返回 500 保底
// ─────────────────────────────────────────────────

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 打印堆栈（生产环境应写入日志文件）
				debug.PrintStack()

				utils.Error(c, http.StatusInternalServerError, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
