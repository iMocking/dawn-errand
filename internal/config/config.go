// Package config 负责从环境变量加载启动配置。
// 业务可管理配置（SMTP/OSS/Casbin 等）通过 SystemSetting 领域存储，不在此处硬编码。
package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 启动配置。仅承载引导与密钥类配置，业务配置走 SystemSetting。
type Config struct {
	AppEnv string // dev / test / prod
	Port   int

	PostgresDSN string

	RedisAddr     string
	RedisPassword string

	JWTSecret        string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	ConfigCipherKey  string // AES-256-GCM 密钥，用于加密 SMTP 密码等敏感配置
	BootstrapToken   string // 初始化令牌
	StorageRootDir   string
	StoragePublicURL string
	StorageMaxSizeMB int64

	// TrustedProxyCIDRS 可信反向代理网段（逗号分隔）。Gin 默认信任所有代理头，
	// 不配置则 X-Forwarded-For 可伪造，IP 维度限流全部失效。"*" 表示信任全部（仅调试用）。
	TrustedProxyCIDRs []string

	// 订单超时任务
	OrderPayTimeoutSec  int64 // 待支付超时（秒）
	OrderAutoConfirmSec int64 // 送达后自动确认（秒）

	// 支付回调签名密钥（真实渠道适配层验签）。
	PaymentCallbackSecret string

	// MetricsToken /metrics 访问令牌；为空时生产环境关闭该端点，非生产环境放行（本机调试）。
	MetricsToken string
}

// Load 从环境变量加载配置；缺省值对应 .env.example。
func Load() *Config {
	return &Config{
		AppEnv: getStr("APP_ENV", "dev"),
		Port:   getInt("HTTP_PORT", 8080),

		PostgresDSN: getStr("POSTGRES_DSN", "host=localhost port=5432 user=dawn password=dawn_dev_only dbname=dawn_errand sslmode=disable TimeZone=UTC"),

		RedisAddr:     getStr("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getStr("REDIS_PASSWORD", "dawn_dev_only"),

		JWTSecret:       getStr("JWT_SECRET", "dev_only_jwt_secret_change_me_0123456789abcdef"),
		AccessTokenTTL:  time.Duration(getInt("ACCESS_TOKEN_TTL", 900)) * time.Second, // 默认 15min；生产建议用 ACCESS_TOKEN_TTL 覆盖（前端已接入 refresh 自动续期）
		RefreshTokenTTL: time.Duration(getInt("REFRESH_TOKEN_TTL", 1209600)) * time.Second,
		ConfigCipherKey: getStr("CONFIG_ENCRYPTION_KEY", "dev_only_config_key_0123456789abcdef0123456789"),
		BootstrapToken:  getStr("BOOTSTRAP_TOKEN", "dev_bootstrap_token_2026"),

		StorageRootDir:   getStr("STORAGE_ROOT_DIR", ".data/uploads"),
		StoragePublicURL: getStr("STORAGE_PUBLIC_BASE_URL", "http://localhost:8080"),
		StorageMaxSizeMB: int64(getInt("STORAGE_MAX_SIZE_MB", 10)),

		TrustedProxyCIDRs: getStrList("TRUSTED_PROXY_CIDRS", []string{"127.0.0.1/32", "::1/128"}),

		OrderPayTimeoutSec:  int64(getInt("ORDER_PAY_TIMEOUT_SEC", 900)),
		OrderAutoConfirmSec: int64(getInt("ORDER_AUTO_CONFIRM_SEC", 86400)),

		PaymentCallbackSecret: getStr("PAYMENT_CALLBACK_SECRET", "dev_payment_callback_secret"),

		MetricsToken: getStr("METRICS_TOKEN", ""),
	}
}

// IsDev 是否开发环境（启用 dev 专用能力，如邮件 outbox 查询）。
func (c *Config) IsDev() bool { return c.AppEnv == "dev" }

// Validate 生产环境启动校验：禁止默认/弱密钥与开发口令。
func (c *Config) Validate() error {
	if c.AppEnv != "prod" {
		return nil
	}
	switch {
	case len(c.JWTSecret) < 32 || containsDev(c.JWTSecret):
		return errors.New("生产环境 JWT_SECRET 必须为 ≥32 字节随机值，且不得使用 dev 默认值")
	case containsDev(c.ConfigCipherKey):
		return errors.New("生产环境 CONFIG_ENCRYPTION_KEY 必须更换")
	case len(c.ConfigCipherKey) < 32:
		// 密钥经 SHA-256 派生，长度本身不报错，但过短意味着熵不足、可被穷举。
		return errors.New("生产环境 CONFIG_ENCRYPTION_KEY 至少 32 字节随机值")
	case c.BootstrapToken == "" || containsDev(c.BootstrapToken):
		return errors.New("生产环境 BOOTSTRAP_TOKEN 必须更换")
	case c.PaymentCallbackSecret == "" || containsDev(c.PaymentCallbackSecret):
		return errors.New("生产环境 PAYMENT_CALLBACK_SECRET 必须更换")
	case containsDev(c.RedisPassword):
		return errors.New("生产环境 Redis 密码不得使用 dev 默认值")
	case containsDev(c.PostgresDSN):
		return errors.New("生产环境数据库连接不得使用 dev 默认密码")
	}
	return nil
}

func containsDev(v string) bool {
	for _, marker := range []string{"dev_only", "dev_bootstrap", "dev_payment", "change_me"} {
		if strings.Contains(v, marker) {
			return true
		}
	}
	return false
}

func getStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getStrList(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}
