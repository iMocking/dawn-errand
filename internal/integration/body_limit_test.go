package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dawn-errand/internal/pkg/errcode"
)

// doRaw 发送原始请求体（用于测试非法/超大 body）。
func doRaw(t *testing.T, method, path, token string, raw string) *resp {
	t.Helper()
	setup()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	var r resp
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		t.Fatalf("decode raw body %s %s: %v", method, path, err)
	}
	return &r
}

// TestRequestBodyLimit 全局请求体上限（R4-1）：公开 JSON 端点超过 1MB 应在读取阶段
// 被截断并返回 413/1303，而不是进入绑定与业务层。
func TestRequestBodyLimit(t *testing.T) {
	setup()

	big := `{"account":"` + strings.Repeat("a", 2<<20) + `","password":"x"}`
	r := doRaw(t, http.MethodPost, "/api/auth/login", "", big)
	if r.Code != errcode.CodeFileTooLarge || !strings.Contains(r.Message, "大小") {
		t.Fatalf("超大请求体应 1303 体积超限, got %d msg=%s", r.Code, r.Message)
	}

	// 上限内的正常请求不受影响（此处应是凭证错误而非体积错误）。
	r2, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account": "no-such@local", "password": "WrongPass1",
	})
	if r2.Code == errcode.CodeFileTooLarge {
		t.Fatal("正常体积被误判为超限")
	}
}

// TestBindErrorNoInternalLeak 绑定失败细节不回显（R4-3）：
// 类型不匹配等解码错误只返回泛化文案，不得暴露字段名/解析器错误。
func TestBindErrorNoInternalLeak(t *testing.T) {
	setup()

	r := doRaw(t, http.MethodPost, "/api/auth/login", "", `{"account":{"x":1},"password":"y"}`)
	if r.Code != errcode.CodeBadRequest {
		t.Fatalf("绑定失败应 1301, got %d msg=%s", r.Code, r.Message)
	}
	for _, leak := range []string{"json", "JSON", "struct", "Account", "account", "field", "EOF", "line"} {
		if strings.Contains(r.Message, leak) {
			t.Fatalf("绑定错误文案泄露内部细节 %q: %s", leak, r.Message)
		}
	}
}
