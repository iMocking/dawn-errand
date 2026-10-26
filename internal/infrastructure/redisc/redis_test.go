package redisc

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"dawn-errand/internal/config"
)

// newTestLocker 连接本地 Redis；不可用则跳过（分布式锁语义必须有真实 Redis 才能验证）。
func newTestLocker(t *testing.T) (*Locker, func()) {
	t.Helper()
	cfg := config.Load()
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		t.Skipf("Redis 不可用，跳过锁用例: %v", err)
	}
	return NewLocker(rdb), func() { _ = rdb.Close() }
}

// TestLockerOwnership 校验锁归属校验：token 不匹配不得删除他人持有的锁。
func TestLockerOwnership(t *testing.T) {
	l, closeFn := newTestLocker(t)
	defer closeFn()
	ctx := context.Background()
	key := "locktest:" + t.Name()
	defer l.rdb.Del(ctx, key)

	tokenA, ok, err := l.TryLock(ctx, key, time.Minute)
	if err != nil || !ok {
		t.Fatalf("A 应抢到锁 ok=%v err=%v", ok, err)
	}
	if tokenA == "" {
		t.Fatal("加锁成功必须返回持有者 token")
	}
	if _, ok, _ := l.TryLock(ctx, key, time.Minute); ok {
		t.Fatal("锁被持有时第二次 TryLock 应失败")
	}

	// 他人 token 释放无效（模拟本实例任务超时、锁已被接管后误删）
	if err := l.Unlock(ctx, key, "some-other-holder"); err != nil {
		t.Fatalf("Unlock(错误 token): %v", err)
	}
	if v, err := l.rdb.Get(ctx, key).Result(); err != nil || v != tokenA {
		t.Fatalf("非持有者不得删锁, got %q err=%v", v, err)
	}

	if err := l.Unlock(ctx, key, tokenA); err != nil {
		t.Fatalf("Unlock(持有者): %v", err)
	}
	if n, err := l.rdb.Exists(ctx, key).Result(); err != nil || n != 0 {
		t.Fatalf("持有者释放后锁应消失, exists=%d err=%v", n, err)
	}
	// 空 token 视为未持锁，直接无操作
	if err := l.Unlock(ctx, key, ""); err != nil {
		t.Fatalf("Unlock(空 token) 应为无操作: %v", err)
	}
}
