package repo

import (
	"errors"
	"strings"
	"testing"
)

type fakeCipher struct{}

func (fakeCipher) Encrypt(plain string) (string, error) { return "v1:" + plain, nil }
func (fakeCipher) Decrypt(ct string) (string, error) {
	if !strings.HasPrefix(ct, "v1:") {
		return "", errors.New("密文校验失败")
	}
	return strings.TrimPrefix(ct, "v1:"), nil
}

func TestRiderSealUnseal(t *testing.T) {
	r := RiderRepo{Cipher: fakeCipher{}}
	sealed := r.sealSecret("110101199001011234")
	if sealed != "v1:110101199001011234" {
		t.Fatalf("应加密: %s", sealed)
	}
	if got := r.unsealSecret(sealed); got != "110101199001011234" {
		t.Fatalf("应解密回原值: %s", got)
	}
	// 存量明文（无 v1: 前缀）原样透传，由上层脱敏兜底。
	if got := r.unsealSecret("110101199001011234"); got != "110101199001011234" {
		t.Fatalf("明文透传异常: %s", got)
	}
	// 无 Cipher 时不处理。
	plain := RiderRepo{}
	if plain.sealSecret("abc") != "abc" || plain.unsealSecret("v1:abc") != "v1:abc" {
		t.Fatal("Cipher 为 nil 时应原样返回")
	}
	// 解密失败（密钥不符/篡改）返回空串而非透出密文。
	rBad := RiderRepo{Cipher: badCipher{}}
	if got := rBad.unsealSecret("v1:anything"); got != "" {
		t.Fatalf("解密失败应返回空串: %s", got)
	}
}

type badCipher struct{}

func (badCipher) Encrypt(s string) (string, error) { return s, nil }
func (badCipher) Decrypt(string) (string, error)   { return "", errors.New("密文校验失败") }
