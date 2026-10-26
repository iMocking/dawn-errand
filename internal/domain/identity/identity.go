// Package identity 身份权限限界上下文：用户、凭证、邮箱验证、密码重置、角色。
package identity

import (
	"regexp"
	"time"

	"dawn-errand/internal/domain/tenant"
)

// 角色（系统内置，与 Casbin policy 对齐）。
const (
	RoleUser          = "user"           // 用户（正式会员）
	RoleSuperAdmin    = "super_admin"    // 系统管理员
	RoleCSManager     = "cs_manager"     // 客服经理
	RoleCSAgent       = "cs_agent"       // 平台客服
	RoleRegionStation = "region_station" // 区域站长
	RoleMerchant      = "merchant"       // 商家
	RoleDriver        = "driver"         // 司机（跑手）
	RoleSystem        = "system"         // 系统级服务身份（非人工，禁登录）
)

// AllRoles 全部内置角色。
var AllRoles = []string{
	RoleSuperAdmin, RoleCSManager, RoleCSAgent, RoleRegionStation,
	RoleMerchant, RoleDriver, RoleUser, RoleSystem,
}

// IsPlatformRole 是否为平台侧管理角色（含区域站长）。
// 用于「代主体上传的私有文件归属在 user/<管理员ID> 下」时的读取判定：
// 客服代商家传的证照，区域站长审核必须能看到；普通会员的证件对站长不可见。
func IsPlatformRole(role string) bool {
	switch role {
	case RoleSuperAdmin, RoleCSManager, RoleCSAgent, RoleRegionStation:
		return true
	default:
		return false
	}
}

// SystemIdentity 系统级身份定义（初始化完成后自动创建，与普通用户同表）。
type SystemIdentity struct {
	Email string
	Name  string
}

// SystemIdentities 内置系统级身份清单（邮箱固定，幂等创建）。
var SystemIdentities = []SystemIdentity{
	{Email: "mailer@system.shuguang", Name: "邮件服务系统"},
	{Email: "sms@system.shuguang", Name: "短信服务系统"},
	{Email: "inmail@system.shuguang", Name: "站内信通知服务系统"},
	{Email: "system@shuguang.local", Name: "基础系统服务"},
}

// 角色中文显示名（DTO 层映射用）。
var RoleDisplayName = map[string]string{
	RoleSuperAdmin:    "系统管理员",
	RoleCSManager:     "客服经理",
	RoleCSAgent:       "平台客服",
	RoleRegionStation: "区域站长",
	RoleMerchant:      "商家",
	RoleDriver:        "司机",
	RoleUser:          "正式会员",
	RoleSystem:        "系统服务",
}

var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

// ValidatePhone 校验中国大陆手机号格式（11 位，1 开头，第二位 3-9）。
func ValidatePhone(phone string) bool { return phoneRe.MatchString(phone) }

// UserStatus 用户状态。
type UserStatus string

const (
	UserStatusActive UserStatus = "active" // 正常
	UserStatusBanned UserStatus = "banned" // 已封禁
)

// User 用户聚合根。身份权限限界上下文核心模型。
type User struct {
	ID            string
	Email         string
	Phone         string
	Name          string // 昵称/姓名
	Avatar        string
	Status        UserStatus
	EmailVerified bool
	Role          string
	// 租户/主体绑定：不同角色绑定不同主体
	RegionID     string // 区域站长/正式会员：所属区域
	MerchantID   string // 商家：所属商家
	RiderID      string // 司机：所属骑手档案
	RegisteredAt int64
	UpdatedAt    int64
}

// IsActive 用户是否可用。
func (u *User) IsActive() bool { return u.Status == UserStatusActive }

// Credential 登录凭证。仅存强哈希，不存明文。
type Credential struct {
	UserID         string
	PasswordHash   string
	SessionVersion int64 // 密码修改/封禁后递增，使旧 JWT 与刷新令牌失效
	UpdatedAt      int64
}

// EmailVerificationToken 邮箱验证令牌。数据库仅存哈希，一次性、带 TTL。
type EmailVerificationToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt int64
	UsedAt    int64 // 0 表示未使用
}

// PasswordResetToken 密码重置令牌。数据库仅存哈希，一次性、带 TTL。
type PasswordResetToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt int64
	UsedAt    int64
}

// UserRepository 用户仓储端口（由 infrastructure 实现）。
type UserRepository interface {
	// Save 保存用户（幂等 upsert）。
	Save(u *User) error
	// FindByID 按 ID 查用户。
	FindByID(id string) (*User, error)
	// FindByEmail 按邮箱查用户。
	FindByEmail(email string) (*User, error)
	// FindByPhone 按手机号查用户。
	FindByPhone(phone string) (*User, error)
	// List 分页查询（keyword 匹配昵称/邮箱/手机号，status 过滤；scope 隔离）。
	List(q UserQuery, sc tenant.Scope) ([]*User, int64, error)
	// UpdateStatus 更新状态并递增会话版本（封禁/解封）。
	UpdateStatus(id string, status UserStatus) error
	// UpdateRole 更新后台角色（角色变更即时生效，下次请求按新角色鉴权）。
	UpdateRole(id string, role string) error
	// CountRegisteredSince 统计指定时间后注册的用户数（scope 隔离）。
	CountRegisteredSince(sc tenant.Scope, since int64) (int64, error)
	// CountStatus 统计指定状态的用户数（scope 隔离）。
	CountStatus(sc tenant.Scope, status UserStatus) (int64, error)
}

// UserQuery 用户分页查询。
type UserQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
}

// CredentialRepository 凭证仓储端口。
type CredentialRepository interface {
	// Save 保存凭证。
	Save(c *Credential) error
	// FindByUserID 按用户 ID 查凭证。
	FindByUserID(userID string) (*Credential, error)
	// BumpSessionVersion 递增会话版本（撤销全部会话）。
	BumpSessionVersion(userID string) (int64, error)
}

// EmailTokenRepository 邮箱验证令牌仓储端口。
type EmailTokenRepository interface {
	// Create 创建令牌记录。
	Create(t *EmailVerificationToken) error
	// FindByHash 按哈希查有效令牌（未使用且未过期）。
	FindByHash(hash string) (*EmailVerificationToken, error)
	// MarkUsed 标记已使用。
	MarkUsed(id string) error
	// DeleteExpired 清理过期记录。
	DeleteExpired(now int64) error
}

// PasswordTokenRepository 密码重置令牌仓储端口。
type PasswordTokenRepository interface {
	// Create 创建令牌记录。
	Create(t *PasswordResetToken) error
	// FindByHash 按哈希查有效令牌。
	FindByHash(hash string) (*PasswordResetToken, error)
	// MarkUsed 标记已使用。
	MarkUsed(id string) error
	// DeleteExpired 清理过期记录。
	DeleteExpired(now int64) error
	// DeleteByUser 删除用户全部重置令牌（改密成功后）。
	DeleteByUser(userID string) error
}

// Now 供领域层测试替换的时间源（默认系统时钟）。
var Now = func() int64 { return time.Now().Unix() }
