package repo

import (
	"context"

	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/payment"
)

// PaymentRepo 实现 order.PaymentRepository。
type PaymentRepo struct{ client *ent.Client }

func (r PaymentRepo) Save(p *order.Payment) error {
	if p.ID == "" {
		p.ID = newULID()
	}
	return r.client.Payment.Create().
		SetID(p.ID).SetOrderID(p.OrderID).SetUserID(p.UserID).
		SetChannel(p.Channel).SetAmount(p.Amount).SetStatus(string(p.Status)).
		SetTradeNo(p.TradeNo).SetPaidAt(p.PaidAt).
		SetCreatedAt(nowSec()).SetUpdatedAt(nowSec()).
		Exec(context.Background())
}

func (r PaymentRepo) FindByOrder(orderID string) (*order.Payment, error) {
	e, err := r.client.Payment.Query().
		Where(payment.OrderID(orderID)).
		Order(ent.Desc(payment.FieldCreatedAt)).
		First(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &order.Payment{
		ID: e.ID, OrderID: e.OrderID, UserID: e.UserID, Channel: e.Channel,
		Amount: e.Amount, Status: order.PaymentStatus(e.Status), TradeNo: e.TradeNo,
		PaidAt: e.PaidAt, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}, nil
}

func (r PaymentRepo) UpdateStatus(id string, status order.PaymentStatus, tradeNo string, paidAt int64) error {
	upd := r.client.Payment.UpdateOneID(id).SetStatus(string(status)).SetUpdatedAt(nowSec())
	if tradeNo != "" {
		upd.SetTradeNo(tradeNo)
	}
	if paidAt > 0 {
		upd.SetPaidAt(paidAt)
	}
	return upd.Exec(context.Background())
}

func (r PaymentRepo) ListPaidBetween(start, end int64) ([]*order.Payment, error) {
	rows, err := r.client.Payment.Query().
		Where(payment.Status(string(order.PaymentPaid)), payment.PaidAtGTE(start), payment.PaidAtLT(end)).
		All(context.Background())
	if err != nil {
		return nil, err
	}
	return toPayments(rows), nil
}

func (r PaymentRepo) ListRefundedBetween(start, end int64) ([]*order.Payment, error) {
	rows, err := r.client.Payment.Query().
		Where(payment.Status(string(order.PaymentRefunded)), payment.UpdatedAtGTE(start), payment.UpdatedAtLT(end)).
		All(context.Background())
	if err != nil {
		return nil, err
	}
	return toPayments(rows), nil
}

func toPayments(rows []*ent.Payment) []*order.Payment {
	out := make([]*order.Payment, 0, len(rows))
	for _, e := range rows {
		out = append(out, &order.Payment{
			ID: e.ID, OrderID: e.OrderID, UserID: e.UserID, Channel: e.Channel,
			Amount: e.Amount, Status: order.PaymentStatus(e.Status), TradeNo: e.TradeNo,
			PaidAt: e.PaidAt, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
		})
	}
	return out
}
