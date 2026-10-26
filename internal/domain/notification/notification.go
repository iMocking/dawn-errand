// Package notification 站内通知。
package notification

import "context"

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
	Save(ctx context.Context, n *Notification) error
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Notification, int64, error)
	CountUnread(ctx context.Context, userID string) (int64, error)
	MarkRead(ctx context.Context, userID, id string, now int64) error
	MarkAllRead(ctx context.Context, userID string, now int64) error
	// DeleteBefore 清理历史通知。
	DeleteBefore(ctx context.Context, before int64) error
}
