package handler

import (
	"context"
	"net/http"

	"shortlink/service"
	"shortlink/utils"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// ShortLinkHandler — HTTP 接口层
// ─────────────────────────────────────────────────

type ShortLinkHandler struct {
	Svc *service.ShortLinkService
}

type ShortenReq struct {
	URL string `json:"url" binding:"required,url"`
}

// Shorten  POST /api/v1/shorten  创建短链接
func (h *ShortLinkHandler) Shorten(c *gin.Context) {
	var req ShortenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误，需要有效的 url 字段")
		return
	}

	link, err := h.Svc.CreateShortLink(context.Background(), req.URL)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建短链接失败")
		return
	}

	utils.Success(c, gin.H{
		"short_code": link.ShortCode,
		"original_url": link.OriginalURL,
	})
}

// Redirect  GET /:code  访问短链接 → 302 重定向
func (h *ShortLinkHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		utils.Error(c, http.StatusBadRequest, "缺少短码")
		return
	}

	originalURL, err := h.Svc.AccessShortLink(context.Background(), shortCode)
	if err != nil {
		if err.Error() == "短链接不存在" {
			utils.Error(c, http.StatusNotFound, "短链接不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "服务异常")
		return
	}

	// 302 临时重定向到原链接
	c.Redirect(http.StatusFound, originalURL)
}

// Stats  GET /api/v1/:code/stats  查看短链接统计
func (h *ShortLinkHandler) Stats(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		utils.Error(c, http.StatusBadRequest, "缺少短码")
		return
	}

	link, count, err := h.Svc.GetStats(context.Background(), shortCode)
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"short_code":   link.ShortCode,
		"original_url": link.OriginalURL,
		"access_count": count,
		"created_at":   link.CreatedAt,
	})
}
