package repo

import (
	"context"

	"dawn-errand/internal/domain/realname"
	"dawn-errand/internal/infrastructure/ent"
	entrealname "dawn-errand/internal/infrastructure/ent/realnameverification"
)

// RealNameRepo 实现 realname.Repository。
type RealNameRepo struct{ client *ent.Client }

func (r RealNameRepo) Save(ctx context.Context, v *realname.Verification) error {
	exists, err := r.client.RealNameVerification.Query().
		Where(entrealname.UserID(v.UserID)).Exist(ctx)
	if err != nil {
		return err
	}
	if v.ID == "" {
		v.ID = newULID()
	}
	if exists {
		_, err := r.client.RealNameVerification.Update().
			Where(entrealname.UserID(v.UserID)).
			SetRealName(v.RealName).SetIDCardCipher(v.IDCardCipher).
			SetStatus(realname.StatusPending).SetRejectReason("").SetReviewedBy("").
			SetAppliedAt(v.AppliedAt).SetReviewedAt(0).SetUpdatedAt(v.UpdatedAt).
			Save(ctx)
		return err
	}
	return r.client.RealNameVerification.Create().
		SetID(v.ID).SetUserID(v.UserID).SetRealName(v.RealName).
		SetIDCardCipher(v.IDCardCipher).SetStatus(realname.StatusPending).
		SetAppliedAt(v.AppliedAt).SetUpdatedAt(v.UpdatedAt).
		Exec(ctx)
}

func toRealName(e *ent.RealNameVerification) *realname.Verification {
	if e == nil {
		return nil
	}
	return &realname.Verification{
		ID:           e.ID,
		UserID:       e.UserID,
		RealName:     e.RealName,
		IDCardCipher: e.IDCardCipher,
		Status:       e.Status,
		RejectReason: e.RejectReason,
		ReviewedBy:   e.ReviewedBy,
		AppliedAt:    e.AppliedAt,
		ReviewedAt:   e.ReviewedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func (r RealNameRepo) FindByID(ctx context.Context, id string) (*realname.Verification, error) {
	e, err := r.client.RealNameVerification.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toRealName(e), nil
}

func (r RealNameRepo) FindByUserID(ctx context.Context, userID string) (*realname.Verification, error) {
	e, err := r.client.RealNameVerification.Query().
		Where(entrealname.UserID(userID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toRealName(e), nil
}

func (r RealNameRepo) List(ctx context.Context, q realname.Query) ([]*realname.Verification, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	bq := r.client.RealNameVerification.Query()
	if q.Status != "" {
		bq = bq.Where(entrealname.Status(q.Status))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := bq.Order(ent.Desc(entrealname.FieldAppliedAt)).
		Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*realname.Verification, 0, len(rows))
	for _, e := range rows {
		out = append(out, toRealName(e))
	}
	return out, int64(total), nil
}

func (r RealNameRepo) UpdateReview(ctx context.Context, id, status, reason, reviewer string, reviewedAt int64) error {
	_, err := r.client.RealNameVerification.Update().
		Where(entrealname.ID(id)).
		SetStatus(status).SetRejectReason(reason).SetReviewedBy(reviewer).
		SetReviewedAt(reviewedAt).SetUpdatedAt(reviewedAt).
		Save(ctx)
	return err
}
