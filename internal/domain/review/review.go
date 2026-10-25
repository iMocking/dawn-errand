// Package review 订单评价。
package review

// Review 评价。
type Review struct {
	ID         string
	OrderID    string
	UserID     string
	MerchantID string
	RiderID    string
	Rating     int
	Tags       []string
	Content    string
	Images     []string
	Reply      string
	RepliedAt  int64
	CreatedAt  int64
}

// Query 评价查询。
type Query struct {
	MerchantID string
	RiderID    string
	Page       int
	PageSize   int
}

// Repository 评价仓储端口。
type Repository interface {
	Save(r *Review) error
	FindByOrder(orderID string) (*Review, error)
	FindByID(id string) (*Review, error)
	List(q Query) ([]*Review, int64, error)
	Reply(id, merchantID, content string, now int64) (bool, error)
}
