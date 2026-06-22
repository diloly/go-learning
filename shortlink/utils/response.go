package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// 统一 API 响应格式
// ─────────────────────────────────────────────────

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, Response{
		Code:    -1,
		Message: msg,
	})
}
