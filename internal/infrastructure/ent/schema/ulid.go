package schema

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"
)

const ulidChars = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

var (
	ulidMu   sync.Mutex
	ulidLast time.Time
)

// newULID 生成 26 字符 ULID（与 pkg/ulidgen 同算法，避免 schema 包循环依赖）。
func newULID() string {
	now := time.Now()
	ulidMu.Lock()
	defer ulidMu.Unlock()
	if !now.After(ulidLast) {
		now = ulidLast.Add(time.Millisecond)
	}
	ulidLast = now

	ts := now.UnixMilli()
	out := make([]byte, 26)
	v := uint64(ts)
	for i := 9; i >= 0; i-- {
		out[i] = ulidChars[v&31]
		v >>= 5
	}
	for i := 10; i < 26; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(32))
		if err != nil {
			out[i] = ulidChars[time.Now().UnixNano()&31]
			continue
		}
		out[i] = ulidChars[n.Int64()]
	}
	return string(out)
}
