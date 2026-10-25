// Package coupon 用户优惠券。
package coupon

// Status 券状态。
type Status string

const (
	StatusUnused  Status = "unused"
	StatusUsed    Status = "used"
	StatusExpired Status = "expired"
)

// Coupon 用户优惠券。
type Coupon struct {
	ID          string
	UserID      string
	MerchantID  string // 空=平台券
	Title       string
	Amount      float64
	MinSpend    float64
	Status      Status
	ExpiresAt   int64
	UsedOrderID string
	CreatedAt   int64
	UpdatedAt   int64
}

// Repository 优惠券仓储端口。
type Repository interface {
	Save(c *Coupon) error
	FindByID(id string) (*Coupon, error)
	ListByUser(userID string) ([]*Coupon, error)
	// ListAll 管理端分页查询。
	ListAll(page, pageSize int, status, userID string) ([]*Coupon, int64, error)
	// MarkUsed 条件占用（unused 且未过期）；0 行返回 false。
	MarkUsed(id, userID, orderID string, now int64) (bool, error)
	// ReleaseByOrder 取消/退款释放订单占用的券。
	ReleaseByOrder(orderID string, now int64) error
	// Expire 将过期未用券置为 expired。
	Expire(userID string, now int64) error
	// ExpireAll 全量过期（定时任务）。
	ExpireAll(now int64) error
}
