package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// 统一 API 响应格式
// ─────────────────────────────────────────────────

type Response struct {
	Code    int         `json:"code"`    // 0=成功，非0=失败
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data,omitempty"`
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, Response{
		Code:    -1,
		Message: msg,
	})
}
