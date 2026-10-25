package admin

import (
	"context"
	"sort"
	"time"

	"dawn-errand/internal/domain/catalog"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/review"
	"dawn-errand/internal/domain/tenant"
)

// MerchantWorkspace 商家详情各 Tab 所需的真实数据。
type MerchantWorkspace struct {
	Merchant   *merchant.Merchant
	Orders     []*order.Order
	Abnormal   []*order.Order
	Categories []*catalog.Category
	Dishes     []*catalog.Dish
	Reviews    []*review.Review
	UserNames  map[string]string
	Trend      []MerchantTrendPoint
}

// MerchantTrendPoint 近 7 日经营趋势。
type MerchantTrendPoint struct {
	Date   string  `json:"date"`
	Orders int64   `json:"orders"`
	Amount float64 `json:"amount"`
}

// Workspace 聚合商家详情页「证照/经营/菜单/评价/异常」数据。
func (s *MerchantsService) Workspace(ctx context.Context, id string, sc tenant.Scope) (*MerchantWorkspace, error) {
	m, err := s.Detail(ctx, id, sc)
	if err != nil {
		return nil, err
	}
	ws := &MerchantWorkspace{Merchant: m, UserNames: map[string]string{}}

	cats, err := s.Repos.Catalog.ListCategories(id)
	if err != nil {
		return nil, err
	}
	dishes, _, err := s.Repos.Catalog.ListDishes(catalog.DishQuery{
		MerchantID: id, IncludeOffSale: true, Page: 1, PageSize: 200,
	})
	if err != nil {
		return nil, err
	}
	reviews, _, err := s.Repos.Reviews.List(review.Query{MerchantID: id, Page: 1, PageSize: 50})
	if err != nil {
		return nil, err
	}
	orders, err := s.Repos.Orders.ListByMerchant(id, 200)
	if err != nil {
		return nil, err
	}
	ws.Categories, ws.Dishes, ws.Reviews, ws.Orders = cats, dishes, reviews, orders

	for _, rv := range reviews {
		if _, ok := ws.UserNames[rv.UserID]; ok {
			continue
		}
		if u, err := s.Repos.Users.FindByID(rv.UserID); err == nil {
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
	return ws, nil
}

func merchantTrend(orders []*order.Order, now time.Time) []MerchantTrendPoint {
	type day struct {
		orders int64
		amount float64
	}
	buckets := make(map[string]*day, 7)
	labels := make([]string, 0, 7)
	for i := 6; i >= 0; i-- {
		key := now.AddDate(0, 0, -i).Format("2006-01-02")
		labels = append(labels, key)
		buckets[key] = &day{}
	}
	for _, o := range orders {
		key := time.Unix(o.CreatedAt, 0).Local().Format("2006-01-02")
		b := buckets[key]
		if b == nil {
			continue
		}
		b.orders++
		b.amount += o.PayAmount
	}
	out := make([]MerchantTrendPoint, 0, len(labels))
	for _, key := range labels {
		out = append(out, MerchantTrendPoint{
			Date: key, Orders: buckets[key].orders, Amount: round2Money(buckets[key].amount),
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out
}
