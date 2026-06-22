package model

import "time"

// ─────────────────────────────────────────────────
// ShortLink — 短链接数据模型
// ─────────────────────────────────────────────────

type ShortLink struct {
	ID          uint64     `gorm:"primarykey;autoIncrement"  json:"id"`
	ShortCode   string     `gorm:"uniqueIndex;size:10;not null" json:"short_code"`
	OriginalURL string     `gorm:"type:text;not null"       json:"original_url"`
	AccessCount int64      `gorm:"default:0"                json:"access_count"`
	ExpiresAt   *time.Time `                                 json:"expires_at,omitempty"`
	CreatedAt   time.Time  `                                 json:"created_at"`
	UpdatedAt   time.Time  `                                 json:"updated_at"`
}

// ─────────────────────────────────────────────────
// Base62 编码 — 将数字 ID 转为短码
//
// 原理: 62 进制（a-z + A-Z + 0-9）
// 6 位短码可表示 62^6 ≈ 568 亿个链接，足够日常使用
// ─────────────────────────────────────────────────

const base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// EncodeBase62 将数字 ID 编码为短码
func EncodeBase62(id uint64) string {
	if id == 0 {
		return string(base62Chars[0])
	}
	var result []byte
	for id > 0 {
		result = append([]byte{base62Chars[id%62]}, result...)
		id /= 62
	}
	return string(result)
}
