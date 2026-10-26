package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/metrics"
	"dawn-errand/internal/pkg/ulidgen"
)

// ClientList 用户端订单列表。
func (s *OrderService) ClientList(ctx context.Context, userID, tab, keyword string, page, size int) ([]*order.Order, int64, error) {
	page, size = normalizePageParams(page, size)
	return s.Repos.Orders.List(ctx, order.Query{
		Page: page, PageSize: size, Keyword: keyword,
		Statuses: order.ClientTabStatuses(tab),
	}, tenant.Scope{Role: identity.RoleUser, UserID: userID})
}

// ClientDetail 用户端订单详情。
func (s *OrderService) ClientDetail(ctx context.Context, userID, orderID string) (*OrderDetail, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.UserID != userID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
	}
	events, err := s.Repos.OrderEvents.ListByOrder(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	return &OrderDetail{Order: o, Events: events}, nil
}

// CancelClient 用户取消订单（待支付/待接单）。
func (s *OrderService) CancelClient(ctx context.Context, userID, orderID, reason string) (*order.Order, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.UserID != userID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
	}
	if o.Status != order.StatusPendingPay && o.Status != order.OrderPending {
		return nil, errcode.New(errcode.CodeOrderState, "当前状态不可取消")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "用户取消"
	}
	return s.cancel(ctx, o, order.ActorClient, userID, reason)
}

// Pay 支付订单（mock/余额）。
func (s *OrderService) Pay(ctx context.Context, userID, orderID, channel string) (*order.Order, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.UserID != userID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
	}
	if o.PaymentStatus == "paid" {
		return o, nil
	}
	if o.Status != order.StatusPendingPay {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可支付")
	}
	channel = strings.TrimSpace(channel)
	switch channel {
	case "balance":
	case "wechat", "alipay", "mock":
		if !s.MockPayment {
			return nil, errcode.New(errcode.CodePaymentFailed, "支付渠道尚未接入，请等待渠道回调")
		}
	default:
		return nil, errcode.New(errcode.CodeBadRequest, "不支持的支付方式")
	}
	now := time.Now().Unix()
	tradeNo := "MOCK" + strings.ToUpper(ulidgen.New())[:20]
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if channel == "balance" {
			if err := deductBalance(ctx, tx, userID, o.PayAmount, now); err != nil {
				return err
			}
		}
		if err := tx.Payments.Save(ctx, &order.Payment{
			OrderID: o.ID, UserID: userID, Channel: channel, Amount: o.PayAmount,
			Status: order.PaymentPaid, TradeNo: tradeNo, PaidAt: now,
		}); err != nil {
			return err
		}
		patch := order.StatusPatch{
			PaymentStatus: "paid", PayMethod: channel,
			FundFlow: fmt.Sprintf("%s 支付 ¥%.2f", channel, o.PayAmount),
		}
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.StatusPendingPay, order.OrderPending, patch, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单状态已变化，请刷新后重试")
		}
		if err := s.record(ctx, tx, o, order.StatusPendingPay, order.OrderPending, order.ActorClient, userID, "支付成功", now); err != nil {
			return err
		}
		s.notifyMerchant(ctx, tx, o, "新订单待接单", "订单 "+o.OrderNo+" 已支付，请及时接单", now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	metrics.OrdersPaidTotal.Inc()
	return s.load(ctx, orderID)
}

// PayByCallback 支付渠道回调确认（幂等；金额必须与服务端应付一致）。
func (s *OrderService) PayByCallback(ctx context.Context, orderID, channel, tradeNo string, amount float64) (*order.Order, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.PaymentStatus == "paid" {
		return o, nil
	}
	if o.Status != order.StatusPendingPay {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可支付")
	}
	if round2(amount) != round2(o.PayAmount) {
		return nil, errcode.New(errcode.CodeAmountMismatch, "回调金额与订单应付不一致")
	}
	now := time.Now().Unix()
	if tradeNo == "" {
		tradeNo = "CB" + strings.ToUpper(ulidgen.New())[:20]
	}
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if err := tx.Payments.Save(ctx, &order.Payment{
			OrderID: o.ID, UserID: o.UserID, Channel: channel, Amount: o.PayAmount,
			Status: order.PaymentPaid, TradeNo: tradeNo, PaidAt: now,
		}); err != nil {
			return err
		}
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.StatusPendingPay, order.OrderPending,
			order.StatusPatch{PaymentStatus: "paid", PayMethod: channel,
				FundFlow: fmt.Sprintf("%s 回调支付 ¥%.2f", channel, o.PayAmount)}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单状态已变化")
		}
		if err := s.record(ctx, tx, o, order.StatusPendingPay, order.OrderPending, order.ActorSystem, "payment", "支付回调成功", now); err != nil {
			return err
		}
		s.notifyMerchant(ctx, tx, o, "新订单待接单", "订单 "+o.OrderNo+" 已支付，请及时接单", now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	metrics.OrdersPaidTotal.Inc()
	return s.load(ctx, orderID)
}

// Confirm 用户确认收货。
func (s *OrderService) Confirm(ctx context.Context, userID, orderID string) (*order.Order, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.UserID != userID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
	}
	if o.Status != order.StatusDelivered {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可确认收货")
	}
	now := time.Now().Unix()
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.StatusDelivered, order.StatusPendingReview,
			order.StatusPatch{ConfirmedAt: now}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单状态已变化，请刷新后重试")
		}
		if err := s.record(ctx, tx, o, order.StatusDelivered, order.StatusPendingReview, order.ActorClient, userID, "确认收货", now); err != nil {
			return err
		}
		s.notifyMerchant(ctx, tx, o, "用户已确认收货", "订单 "+o.OrderNo+" 已确认收货", now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.load(ctx, orderID)
}

// ClientActions 用户端可执行动作（服务端下发）。
func ClientActions(o *order.Order) []string {
	switch o.Status {
	case order.StatusPendingPay:
		return []string{"pay", "cancel"}
	case order.OrderPending:
		return []string{"cancel"}
	case order.StatusPreparing, order.StatusReady, order.OrderDelivering:
		return []string{"track"}
	case order.StatusDelivered:
		return []string{"confirm", "track"}
	case order.StatusPendingReview:
		// 评价接口在 M4 落地，此处暂不下发不可用动作。
		return nil
	}
	return nil
}
