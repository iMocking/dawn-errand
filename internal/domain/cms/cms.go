// Package cms 内容与运营配置限界上下文：公告、费率规则。
package cms

import "dawn-errand/internal/domain/tenant"

// AnnouncementStatus 公告状态。
type AnnouncementStatus string

const (
	AnnPublished AnnouncementStatus = "已发布"
	AnnScheduled AnnouncementStatus = "待发布"
	AnnReviewing AnnouncementStatus = "审核中"
	AnnOffline   AnnouncementStatus = "已下线"
)

// AnnouncementStat 推送统计。
type AnnouncementStat struct {
	Label     string
	Value     string
	Highlight bool
}

// Announcement 公告聚合根。
type Announcement struct {
	ID          string
	Title       string
	Type        string // 费率 / 活动 / 提醒 / 协议
	TypeTone    string // orange / blue / yellow / purple
	Audience    string
	Status      AnnouncementStatus
	PublishedAt int64
	PlannedAt   int64
	ActionName  string
	Content     string
	Tags        []string
	Stats       []AnnouncementStat
	CreatedBy   string
	CreatedAt   int64
	UpdatedAt   int64
}

// AnnouncementQuery 公告查询。
type AnnouncementQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
}

// FeeRuleStatus 费率规则状态。
type FeeRuleStatus string

const (
	FeeEnabled  FeeRuleStatus = "已启用"
	FeeDisabled FeeRuleStatus = "已停用"
)

// FeeRule 费率规则行。
type FeeRule struct {
	ID              string
	Category        string // summary / basic / ladder / special / scope
	Name            string
	Status          FeeRuleStatus
	Value           string
	Desc            string
	Accent          bool
	Highlight       bool
	Rule            string
	Example         string
	Fee             string
	RiderShare      string
	IconColor       string
	Version         string
	UpdatedAt       int64
	NextEffectiveAt int64
	Enabled         bool
}

// AnnouncementRepository 公告仓储端口。
type AnnouncementRepository interface {
	Save(a *Announcement) error
	FindByID(id string) (*Announcement, error)
	List(q AnnouncementQuery, scope tenant.Scope) ([]*Announcement, int64, error)
	// ListDueScheduled 待发布的到期公告（cron 调度发布用）。
	ListDueScheduled(now int64) ([]*Announcement, error)
	// MarkPublished 标记已发布。
	MarkPublished(id string, now int64) error
	// UpdateStatus 更新状态（下线/撤回/发布）。
	UpdateStatus(id string, status AnnouncementStatus, now int64) error
}

// FeeRuleRepository 费率规则仓储端口。
type FeeRuleRepository interface {
	Save(r *FeeRule) error
	List(scope tenant.Scope) ([]*FeeRule, error)
	FindByID(id string) (*FeeRule, error)
	Delete(id string) error
}
