package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"dawn-errand/internal/application"
)

// R3-9 回归：仅凭「密钥正确 + 未过期」不足以放行，iss/aud 必须匹配，
// 否则同密钥签发的其他用途令牌（邮件链接、内部任务票据）会被当成访问令牌。
func TestParseAccessTokenValidatesIssuerAndAudience(t *testing.T) {
	m := NewManager("test_secret_value_for_unit_test_0123456789", nil)
	tc := application.TokenClaims{UserID: "U1", Role: "user", SessionVersion: 1}

	good, err := m.IssueAccessToken(tc, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := m.ParseAccessToken(good)
	if err != nil || parsed.UserID != "U1" {
		t.Fatalf("正常令牌应解析成功: %+v err=%v", parsed, err)
	}

	// 同密钥、同有效期，但 iss/aud 不同
	bad := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role: "user", SessionVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "U1",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			Issuer:    "someone-else",
			Audience:  jwt.ClaimStrings{"other-app"},
		},
	})
	tokStr, err := bad.SignedString(m.secret)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.ParseAccessToken(tokStr); err != ErrTokenInvalid {
		t.Fatalf("iss/aud 不匹配应 ErrTokenInvalid, got %v", err)
	}
}
