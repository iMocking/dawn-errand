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

// CodeStore 基于 Redis 的一次性验证码存储，实现 application.CodeStore 端口。
type CodeStore struct {
	rdb *redis.Client
}

// NewCodeStore 构建验证码存储。
func NewCodeStore(rdb *redis.Client) *CodeStore { return &CodeStore{rdb: rdb} }

// Set 写入验证码并设置过期时间。
func (s *CodeStore) Set(ctx context.Context, key, code string, ttl time.Duration) error {
	return s.rdb.Set(ctx, key, code, ttl).Err()
}

// Get 读取验证码；不存在返回空串（不报错）。
func (s *CodeStore) Get(ctx context.Context, key string) (string, error) {
	v, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return v, err
}

// Delete 删除验证码（验证成功后作废）。
func (s *CodeStore) Delete(ctx context.Context, key string) error {
	return s.rdb.Del(ctx, key).Err()
}

// Locker 基于 Redis SETNX 的分布式锁，实现 application.Locker。
type Locker struct {
	rdb *redis.Client
}

// NewLocker 构建分布式锁。
func NewLocker(rdb *redis.Client) *Locker { return &Locker{rdb: rdb} }

// TryLock 尝试加锁（SET NX + TTL）。
func (l *Locker) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return l.rdb.SetNX(ctx, key, "1", ttl).Result()
}

// Unlock 释放锁。
func (l *Locker) Unlock(ctx context.Context, key string) error {
	return l.rdb.Del(ctx, key).Err()
}

// GeoStore 基于 Redis GEO 的骑手位置索引。
type GeoStore struct {
	rdb *redis.Client
}

// NewGeoStore 构建位置索引。
func NewGeoStore(rdb *redis.Client) *GeoStore { return &GeoStore{rdb: rdb} }

// SetRiderLocation 写入骑手位置（同区域 GEO 集合，24h 过期）。
func (g *GeoStore) SetRiderLocation(ctx context.Context, regionID, riderID string, lat, lng float64) error {
	key := "geo:riders:" + regionID
	if err := g.rdb.GeoAdd(ctx, key, &redis.GeoLocation{
		Name: riderID, Longitude: lng, Latitude: lat,
	}).Err(); err != nil {
		return err
	}
	return g.rdb.Expire(ctx, key, 24*time.Hour).Err()
}
