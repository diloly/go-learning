package middleware

import (
	"strings"

	"net/http"

	"mini-blog/utils"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// Auth — JWT 鉴权中间件
// 从 Authorization: Bearer <token> 中解析用户身份
// ─────────────────────────────────────────────────

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "缺少认证令牌")
			c.Abort()
			return
		}

		// 格式: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c, http.StatusUnauthorized, "认证格式错误，使用 Bearer <token>")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "令牌无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息注入上下文，后续 handler 可通过 c.Get() 获取
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
