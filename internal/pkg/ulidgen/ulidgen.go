// Package ulidgen 生成 ULID 主键。
package ulidgen

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"
)

const chars = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

var (
	mu   sync.Mutex
	last time.Time
)

// New 生成 26 字符 ULID。
func New() string {
	now := time.Now()
	mu.Lock()
	defer mu.Unlock()
	if !now.After(last) {
		now = last.Add(time.Millisecond)
	}
	last = now

	ts := now.UnixMilli()
	out := make([]byte, 26)
	// Crockford base32 编码时间戳（48 bit → 10 字符）
	v := uint64(ts)
	for i := 9; i >= 0; i-- {
		out[i] = chars[v&31]
		v >>= 5
	}
	// 随机部分 16 字符（80 bit 熵）
	for i := 10; i < 26; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(32))
		if err != nil {
			out[i] = chars[time.Now().UnixNano()&31]
			continue
		}
		out[i] = chars[n.Int64()]
	}
	return string(out)
}
