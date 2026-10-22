// Package requestid 提供请求 ID 的生成、透传与上下文存取。
package requestid

import (
	"context"
	"math/rand"
	"time"
)

type ctxKey struct{}

// 生成 ULID 风格 26 字符请求 ID（纯内存实现，与全局 ULID 生成器同字符集）。
const chars = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// New 生成请求 ID。
func New() string {
	b := make([]byte, 26)
	// 时间前缀（前 10 位，毫秒级）保证趋势可读
	ms := time.Now().UnixMilli()
	b[0] = chars[(ms>>35)&31]
	b[1] = chars[(ms>>30)&31]
	b[2] = chars[(ms>>25)&31]
	b[3] = chars[(ms>>20)&31]
	b[4] = chars[(ms>>15)&31]
	b[5] = chars[(ms>>10)&31]
	b[6] = chars[(ms>>5)&31]
	b[7] = chars[ms&31]
	for i := 8; i < 26; i++ {
		b[i] = chars[rng.Intn(32)]
	}
	return string(b)
}

// WithID 将请求 ID 注入上下文。
func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// From 从上下文取请求 ID；不存在返回空串。
func From(ctx context.Context) string {
	v, _ := ctx.Value(ctxKey{}).(string)
	return v
}
