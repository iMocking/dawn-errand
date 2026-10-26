// Package application 定义应用层端口（依赖倒置：infrastructure 实现这些端口）。
package application

import (
	"context"
	"errors"
	"time"

	"dawn-errand/internal/domain/address"
	"dawn-errand/internal/domain/catalog"
	"dawn-errand/internal/domain/cms"
	"dawn-errand/internal/domain/complaint"
	"dawn-errand/internal/domain/coupon"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/notification"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/realname"
	"dawn-errand/internal/domain/refund"
	"dawn-errand/internal/domain/review"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/system"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/tracking"
	"dawn-errand/internal/domain/userprofile"
	"dawn-errand/internal/domain/wallet"
)

// ErrNotFound 通用未找到错误（仓储层统一映射）。
var ErrNotFound = errors.New("not found")

// AuditEntry 审计日志条目。
type AuditEntry struct {
	UserID    string
	Action    string // 动作：login / password.change / config.update / bootstrap / permission.change ...
	Object    string
	Result    string // success / fail
	RequestID string
	Operator  string
	IP        string
}

// AuditLogger 审计日志端口（不阻塞主流程）。
type AuditLogger interface {
	Record(ctx context.Context, e AuditEntry)
}

// RepoUnit 仓储集合（普通客户端绑定）。
type RepoUnit struct {
	Users          identity.UserRepository
	Credentials    identity.CredentialRepository
	EmailTokens    identity.EmailTokenRepository
	PasswordTokens identity.PasswordTokenRepository
	Regions        tenant.RegionRepository
	Settings       system.SettingRepository
	Merchants      merchant.MerchantRepository
	Riders         rider.RiderRepository
	Orders         order.OrderRepository
	OrderEvents    order.EventRepository
	Payments       order.PaymentRepository
	Complaints     complaint.ComplaintRepository
	Withdrawals    finance.WithdrawalRepository
	Settlements    finance.SettlementRepository
	Announcements  cms.AnnouncementRepository
	FeeRules       cms.FeeRuleRepository
	Profiles       userprofile.ProfileRepository
	RealNames      realname.Repository
	Catalog        catalog.Repository
	Addresses      address.Repository
	Locations      tracking.Repository
	Notifications  notification.Repository
	Wallets        wallet.Repository
	Reviews        review.Repository
	Refunds        refund.Repository
	Coupons        coupon.Repository
	Audit          AuditLogger
}

// TxRunner 事务运行器：跨聚合一致性 MUST 在 application 层通过事务执行。
type TxRunner interface {
	// RunInTx 在事务内执行 fn，返回 tx 绑定的仓储集合。
	RunInTx(ctx context.Context, fn func(tx *RepoUnit) error) error
}

// MailerPort 邮件发送端口（infrastructure 实现 SMTP；dev 可写 outbox）。
type MailerPort interface {
	// Send 发送 HTML 邮件。
	Send(ctx context.Context, to, subject, html string) error
	// TestConnection 测试 SMTP 连接。
	TestConnection(ctx context.Context) error
}

// StoragePort 对象存储端口（本地存储/OSS 可切换实现）。
type StoragePort interface {
	// Put 保存文件，返回公开访问 URL。
	Put(ctx context.Context, data []byte, filename string, mime string) (string, error)
	// PutPrivate 保存到私有桶（证件等 PII），返回受控访问 URL；rel 为桶内相对路径。
	PutPrivate(ctx context.Context, data []byte, rel, filename, mime string) (url string, relOut string, err error)
	// ReadPrivate 读取私有桶对象；实现必须拒绝路径穿越。
	ReadPrivate(ctx context.Context, rel string) (data []byte, mime string, err error)
	// TestConnection 测试存储可用性。
	TestConnection(ctx context.Context) error
}

// PasswordHasher 密码强哈希端口（Argon2id）。
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

// TokenClaims 访问令牌声明。
type TokenClaims struct {
	UserID         string
	Role           string
	SessionVersion int64
}

// TokenManager 令牌端口：访问令牌 JWT + 刷新令牌轮换。
type TokenManager interface {
	// IssueAccessToken 签发访问令牌。
	IssueAccessToken(claims TokenClaims, ttl time.Duration) (string, error)
	// ParseAccessToken 校验并解析访问令牌。
	ParseAccessToken(token string) (*TokenClaims, error)
	// IssueRefreshToken 签发刷新令牌；返回明文与哈希。
	IssueRefreshToken(ctx context.Context, userID string, ttl time.Duration) (string, string, error)
	// UserIDForRefreshToken 由明文刷新令牌解析用户 ID。
	UserIDForRefreshToken(ctx context.Context, token string) (string, error)
	// VerifyRefreshToken 校验刷新令牌有效。
	VerifyRefreshToken(ctx context.Context, userID, token string) error
	// RevokeRefreshTokens 撤销用户全部刷新令牌（SCAN 遍历，并清理 dawn:rt:* 反向索引）。
	RevokeRefreshTokens(ctx context.Context, userID string) error
	// RevokeRefreshToken 撤销单个刷新令牌。
	RevokeRefreshToken(ctx context.Context, userID, token string) error
}

// ConfigCipher 配置敏感值加密端口（AES-256-GCM）。
type ConfigCipher interface {
	Encrypt(plain string) (string, error)
	Decrypt(cipher string) (string, error)
}

// CodeStore 一次性验证码存储端口（Redis 实现，手机号验证码登录用）。
type CodeStore interface {
	// Set 写入验证码并设置过期时间。
	Set(ctx context.Context, key, code string, ttl time.Duration) error
	// Get 读取验证码（不存在返回空串）。
	Get(ctx context.Context, key string) (string, error)
	// Delete 删除验证码（验证成功后作废）。
	Delete(ctx context.Context, key string) error
}

// RateLimiter 限流端口（Redis 实现）。
type RateLimiter interface {
	// Allow 判定是否放行；limit 为窗口内最大次数。
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// Locker 分布式锁端口（Redis SETNX 实现，用于定时任务单实例执行）。
type Locker interface {
	// TryLock 尝试加锁；成功返回持有者 token（Unlock 时用于校验归属），未抢到返回 ok=false。
	TryLock(ctx context.Context, key string, ttl time.Duration) (token string, ok bool, err error)
	// Unlock 仅当持有者 token 与锁当前值一致时释放，避免删掉他人（超时后被接管）的锁。
	Unlock(ctx context.Context, key, token string) error
}

// GeoStore 骑手位置地理索引（Redis GEO，供派单/附近查询扩展）。
type GeoStore interface {
	SetRiderLocation(ctx context.Context, regionID, riderID string, lat, lng float64) error
}

// PayoutGateway 提现打款端口（Mock/银行/微信代付可替换）。
type PayoutGateway interface {
	// Payout 发起打款，返回渠道回执号。
	Payout(ctx context.Context, withdrawalID, partyType, partyID string, amount float64) (string, error)
}

// PolicyManager 权限策略管理端口（Casbin 实现）。
type PolicyManager interface {
	Enforce(role, obj, act string) (bool, error)
	PermissionsForRole(role string) []string
	// SavePolicies 全量替换策略并重载。
	SavePolicies(policies [][]string) error
	// Policies 返回当前策略（含 ptype 前缀）。
	Policies() [][]string
}

// MailEntry dev 邮件发件箱条目。
type MailEntry struct {
	ID        string
	ToEmail   string
	Subject   string
	Body      string
	CreatedAt int64
}

// OutboxProvider dev 邮件发件箱查询。
type OutboxProvider func(ctx context.Context, limit int) ([]MailEntry, error)

type reqIDKey struct{}
type ipKey struct{}

// WithRequestMeta 注入 request_id 与 IP（认证中间件调用，供审计使用）。
func WithRequestMeta(ctx context.Context, requestID, ip string) context.Context {
	ctx = context.WithValue(ctx, reqIDKey{}, requestID)
	return context.WithValue(ctx, ipKey{}, ip)
}

// RequestID 从上下文读取 request_id。
func RequestID(ctx context.Context) string {
	if v, _ := ctx.Value(reqIDKey{}).(string); v != "" {
		return v
	}
	return ""
}

// ClientIP 从上下文读取客户端 IP。
func ClientIP(ctx context.Context) string {
	if v, _ := ctx.Value(ipKey{}).(string); v != "" {
		return v
	}
	return ""
}
