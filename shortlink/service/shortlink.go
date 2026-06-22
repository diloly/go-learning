package service

import (
	"context"
	"errors"
	"log"

	"shortlink/model"
	"shortlink/store"
)

// ─────────────────────────────────────────────────
// ShortLinkService — 核心业务逻辑
//
// 职责: 协调 DB 持久化 + Redis 缓存 + 计数器
// ─────────────────────────────────────────────────

type ShortLinkService struct {
	DB    *store.DBStore
	Cache *store.RedisStore
}

// CreateShortLink 创建短链接
// 流程: 写入 MySQL → 生成短码 → 预热 Redis 缓存
func (s *ShortLinkService) CreateShortLink(ctx context.Context, originalURL string) (*model.ShortLink, error) {
	link, err := s.DB.Create(originalURL)
	if err != nil {
		return nil, err
	}

	// 预热缓存：创建即缓存，首次访问不 miss
	if err := s.Cache.SetURL(ctx, link.ShortCode, link.OriginalURL); err != nil {
		// 缓存失败不影响主流程，仅打日志
		log.Printf("[warn] 缓存预热失败: short_code=%s, err=%v", link.ShortCode, err)
	}

	return link, nil
}

// AccessShortLink 访问短链接（核心路径）
//
// 流程:
//   1. Redis 查询缓存 → hit 直接返回
//   2. 缓存 miss → MySQL 查询 → 回填缓存
//   3. Redis INCR 原子计数
//   4. MySQL 更新计数（异步 goroutine，不阻塞响应）
//
// 返回 originalURL 和 error
func (s *ShortLinkService) AccessShortLink(ctx context.Context, shortCode string) (string, error) {
	// ── Step 1: 查缓存 ──
	originalURL, err := s.Cache.GetURL(ctx, shortCode)
	if err == nil && originalURL != "" {
		// 缓存命中，计数后直接返回
		go func() { _ = s.asyncCount(ctx, shortCode) }()
		return originalURL, nil
	}

	// ── Step 2: 缓存 miss，查数据库 ──
	link, err := s.DB.GetByShortCode(shortCode)
	if err != nil {
		return "", err
	}
	if link == nil {
		return "", errors.New("短链接不存在")
	}

	// ── Step 3: 回填缓存 ──
	if err := s.Cache.SetURL(ctx, link.ShortCode, link.OriginalURL); err != nil {
		log.Printf("[warn] 缓存回填失败: short_code=%s, err=%v", shortCode, err)
	}

	// ── Step 4: 异步计数 ──
	go func() { _ = s.asyncCount(ctx, shortCode) }()

	return link.OriginalURL, nil
}

// asyncCount 异步更新 Redis 计数 + MySQL 计数
func (s *ShortLinkService) asyncCount(ctx context.Context, shortCode string) error {
	if _, err := s.Cache.IncrCount(ctx, shortCode); err != nil {
		log.Printf("[warn] Redis 计数失败: short_code=%s, err=%v", shortCode, err)
	}
	if err := s.DB.IncrementCount(shortCode); err != nil {
		log.Printf("[warn] MySQL 计数更新失败: short_code=%s, err=%v", shortCode, err)
	}
	return nil
}

// GetStats 获取短链接统计信息
// 合并 MySQL 元数据 + Redis 实时计数
func (s *ShortLinkService) GetStats(ctx context.Context, shortCode string) (*model.ShortLink, int64, error) {
	link, err := s.DB.GetStats(shortCode)
	if err != nil {
		return nil, 0, err
	}
	if link == nil {
		return nil, 0, errors.New("短链接不存在")
	}

	// 补充 Redis 实时计数（如果 Redis 有更最新的值）
	redisCount, err := s.Cache.GetCount(ctx, shortCode)
	if err == nil && redisCount > link.AccessCount {
		link.AccessCount = redisCount
	}

	return link, link.AccessCount, nil
}
