package handler

import (
	"net/http"
	"strconv"

	"mini-blog/model"
	"mini-blog/service"
	"mini-blog/utils"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// ArticleHandler — 文章 HTTP 接口
// ─────────────────────────────────────────────────

type ArticleHandler struct {
	Svc *service.ArticleService
}

type CreateArticleReq struct {
	Title      string `json:"title" binding:"required,max=200"`
	Content    string `json:"content" binding:"required"`
	Summary    string `json:"summary" binding:"max=500"`
	CategoryID uint   `json:"category_id" binding:"required"`
}

type UpdateArticleReq struct {
	Title      string `json:"title" binding:"required,max=200"`
	Content    string `json:"content" binding:"required"`
	Summary    string `json:"summary" binding:"max=500"`
	CategoryID uint   `json:"category_id" binding:"required"`
}

// POST /api/v1/articles — 创建文章（需认证）
func (h *ArticleHandler) Create(c *gin.Context) {
	var req CreateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")

	article := &model.Article{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		UserID:     userID.(uint),
		CategoryID: req.CategoryID,
	}

	if err := h.Svc.Create(article); err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}

	utils.Success(c, article)
}

// GET /api/v1/articles/:id — 获取单篇文章
func (h *ArticleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的文章 ID")
		return
	}

	article, err := h.Svc.GetByID(uint(id))
	if err != nil {
		utils.Error(c, http.StatusNotFound, err.Error())
		return
	}

	utils.Success(c, article)
}

// PUT /api/v1/articles/:id — 更新文章（需认证）
func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的文章 ID")
		return
	}

	var req UpdateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")

	article := &model.Article{
		ID:         uint(id),
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		UserID:     userID.(uint),
		CategoryID: req.CategoryID,
	}

	if err := h.Svc.Update(article); err != nil {
		utils.Error(c, http.StatusForbidden, err.Error())
		return
	}

	utils.Success(c, article)
}

// DELETE /api/v1/articles/:id — 删除文章（需认证，仅作者可删）
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的文章 ID")
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.Svc.Delete(uint(id), userID.(uint)); err != nil {
		utils.Error(c, http.StatusForbidden, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GET /api/v1/articles — 分页查询文章列表
func (h *ArticleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	categoryID, _ := strconv.ParseUint(c.DefaultQuery("category_id", "0"), 10, 64)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	articles, total, err := h.Svc.List(page, pageSize, uint(categoryID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      articles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
