package app

import (
	"context"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
)

// StoreList 商家端订单列表。
func (s *OrderService) StoreList(ctx context.Context, merchantID, tab, keyword string, page, size int) ([]*order.Order, int64, error) {
	page, size = normalizePageParams(page, size)
	return s.Repos.Orders.List(ctx, order.Query{
		Page: page, PageSize: size, Keyword: keyword,
		Statuses: order.StoreTabStatuses(tab),
	}, tenant.Scope{Role: identity.RoleMerchant, MerchantID: merchantID})
}

// StoreDetail 商家端订单详情。
func (s *OrderService) StoreDetail(ctx context.Context, merchantID, orderID string) (*OrderDetail, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.MerchantID != merchantID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotOwned, "订单不存在")
	}
	events, err := s.Repos.OrderEvents.ListByOrder(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	return &OrderDetail{Order: o, Events: events}, nil
}

// Accept 商家接单。
func (s *OrderService) Accept(ctx context.Context, merchantID, orderID, operator string) (*order.Order, error) {
	o, err := s.storeOrder(ctx, merchantID, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != order.OrderPending {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可接单")
	}
	now := time.Now().Unix()
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.OrderPending, order.StatusPreparing,
			order.StatusPatch{AcceptedAt: now}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单已被处理")
		}
		if err := s.record(ctx, tx, o, order.OrderPending, order.StatusPreparing, order.ActorStore, operator, "商家接单", now); err != nil {
			return err
		}
		s.notify(ctx, tx, o.UserID, "商家已接单", "订单 "+o.OrderNo+" 商家已接单，正在备餐", "order", o.ID, now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.load(ctx, orderID)
}

// Reject 商家拒单（全额退款 + 释放库存）。
func (s *OrderService) Reject(ctx context.Context, merchantID, orderID, reason, operator string) (*order.Order, error) {
	o, err := s.storeOrder(ctx, merchantID, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != order.OrderPending {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可拒单")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "商家拒单"
	}
	return s.cancel(ctx, o, order.ActorStore, operator, "商家拒单："+reason)
}

// Ready 商家出餐。
func (s *OrderService) Ready(ctx context.Context, merchantID, orderID, operator string) (*order.Order, error) {
	o, err := s.storeOrder(ctx, merchantID, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != order.StatusPreparing {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可出餐")
	}
	now := time.Now().Unix()
	pickupCode, deliveryCode := randomCode(), randomCode()
	for deliveryCode == pickupCode {
		deliveryCode = randomCode()
	}
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.StatusPreparing, order.StatusReady,
			order.StatusPatch{
				ReadyAt: now, PickupCode: pickupCode, DeliveryCode: deliveryCode,
			}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单已被处理")
		}
		return s.record(ctx, tx, o, order.StatusPreparing, order.StatusReady, order.ActorStore, operator, "出餐完成", now)
	})
	if err != nil {
		return nil, err
	}
	return s.load(ctx, orderID)
}

func (s *OrderService) storeOrder(ctx context.Context, merchantID, orderID string) (*order.Order, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.MerchantID != merchantID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotOwned, "订单不存在")
	}
	return o, nil
}

// StoreActions 商家端可执行动作。
func StoreActions(o *order.Order) []string {
	switch o.Status {
	case order.OrderPending:
		return []string{"accept", "reject"}
	case order.StatusPreparing:
		return []string{"ready"}
	case order.StatusReady, order.OrderDelivering:
		return []string{"detail"}
	}
	return nil
}
