package store

import (
	"context"
	"fmt"
	"time"

	"shortlink/config"

	"github.com/redis/go-redis/v9"
)

// ─────────────────────────────────────────────────
// RedisStore — Redis 缓存层
//
// 缓存策略:
//   - 短链接缓存: shortlink:{code} → original_url（带 TTL）
//   - 访问计数:   shortlink:count:{code} → 累计访问次数（用 INCR）
// ─────────────────────────────────────────────────

type RedisStore struct {
	Client  *redis.Client
	TTL     time.Duration // 缓存过期时间
}

func NewRedisStore(cfg config.RedisConfig, ttl time.Duration) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 连接失败: %w", err)
	}

	fmt.Println("✅ Redis 连接成功")
	return &RedisStore{Client: client, TTL: ttl}, nil
}

// ── 短链接缓存 ─────────────────────────────────

// GetURL 从缓存获取原始链接
func (r *RedisStore) GetURL(ctx context.Context, shortCode string) (string, error) {
	return r.Client.Get(ctx, "shortlink:"+shortCode).Result()
}

// SetURL 写入缓存（带 TTL）
func (r *RedisStore) SetURL(ctx context.Context, shortCode, originalURL string) error {
	return r.Client.Set(ctx, "shortlink:"+shortCode, originalURL, r.TTL).Err()
}

// ── 访问计数 ────────────────────────────────────

// IncrCount 增加一次访问计数（原子操作）
func (r *RedisStore) IncrCount(ctx context.Context, shortCode string) (int64, error) {
	return r.Client.Incr(ctx, "shortlink:count:"+shortCode).Result()
}

// GetCount 获取当前访问计数
func (r *RedisStore) GetCount(ctx context.Context, shortCode string) (int64, error) {
	return r.Client.Get(ctx, "shortlink:count:"+shortCode).Int64()
}

// ── 生命周期 ────────────────────────────────────

// Close 关闭 Redis 连接
func (r *RedisStore) Close() error {
	return r.Client.Close()
}
