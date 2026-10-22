package repo

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/infrastructure/ent"
	entorder "dawn-errand/internal/infrastructure/ent/order"
)

// OrderRepo 实现 order.OrderRepository。
type OrderRepo struct{ client *ent.Client }

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func (r OrderRepo) Save(o *order.Order) error {
	ctx := context.Background()
	if o.ID == "" {
		o.ID = newULID()
	}
	err := r.client.Order.Create().
		SetID(o.ID).SetOrderNo(o.OrderNo).SetRegionID(o.RegionID).
		SetMerchantID(o.MerchantID).SetUserID(o.UserID).SetRiderID(o.RiderID).
		SetType(string(o.Type)).SetStatus(string(o.Status)).SetRoute(o.Route).
		SetUserName(o.UserName).SetRiderName(o.RiderName).SetAmount(o.Amount).
		SetCreatedAt(o.CreatedAt).SetChannel(o.Channel).SetEstimatedArrival(o.EstimatedArrival).
		SetRemark(o.Remark).
		SetGoods(mustJSON(o.Goods)).SetFees(mustJSON(o.Fees)).SetSteps(mustJSON(o.Steps)).
		SetMerchant(mustJSON(o.Merchant)).SetUser(mustJSON(o.User)).SetRider(mustJSON(o.Rider)).
		SetPayAmount(o.PayAmount).SetPayMethod(o.PayMethod).SetFundFlow(o.FundFlow).
		Exec(ctx)
	return err
}

func toOrder(e *ent.Order) *order.Order {
	if e == nil {
		return nil
	}
	o := &order.Order{
		ID: e.ID, OrderNo: e.OrderNo, RegionID: e.RegionID,
		MerchantID: e.MerchantID, UserID: e.UserID, RiderID: e.RiderID,
		Type: order.OrderType(e.Type), Status: order.OrderStatus(e.Status),
		Route: e.Route, UserName: e.UserName, RiderName: e.RiderName, Amount: e.Amount,
		CreatedAt: e.CreatedAt, Channel: e.Channel, EstimatedArrival: e.EstimatedArrival,
		Remark: e.Remark, PayAmount: e.PayAmount, PayMethod: e.PayMethod, FundFlow: e.FundFlow,
	}
	_ = json.Unmarshal(e.Goods, &o.Goods)
	_ = json.Unmarshal(e.Fees, &o.Fees)
	_ = json.Unmarshal(e.Steps, &o.Steps)
	_ = json.Unmarshal(e.Merchant, &o.Merchant)
	_ = json.Unmarshal(e.User, &o.User)
	_ = json.Unmarshal(e.Rider, &o.Rider)
	return o
}

func (r OrderRepo) FindByID(id string) (*order.Order, error) {
	e, err := r.client.Order.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toOrder(e), nil
}

func (r OrderRepo) List(q order.Query, sc tenant.Scope) ([]*order.Order, int64, error) {
	ctx := context.Background()
	bq := r.client.Order.Query()
	scopeOrder(bq, sc)
	if q.UserID != "" {
		bq.Where(entorder.UserID(q.UserID))
	}
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entorder.Status(s))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(entorder.Or(
			entorder.OrderNoContainsFold(k),
			entorder.UserNameContainsFold(k),
			entorder.RiderNameContainsFold(k),
			entorder.RouteContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entorder.FieldCreatedAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*order.Order, 0, len(rows))
	for _, e := range rows {
		out = append(out, toOrder(e))
	}
	return out, int64(total), nil
}

func (r OrderRepo) ListByUser(userID string, limit int) ([]*order.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.client.Order.Query().
		Where(entorder.UserID(userID)).
		Order(ent.Desc(entorder.FieldCreatedAt)).
		Limit(limit).
		All(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]*order.Order, 0, len(rows))
	for _, e := range rows {
		out = append(out, toOrder(e))
	}
	return out, nil
}

func (r OrderRepo) Stats(sc tenant.Scope, now int64) (*order.Stats, error) {
	ctx := context.Background()
	dayStart := startOfDayUTC(now)
	base := func() *ent.OrderQuery {
		bq := r.client.Order.Query()
		scopeOrder(bq, sc)
		return bq
	}
	// 今日
	today := base().Where(entorder.CreatedAtGTE(dayStart))
	todayCount, err := today.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	todayGMV, err := sumFloat64(ctx, today.Clone(), entorder.FieldPayAmount)
	if err != nil {
		return nil, err
	}
	// 昨日
	yesterdayStart := dayStart - 86400
	yesterday := base().Where(entorder.CreatedAtGTE(yesterdayStart), entorder.CreatedAtLT(dayStart))
	yCount, err := yesterday.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	yGMV, err := sumFloat64(ctx, yesterday.Clone(), entorder.FieldPayAmount)
	if err != nil {
		return nil, err
	}
	// 准时率（履约率）：已完成 / 非取消
	nonCanceled := base().Where(entorder.StatusNEQ(string(order.OrderCanceled)))
	done, err := base().Where(entorder.Status(string(order.OrderDone))).Count(ctx)
	if err != nil {
		return nil, err
	}
	nc, err := nonCanceled.Count(ctx)
	if err != nil {
		return nil, err
	}
	rate := 0.0
	if nc > 0 {
		rate = float64(done) / float64(nc) * 100
	}
	return &order.Stats{
		TodayOrders: int64(todayCount), TodayGMV: todayGMV,
		YesterdayOrders: int64(yCount), YesterdayGMV: yGMV, OnTimeRate: rate,
	}, nil
}

func (r OrderRepo) Trend7d(sc tenant.Scope, now int64) ([]order.DayTrend, error) {
	ctx := context.Background()
	start := daysAgoStart(now, 6)
	bq := r.client.Order.Query()
	scopeOrder(bq, sc)
	rows, err := bq.Where(entorder.CreatedAtGTE(start)).
		Select(entorder.FieldCreatedAt, entorder.FieldPayAmount).
		All(ctx)
	if err != nil {
		return nil, err
	}
	byDay := map[string]*order.DayTrend{}
	for i := 6; i >= 0; i-- {
		t := time.Unix(start+int64(i)*86400, 0).UTC()
		key := t.Format("01/02")
		byDay[key] = &order.DayTrend{Day: key}
	}
	for _, e := range rows {
		t := time.Unix(e.CreatedAt, 0).UTC()
		key := t.Format("01/02")
		if d, ok := byDay[key]; ok {
			d.Orders++
			d.GMV += e.PayAmount
		}
	}
	out := make([]order.DayTrend, 0, 7)
	for i := 6; i >= 0; i-- {
		t := time.Unix(start+int64(i)*86400, 0).UTC()
		key := t.Format("01/02")
		out = append(out, *byDay[key])
	}
	return out, nil
}

func (r OrderRepo) CountByRegion(regionID string, now int64) (int64, error) {
	n, err := r.client.Order.Query().
		Where(entorder.RegionID(regionID), entorder.CreatedAtGTE(startOfDayUTC(now))).
		Count(context.Background())
	return int64(n), err
}

func (r OrderRepo) CountToday(now int64) (int64, error) {
	n, err := r.client.Order.Query().
		Where(entorder.CreatedAtGTE(startOfDayUTC(now))).
		Count(context.Background())
	return int64(n), err
}

func (r OrderRepo) SumTodayGMV(now int64) (float64, error) {
	return sumFloat64(context.Background(), r.client.Order.Query().
		Where(entorder.CreatedAtGTE(startOfDayUTC(now))), entorder.FieldPayAmount)
}

func (r OrderRepo) CountUserOrders(userID string) (int64, error) {
	n, err := r.client.Order.Query().Where(entorder.UserID(userID)).Count(context.Background())
	return int64(n), err
}

func (r OrderRepo) CountUserMonth(userID string, now int64) (int64, error) {
	n, err := r.client.Order.Query().
		Where(entorder.UserID(userID), entorder.CreatedAtGTE(daysAgoStart(now, 30))).
		Count(context.Background())
	return int64(n), err
}

func (r OrderRepo) CountUserSpent(userID string) (float64, error) {
	return sumFloat64(context.Background(), r.client.Order.Query().
		Where(entorder.UserID(userID)), entorder.FieldPayAmount)
}

func (r OrderRepo) CountActiveMonth(sc tenant.Scope, now int64) (int64, error) {
	ctx := context.Background()
	bq := r.client.Order.Query()
	scopeOrder(bq, sc)
	var rows []struct {
		UserID string `sql:"user_id"`
		Count  int
	}
	if err := bq.Where(entorder.CreatedAtGTE(daysAgoStart(now, 30))).
		GroupBy(entorder.FieldUserID).Aggregate(ent.Count()).Scan(ctx, &rows); err != nil {
		return 0, err
	}
	return int64(len(rows)), nil
}

func (r OrderRepo) CountRecentActiveUsers(now int64) (int64, error) {
	var rows []struct {
		UserID string `sql:"user_id"`
		Count  int
	}
	if err := r.client.Order.Query().
		Where(entorder.CreatedAtGTE(daysAgoStart(now, 7))).
		GroupBy(entorder.FieldUserID).Aggregate(ent.Count()).Scan(context.Background(), &rows); err != nil {
		return 0, err
	}
	return int64(len(rows)), nil
}

func (r OrderRepo) CountByType(sc tenant.Scope) (map[order.OrderType]int64, error) {
	ctx := context.Background()
	bq := r.client.Order.Query()
	scopeOrder(bq, sc)
	var rows []struct {
		Type  string
		Count int
	}
	if err := bq.GroupBy(entorder.FieldType).Aggregate(ent.Count()).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := map[order.OrderType]int64{}
	for _, row := range rows {
		out[order.OrderType(row.Type)] = int64(row.Count)
	}
	return out, nil
}

func (r OrderRepo) CountAll(sc tenant.Scope) (int64, error) {
	bq := r.client.Order.Query()
	scopeOrder(bq, sc)
	n, err := bq.Count(context.Background())
	return int64(n), err
}
