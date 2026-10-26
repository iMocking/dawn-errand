package repo

import (
	"context"

	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/infrastructure/ent"
	entorder "dawn-errand/internal/infrastructure/ent/order"
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

// SummarizeRecon 聚合支付时间窗口内的对账口径。
// 订单侧用分片批量查询替代逐单 FindByID（原实现每笔支付一次查询，属 N+1）；
// regionID 非空时只统计该区域的订单，跨区数据不计入，实现区域隔离。
func (r PaymentRepo) SummarizeRecon(start, end int64, regionID string) (*order.ReconSummary, error) {
	ctx := context.Background()
	res := &order.ReconSummary{}

	paid, err := r.client.Payment.Query().
		Where(payment.Status(string(order.PaymentPaid)), payment.PaidAtGTE(start), payment.PaidAtLT(end)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(paid))
	paidOrderIDs := make([]string, 0, len(paid))
	for _, p := range paid {
		res.PaidCount++
		res.PaidAmount += p.Amount
		if !seen[p.OrderID] {
			seen[p.OrderID] = true
			paidOrderIDs = append(paidOrderIDs, p.OrderID)
		}
	}
	paidAmountByOrder, err := r.orderPayAmounts(ctx, paidOrderIDs, regionID)
	if err != nil {
		return nil, err
	}
	for _, p := range paid {
		amt, ok := paidAmountByOrder[p.OrderID]
		if !ok {
			res.MissingOrders++
			continue
		}
		res.PaidOrderCount++
		res.OrderAmount += amt
	}

	refunded, err := r.client.Payment.Query().
		Where(payment.Status(string(order.PaymentRefunded)), payment.UpdatedAtGTE(start), payment.UpdatedAtLT(end)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	refundOrderIDs := make([]string, 0, len(refunded))
	for _, p := range refunded {
		refundOrderIDs = append(refundOrderIDs, p.OrderID)
	}
	refundVisible, err := r.orderPayAmounts(ctx, refundOrderIDs, regionID)
	if err != nil {
		return nil, err
	}
	for _, p := range refunded {
		if regionID != "" {
			if _, ok := refundVisible[p.OrderID]; !ok {
				continue
			}
		}
		res.RefundCount++
		res.RefundAmount += p.Amount
	}
	return res, nil
}

// orderPayAmounts 批量读取订单应付金额；regionID 非空时按区域过滤。
func (r PaymentRepo) orderPayAmounts(ctx context.Context, ids []string, regionID string) (map[string]float64, error) {
	out := make(map[string]float64, len(ids))
	const chunk = 500
	for i := 0; i < len(ids); i += chunk {
		stop := min(i+chunk, len(ids))
		q := r.client.Order.Query().Where(entorder.IDIn(ids[i:stop]...))
		if regionID != "" {
			q = q.Where(entorder.RegionID(regionID))
		}
		rows, err := q.All(ctx)
		if err != nil {
			return nil, err
		}
		for _, o := range rows {
			out[o.ID] = o.PayAmount
		}
	}
	return out, nil
}
