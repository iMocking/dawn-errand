// Package admin 管理端业务应用服务。
package admin

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/complaint"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
)

// DashboardData 数据总览原始数据（DTO 层格式化）。
type DashboardData struct {
	TodayOrders       int64
	TodayGMV          float64
	YesterdayOrders   int64
	YesterdayGMV      float64
	OnlineRiders      int64
	TotalRiders       int64
	Merchants         int64
	MerchantsWeekNew  int64
	OnTimeRate        float64
	Trend             []order.DayTrend
	Distribution      map[order.OrderType]int64
	RecentOrders      []*order.Order
	TodoMerchantAudit int64
	TodoRiderAudit    int64
	TodoWithdrawAudit int64
	TodoComplaints    int64
}

// DashboardService 数据总览。
type DashboardService struct {
	Repos *application.RepoUnit
}

// Get 获取总览数据（scope 强制隔离）。
func (s *DashboardService) Get(ctx context.Context, sc tenant.Scope) (*DashboardData, error) {
	now := time.Now().Unix()
	os, err := s.Repos.Orders.Stats(sc, now)
	if err != nil {
		return nil, err
	}
	rs, err := s.Repos.Riders.Stats(sc)
	if err != nil {
		return nil, err
	}
	ms, err := s.Repos.Merchants.Stats(sc)
	if err != nil {
		return nil, err
	}
	trend, err := s.Repos.Orders.Trend7d(sc, now)
	if err != nil {
		return nil, err
	}
	dist, err := s.Repos.Orders.CountByType(sc)
	if err != nil {
		return nil, err
	}
	recent, _, err := s.Repos.Orders.List(order.Query{Page: 1, PageSize: 8}, sc)
	if err != nil {
		return nil, err
	}
	wd, err := s.Repos.Withdrawals.Stats(sc, now)
	if err != nil {
		return nil, err
	}
	_, pendingComplaints, err := s.Repos.Complaints.List(complaint.Query{
		Page: 1, PageSize: 1, Status: "待处理",
	}, sc)
	if err != nil {
		return nil, err
	}

	return &DashboardData{
		TodayOrders: os.TodayOrders, TodayGMV: os.TodayGMV,
		YesterdayOrders: os.YesterdayOrders, YesterdayGMV: os.YesterdayGMV,
		OnlineRiders: rs.Online, TotalRiders: rs.Total,
		Merchants: ms.Total, MerchantsWeekNew: ms.WeekNew,
		OnTimeRate: os.OnTimeRate, Trend: trend, Distribution: dist,
		RecentOrders:      recent,
		TodoMerchantAudit: ms.Pending, TodoRiderAudit: rs.PendingAudit,
		TodoWithdrawAudit: wd.WithdrawPending, TodoComplaints: pendingComplaints,
	}, nil
}

// OrdersService 订单。
type OrdersService struct {
	Repos *application.RepoUnit
}

// List 订单列表。
func (s *OrdersService) List(ctx context.Context, q order.Query, sc tenant.Scope) ([]*order.Order, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 200 {
		q.PageSize = 10
	}
	return s.Repos.Orders.List(q, sc)
}

// CountByRegion 区域今日订单数（settings 页统计用）。
func (s *OrdersService) CountByRegion(ctx context.Context, regionID string, now int64) (int64, error) {
	return s.Repos.Orders.CountByRegion(regionID, now)
}

// Events 订单事件时间线（后台详情用）。
func (s *OrdersService) Events(ctx context.Context, id string) ([]*order.Event, error) {
	return s.Repos.OrderEvents.ListByOrder(id)
}

// Detail 订单详情（对象级数据隔离）。
func (s *OrdersService) Detail(ctx context.Context, id string, sc tenant.Scope) (*order.Order, error) {
	o, err := s.Repos.Orders.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
		}
		return nil, err
	}
	if !orderAccessOK(sc, o) {
		return nil, errcode.NewStatus(403, errcode.CodeTenantScopeDenied, "无权访问该订单")
	}
	return o, nil
}

func orderAccessOK(sc tenant.Scope, o *order.Order) bool {
	switch {
	case sc.IsGlobal():
		return true
	case sc.RegionOnly():
		return o.RegionID == sc.RegionID
	case sc.MerchantOnly():
		return o.MerchantID == sc.MerchantID
	case sc.RiderOnly():
		return o.RiderID == sc.RiderID
	case sc.UserOnly():
		return o.UserID == sc.UserID
	}
	// 默认拒绝：未识别角色不得放行（此前 return true 等于全量可见）。
	return false
}

// merchantAccessOK 商家对象级隔离。
func merchantAccessOK(sc tenant.Scope, m *merchant.Merchant) bool {
	switch {
	case sc.IsGlobal():
		return true
	case sc.RegionOnly():
		return m.RegionID == sc.RegionID
	case sc.MerchantOnly():
		return m.ID == sc.MerchantID
	}
	return false
}

// riderAccessOK 骑手对象级隔离。
func riderAccessOK(sc tenant.Scope, r *rider.Rider) bool {
	switch {
	case sc.IsGlobal():
		return true
	case sc.RegionOnly():
		return r.RegionID == sc.RegionID
	case sc.RiderOnly():
		return r.ID == sc.RiderID
	}
	return false
}

// tenantDenied 统一的租户越权错误。
func tenantDenied() error {
	return errcode.NewStatus(403, errcode.CodeTenantScopeDenied, "无权访问该资源")
}
