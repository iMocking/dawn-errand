// Package token 访问令牌（JWT）与刷新令牌（Redis，轮换）实现。
package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"dawn-errand/internal/application"
)

var (
	ErrTokenInvalid = errors.New("token 无效")
	ErrTokenExpired = errors.New("token 已过期")
)

type claims struct {
	Role           string `json:"role"`
	SessionVersion int64  `json:"ver"`
	jwt.RegisteredClaims
}

// Manager 实现 application.TokenManager 端口。
type Manager struct {
	secret    []byte
	redis     *redis.Client
	keyPrefix string
}

// NewManager 构建令牌管理器。
func NewManager(secret string, rdb *redis.Client) *Manager {
	return &Manager{secret: []byte(secret), redis: rdb, keyPrefix: "dawn:refresh:"}
}

// 令牌签发方与接收方：校验时可拒掉同密钥签发的其他用途令牌（如邮件链接、第三方回调）。
const (
	tokenIssuer   = "dawn-errand"
	tokenAudience = "dawn-errand-api"
)

// IssueAccessToken 签发 JWT 访问令牌。
func (m *Manager) IssueAccessToken(c application.TokenClaims, ttl time.Duration) (string, error) {
	now := time.Now()
	cl := claims{
		Role:           c.Role,
		SessionVersion: c.SessionVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   c.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    tokenIssuer,
			Audience:  jwt.ClaimStrings{tokenAudience},
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, cl).SignedString(m.secret)
}

// ParseAccessToken 校验并解析访问令牌。
// 除签名与过期外，还校验 iss/aud：仅凭「密钥正确 + 未过期」放行会让同密钥签发的其他令牌被当作访问令牌。
func (m *Manager) ParseAccessToken(tokenStr string) (*application.TokenClaims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return m.secret, nil
	}, jwt.WithIssuer(tokenIssuer), jwt.WithAudience(tokenAudience))
	if err != nil {
		return nil, ErrTokenInvalid
	}
	cl, ok := tok.Claims.(*claims)
	if !ok || !tok.Valid {
		return nil, ErrTokenInvalid
	}
	if err := validateExp(cl.RegisteredClaims.ExpiresAt); err != nil {
		return nil, err
	}
	return &application.TokenClaims{
		UserID:         cl.Subject,
		Role:           cl.Role,
		SessionVersion: cl.SessionVersion,
	}, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func validateExp(exp *jwt.NumericDate) error {
	if exp == nil {
		return ErrTokenInvalid
	}
	if time.Now().After(exp.Time) {
		return ErrTokenExpired
	}
	return nil
}

func (m *Manager) key(userID, hash string) string { return m.keyPrefix + userID + ":" + hash }

// rtKey 刷新令牌哈希 → 用户 ID 的反向索引（供公开 refresh 接口解析用户）。
func (m *Manager) rtKey(hash string) string { return "dawn:rt:" + hash }

// IssueRefreshToken 签发刷新令牌，明文返回给客户端，Redis 仅存哈希。
func (m *Manager) IssueRefreshToken(ctx context.Context, userID string, ttl time.Duration) (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	plain := base64.RawURLEncoding.EncodeToString(raw)
	h := hashToken(plain)
	pipe := m.redis.TxPipeline()
	pipe.Set(ctx, m.key(userID, h), "1", ttl)
	pipe.Set(ctx, m.rtKey(h), userID, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", "", err
	}
	return plain, h, nil
}

// UserIDForRefreshToken 由明文刷新令牌解析用户 ID（public refresh 入口使用）。
func (m *Manager) UserIDForRefreshToken(ctx context.Context, plain string) (string, error) {
	h := hashToken(plain)
	uid, err := m.redis.Get(ctx, m.rtKey(h)).Result()
	if err == redis.Nil {
		return "", ErrTokenInvalid
	}
	if err != nil {
		return "", err
	}
	return uid, nil
}

// VerifyRefreshToken 校验刷新令牌。
func (m *Manager) VerifyRefreshToken(ctx context.Context, userID, plain string) error {
	h := hashToken(plain)
	n, err := m.redis.Exists(ctx, m.key(userID, h)).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrTokenInvalid
	}
	return nil
}

// RevokeRefreshTokens 撤销用户全部刷新令牌。
// 使用 SCAN 而非 KEYS：KEYS 是 O(N) 阻塞命令，会冻结整个 Redis 实例。
// 同时清理 dawn:rt:<hash> 反向索引，避免残留索引让 public refresh 入口误判令牌仍有效。
func (m *Manager) RevokeRefreshTokens(ctx context.Context, userID string) error {
	var (
		cursor  uint64
		batch   []string
		err     error
		matched []string
	)
	for {
		batch, cursor, err = m.redis.Scan(ctx, cursor, m.keyPrefix+userID+":*", 200).Result()
		if err != nil {
			return err
		}
		matched = append(matched, batch...)
		if cursor == 0 {
			break
		}
	}
	if len(matched) == 0 {
		return nil
	}
	prefix := m.keyPrefix + userID + ":"
	del := make([]string, 0, len(matched)*2)
	for _, k := range matched {
		del = append(del, k)
		if h := strings.TrimPrefix(k, prefix); h != "" && h != k {
			del = append(del, m.rtKey(h))
		}
	}
	// 分批删除，避免单命令携带过多 key 造成长时间阻塞。
	for i := 0; i < len(del); i += 200 {
		end := min(i+200, len(del))
		if err := m.redis.Del(ctx, del[i:end]...).Err(); err != nil {
			return err
		}
	}
	return nil
}

// RevokeRefreshToken 撤销单个刷新令牌。
func (m *Manager) RevokeRefreshToken(ctx context.Context, userID, plain string) error {
	h := hashToken(plain)
	pipe := m.redis.TxPipeline()
	pipe.Del(ctx, m.key(userID, h))
	pipe.Del(ctx, m.rtKey(h))
	_, err := pipe.Exec(ctx)
	return err
}
