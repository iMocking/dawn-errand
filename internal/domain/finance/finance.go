// Package finance 资金结算限界上下文：提现审核、结算流水。
package finance

import "dawn-errand/internal/domain/tenant"

// WithdrawStatus 提现状态。
type WithdrawStatus string

const (
	WithdrawPending  WithdrawStatus = "待审核"
	WithdrawApproved WithdrawStatus = "已通过"
	WithdrawRejected WithdrawStatus = "已驳回"
)

// Event 提现时间线事件（JSON 小写，直接对齐管理端 DTO）。
type Event struct {
	Time     int64  `json:"time"`
	Text     string `json:"text"`
	Operator string `json:"operator"`
}

// Withdrawal 提现单聚合根。
type Withdrawal struct {
	ID       string
	RegionID string

	PartyType     string // rider 骑手 / merchant 商家
	PartyID       string
	PartyName     string
	PartyNo       string
	BankInfo      string
	Amount        float64
	Fee           float64
	FeeNote       string
	Status        WithdrawStatus
	Kind          string // 骑手提现 / 商家结算
	Applicant     string
	AccountHolder string
	WithdrawType  string
	ArriveExpect  string
	Purpose       string
	Balance       float64
	RemainBalance float64
	MonthOrders   int64
	MonthIncome   float64
	RiskItems     []string
	RiskWarning   string
	RejectReason  string
	PayoutRef     string
	PayoutAt      int64
	CreatedAt     int64
	Timeline      []Event
}

// SettlementType 结算类型。
type SettlementType string

const (
	SettlementMerchant SettlementType = "商家结算"
	SettlementRider    SettlementType = "骑手结算"
	SettlementRefund   SettlementType = "用户退款"
)

// SettlementStatus 结算状态。
type SettlementStatus string

const (
	SettlementArrived    SettlementStatus = "已到账"
	SettlementProcessing SettlementStatus = "处理中"
)

// Settlement 结算流水。
type Settlement struct {
	ID        string
	RegionID  string
	Party     string
	PartyType string
	PartyID   string
	Type      SettlementType
	Amount    float64
	Status    SettlementStatus
	Period    string
	SettledAt int64
	Action    string
}

// WithdrawQuery 提现列表查询。
type WithdrawQuery struct {
	Page      int
	PageSize  int
	Keyword   string
	Status    string
	PartyType string
	PartyID   string
}

// SettlementQuery 结算流水查询。
type SettlementQuery struct {
	Page      int
	PageSize  int
	Keyword   string
	Status    string
	Type      string
	PartyType string
	PartyID   string
}

// Stats 财务页顶部汇总。
type Stats struct {
	TodayIncome     float64
	YesterdayIncome float64
	PendingBalance  float64
	WithdrawPending int64
	AbnormalBills   int64
	AbnormalAmount  float64
}

// SettlementStats 结算页顶部汇总。
type SettlementStats struct {
	TodaySettled     float64
	YesterdaySettled float64
	WeekTotal        float64
	MonthTotal       float64
	PendingAmount    float64
	PendingCount     int64
}

// WithdrawalRepository 提现仓储端口。
type WithdrawalRepository interface {
	Save(w *Withdrawal) error
	FindByID(id string) (*Withdrawal, error)
	List(q WithdrawQuery, scope tenant.Scope) ([]*Withdrawal, int64, error)
	// Audit 审核通过/驳回。
	Audit(id string, approved bool, remark, operator string, now int64) error
	// Stats 汇总统计（scope 隔离）。
	Stats(scope tenant.Scope, now int64) (*Stats, error)
	// CountPending 待审核数。
	CountPending() (int64, error)
	// CountByPartySince 主体在指定时间后的提现申请次数（每日次数限制）。
	CountByPartySince(partyType, partyID string, since int64) (int64, error)
	// SumPendingByParty 主体待审核提现金额。
	SumPendingByParty(partyType, partyID string) (float64, error)
	// MarkPayout 记录打款回执（时间线 + 回执号）。
	MarkPayout(id, ref string, now int64) error
}

// SettlementRepository 结算流水仓储端口。
type SettlementRepository interface {
	Save(s *Settlement) error
	List(q SettlementQuery, scope tenant.Scope) ([]*Settlement, int64, error)
	// Stats 汇总（scope 隔离）。
	Stats(scope tenant.Scope, now int64) (*SettlementStats, error)
	// SumPendingByParty 主体待结算金额。
	SumPendingByParty(partyType, partyID string) (float64, error)
}
