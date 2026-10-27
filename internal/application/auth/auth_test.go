package auth

import (
	"encoding/hex"
	"regexp"
	"strings"
	"testing"

	"dawn-errand/internal/pkg/errcode"
)

// R5-5：认证入口的格式校验与令牌生成是纯函数，先在这里锁死，
// 登录/注册/重置的端到端行为由 integration 用例覆盖。

func TestValidateEmail(t *testing.T) {
	ok := []string{"a@b.co", "user.name+tag@sub.example.com", "jsong@dawnerrand.local"}
	bad := []string{"", "a@b", "a b@c.com", "@b.com", "a@.com", "a@b.", "a@@b.com", "a@b@c.com", " a@b.com", "a@b.com "}
	for _, e := range ok {
		if !ValidateEmail(e) {
			t.Errorf("合法邮箱被拒: %q", e)
		}
	}
	for _, e := range bad {
		if ValidateEmail(e) {
			t.Errorf("非法邮箱被放行: %q", e)
		}
	}
}

// R5-5：密码策略是爆破成本的第一道闸，长度上界同时防止 bcrypt 72 字节截断。
func TestValidatePassword(t *testing.T) {
	ok := []string{"Client123456", "abcd1234", "Passw0rd!", strings.Repeat("a", 4) + strings.Repeat("1", 4), strings.Repeat("a", 64) + "1"}
	bad := []string{"", "abc1234", "abcdefgh", "12345678", strings.Repeat("a", 73) + "1"}
	for _, p := range ok {
		if err := ValidatePassword(p); err != nil {
			t.Errorf("合法密码被拒: %q -> %v", p, err)
		}
	}
	for _, p := range bad {
		err := ValidatePassword(p)
		if err == nil {
			t.Errorf("弱密码被放行: %q", p)
			continue
		}
		if e, as := err.(*errcode.Error); !as || e.Code != errcode.CodePasswordPolicy {
			t.Errorf("密码策略错误码错误: %v", err)
		}
	}
	// 边界：恰好 8 位与 72 位
	if err := ValidatePassword("abcd1234"); err != nil {
		t.Errorf("8 位边界应通过: %v", err)
	}
	if err := ValidatePassword(strings.Repeat("a", 71) + "1"); err != nil {
		t.Errorf("72 位边界应通过: %v", err)
	}
}

// R5-5：证件号掩码是对外展示的唯一形态，短输入必须整段遮蔽而不是回显原文。
func TestMaskIDCard(t *testing.T) {
	cases := map[string]string{
		"110101199003077219": "110101********7219",
		"11010119900307721X": "110101********721X",
		"110101900307721":    "110101********7721",
		"12345678":           "******",
		"1234567":            "******",
		"":                   "******",
		"张三":                 "******",
	}
	for in, want := range cases {
		if got := maskIDCard(in); got != want {
			t.Errorf("maskIDCard(%q)=%q want=%q", in, got, want)
		}
	}
	// 掩码结果里不得残留中间 8 位（出生日期段）
	for _, got := range []string{maskIDCard("110101199003077219"), maskIDCard("11010119900307721X")} {
		if strings.Contains(got, "19900307") {
			t.Errorf("掩码泄漏出生日期段: %s", got)
		}
	}
}

// R5-5：刷新令牌/短信验证码必须是加密随机且长度稳定，可预测即等于会话劫持。
func TestTokenAndCodeGenerators(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		tk := newToken()
		if len(tk) < 32 {
			t.Fatalf("刷新令牌过短，疑似随机源降级: %q", tk)
		}
		if seen[tk] {
			t.Fatalf("令牌重复: %s", tk)
		}
		seen[tk] = true
	}
	if strings.ContainsAny(newToken(), "+/=") {
		t.Error("令牌不应使用需要再编码的 base64 字符集")
	}

	seenCode := map[string]bool{}
	for i := 0; i < 300; i++ {
		c := smsCode()
		if len(c) != 6 || strings.Trim(c, "0123456789") != "" {
			t.Fatalf("短信验证码需为 6 位数字, got=%q", c)
		}
		seenCode[c] = true
	}
	if len(seenCode) < 200 {
		t.Errorf("验证码重复率过高，疑似非加密随机: %d/300", len(seenCode))
	}

	h := hashToken("secret-token")
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(h) {
		t.Errorf("令牌哈希应为 sha256 十六进制, got=%q", h)
	}
	if _, err := hex.DecodeString(h); err != nil {
		t.Errorf("哈希非合法十六进制: %v", err)
	}
	if hashToken("secret-token") != h || hashToken("other") == h {
		t.Error("哈希不稳定或碰撞")
	}
	if strings.Contains(h, "secret") {
		t.Error("哈希结果包含原文")
	}
}
