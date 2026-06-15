package model

import "time"

// ─────────────────────────────────────────────────
// User — 用户模型
// ─────────────────────────────────────────────────

type User struct {
	ID        uint      `gorm:"primarykey"         json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string    `gorm:"size:255;not null"  json:"-"`          // JSON 序列化时隐藏
	Nickname  string    `gorm:"size:50"            json:"nickname"`
	CreatedAt time.Time `                         json:"created_at"`
	UpdatedAt time.Time `                         json:"updated_at"`
}
