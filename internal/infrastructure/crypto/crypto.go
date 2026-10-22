// Package crypto 配置敏感值 AES-256-GCM 加解密。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

// Cipher 实现 application.ConfigCipher 端口。
type Cipher struct {
	key [32]byte
}

// New 构建加密器；密钥不足 32 字节时用 SHA-256 派生，避免运行时错误。
func New(key string) *Cipher {
	var k [32]byte
	sum := sha256.Sum256([]byte(key))
	copy(k[:], sum[:])
	return &Cipher{key: k}
}

// Encrypt 加密：v1:base64(nonce || ciphertext)。
func (c *Cipher) Encrypt(plain string) (string, error) {
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return "v1:" + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt 解密 v1 格式密文。
func (c *Cipher) Decrypt(ciphertext string) (string, error) {
	if !strings.HasPrefix(ciphertext, "v1:") {
		return "", errors.New("不支持的密文格式")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, "v1:"))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("密文长度非法")
	}
	nonce, sealed := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", errors.New("密文校验失败")
	}
	return string(plain), nil
}
