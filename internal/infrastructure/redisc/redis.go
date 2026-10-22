// Package redisc 提供 Redis 客户端与限流实现。
package redisc

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewClient 构建 Redis 客户端并 Ping 校验连通性。
func NewClient(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

// Limiter 基于 Redis 的固定窗口限流，实现 application.RateLimiter 端口。
type Limiter struct {
	rdb *redis.Client
}

// NewLimiter 构建限流器。
func NewLimiter(rdb *redis.Client) *Limiter { return &Limiter{rdb: rdb} }

// Allow 固定窗口计数；超过 limit 拒绝。
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	pipe := l.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return incr.Val() <= int64(limit), nil
}
