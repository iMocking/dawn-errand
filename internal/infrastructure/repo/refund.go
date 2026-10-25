package repo

import (
	"context"

	"dawn-errand/internal/domain/refund"
	"dawn-errand/internal/infrastructure/ent"
	entrefund "dawn-errand/internal/infrastructure/ent/refund"
)

// RefundRepo 实现 refund.Repository。
type RefundRepo struct{ client *ent.Client }

func (r RefundRepo) Save(v *refund.Refund) error {
	if v.ID == "" {
		v.ID = newULID()
	}
	return r.client.Refund.Create().
		SetID(v.ID).SetOrderID(v.OrderID).SetUserID(v.UserID).SetMerchantID(v.MerchantID).
		SetAmount(v.Amount).SetReason(v.Reason).SetStatus(string(v.Status)).
		SetPrevStatus(v.PrevStatus).SetOperator(v.Operator).SetRejectReason(v.RejectReason).
		SetCreatedAt(v.CreatedAt).SetHandledAt(v.HandledAt).
		Exec(context.Background())
}

func (r RefundRepo) FindByOrder(orderID string) (*refund.Refund, error) {
	e, err := r.client.Refund.Query().
		Where(entrefund.OrderID(orderID)).
		Order(ent.Desc(entrefund.FieldCreatedAt)).First(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toRefund(e), nil
}

func (r RefundRepo) FindByID(id string) (*refund.Refund, error) {
	e, err := r.client.Refund.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toRefund(e), nil
}

func (r RefundRepo) Update(id string, status refund.Status, operator, rejectReason string, handledAt int64) error {
	upd := r.client.Refund.UpdateOneID(id).
		SetStatus(string(status)).SetOperator(operator).SetHandledAt(handledAt)
	if rejectReason != "" {
		upd.SetRejectReason(rejectReason)
	}
	return upd.Exec(context.Background())
}

func toRefund(e *ent.Refund) *refund.Refund {
	return &refund.Refund{
		ID: e.ID, OrderID: e.OrderID, UserID: e.UserID, MerchantID: e.MerchantID,
		Amount: e.Amount, Reason: e.Reason, Status: refund.Status(e.Status),
		PrevStatus: e.PrevStatus, Operator: e.Operator, RejectReason: e.RejectReason,
		CreatedAt: e.CreatedAt, HandledAt: e.HandledAt,
	}
}
