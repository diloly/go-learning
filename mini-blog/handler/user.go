package handler

import (
	"net/http"

	"mini-blog/service"
	"mini-blog/utils"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────────
// UserHandler — 用户 HTTP 接口
// ─────────────────────────────────────────────────

type UserHandler struct {
	Svc *service.UserService
}

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname" binding:"required,min=1,max=50"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// POST /api/v1/register
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := h.Svc.Register(req.Username, req.Password, req.Nickname)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
	})
}

// POST /api/v1/login
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	// 验证用户名密码
	token, user, err := h.Svc.Login(req.Username, req.Password)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	// token 在 service 没生成(因为 service 不含 jwt 配置)，在 handler 生成
	// 实际上我们调整: 让 Login 只验证，token 在这里生成
	if token == "" {
		token, err = utils.GenerateToken(user.ID, user.Username, 24)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "生成令牌失败")
			return
		}
	}

	utils.Success(c, gin.H{
		"token":    token,
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
	})
}
