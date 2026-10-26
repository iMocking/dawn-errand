package repo

import (
	"context"

	"dawn-errand/internal/domain/coupon"
	"dawn-errand/internal/infrastructure/ent"
	entcoupon "dawn-errand/internal/infrastructure/ent/usercoupon"
)

// CouponRepo 实现 coupon.Repository。
type CouponRepo struct{ client *ent.Client }

func (r CouponRepo) Save(ctx context.Context, c *coupon.Coupon) error {
	if c.ID == "" {
		c.ID = newULID()
	}
	exists, err := r.client.UserCoupon.Query().Where(entcoupon.ID(c.ID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		err = r.client.UserCoupon.UpdateOneID(c.ID).
			SetUserID(c.UserID).SetMerchantID(c.MerchantID).SetTitle(c.Title).
			SetAmount(c.Amount).SetMinSpend(c.MinSpend).SetStatus(string(c.Status)).
			SetExpiresAt(c.ExpiresAt).SetUsedOrderID(c.UsedOrderID).
			SetUpdatedAt(nowSec()).Exec(ctx)
		return err
	}
	return r.client.UserCoupon.Create().
		SetID(c.ID).SetUserID(c.UserID).SetMerchantID(c.MerchantID).SetTitle(c.Title).
		SetAmount(c.Amount).SetMinSpend(c.MinSpend).SetStatus(string(c.Status)).
		SetExpiresAt(c.ExpiresAt).SetUsedOrderID(c.UsedOrderID).
		SetCreatedAt(nowSec()).SetUpdatedAt(nowSec()).Exec(ctx)
}

func (r CouponRepo) FindByID(ctx context.Context, id string) (*coupon.Coupon, error) {
	e, err := r.client.UserCoupon.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toCoupon(e), nil
}

func (r CouponRepo) ListByUser(ctx context.Context, userID string) ([]*coupon.Coupon, error) {
	rows, err := r.client.UserCoupon.Query().
		Where(entcoupon.UserID(userID)).
		Order(ent.Desc(entcoupon.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*coupon.Coupon, 0, len(rows))
	for _, e := range rows {
		out = append(out, toCoupon(e))
	}
	return out, nil
}

func (r CouponRepo) ListAll(ctx context.Context, page, pageSize int, status, userID string) ([]*coupon.Coupon, int64, error) {
	q := r.client.UserCoupon.Query()
	if status != "" {
		q = q.Where(entcoupon.Status(status))
	}
	if userID != "" {
		q = q.Where(entcoupon.UserID(userID))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(entcoupon.FieldCreatedAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*coupon.Coupon, 0, len(rows))
	for _, e := range rows {
		out = append(out, toCoupon(e))
	}
	return out, int64(total), nil
}

func (r CouponRepo) MarkUsed(ctx context.Context, id, userID, orderID string, now int64) (bool, error) {
	affected, err := r.client.UserCoupon.Update().
		Where(entcoupon.ID(id), entcoupon.UserID(userID),
			entcoupon.Status(string(coupon.StatusUnused)), entcoupon.ExpiresAtGT(now)).
		SetStatus(string(coupon.StatusUsed)).SetUsedOrderID(orderID).SetUpdatedAt(now).
		Save(ctx)
	return affected > 0, err
}

func (r CouponRepo) ReleaseByOrder(ctx context.Context, orderID string, now int64) error {
	_, err := r.client.UserCoupon.Update().
		Where(entcoupon.UsedOrderID(orderID)).
		SetStatus(string(coupon.StatusUnused)).SetUsedOrderID("").SetUpdatedAt(now).
		Save(ctx)
	return err
}

func (r CouponRepo) Expire(ctx context.Context, userID string, now int64) error {
	_, err := r.client.UserCoupon.Update().
		Where(entcoupon.UserID(userID), entcoupon.Status(string(coupon.StatusUnused)),
			entcoupon.ExpiresAtLTE(now)).
		SetStatus(string(coupon.StatusExpired)).SetUpdatedAt(now).
		Save(ctx)
	return err
}

func (r CouponRepo) ExpireAll(ctx context.Context, now int64) error {
	_, err := r.client.UserCoupon.Update().
		Where(entcoupon.Status(string(coupon.StatusUnused)),
			entcoupon.ExpiresAtGT(0), entcoupon.ExpiresAtLTE(now)).
		SetStatus(string(coupon.StatusExpired)).SetUpdatedAt(now).
		Save(ctx)
	return err
}

func toCoupon(e *ent.UserCoupon) *coupon.Coupon {
	return &coupon.Coupon{
		ID: e.ID, UserID: e.UserID, MerchantID: e.MerchantID, Title: e.Title,
		Amount: e.Amount, MinSpend: e.MinSpend, Status: coupon.Status(e.Status),
		ExpiresAt: e.ExpiresAt, UsedOrderID: e.UsedOrderID,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}
