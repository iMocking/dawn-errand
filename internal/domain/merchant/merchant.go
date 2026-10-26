// Package merchant 商家限界上下文。
package merchant

import "context"

import "dawn-errand/internal/domain/tenant"

// MerchantStatus 商家营业状态。
type MerchantStatus string

const (
	MerchantStatusOpen   MerchantStatus = "open"   // 营业中
	MerchantStatusRest   MerchantStatus = "rest"   // 休息中
	MerchantStatusBanned MerchantStatus = "banned" // 已封禁
)

// AuditStatus 入驻审核状态。
type AuditStatus string

const (
	AuditPending  AuditStatus = "pending"  // 待审核
	AuditApproved AuditStatus = "approved" // 已通过
	AuditRejected AuditStatus = "rejected" // 已驳回
)

// Application 商家入驻完整资料（资质/结算/账号）。
type Application struct {
	LicenseNo        string   `json:"licenseNo"`
	LicenseImage     string   `json:"licenseImage"`
	LegalName        string   `json:"legalName"`
	LegalIDCard      string   `json:"legalIdCard"`
	IDCardFront      string   `json:"idCardFront"`
	IDCardBack       string   `json:"idCardBack"`
	FoodLicenseNo    string   `json:"foodLicenseNo"`
	FoodLicenseImage string   `json:"foodLicenseImage"`
	StorefrontImages []string `json:"storefrontImages"`
	BankName         string   `json:"bankName"`
	BankAccountName  string   `json:"bankAccountName"`
	BankAccountNo    string   `json:"bankAccountNo"`
	AccountEmail     string   `json:"accountEmail"`
	AccountPhone     string   `json:"accountPhone"`
	Notes            string   `json:"notes"`
}

// Merchant 商家聚合根。
type Merchant struct {
	ID          string
	RegionID    string
	OwnerUserID string

	Name           string
	Category       string
	Address        string
	AddressSub     string
	Contact        string
	Phone          string
	OpenHours      string
	DeliveryRange  string
	Commission     float64 // 平台佣金比例 %
	BankAccount    string
	DeliveryMethod string

	// 门店位置与经营设置（App 商家端）。
	Lat          float64
	Lng          float64
	AutoAccept   bool
	MinPrice     float64
	DeliveryFee  float64
	PackingFee   float64
	Announcement string
	ServicePhone string

	Status       MerchantStatus
	AuditStatus  AuditStatus
	ApplyNo      string
	Operator     string // 申请人姓名
	AppliedAt    int64
	CertLabels   []string // 证照类型标签 license/idcard/storefront/foodlicense/pharmacylicense
	RejectReason string

	// 入驻资料与认证。
	Application Application
	SubmittedAt int64
	VerifiedAt  int64
	VerifiedBy  string

	Rating         float64
	MonthlyOrders  int64
	MonthlyRevenue float64
	AvgOrder       float64
	GoodRate       float64
	JoinedAt       int64
	LastUpdatedAt  int64
	Completeness   int
	ReviewCount    int
	LastOperation  string

	// 今日快照（运营指标）
	TodayOrders     int64
	TodayRevenue    float64
	OnTimeRate      float64
	PendingAbnormal int
}

// MerchantQuery 商家列表查询。
type MerchantQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string // 营业状态（列表页）
}

// AuditQuery 审核列表查询。
type AuditQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string // 审核状态
}

// Stats 商家模块顶部聚合。
type Stats struct {
	Total    int64
	WeekNew  int64
	Open     int64
	OpenRate float64
	Pending  int64
	Closed   int64 // 休息 + 封禁
}

// MerchantRepository 商家仓储端口。
type MerchantRepository interface {
	Save(ctx context.Context, m *Merchant) error
	FindByID(ctx context.Context, id string) (*Merchant, error)
	// List 商家列表（scope 强制数据隔离）。
	List(ctx context.Context, q MerchantQuery, scope tenant.Scope) ([]*Merchant, int64, error)
	// ListForClient C 端商家列表（区域内、审核通过、未封禁；营业中优先）。
	ListForClient(ctx context.Context, regionID, keyword, category string, page, pageSize int) ([]*Merchant, int64, error)
	// AuditList 入驻审核列表（scope 隔离）。
	AuditList(ctx context.Context, q AuditQuery, scope tenant.Scope) ([]*Merchant, int64, error)
	// UpdateAudit 审核处理。
	UpdateAudit(ctx context.Context, id string, approved bool, reason, operator string, now int64) error
	// UpdateStatus 变更营业状态。
	UpdateStatus(ctx context.Context, id string, status MerchantStatus, now int64) error
	// Stats 聚合统计（scope 隔离）。
	Stats(ctx context.Context, scope tenant.Scope) (*Stats, error)
	// CountByRegion 统计各区域商家数（settings 用）。
	CountByRegion(ctx context.Context, regionID string) (int64, error)
	// OnlineCountByRegion 统计区域在线商家数。
	OnlineCountByRegion(ctx context.Context, regionID string) (int64, error)
	// CountAll 全局商家数。
	CountAll(ctx context.Context) (int64, error)
	// CountWeekNew 本周新增（入驻）数。
	CountWeekNew(ctx context.Context, now int64) (int64, error)
}
