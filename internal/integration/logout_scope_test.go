package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

type tokenPair struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// mustRegisterAndLogin 注册并返回一对令牌（邮箱需先验证）。
func mustRegisterAndLogin(t *testing.T, email, password string) tokenPair {
	t.Helper()
	setup()
	clearRateLimitKeys()
	cleanUser(t, email)
	if reg, _ := do(t, http.MethodPost, "/api/auth/register", "", map[string]string{
		"email": email, "password": password, "name": "会话测试",
	}); reg.Code != 0 {
		t.Fatalf("register code=%d msg=%s", reg.Code, reg.Message)
	}
	vt := mailToken(t, email, "验证您的邮箱")
	if vr, _ := do(t, http.MethodGet, "/api/auth/email/verify?token="+vt, "", nil); vr.Code != 0 {
		t.Fatalf("verify code=%d", vr.Code)
	}
	return mustLogin(t, email, password)
}

func mustLogin(t *testing.T, email, password string) tokenPair {
	t.Helper()
	setup()
	r, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account": email, "password": password,
	})
	if r.Code == 1007 {
		clearRateLimitKeys()
		r, _ = do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
			"account": email, "password": password,
		})
	}
	if r.Code != 0 {
		t.Fatalf("login code=%d msg=%s", r.Code, r.Message)
	}
	var out tokenPair
	if err := json.Unmarshal(r.Body, &out); err != nil {
		t.Fatal(err)
	}
	if out.Token == "" || out.RefreshToken == "" {
		t.Fatal("login 未返回令牌")
	}
	return out
}

// TestLogoutScopeSemantics 校验登出双语义：
// scope=current 只吊销当前设备刷新令牌，其他端会话保持有效；
// 缺省为全端登出，所有 access token 与 refresh 令牌（含 dawn:rt:* 反向索引）失效。
func TestLogoutScopeSemantics(t *testing.T) {
	email := "logout-scope@shuguang.local"
	password := "LogoutPass123"
	first := mustRegisterAndLogin(t, email, password)
	second := mustLogin(t, email, password)

	// 仅当前设备登出
	if lo, _ := do(t, http.MethodPost, "/api/auth/logout", first.Token, map[string]string{
		"refresh_token": first.RefreshToken, "scope": "current",
	}); lo.Code != 0 {
		t.Fatalf("scope=current logout code=%d msg=%s", lo.Code, lo.Message)
	}
	if rf, _ := do(t, http.MethodPost, "/api/auth/refresh", "", map[string]string{
		"refresh_token": first.RefreshToken,
	}); rf.Code != 1006 {
		t.Fatalf("当前设备 refresh 应失效 1006, got %d", rf.Code)
	}
	// 其他端会话不受影响
	if me, _ := do(t, http.MethodGet, "/api/auth/me", second.Token, nil); me.Code != 0 {
		t.Fatalf("另一设备 access token 应仍有效, got code=%d", me.Code)
	}
	if rf, _ := do(t, http.MethodPost, "/api/auth/refresh", "", map[string]string{
		"refresh_token": second.RefreshToken,
	}); rf.Code != 0 {
		t.Fatalf("另一设备 refresh 应仍有效, got code=%d msg=%s", rf.Code, rf.Message)
	}

	// 全端登出（缺省语义）
	if lo, _ := do(t, http.MethodPost, "/api/auth/logout", second.Token, nil); lo.Code != 0 {
		t.Fatalf("全端 logout code=%d msg=%s", lo.Code, lo.Message)
	}
	if me, _ := do(t, http.MethodGet, "/api/auth/me", second.Token, nil); me.Code != 1006 {
		t.Fatalf("全端登出后 access token 应失效 1006, got %d", me.Code)
	}
	// 反向索引随 SCAN 一并清理，否则 public refresh 入口仍能解析出用户
	if rf, _ := do(t, http.MethodPost, "/api/auth/refresh", "", map[string]string{
		"refresh_token": second.RefreshToken,
	}); rf.Code != 1006 {
		t.Fatalf("全端登出后 refresh 应失效 1006, got %d", rf.Code)
	}
}

// TestRefreshAllDevices 校验 refresh 的 all_devices 语义：
// 刷新同时踢出其他设备，当前设备拿到新的一对令牌并继续可用。
func TestRefreshAllDevices(t *testing.T) {
	email := "refresh-all@shuguang.local"
	password := "RefreshAll123"
	first := mustRegisterAndLogin(t, email, password)
	second := mustLogin(t, email, password)

	rf, _ := do(t, http.MethodPost, "/api/auth/refresh", "", map[string]any{
		"refresh_token": first.RefreshToken, "all_devices": true,
	})
	if rf.Code != 0 {
		t.Fatalf("all_devices refresh code=%d msg=%s", rf.Code, rf.Message)
	}
	var fresh tokenPair
	if err := json.Unmarshal(rf.Body, &fresh); err != nil {
		t.Fatal(err)
	}
	if fresh.Token == "" || fresh.RefreshToken == "" || fresh.RefreshToken == first.RefreshToken {
		t.Fatal("all_devices refresh 应签发新令牌")
	}
	// 旧刷新令牌与其他设备令牌全部失效
	for _, old := range []string{first.RefreshToken, second.RefreshToken} {
		if again, _ := do(t, http.MethodPost, "/api/auth/refresh", "", map[string]string{
			"refresh_token": old,
		}); again.Code != 1006 {
			t.Fatalf("全端刷新后旧 refresh 应失效 1006, got %d", again.Code)
		}
	}
	if me, _ := do(t, http.MethodGet, "/api/auth/me", second.Token, nil); me.Code != 1006 {
		t.Fatalf("其他设备 access token 应随会话版本提升失效, got %d", me.Code)
	}
	// 当前设备新令牌可用
	if me, _ := do(t, http.MethodGet, "/api/auth/me", fresh.Token, nil); me.Code != 0 {
		t.Fatalf("当前设备新 access token 应有效, got %d", me.Code)
	}
}
