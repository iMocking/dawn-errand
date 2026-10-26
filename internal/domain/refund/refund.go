// Package refund 退款/售后单。
package refund

import "context"

// Status 售后状态。
type Status string

const (
	StatusRequested Status = "requested"
	StatusApproved  Status = "approved"
	StatusRejected  Status = "rejected"
	StatusRefunded  Status = "refunded"
)

// Refund 售后单。
type Refund struct {
	ID           string
	OrderID      string
	UserID       string
	MerchantID   string
	Amount       float64
	Reason       string
	Status       Status
	PrevStatus   string
	Operator     string
	RejectReason string
	CreatedAt    int64
	HandledAt    int64
}

// Repository 售后台账端口。
type Repository interface {
	Save(ctx context.Context, r *Refund) error
	FindByOrder(ctx context.Context, orderID string) (*Refund, error)
	FindByID(ctx context.Context, id string) (*Refund, error)
	Update(ctx context.Context, id string, status Status, operator, rejectReason string, handledAt int64) error
}
