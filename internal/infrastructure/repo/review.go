package repo

import (
	"context"

	"dawn-errand/internal/domain/review"
	"dawn-errand/internal/infrastructure/ent"
	entreview "dawn-errand/internal/infrastructure/ent/review"
)

// ReviewRepo 实现 review.Repository。
type ReviewRepo struct{ client *ent.Client }

func (r ReviewRepo) Save(ctx context.Context, v *review.Review) error {
	if v.ID == "" {
		v.ID = newULID()
	}
	return r.client.Review.Create().
		SetID(v.ID).SetOrderID(v.OrderID).SetUserID(v.UserID).
		SetMerchantID(v.MerchantID).SetRiderID(v.RiderID).
		SetRating(v.Rating).SetTags(v.Tags).SetContent(v.Content).SetImages(v.Images).
		SetCreatedAt(v.CreatedAt).
		Exec(ctx)
}

func (r ReviewRepo) FindByOrder(ctx context.Context, orderID string) (*review.Review, error) {
	e, err := r.client.Review.Query().Where(entreview.OrderID(orderID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toReview(e), nil
}

func (r ReviewRepo) FindByID(ctx context.Context, id string) (*review.Review, error) {
	e, err := r.client.Review.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toReview(e), nil
}

func (r ReviewRepo) List(ctx context.Context, q review.Query) ([]*review.Review, int64, error) {
	bq := r.client.Review.Query()
	if q.MerchantID != "" {
		bq.Where(entreview.MerchantID(q.MerchantID))
	}
	if q.RiderID != "" {
		bq.Where(entreview.RiderID(q.RiderID))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entreview.FieldCreatedAt)).
		Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*review.Review, 0, len(rows))
	for _, e := range rows {
		out = append(out, toReview(e))
	}
	return out, int64(total), nil
}

func (r ReviewRepo) Reply(ctx context.Context, id, merchantID, content string, now int64) (bool, error) {
	affected, err := r.client.Review.Update().
		Where(entreview.ID(id), entreview.MerchantID(merchantID)).
		SetReply(content).SetRepliedAt(now).Save(ctx)
	return affected > 0, err
}

func toReview(e *ent.Review) *review.Review {
	return &review.Review{
		ID: e.ID, OrderID: e.OrderID, UserID: e.UserID,
		MerchantID: e.MerchantID, RiderID: e.RiderID,
		Rating: e.Rating, Tags: e.Tags, Content: e.Content, Images: e.Images,
		Reply: e.Reply, RepliedAt: e.RepliedAt, CreatedAt: e.CreatedAt,
	}
}
