package service

import (
	"errors"

	"mini-blog/model"

	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────
// ArticleService — 文章业务逻辑
// ─────────────────────────────────────────────────

type ArticleService struct {
	DB *gorm.DB
}

// Create 创建文章
func (s *ArticleService) Create(article *model.Article) error {
	return s.DB.Create(article).Error
}

// GetByID 根据 ID 获取文章（含作者、分类信息）
func (s *ArticleService) GetByID(id uint) (*model.Article, error) {
	var article model.Article
	err := s.DB.
		Preload("User").
		Preload("Category").
		First(&article, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, err
	}
	return &article, nil
}

// Update 更新文章（只更新非零字段）
func (s *ArticleService) Update(article *model.Article) error {
	// 校验文章存在且属于当前用户
	var existing model.Article
	if err := s.DB.First(&existing, article.ID).Error; err != nil {
		return errors.New("文章不存在")
	}
	if existing.UserID != article.UserID {
		return errors.New("无权修改此文章")
	}

	return s.DB.Model(&existing).Updates(map[string]interface{}{
		"title":       article.Title,
		"content":     article.Content,
		"summary":     article.Summary,
		"category_id": article.CategoryID,
	}).Error
}

// Delete 删除文章（仅作者可删）
func (s *ArticleService) Delete(id, userID uint) error {
	result := s.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Article{})
	if result.RowsAffected == 0 {
		return errors.New("文章不存在或无权限删除")
	}
	return result.Error
}

// List 分页查询文章列表
func (s *ArticleService) List(page, pageSize int, categoryID uint) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := s.DB.Model(&model.Article{})
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	// 先查总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 再查分页数据
	err := query.
		Preload("User").
		Preload("Category").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&articles).Error

	return articles, total, err
}
