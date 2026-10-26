package repo

import (
	"context"

	"dawn-errand/internal/domain/notification"
	"dawn-errand/internal/infrastructure/ent"
	entnotification "dawn-errand/internal/infrastructure/ent/notification"
)

// NotificationRepo 实现 notification.Repository。
type NotificationRepo struct{ client *ent.Client }

func (r NotificationRepo) Save(ctx context.Context, n *notification.Notification) error {
	if n.ID == "" {
		n.ID = newULID()
	}
	return r.client.Notification.Create().
		SetID(n.ID).SetUserID(n.UserID).SetTitle(n.Title).SetContent(n.Content).
		SetType(n.Type).SetRefID(n.RefID).SetReadAt(n.ReadAt).SetCreatedAt(n.CreatedAt).
		Exec(ctx)
}

func (r NotificationRepo) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*notification.Notification, int64, error) {
	bq := r.client.Notification.Query().Where(entnotification.UserID(userID))
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(page, pageSize)
	rows, err := bq.Order(ent.Desc(entnotification.FieldCreatedAt)).
		Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*notification.Notification, 0, len(rows))
	for _, e := range rows {
		out = append(out, &notification.Notification{
			ID: e.ID, UserID: e.UserID, Title: e.Title, Content: e.Content,
			Type: e.Type, RefID: e.RefID, ReadAt: e.ReadAt, CreatedAt: e.CreatedAt,
		})
	}
	return out, int64(total), nil
}

func (r NotificationRepo) CountUnread(ctx context.Context, userID string) (int64, error) {
	n, err := r.client.Notification.Query().
		Where(entnotification.UserID(userID), entnotification.ReadAt(0)).
		Count(ctx)
	return int64(n), err
}

func (r NotificationRepo) MarkRead(ctx context.Context, userID, id string, now int64) error {
	_, err := r.client.Notification.Update().
		Where(entnotification.ID(id), entnotification.UserID(userID)).
		SetReadAt(now).Save(ctx)
	return err
}

func (r NotificationRepo) MarkAllRead(ctx context.Context, userID string, now int64) error {
	_, err := r.client.Notification.Update().
		Where(entnotification.UserID(userID), entnotification.ReadAt(0)).
		SetReadAt(now).Save(ctx)
	return err
}

func (r NotificationRepo) DeleteBefore(ctx context.Context, before int64) error {
	_, err := r.client.Notification.Delete().
		Where(entnotification.CreatedAtLT(before)).Exec(ctx)
	return err
}
