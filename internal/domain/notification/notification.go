// Package notification 站内通知。
package notification

// Notification 站内通知。
type Notification struct {
	ID        string
	UserID    string
	Title     string
	Content   string
	Type      string // order / system / audit
	RefID     string // 关联订单等
	ReadAt    int64
	CreatedAt int64
}

// Repository 通知仓储端口。
type Repository interface {
	Save(n *Notification) error
	ListByUser(userID string, page, pageSize int) ([]*Notification, int64, error)
	CountUnread(userID string) (int64, error)
	MarkRead(userID, id string, now int64) error
	MarkAllRead(userID string, now int64) error
	// DeleteBefore 清理历史通知。
	DeleteBefore(before int64) error
}
