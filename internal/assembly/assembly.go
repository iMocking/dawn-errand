// Package assembly 依赖装配：将 infrastructure 实现注入 application/interfaces，
// 供 cmd/server 与 cmd/seed 复用。
package assembly

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"dawn-errand/internal/application"
	"dawn-errand/internal/application/admin"
	"dawn-errand/internal/application/auth"
	"dawn-errand/internal/application/system"
	"dawn-errand/internal/config"
	sysdomain "dawn-errand/internal/domain/system"
	"dawn-errand/internal/infrastructure/casbin"
	"dawn-errand/internal/infrastructure/crypto"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/mailer"
	"dawn-errand/internal/infrastructure/passwd"
	"dawn-errand/internal/infrastructure/redisc"
	"dawn-errand/internal/infrastructure/repo"
	"dawn-errand/internal/infrastructure/storage"
	"dawn-errand/internal/infrastructure/token"
	"dawn-errand/internal/interfaces/http"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/log"
)

// init 注册 pgx 驱动为 "postgres"，供 ent.Open("postgres", dsn) 使用。
func init() {
	sql.Register("postgres", stdlib.GetDefaultDriver())
}

// App 全部装配结果。
type App struct {
	Cfg      *config.Config
	Logger   *zap.Logger
	Client   *ent.Client
	Rdb      *redis.Client
	Store    *repo.Store
	Enforcer *casbin.Enforcer
	Sender   *mailer.Sender
	Auth     *auth.Service
	System   *system.Service
	Handler  *http.Handler
	MW       *middleware.Middleware
}

// smtpConfigProvider 读取已保存的 SMTP 配置（解密密码；未配置返回空配置）。
func smtpConfigProvider(cipher *crypto.Cipher, store *repo.Store) func() (*sysdomain.SMTPConfig, error) {
	return func() (*sysdomain.SMTPConfig, error) {
		rec, err := store.Repos().Settings.Get(sysdomain.KeySMTP)
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				return &sysdomain.SMTPConfig{}, nil
			}
			return nil, err
		}
		var cfg sysdomain.SMTPConfig
		if err := json.Unmarshal([]byte(rec.ValueJSON), &cfg); err != nil {
			return nil, err
		}
		if cfg.Password != "" && cfg.Password != "******" {
			if plain, err := cipher.Decrypt(cfg.Password); err == nil {
				cfg.Password = plain
			}
		}
		return &cfg, nil
	}
}

// localStorageProvider 读取已保存的本地存储配置。
func localStorageProvider(store *repo.Store) func() *sysdomain.LocalStorageConfig {
	return func() *sysdomain.LocalStorageConfig {
		rec, err := store.Repos().Settings.Get(sysdomain.KeyLocal)
		if err != nil {
			return &sysdomain.LocalStorageConfig{RootDir: ".data/uploads"}
		}
		var cfg sysdomain.LocalStorageConfig
		if err := json.Unmarshal([]byte(rec.ValueJSON), &cfg); err != nil {
			return &sysdomain.LocalStorageConfig{RootDir: ".data/uploads"}
		}
		return &cfg
	}
}

// outboxProvider dev 邮件发件箱查询。
func outboxProvider(client *ent.Client) func(ctx context.Context, limit int) ([]application.MailEntry, error) {
	return func(ctx context.Context, limit int) ([]application.MailEntry, error) {
		if limit < 1 || limit > 200 {
			limit = 50
		}
		rows, err := client.MailOutbox.Query().Order(ent.Desc("created_at")).Limit(limit).All(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]application.MailEntry, 0, len(rows))
		for _, r := range rows {
			out = append(out, application.MailEntry{
				ID: r.ID, ToEmail: r.ToEmail, Subject: r.Subject, Body: r.Body, CreatedAt: r.CreatedAt,
			})
		}
		return out, nil
	}
}

// Build 装配全部依赖。
func Build(cfg *config.Config) (*App, error) {
	logger, err := log.New(cfg.AppEnv)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	client, err := ent.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	rdb, err := redisc.NewClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		return nil, fmt.Errorf("open redis: %w", err)
	}

	store := repo.NewStore(client)
	repos := store.Repos()

	enforcer, err := casbin.NewEnforcer(client)
	if err != nil {
		return nil, fmt.Errorf("init casbin: %w", err)
	}

	cipher := crypto.New(cfg.ConfigCipherKey)
	hasher := passwd.Hasher{}
	tokens := token.NewManager(cfg.JWTSecret, rdb)
	limiter := redisc.NewLimiter(rdb)
	sender := mailer.NewSender(client, smtpConfigProvider(cipher, store))
	_ = storage.NewLocal(localStorageProvider(store))

	authSvc := &auth.Service{
		Repos: repos, Tx: store, Hasher: hasher, Tokens: tokens,
		Mailer: sender, Cfg: cfg, Enforcer: enforcer, Limiter: limiter,
	}
	systemSvc := &system.Service{
		Repos: repos, Tx: store, Hasher: hasher, Cipher: cipher,
		Policies: enforcer, Cfg: cfg, Outbox: outboxProvider(client),
	}

	mw := &middleware.Middleware{
		Logger: logger, Tokens: tokens, Users: repos.Users,
		Creds: repos.Credentials, Limiter: limiter, Enforcer: enforcer,
	}

	h := &http.Handler{
		Cfg: cfg, MW: mw, Auth: authSvc, System: systemSvc,
		Dashboard:  &admin.DashboardService{Repos: repos},
		Orders:     &admin.OrdersService{Repos: repos},
		Merchants:  &admin.MerchantsService{Repos: repos},
		Riders:     &admin.RidersService{Repos: repos},
		Users:      &admin.UsersService{Repos: repos},
		Finance:    &admin.FinanceService{Repos: repos},
		Complaints: &admin.ComplaintsService{Repos: repos},
		Settings:   &admin.SettingsService{Repos: repos},
	}

	return &App{
		Cfg: cfg, Logger: logger, Client: client, Rdb: rdb,
		Store: store, Enforcer: enforcer, Sender: sender,
		Auth: authSvc, System: systemSvc, Handler: h, MW: mw,
	}, nil
}
