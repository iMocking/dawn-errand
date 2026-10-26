package app

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/notification"
)

// NotificationService 站内通知用例。
type NotificationService struct {
	Repos *application.RepoUnit
}

// List 通知列表。
func (s *NotificationService) List(ctx context.Context, userID string, page, size int) ([]*notification.Notification, int64, error) {
	page, size = normalizePageParams(page, size)
	return s.Repos.Notifications.ListByUser(ctx, userID, page, size)
}

// UnreadCount 未读数。
func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.Repos.Notifications.CountUnread(ctx, userID)
}

// MarkRead 标记单条已读。
func (s *NotificationService) MarkRead(ctx context.Context, userID, id string) error {
	return s.Repos.Notifications.MarkRead(ctx, userID, id, time.Now().Unix())
}

// MarkAllRead 全部已读。
func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.Repos.Notifications.MarkAllRead(ctx, userID, time.Now().Unix())
}
