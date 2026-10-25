// Package assembly 依赖装配：将 infrastructure 实现注入 application/interfaces，
// 供 cmd/server 与 cmd/seed 复用。
package assembly

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"dawn-errand/internal/application"
	"dawn-errand/internal/application/admin"
	appsvc "dawn-errand/internal/application/app"
	"dawn-errand/internal/application/auth"
	"dawn-errand/internal/application/jobs"
	"dawn-errand/internal/application/system"
	"dawn-errand/internal/config"
	sysdomain "dawn-errand/internal/domain/system"
	"dawn-errand/internal/infrastructure/casbin"
	"dawn-errand/internal/infrastructure/crypto"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/mailer"
	"dawn-errand/internal/infrastructure/passwd"
	"dawn-errand/internal/infrastructure/payout"
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
	Jobs     *jobs.Runner
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

// localStorageProvider 读取已保存的本地存储配置；未配置时回退到启动配置。
func localStorageProvider(store *repo.Store, cfg *config.Config) func() *sysdomain.LocalStorageConfig {
	return func() *sysdomain.LocalStorageConfig {
		def := &sysdomain.LocalStorageConfig{
			RootDir:       cfg.StorageRootDir,
			PublicBaseURL: cfg.StoragePublicURL,
			MaxSizeMB:     cfg.StorageMaxSizeMB,
		}
		rec, err := store.Repos().Settings.Get(sysdomain.KeyLocal)
		if err != nil {
			return def
		}
		var c sysdomain.LocalStorageConfig
		if err := json.Unmarshal([]byte(rec.ValueJSON), &c); err != nil {
			return def
		}
		if c.RootDir == "" {
			c.RootDir = def.RootDir
		}
		if c.PublicBaseURL == "" {
			c.PublicBaseURL = def.PublicBaseURL
		}
		if c.MaxSizeMB <= 0 {
			c.MaxSizeMB = def.MaxSizeMB
		}
		return &c
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
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置校验失败: %w", err)
	}
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
	fileCfg := localStorageProvider(store, cfg)()
	fileStore := storage.NewLocal(localStorageProvider(store, cfg))
	storageMount := strings.Trim(fileCfg.PathPrefix, "/")
	if storageMount == "" {
		storageMount = "uploads"
	}

	authSvc := &auth.Service{
		Repos: repos, Tx: store, Hasher: hasher, Tokens: tokens,
		Mailer: sender, Cipher: cipher, Cfg: cfg, Enforcer: enforcer, Limiter: limiter,
		Codes: redisc.NewCodeStore(rdb),
	}
	systemSvc := &system.Service{
		Repos: repos, Tx: store, Hasher: hasher, Cipher: cipher,
		Policies: enforcer, Cfg: cfg, Outbox: outboxProvider(client),
	}

	mw := &middleware.Middleware{
		Logger: logger, Tokens: tokens, Users: repos.Users,
		Creds: repos.Credentials, Limiter: limiter, Enforcer: enforcer,
	}

	orderSvc := &appsvc.OrderService{Repos: repos, Tx: store, MockPayment: cfg.IsDev()}
	trackingSvc := &appsvc.TrackingService{Repos: repos, Geo: redisc.NewGeoStore(rdb)}
	notifySvc := &appsvc.NotificationService{Repos: repos}
	walletSvc := &appsvc.WalletService{Repos: repos}
	var payoutGw application.PayoutGateway
	if cfg.IsDev() {
		payoutGw = payout.NewMock()
	}
	financeAppSvc := &appsvc.FinanceAppService{Repos: repos, Tx: store, Wallet: walletSvc, Payout: payoutGw}
	reviewSvc := &appsvc.ReviewService{Repos: repos, Tx: store}
	refundSvc := &appsvc.RefundService{Repos: repos, Tx: store}
	couponSvc := &appsvc.CouponService{Repos: repos}
	jobsRunner := &jobs.Runner{
		Repos: repos, Orders: orderSvc, Locker: redisc.NewLocker(rdb),
		PayTimeout:  time.Duration(cfg.OrderPayTimeoutSec) * time.Second,
		AutoConfirm: time.Duration(cfg.OrderAutoConfirmSec) * time.Second,
		BatchLimit:  100,
	}

	h := &http.Handler{
		Cfg: cfg, MW: mw, Auth: authSvc, System: systemSvc,
		Portal:        &appsvc.Service{Repos: repos, Tx: store},
		ClientApp:     &appsvc.ClientService{Repos: repos, Tx: store},
		StoreApp:      &appsvc.StoreService{Repos: repos, Tx: store},
		OrderApp:      orderSvc,
		Tracking:      trackingSvc,
		Notifications: notifySvc,
		Wallet:        walletSvc,
		FinanceApp:    financeAppSvc,
		Reviews:       reviewSvc,
		Refunds:       refundSvc,
		Coupons:       couponSvc,
		Storage:       fileStore,
		StorageMount:  storageMount,
		Accounts:      &admin.AccountsService{Repos: repos, Tx: store, Hasher: hasher},
		Dashboard:     &admin.DashboardService{Repos: repos},
		Orders:        &admin.OrdersService{Repos: repos},
		Merchants:     &admin.MerchantsService{Repos: repos},
		Riders:        &admin.RidersService{Repos: repos},
		Users:         &admin.UsersService{Repos: repos, Tx: store, Cipher: cipher},
		Finance:       &admin.FinanceService{Repos: repos, Tx: store, Payout: payoutGw},
		Complaints:    &admin.ComplaintsService{Repos: repos},
		Settings:      &admin.SettingsService{Repos: repos},
		CouponsAdmin:  &admin.CouponService{Repos: repos},
	}
	h.Ready = func(ctx context.Context) error {
		if _, err := client.User.Query().Limit(1).Exist(ctx); err != nil {
			return err
		}
		return rdb.Ping(ctx).Err()
	}

	return &App{
		Cfg: cfg, Logger: logger, Client: client, Rdb: rdb,
		Store: store, Enforcer: enforcer, Sender: sender,
		Auth: authSvc, System: systemSvc, Jobs: jobsRunner, Handler: h, MW: mw,
	}, nil
}
