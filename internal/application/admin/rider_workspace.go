package admin

import (
	"context"
	"sort"
	"time"

	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/review"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
)

// RiderWorkspace 骑手详情各 Tab 数据。
type RiderWorkspace struct {
	Rider     *rider.Rider
	Orders    []*order.Order
	Abnormal  []*order.Order
	Reviews   []*review.Review
	UserNames map[string]string
	Trend     []MerchantTrendPoint
	Records   []RiderRecord
}

// RiderRecord 奖惩/操作记录（由审核、异常订单、最近操作聚合）。
type RiderRecord struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Tone   string `json:"tone"` // success / warning / info
	Time   int64  `json:"time"`
}

// Workspace 聚合骑手详情页「配送统计/绩效/证件/奖惩」数据。
func (s *RidersService) Workspace(ctx context.Context, id string, sc tenant.Scope) (*RiderWorkspace, error) {
	r, err := s.Detail(ctx, id, sc)
	if err != nil {
		return nil, err
	}
	ws := &RiderWorkspace{Rider: r, UserNames: map[string]string{}}
	orders, err := s.Repos.Orders.ListByRider(ctx, id, 200)
	if err != nil {
		return nil, err
	}
	reviews, _, err := s.Repos.Reviews.List(ctx, review.Query{RiderID: id, Page: 1, PageSize: 50})
	if err != nil {
		return nil, err
	}
	ws.Orders, ws.Reviews = orders, reviews
	for _, rv := range reviews {
		if _, ok := ws.UserNames[rv.UserID]; ok {
			continue
		}
		if u, err := s.Repos.Users.FindByID(ctx, rv.UserID); err == nil {
			ws.UserNames[rv.UserID] = u.Name
		}
	}
	for _, o := range orders {
		switch o.Status {
		case order.OrderAbnormal, order.OrderCookingOot, order.OrderCanceled,
			order.StatusAfterSale, order.StatusRefunded:
			ws.Abnormal = append(ws.Abnormal, o)
		}
	}
	ws.Trend = merchantTrend(orders, time.Now())

	if r.AuditStatus == rider.AuditApproved {
		ws.Records = append(ws.Records, RiderRecord{
			Title: "入驻审核通过", Detail: "认证通过，骑手端账号生效", Tone: "success", Time: r.JoinedAt,
		})
	}
	for _, o := range ws.Abnormal {
		ws.Records = append(ws.Records, RiderRecord{
			Title: "订单" + string(o.Status), Detail: o.OrderNo + " · " + o.Route,
			Tone: "warning", Time: o.CreatedAt,
		})
	}
	if r.LastAction != "" {
		ws.Records = append(ws.Records, RiderRecord{
			Title: r.LastAction, Detail: "最近操作", Tone: "info", Time: r.LastActionAt,
		})
	}
	sort.SliceStable(ws.Records, func(i, j int) bool { return ws.Records[i].Time > ws.Records[j].Time })
	return ws, nil
}
