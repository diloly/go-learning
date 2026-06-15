package service

import (
	"errors"

	"mini-blog/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────
// UserService — 用户业务逻辑
// ─────────────────────────────────────────────────

type UserService struct {
	DB *gorm.DB
}

// Register 注册：密码加密（bcrypt）→ 写入数据库
func (s *UserService) Register(username, password, nickname string) (*model.User, error) {
	// 检查用户名是否已存在
	var existing model.User
	if err := s.DB.Where("username = ?", username).First(&existing); err == nil {
		return nil, errors.New("用户名已存在")
	}

	// bcrypt 加密密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		Username: username,
		Password: string(hashed),
		Nickname: nickname,
	}

	if err := s.DB.Create(user).Error; err != nil {
		return nil, errors.New("创建用户失败")
	}

	return user, nil
}

// Login 登录：验证密码 → 返回 JWT Token
func (s *UserService) Login(username, password string) (string, *model.User, error) {
	var user model.User
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 比对密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// token 交给 handler 层生成（因为 handler 知道 jwt 配置）
	return "", &user, nil
}

// GetByID 按 ID 查找用户
func (s *UserService) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := s.DB.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}
