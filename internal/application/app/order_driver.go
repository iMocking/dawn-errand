package app

import (
	"context"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
)

// DriverList 骑手端我的订单。
func (s *OrderService) DriverList(ctx context.Context, riderID, tab string, page, size int) ([]*order.Order, int64, error) {
	page, size = normalizePageParams(page, size)
	return s.Repos.Orders.List(ctx, order.Query{
		Page: page, PageSize: size, Statuses: order.DriverTabStatuses(tab),
	}, tenant.Scope{Role: identity.RoleDriver, RiderID: riderID})
}

// DriverDetail 骑手端订单详情。
func (s *OrderService) DriverDetail(ctx context.Context, riderID, orderID string) (*OrderDetail, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.RiderID != riderID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
	}
	events, err := s.Repos.OrderEvents.ListByOrder(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	return &OrderDetail{Order: o, Events: events}, nil
}

// Pool 抢单池（按骑手位置计算取货距离）。
func (s *OrderService) Pool(ctx context.Context, riderID string, lat, lng float64, page, size int) ([]*PoolItem, int64, error) {
	r, err := s.Repos.Riders.FindByID(ctx, riderID)
	if err != nil {
		return nil, 0, errcode.NewStatus(404, errcode.CodeRiderNotFound, "骑手不存在")
	}
	if r.Status == "frozen" {
		return nil, 0, errcode.New(errcode.CodeRiderOffline, "账号已冻结")
	}
	page, size = normalizePageParams(page, size)
	rows, total, err := s.Repos.Orders.ListPool(ctx, r.RegionID, page, size)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*PoolItem, 0, len(rows))
	for _, o := range rows {
		item := &PoolItem{Order: o}
		if validCoord(lat, lng) && validCoord(o.PickupLat, o.PickupLng) {
			item.DistanceKM = round2(haversine(lat, lng, o.PickupLat, o.PickupLng))
			item.DeliveryMinutes = estimateMinutes(item.DistanceKM)
		}
		out = append(out, item)
	}
	return out, total, nil
}

// SetDriverStatus 骑手上线/下线。
func (s *OrderService) SetDriverStatus(ctx context.Context, riderID string, online bool, operator string) (*rider.Rider, error) {
	r, err := s.Repos.Riders.FindByID(ctx, riderID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeRiderNotFound, "骑手不存在")
	}
	if r.AuditStatus != rider.AuditApproved {
		return nil, errcode.New(errcode.CodeRiderStatus, "骑手账号尚未通过审核")
	}
	if r.Status == rider.RiderFrozen {
		return nil, errcode.New(errcode.CodeRiderStatus, "账号已冻结，请联系平台")
	}
	status := rider.RiderOffline
	if online {
		status = rider.RiderOnline
	}
	if err := s.Repos.Riders.UpdateStatus(ctx, riderID, status, time.Now().Unix()); err != nil {
		return nil, err
	}
	r.Status = status
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "driver.status", Object: riderID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx), Operator: operator,
	})
	return r, nil
}

// Grab 抢单：单条条件更新保证并发唯一。
func (s *OrderService) Grab(ctx context.Context, riderID, orderID string) (*order.Order, error) {
	r, err := s.Repos.Riders.FindByID(ctx, riderID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeRiderNotFound, "骑手不存在")
	}
	if r.Status != "online" && r.Status != "busy" {
		return nil, errcode.New(errcode.CodeRiderOffline, "请先上线再抢单")
	}
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.RegionID != r.RegionID {
		return nil, errcode.New(errcode.CodeRiderOutOfRange, "订单不在您的服务区域")
	}
	now := time.Now().Unix()
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		ok, err := tx.Orders.AssignRider(ctx, o.ID, riderID, r.Name, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.NewStatus(409, errcode.CodeOrderGrabbed, "手慢了，订单已被抢")
		}
		if err := s.record(ctx, tx, o, order.StatusReady, order.StatusReady, order.ActorDriver, riderID, "骑手抢单", now); err != nil {
			return err
		}
		s.notify(ctx, tx, o.UserID, "骑手已接单", "订单 "+o.OrderNo+" 骑手 "+r.Name+" 已接单", "order", o.ID, now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.load(ctx, orderID)
}

// Pickup 骑手取货。
func (s *OrderService) Pickup(ctx context.Context, riderID, orderID, code string) (*order.Order, error) {
	o, err := s.driverOrder(ctx, riderID, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != order.StatusReady {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可取货")
	}
	if o.PickupCode != "" && strings.TrimSpace(code) != o.PickupCode {
		return nil, errcode.New(errcode.CodeDeliveryCodeFail, "取货码不正确")
	}
	now := time.Now().Unix()
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.StatusReady, order.OrderDelivering,
			order.StatusPatch{PickedAt: now}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单已被处理")
		}
		return s.record(ctx, tx, o, order.StatusReady, order.OrderDelivering, order.ActorDriver, riderID, "骑手取货", now)
	})
	if err != nil {
		return nil, err
	}
	return s.load(ctx, orderID)
}

// Deliver 骑手送达。
func (s *OrderService) Deliver(ctx context.Context, riderID, orderID, code string) (*order.Order, error) {
	o, err := s.driverOrder(ctx, riderID, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != order.OrderDelivering {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可送达")
	}
	if o.DeliveryCode != "" && strings.TrimSpace(code) != o.DeliveryCode {
		return nil, errcode.New(errcode.CodeDeliveryCodeFail, "送达码不正确")
	}
	now := time.Now().Unix()
	earning := round2(o.DeliveryFee * riderShareRate)
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.OrderDelivering, order.StatusDelivered,
			order.StatusPatch{DeliveredAt: now, RiderEarning: &earning}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单已被处理")
		}
		if err := s.record(ctx, tx, o, order.OrderDelivering, order.StatusDelivered, order.ActorDriver, riderID, "骑手送达", now); err != nil {
			return err
		}
		if err := creditWallet(ctx, tx, "rider", riderID, earning, "delivery_income", o.ID, "配送收入", now); err != nil {
			return err
		}
		if r, err := tx.Riders.FindByID(ctx, riderID); err == nil {
			_ = tx.Settlements.Save(ctx, &finance.Settlement{
				RegionID: o.RegionID, Party: r.Name, PartyType: "rider", PartyID: riderID,
				Type: finance.SettlementRider, Amount: earning, Status: finance.SettlementArrived,
				Period: time.Unix(now, 0).UTC().Format("2006-01-02"), SettledAt: now,
				Action: "配送完成自动入账",
			})
		}
		s.notify(ctx, tx, o.UserID, "订单已送达", "订单 "+o.OrderNo+" 已送达，请确认收货", "order", o.ID, now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.load(ctx, orderID)
}

func (s *OrderService) driverOrder(ctx context.Context, riderID, orderID string) (*order.Order, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.RiderID != riderID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
	}
	return o, nil
}

// DriverActions 骑手端可执行动作。
func DriverActions(o *order.Order, riderID string) []string {
	switch o.Status {
	case order.StatusReady:
		if o.RiderID == "" {
			return []string{"grab"}
		}
		if o.RiderID == riderID {
			return []string{"pickup"}
		}
	case order.OrderDelivering:
		if o.RiderID == riderID {
			return []string{"deliver"}
		}
	}
	return nil
}
