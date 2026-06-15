package handler

import (
	"net/http"

	"mini-blog/service"
	"mini-blog/utils"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// CategoryHandler — 分类 HTTP 接口
// ─────────────────────────────────────────────────

type CategoryHandler struct {
	Svc *service.CategoryService
}

type CreateCategoryReq struct {
	Name string `json:"name" binding:"required,max=50"`
}

// GET /api/v1/categories — 获取全部分类
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.Svc.List()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取分类失败")
		return
	}
	utils.Success(c, categories)
}

// POST /api/v1/categories — 创建分类（需认证）
func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	category, err := h.Svc.Create(req.Name)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "创建分类失败，可能已存在")
		return
	}

	utils.Success(c, category)
}
