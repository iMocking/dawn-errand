// Package assembly 依赖装配：将 infrastructure 实现注入 application/interfaces，
// 供 cmd/server 与 cmd/seed 复用。
package assembly

import (
	"context"
	"database/sql"
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
	"dawn-errand/internal/infrastructure/casbin"
	"dawn-errand/internal/infrastructure/crypto"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/mailer"
	"dawn-errand/internal/infrastructure/passwd"
	"dawn-errand/internal/infrastructure/payout"
	"dawn-errand/internal/infrastructure/redisc"
	"dawn-errand/internal/infrastructure/repo"
	"dawn-errand/internal/infrastructure/storage"
	"dawn-errand/internal/infrastructure/syscfg"
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

// Build 装配全部依赖。ctx 为进程根上下文：装配期的配置读取（如本地存储挂载点）用它，
// 请求期的 ctx 由 gin 逐请求注入，不在这里传递。
func Build(ctx context.Context, cfg *config.Config) (*App, error) {
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
	store.Cipher = cipher // 敏感字段（骑手身份证）透明加密
	hasher := passwd.Hasher{}
	tokens := token.NewManager(cfg.JWTSecret, rdb)
	limiter := redisc.NewLimiter(rdb)
	sender := mailer.NewSender(client, syscfg.SMTP(cipher, store))
	localStorage := syscfg.LocalStorage(store, cfg)
	fileCfg := localStorage(ctx)
	fileStore := storage.NewLocal(localStorage)
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
		Policies: enforcer, Cfg: cfg, Outbox: syscfg.Outbox(client),
	}

	mw := &middleware.Middleware{
		Logger: logger, Tokens: tokens, Users: repos.Users,
		Creds: repos.Credentials, Limiter: limiter, Enforcer: enforcer,
		Merchants: repos.Merchants, Riders: repos.Riders,
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
	accountsSvc := &admin.AccountsService{Repos: repos, Tx: store, Hasher: hasher}
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
		Accounts:      accountsSvc,
		Dashboard:     &admin.DashboardService{Repos: repos},
		Orders:        &admin.OrdersService{Repos: repos},
		Merchants:     &admin.MerchantsService{Repos: repos, Accounts: accountsSvc},
		Riders:        &admin.RidersService{Repos: repos, Accounts: accountsSvc},
		Users:         &admin.UsersService{Repos: repos, Tx: store, Cipher: cipher},
		Finance:       &admin.FinanceService{Repos: repos, Tx: store, Payout: payoutGw},
		Complaints:    &admin.ComplaintsService{Repos: repos},
		Settings:      &admin.SettingsService{Repos: repos},
		CouponsAdmin:  &admin.CouponService{Repos: repos},
		Workspace:     &admin.WorkspaceService{Repos: repos},
	}
	h.Ready = syscfg.Ready(client, rdb)

	return &App{
		Cfg: cfg, Logger: logger, Client: client, Rdb: rdb,
		Store: store, Enforcer: enforcer, Sender: sender,
		Auth: authSvc, System: systemSvc, Jobs: jobsRunner, Handler: h, MW: mw,
	}, nil
}
