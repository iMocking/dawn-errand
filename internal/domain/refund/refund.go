// Package refund 退款/售后单。
package refund

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
	Save(r *Refund) error
	FindByOrder(orderID string) (*Refund, error)
	FindByID(id string) (*Refund, error)
	Update(id string, status Status, operator, rejectReason string, handledAt int64) error
}
