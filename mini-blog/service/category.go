package service

import (
	"mini-blog/model"

	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────
// CategoryService — 分类业务逻辑
// ─────────────────────────────────────────────────

type CategoryService struct {
	DB *gorm.DB
}

// List 获取全部分类列表
func (s *CategoryService) List() ([]model.Category, error) {
	var categories []model.Category
	err := s.DB.Order("id ASC").Find(&categories).Error
	return categories, err
}

// Create 创建分类
func (s *CategoryService) Create(name string) (*model.Category, error) {
	category := &model.Category{Name: name}
	err := s.DB.Create(category).Error
	return category, err
}
