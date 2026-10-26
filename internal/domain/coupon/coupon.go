// Package coupon 用户优惠券。
package coupon

import "context"

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
	Save(ctx context.Context, c *Coupon) error
	FindByID(ctx context.Context, id string) (*Coupon, error)
	ListByUser(ctx context.Context, userID string) ([]*Coupon, error)
	// ListAll 管理端分页查询。
	ListAll(ctx context.Context, page, pageSize int, status, userID string) ([]*Coupon, int64, error)
	// MarkUsed 条件占用（unused 且未过期）；0 行返回 false。
	MarkUsed(ctx context.Context, id, userID, orderID string, now int64) (bool, error)
	// ReleaseByOrder 取消/退款释放订单占用的券。
	ReleaseByOrder(ctx context.Context, orderID string, now int64) error
	// Expire 将过期未用券置为 expired。
	Expire(ctx context.Context, userID string, now int64) error
	// ExpireAll 全量过期（定时任务）。
	ExpireAll(ctx context.Context, now int64) error
}
