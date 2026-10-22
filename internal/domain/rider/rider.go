// Package rider 骑手（跑手）限界上下文。
package rider

import "dawn-errand/internal/domain/tenant"

// RiderStatus 在线状态。
type RiderStatus string

const (
	RiderOnline  RiderStatus = "online"  // 在线
	RiderBusy    RiderStatus = "busy"    // 忙碌
	RiderOffline RiderStatus = "offline" // 离线
	RiderFrozen  RiderStatus = "frozen"  // 已冻结
)

// AuditStatus 入驻审核状态。
type AuditStatus string

const (
	AuditPending  AuditStatus = "pending"
	AuditApproved AuditStatus = "approved"
	AuditRejected AuditStatus = "rejected"
)

// Rider 骑手聚合根。
type Rider struct {
	ID       string
	RegionID string
	UserID   string

	Name    string
	Phone   string
	Area    string
	Tool    string // 配送工具：电动两轮车/电动三轮车/摩托车/步行
	Level   string // 等级，如 Lv.5
	IDCard  string
	Vehicle string
	Account string

	Status      RiderStatus
	AuditStatus AuditStatus
	ApplyNo     string
	SubmittedAt int64
	OrderTypes  string
	IDCheck     string // 已通过 / 待补充
	NoticeType  string // none / warning
	Notice      string
	DocCount    int
	Docs        string

	ServiceType      string // 外卖/帮送
	JoinedAt         int64
	ServiceMonths    int
	Rating           float64
	TodayOrders      int64
	TodayOrdersDelta string
	OnlineHours      float64
	MonthIncome      float64
	OnTimeRate       float64
	ComplaintsToday  int
	AvgMinutes       float64
	LastAction       string
	LastActionAt     int64
}

// RiderQuery 骑手列表查询。
type RiderQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
}

// AuditQuery 审核列表查询。
type AuditQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
}

// Stats 骑手模块顶部聚合。
type Stats struct {
	Total        int64
	Online       int64
	OnlineRate   float64
	PendingAudit int64
	Frozen       int64
	WeekDelta    string
	Audit        AuditStats
}

// AuditStats 审核 tab 汇总。
type AuditStats struct {
	TodayNew      int64
	Approved      int64
	Rejected      int64
	MonthApproved int64
	MomRate       string
}

// RiderRepository 骑手仓储端口。
type RiderRepository interface {
	Save(r *Rider) error
	FindByID(id string) (*Rider, error)
	List(q RiderQuery, scope tenant.Scope) ([]*Rider, int64, error)
	AuditList(q AuditQuery, scope tenant.Scope) ([]*Rider, int64, error)
	UpdateAudit(id string, approved bool, reason, operator string, now int64) error
	UpdateStatus(id string, status RiderStatus, now int64) error
	Stats(scope tenant.Scope) (*Stats, error)
	CountByRegion(regionID string) (int64, error)
	ActiveCountByRegion(regionID string) (int64, error)
	OnlineCount() (int64, error)
	TotalCount() (int64, error)
}
