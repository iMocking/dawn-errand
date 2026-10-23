// Package config 负责从环境变量加载启动配置。
// 业务可管理配置（SMTP/OSS/Casbin 等）通过 SystemSetting 领域存储，不在此处硬编码。
package config

import (
	"os"
	"strconv"
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
	}
}

// IsDev 是否开发环境（启用 dev 专用能力，如邮件 outbox 查询）。
func (c *Config) IsDev() bool { return c.AppEnv == "dev" }

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
