// Package system 系统配置限界上下文：可管理配置（SMTP/OSS/本地存储/Casbin）与首次启动初始化。
package system

// SettingKey 配置键。
type SettingKey string

const (
	KeySMTP          SettingKey = "smtp"
	KeySMS           SettingKey = "sms"
	KeyOSS           SettingKey = "oss"
	KeyLocal         SettingKey = "storage.local"
	KeyCasbin        SettingKey = "casbin"
	KeyWithdrawRule  SettingKey = "withdraw.rule"
	KeyBootstrap     SettingKey = "bootstrap"
)

// SettingStatus 配置启用状态。
type SettingStatus string

const (
	SettingEnabled  SettingStatus = "enabled"
	SettingDisabled SettingStatus = "disabled"
)

// SystemSetting 配置聚合。敏感值（密码/密钥）以密文存储，API 返回脱敏。
type SystemSetting struct {
	Key       SettingKey
	Status    SettingStatus
	ValueJSON string // 配置 JSON（敏感字段已加密）
	UpdatedBy string
	UpdatedAt int64
}

// SMTPConfig SMTP 配置。
type SMTPConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password,omitempty"` // 密文
	FromEmail  string `json:"from_email"`
	FromName   string `json:"from_name"`
	Encryption string `json:"encryption"` // tls / ssl / none
	TimeoutSec int    `json:"timeout_sec"`
}

// SMSConfig 短信服务配置。
type SMSConfig struct {
	Provider     string `json:"provider"` // aliyun / tencent / other
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key,omitempty"` // 密文
	SignName     string `json:"sign_name"`
	TemplateCode string `json:"template_code"`
}

// WithdrawRuleConfig 提现规则配置。
type WithdrawRuleConfig struct {
	Enabled     bool    `json:"enabled"`      // 是否启用提现
	FeeRate     float64 `json:"fee_rate"`     // 提现手续费率（0.006 = 0.6%）
	MinAmount   float64 `json:"min_amount"`   // 单笔最低提现金额（元）
	MaxAmount   float64 `json:"max_amount"`   // 单笔最高提现金额（元）
	DailyLimit  int     `json:"daily_limit"`  // 每日提现次数上限
	Mode        string  `json:"mode"`         // auto 自动审核 / manual 人工审核
	SettleDays  int     `json:"settle_days"`  // 到账周期（T+N 个工作日）
}

// OSSConfig 对象存储配置。
type OSSConfig struct {
	Provider      string `json:"provider"` // oss / s3 ...
	Endpoint      string `json:"endpoint"`
	Region        string `json:"region"`
	Bucket        string `json:"bucket"`
	AccessKey     string `json:"access_key"`
	SecretKey     string `json:"secret_key,omitempty"` // 密文
	PublicBaseURL string `json:"public_base_url"`
	PathPrefix    string `json:"path_prefix"`
}

// LocalStorageConfig 本地图片存储配置。
type LocalStorageConfig struct {
	RootDir       string   `json:"root_dir"`
	PublicBaseURL string   `json:"public_base_url"`
	MaxSizeMB     int64    `json:"max_size_mb"`
	AllowedMIME   []string `json:"allowed_mime"`
	PathPrefix    string   `json:"path_prefix"`
}

// CasbinConfig Casbin 身份配置。
type CasbinConfig struct {
	Model       string   `json:"model"`
	Policy      []string `json:"policy"`       // p/g 行
	PolicyStore string   `json:"policy_store"` // file / db
}

// BootstrapState 初始化状态。
type BootstrapState struct {
	Initialized bool
	AdminUserID string
	InitAt      int64
}

// SettingRepository 配置仓储端口。
type SettingRepository interface {
	// Get 读取配置。
	Get(key SettingKey) (*SystemSetting, error)
	// Upsert 保存配置。
	Upsert(s *SystemSetting) error
	// GetBootstrap 读取初始化状态。
	GetBootstrap() (*BootstrapState, error)
	// MarkInitialized 写入初始化完成标记（幂等：已存在则不覆盖）。
	MarkInitialized(adminUserID string, now int64) error
}
