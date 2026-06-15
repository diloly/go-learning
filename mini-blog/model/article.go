package model

import "time"

// ─────────────────────────────────────────────────
// Category — 文章分类
// ─────────────────────────────────────────────────

type Category struct {
	ID        uint      `gorm:"primarykey"         json:"id"`
	Name      string    `gorm:"uniqueIndex;size:50;not null" json:"name"`
	CreatedAt time.Time `                         json:"created_at"`
	UpdatedAt time.Time `                         json:"updated_at"`
}

// ─────────────────────────────────────────────────
// Article — 文章模型
// ─────────────────────────────────────────────────

type Article struct {
	ID         uint      `gorm:"primarykey"         json:"id"`
	Title      string    `gorm:"size:200;not null"  json:"title"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Summary    string    `gorm:"size:500"           json:"summary"`
	UserID     uint      `gorm:"not null;index"     json:"user_id"`
	User       User      `gorm:"foreignKey:UserID"  json:"user,omitempty"`
	CategoryID uint      `gorm:"not null;index"     json:"category_id"`
	Category   Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CreatedAt  time.Time `                          json:"created_at"`
	UpdatedAt  time.Time `                          json:"updated_at"`
}
