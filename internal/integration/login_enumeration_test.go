package integration

import (
	"net/http"
	"testing"
	"time"
)

// R3-3 回归：登录路径不因「账号是否存在」暴露差异。
// 响应码/文案必须一致；账号不存在分支要执行占位 Argon2 校验，
// 否则探测者可用耗时差枚举有效账号（真实校验约需数十毫秒，不存在分支近乎瞬时）。
func TestLoginNoEnumerationSignal(t *testing.T) {
	const known = "client.test@dawnerrand.local"
	setup()
	clearRateLimitKeys() // 耗时断言不应被限流键残留干扰

	unknownStart := time.Now()
	unknown, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account": "no-such-user@shuguang.local", "password": "WrongPass123",
	})
	unknownCost := time.Since(unknownStart)

	wrongStart := time.Now()
	wrong, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account": known, "password": "WrongPass123",
	})
	wrongCost := time.Since(wrongStart)

	if unknown.Code != 1002 || wrong.Code != 1002 {
		t.Fatalf("两条路径应同为 1002, got unknown=%d wrong=%d", unknown.Code, wrong.Code)
	}
	if unknown.Message != wrong.Message {
		t.Fatalf("文案泄露账号存在性: %q vs %q", unknown.Message, wrong.Message)
	}
	if unknownCost < 10*time.Millisecond {
		t.Fatalf("账号不存在分支未执行占位密码校验, 耗时=%s", unknownCost)
	}
	// 占位校验与真实校验同量级：不存在分支不应比命中分支慢一个数量级以上。
	if unknownCost > wrongCost*10+time.Second {
		t.Fatalf("耗时差异过大 unknown=%s wrong=%s", unknownCost, wrongCost)
	}
}
