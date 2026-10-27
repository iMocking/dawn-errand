package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/coupon"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/pkg/errcode"
)

// R5-5 订单状态机单测（商家端接单→出餐、骑手端取货→送达）。
//
// integration 用例走真实 HTTP + 数据库，验证的是「链路通」；这一层验证的是
// 「每一步只能靠条件更新推进，且事件/资金/库存与状态同生共死」——条件更新
// 返回 0 行时，事务里不能留下任何半截副作用。替身按仓储语义复刻（见 fakes_test.go）。

// storePendingOrder 造一张已支付、待商家接单的外卖单。
func storePendingOrder(u *unit, id, merchantID string) *order.Order {
	o := &order.Order{
		ID: id, UserID: "U1", MerchantID: merchantID, RegionID: "RG1", OrderNo: "N" + id,
		Type: order.OrderTypeFood, Status: order.OrderPending,
		PaymentStatus: "paid", PayMethod: "balance",
		Amount: 50, PayAmount: 56, DeliveryFee: 6, PackingFee: 2,
		Goods: []order.OrderGoods{{DishID: "D1", Name: "面", Count: 2, Price: 25}},
	}
	u.orders.byID[id] = o
	return o
}

func TestStoreAcceptGuardsAndEffects(t *testing.T) {
	ctx := context.Background()

	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	u.merch.byID["M1"] = &merchant.Merchant{ID: "M1", Name: "小店", OwnerUserID: "MU1"}
	storePendingOrder(u, "O1", "M1")

	// 越权接单：订单归属校验先于状态校验，且回 404 不泄露存在性。
	if code := codeOf(t, mustErr(s.Accept(ctx, "M9", "O1", "op"))); code != errcode.CodeOrderNotOwned {
		t.Errorf("他人订单接单应 404, got=%d", code)
	}
	if len(u.evts.saved) != 0 {
		t.Error("越权接单却落了事件")
	}

	got, err := s.Accept(ctx, "M1", "O1", "op-1")
	if err != nil {
		t.Fatalf("接单失败: %v", err)
	}
	if got.Status != order.StatusPreparing || got.AcceptedAt == 0 {
		t.Errorf("接单后状态/时间错误: %+v", got)
	}
	if len(u.evts.saved) != 1 {
		t.Fatalf("接单事件条数错误: %d", len(u.evts.saved))
	}
	if e := u.evts.saved[0]; e.FromStatus != order.OrderPending || e.ToStatus != order.StatusPreparing ||
		e.ActorType != order.ActorStore || e.ActorID != "op-1" {
		t.Errorf("接单事件内容错误: %+v", e)
	}
	if len(u.noti.saved) != 1 || u.noti.saved[0].UserID != "U1" {
		t.Errorf("接单未通知下单用户: %+v", u.noti.saved)
	}

	// 重复接单：状态已推进，必须拒绝而不是再发一次事件/通知。
	if code := codeOf(t, mustErr(s.Accept(ctx, "M1", "O1", "op-1"))); code != errcode.CodeOrderState {
		t.Errorf("重复接单应报状态不允许, got=%d", code)
	}
	if len(u.evts.saved) != 1 || len(u.noti.saved) != 1 {
		t.Error("重复接单产生了重复事件/通知")
	}

	// 并发抢占：条件更新 0 行时整步失败，且不得追加事件。
	u2 := newUnit()
	s2 := &OrderService{Repos: u2.repos, Tx: u2.tx}
	storePendingOrder(u2, "O2", "M2")
	u2.orders.updateOK = false
	if code := codeOf(t, mustErr(s2.Accept(ctx, "M2", "O2", "op"))); code != errcode.CodeOrderState {
		t.Errorf("并发接单应报订单已被处理, got=%d", code)
	}
	if len(u2.evts.saved) != 0 {
		t.Errorf("并发失败仍落事件: %+v", u2.evts.saved)
	}
	if u2.orders.byID["O2"].Status != order.OrderPending {
		t.Error("并发失败改动了订单状态")
	}
}

func TestStoreReadyIssuesPickupCodes(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	o := storePendingOrder(u, "O1", "M1")
	o.Status = order.StatusPreparing

	got, err := s.Ready(ctx, "M1", "O1", "op-1")
	if err != nil {
		t.Fatalf("出餐失败: %v", err)
	}
	if got.Status != order.StatusReady || got.ReadyAt == 0 {
		t.Errorf("出餐后状态/时间错误: %+v", got)
	}
	if len(got.PickupCode) != 4 || len(got.DeliveryCode) != 4 || got.PickupCode == got.DeliveryCode {
		t.Errorf("取送码应为两个不同的 4 位码: pickup=%q delivery=%q", got.PickupCode, got.DeliveryCode)
	}
	if e := u.evts.saved[len(u.evts.saved)-1]; e.ToStatus != order.StatusReady || e.ActorType != order.ActorStore {
		t.Errorf("出餐事件错误: %+v", e)
	}

	// 非备餐中不可出餐（已出餐的单再点一次）。
	if code := codeOf(t, mustErr(s.Ready(ctx, "M1", "O1", "op-1"))); code != errcode.CodeOrderState {
		t.Errorf("重复出餐应报状态不允许, got=%d", code)
	}
	// 越权。
	if code := codeOf(t, mustErr(s.Ready(ctx, "M9", "O1", "op-1"))); code != errcode.CodeOrderNotOwned {
		t.Errorf("他人订单出餐应 404, got=%d", code)
	}
	// 非备餐中（待接单）不可直接出餐。
	u3 := newUnit()
	s3 := &OrderService{Repos: u3.repos, Tx: u3.tx}
	storePendingOrder(u3, "O3", "M3")
	if code := codeOf(t, mustErr(s3.Ready(ctx, "M3", "O3", "op"))); code != errcode.CodeOrderState {
		t.Errorf("待接单不可直接出餐, got=%d", code)
	}

	// 并发：条件更新失败不得写码、不得落事件。
	u2 := newUnit()
	s2 := &OrderService{Repos: u2.repos, Tx: u2.tx}
	o2 := storePendingOrder(u2, "O2", "M2")
	o2.Status = order.StatusPreparing
	u2.orders.updateOK = false
	if code := codeOf(t, mustErr(s2.Ready(ctx, "M2", "O2", "op"))); code != errcode.CodeOrderState {
		t.Errorf("并发出餐应报订单已被处理, got=%d", code)
	}
	if o2.PickupCode != "" || o2.DeliveryCode != "" {
		t.Errorf("并发出餐仍写入了取送码: %+v", o2)
	}
	if len(u2.evts.saved) != 0 {
		t.Error("并发出餐仍落了事件")
	}
}

func TestStoreRejectRefundsRestocksAndReleasesCoupon(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	o := storePendingOrder(u, "O1", "M1")
	o.CouponID = "C1"
	u.catalog.stocks["D1"] = 3 // 下单扣了 2，现剩 3；拒单应退回 2
	u.prof.balances["U1"] = 0
	u.coupon.byID["C1"] = &coupon.Coupon{ID: "C1", UserID: "U1", Status: coupon.StatusUsed, UsedOrderID: "O1"}
	u.pay.byOrder["O1"] = &order.Payment{ID: "P1", OrderID: "O1", Status: order.PaymentPaid}

	got, err := s.Reject(ctx, "M1", "O1", "  后厨忙不过来  ", "op-1")
	if err != nil {
		t.Fatalf("拒单失败: %v", err)
	}
	if got.Status != order.OrderCanceled {
		t.Errorf("拒单后状态应为已取消, got=%q", got.Status)
	}
	if got.CancelReason != "商家拒单：后厨忙不过来" {
		t.Errorf("拒单原因未裁剪/拼接: %q", got.CancelReason)
	}
	if u.prof.balances["U1"] != 56 {
		t.Errorf("已支付订单拒单未退回余额: %v", u.prof.balances["U1"])
	}
	if got.PaymentStatus != "refunded" {
		t.Errorf("退款状态未回写: %q", got.PaymentStatus)
	}
	if u.catalog.stocks["D1"] != 5 {
		t.Errorf("拒单未释放库存: %v", u.catalog.stocks["D1"])
	}
	if len(u.coupon.released) != 1 || u.coupon.released[0] != "O1" {
		t.Errorf("拒单未释放优惠券: %v", u.coupon.released)
	}
	if len(u.pay.statuses) != 1 || u.pay.statuses[0] != string(order.PaymentRefunded) {
		t.Errorf("支付单未标退款: %v", u.pay.statuses)
	}
	if e := u.evts.saved[len(u.evts.saved)-1]; e.ActorType != order.ActorStore || e.ToStatus != order.OrderCanceled {
		t.Errorf("拒单事件错误: %+v", e)
	}
	// 商家操作方 → 通知用户（不是通知商家自己）。
	if len(u.noti.saved) != 1 || u.noti.saved[0].UserID != "U1" {
		t.Errorf("拒单应通知用户: %+v", u.noti.saved)
	}

	// 空原因回落默认文案。
	u2 := newUnit()
	s2 := &OrderService{Repos: u2.repos, Tx: u2.tx}
	storePendingOrder(u2, "O2", "M2").PaymentStatus = "unpaid"
	got2, err := s2.Reject(ctx, "M2", "O2", "   ", "op")
	if err != nil {
		t.Fatalf("空原因拒单失败: %v", err)
	}
	if got2.CancelReason != "商家拒单：商家拒单" {
		t.Errorf("空原因默认文案错误: %q", got2.CancelReason)
	}
	if got2.PaymentStatus != "unpaid" {
		t.Errorf("未支付单不应被标退款: %q", got2.PaymentStatus)
	}

	// 备餐中不可拒单（只能走售后/取消流程，避免「已开工还全额退」）。
	u3 := newUnit()
	s3 := &OrderService{Repos: u3.repos, Tx: u3.tx}
	storePendingOrder(u3, "O3", "M3").Status = order.StatusPreparing
	if code := codeOf(t, mustErr(s3.Reject(ctx, "M3", "O3", "不做了", "op"))); code != errcode.CodeOrderState {
		t.Errorf("备餐中拒单应报状态不允许, got=%d", code)
	}
	if len(u3.noti.saved) != 0 {
		t.Error("被拒绝的拒单仍发了通知")
	}
}

func TestDriverPickupDeliverSettlesEarning(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	u.riders.byID["D1"] = &rider.Rider{ID: "D1", Name: "飞毛腿", RegionID: "RG1", Status: rider.RiderOnline}
	u.orders.byID["O1"] = &order.Order{
		ID: "O1", UserID: "U1", MerchantID: "M1", RegionID: "RG1", RiderID: "D1", OrderNo: "N1",
		Status: order.StatusReady, DeliveryFee: 6, PickupCode: "1234", DeliveryCode: "5678",
	}

	// 越权：不是自己的单，取货/送达都必须 404。
	if code := codeOf(t, mustErr(s.Pickup(ctx, "D9", "O1", "1234"))); code != errcode.CodeOrderNotFound {
		t.Errorf("他人订单取货应 404, got=%d", code)
	}
	if code := codeOf(t, mustErr(s.Pickup(ctx, "D1", "O1", "9999"))); code != errcode.CodeDeliveryCodeFail {
		t.Errorf("错误取货码应报校验失败, got=%d", code)
	}
	if u.orders.byID["O1"].Status != order.StatusReady {
		t.Error("取货码错误却推进了状态")
	}

	got, err := s.Pickup(ctx, "D1", "O1", " 1234 ")
	if err != nil {
		t.Fatalf("取货失败: %v", err)
	}
	if got.Status != order.OrderDelivering || got.PickedAt == 0 {
		t.Errorf("取货后状态/时间错误: %+v", got)
	}
	if e := u.evts.saved[len(u.evts.saved)-1]; e.ActorType != order.ActorDriver || e.ToStatus != order.OrderDelivering {
		t.Errorf("取货事件错误: %+v", e)
	}

	// 未送达不可再取。
	if code := codeOf(t, mustErr(s.Pickup(ctx, "D1", "O1", "1234"))); code != errcode.CodeOrderState {
		t.Errorf("配送中不可再取货, got=%d", code)
	}
	// 送达码错误
	if code := codeOf(t, mustErr(s.Deliver(ctx, "D1", "O1", "0000"))); code != errcode.CodeDeliveryCodeFail {
		t.Errorf("错误送达码应报校验失败, got=%d", code)
	}

	done, err := s.Deliver(ctx, "D1", "O1", "5678")
	if err != nil {
		t.Fatalf("送达失败: %v", err)
	}
	if done.Status != order.StatusDelivered || done.DeliveredAt == 0 {
		t.Errorf("送达后状态/时间错误: %+v", done)
	}
	// 骑手分成 = 配送费 × 0.8，必须按分取整后落钱包与结算流水。
	if want := round2(6 * riderShareRate); done.RiderEarning != want {
		t.Errorf("骑手收入错误: got=%v want=%v", done.RiderEarning, want)
	}
	acc := u.wallet.accounts[walletKey("rider", "D1")]
	if acc == nil || acc.Balance != 4.8 {
		t.Errorf("配送收入未入钱包: %+v", acc)
	}
	if len(u.wallet.txns) != 1 || u.wallet.txns[0].Type != "delivery_income" {
		t.Errorf("配送收入流水错误: %+v", u.wallet.txns)
	}
	if len(u.settle.saved) != 1 {
		t.Fatalf("结算流水缺失: %d", len(u.settle.saved))
	}
	if st := u.settle.saved[0]; st.Amount != 4.8 || st.Party != "飞毛腿" || st.PartyType != "rider" {
		t.Errorf("结算流水内容错误: %+v", st)
	}
	if len(u.noti.saved) != 1 || u.noti.saved[0].UserID != "U1" {
		t.Errorf("送达未通知用户: %+v", u.noti.saved)
	}
	// 已送达不可重复送达。
	if code := codeOf(t, mustErr(s.Deliver(ctx, "D1", "O1", "5678"))); code != errcode.CodeOrderState {
		t.Errorf("重复送达应报状态不允许, got=%d", code)
	}

	// 取送码为空（如跑腿单未生成）时不拦校验，但状态守卫仍然生效。
	u2 := newUnit()
	s2 := &OrderService{Repos: u2.repos, Tx: u2.tx}
	u2.riders.byID["D2"] = &rider.Rider{ID: "D2", Name: "小马", Status: rider.RiderOnline}
	u2.orders.byID["O2"] = &order.Order{ID: "O2", UserID: "U2", RegionID: "RG1", RiderID: "D2",
		OrderNo: "N2", Status: order.StatusReady, DeliveryFee: 5}
	if _, err := s2.Pickup(ctx, "D2", "O2", ""); err != nil {
		t.Errorf("无取货码时应放行: %v", err)
	}
	u2.orders.byID["O3"] = &order.Order{ID: "O3", RiderID: "D2", Status: order.StatusReady, DeliveryFee: 5}
	if code := codeOf(t, mustErr(s2.Deliver(ctx, "D2", "O3", ""))); code != errcode.CodeOrderState {
		t.Errorf("未取货不可送达, got=%d", code)
	}
}

func TestDriverSetStatusGuards(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}

	if code := codeOf(t, mustSetStatusErr(s.SetDriverStatus(ctx, "D404", true, "op"))); code != errcode.CodeRiderNotFound {
		t.Errorf("骑手不存在应 404, got=%d", code)
	}
	u.riders.byID["D1"] = &rider.Rider{ID: "D1", AuditStatus: rider.AuditPending, Status: rider.RiderOffline}
	if code := codeOf(t, mustSetStatusErr(s.SetDriverStatus(ctx, "D1", true, "op"))); code != errcode.CodeRiderStatus {
		t.Errorf("未审核骑手不应能上线, got=%d", code)
	}
	u.riders.byID["D1"].AuditStatus = rider.AuditApproved
	u.riders.byID["D1"].Status = rider.RiderFrozen
	if code := codeOf(t, mustSetStatusErr(s.SetDriverStatus(ctx, "D1", true, "op"))); code != errcode.CodeRiderStatus {
		t.Errorf("冻结骑手不应能上线, got=%d", code)
	}

	u.riders.byID["D1"].Status = rider.RiderOffline
	got, err := s.SetDriverStatus(ctx, "D1", true, "op-1")
	if err != nil {
		t.Fatalf("上线失败: %v", err)
	}
	if got.Status != rider.RiderOnline {
		t.Errorf("上线后状态错误: %q", got.Status)
	}
	if len(u.riders.statuses) != 1 || u.riders.statuses[0] != rider.RiderOnline {
		t.Errorf("上线未写库: %v", u.riders.statuses)
	}
	if len(u.audit.entries) != 1 || u.audit.entries[0].Action != "driver.status" {
		t.Errorf("上线未留审计: %+v", u.audit.entries)
	}

	// 下线：直接置 offline。
	if _, err := s.SetDriverStatus(ctx, "D1", false, "op-1"); err != nil {
		t.Fatalf("下线失败: %v", err)
	}
	if got.Status != rider.RiderOffline {
		t.Errorf("下线后状态错误: %q", got.Status)
	}
	// 仓储写失败必须上抛，避免「界面显示已上线、库里还是离线」。
	u.riders.updateErr = application.ErrNotFound
	if _, err := s.SetDriverStatus(ctx, "D1", true, "op"); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("写库失败应原样上抛, got=%v", err)
	}
}

func TestOrderTabsAndScopePassThrough(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	u.orders.byID["O1"] = &order.Order{ID: "O1", UserID: "U1", MerchantID: "M1", RiderID: "D1", Status: order.StatusPreparing}
	u.orders.byID["O2"] = &order.Order{ID: "O2", UserID: "U1", MerchantID: "M2", RiderID: "D2", Status: order.StatusPreparing}
	u.orders.byID["O3"] = &order.Order{ID: "O3", UserID: "U2", MerchantID: "M1", Status: order.OrderPending}

	// R3-5：租户隔离必须靠下推 scope，而不是取全量后内存过滤。
	rows, total, err := s.StoreList(ctx, "M1", "preparing", "", 0, 0)
	if err != nil {
		t.Fatalf("商家列表失败: %v", err)
	}
	if total != 1 || len(rows) != 1 || rows[0].ID != "O1" {
		t.Errorf("商家列表越界: total=%d rows=%+v", total, rows)
	}
	if u.orders.lastScope.Role != identity.RoleMerchant || u.orders.lastScope.MerchantID != "M1" {
		t.Errorf("商家列表未下推 scope: %+v", u.orders.lastScope)
	}
	if q := u.orders.lastQuery; q.Page != 1 || q.PageSize != 20 {
		t.Errorf("分页参数未归一: page=%d size=%d", q.Page, q.PageSize)
	}
	if len(u.orders.lastQuery.Statuses) != 1 || u.orders.lastQuery.Statuses[0] != string(order.StatusPreparing) {
		t.Errorf("Tab 未转换为状态清单: %v", u.orders.lastQuery.Statuses)
	}

	rows, _, err = s.DriverList(ctx, "D1", "delivering", 1, 10)
	if err != nil {
		t.Fatalf("骑手列表失败: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("骑手列表应按骑手过滤, got=%+v", rows)
	}
	if u.orders.lastScope.Role != identity.RoleDriver || u.orders.lastScope.RiderID != "D1" {
		t.Errorf("骑手列表未下推 scope: %+v", u.orders.lastScope)
	}

	rows, total, err = s.ClientList(ctx, "U1", "ongoing", "", 1, 10)
	if err != nil {
		t.Fatalf("用户列表失败: %v", err)
	}
	if total != 2 || len(rows) != 2 {
		t.Errorf("用户列表条数错误: total=%d rows=%d", total, len(rows))
	}
	if u.orders.lastScope.Role != identity.RoleUser || u.orders.lastScope.UserID != "U1" {
		t.Errorf("用户列表未下推 scope: %+v", u.orders.lastScope)
	}

	// 详情：归属校验 + 时间线。
	u.evts.saved = []*order.Event{{OrderID: "O1", FromStatus: order.OrderPending, ToStatus: order.StatusPreparing}}
	d, err := s.ClientDetail(ctx, "U1", "O1")
	if err != nil {
		t.Fatalf("用户详情失败: %v", err)
	}
	if d.Order.ID != "O1" || len(d.Events) != 1 {
		t.Errorf("用户详情/事件错误: %+v", d)
	}
	if code := codeOf(t, mustDetailErr(s.ClientDetail(ctx, "U2", "O1"))); code != errcode.CodeOrderNotFound {
		t.Errorf("他人订单详情应 404, got=%d", code)
	}
	if code := codeOf(t, mustDetailErr(s.StoreDetail(ctx, "M2", "O1"))); code != errcode.CodeOrderNotOwned {
		t.Errorf("非本店详情应 404, got=%d", code)
	}
	if code := codeOf(t, mustDetailErr(s.DriverDetail(ctx, "D2", "O1"))); code != errcode.CodeOrderNotFound {
		t.Errorf("非本人配送单详情应 404, got=%d", code)
	}
	if code := codeOf(t, mustDetailErr(s.ClientDetail(ctx, "U1", "MISSING"))); code != errcode.CodeOrderNotFound {
		t.Errorf("订单不存在应 404, got=%d", code)
	}
}

func TestOrderActionsFollowStatus(t *testing.T) {
	cases := []struct {
		status  order.OrderStatus
		riderID string
		client  []string
		store   []string
		driver  []string
	}{
		{order.StatusPendingPay, "", []string{"pay", "cancel"}, nil, nil},
		{order.OrderPending, "", []string{"cancel"}, []string{"accept", "reject"}, nil},
		{order.StatusPreparing, "", []string{"track"}, []string{"ready"}, nil},
		{order.StatusReady, "", []string{"track"}, []string{"detail"}, []string{"grab"}},
		{order.StatusReady, "D1", []string{"track"}, []string{"detail"}, []string{"pickup"}},
		{order.OrderDelivering, "D1", []string{"track"}, []string{"detail"}, []string{"deliver"}},
		{order.StatusDelivered, "D1", []string{"confirm", "track"}, nil, nil},
		{order.StatusPendingReview, "D1", nil, nil, nil},
		{order.OrderCanceled, "", nil, nil, nil},
	}
	for _, c := range cases {
		o := &order.Order{Status: c.status, RiderID: c.riderID}
		if got := ClientActions(o); strings.Join(got, ",") != strings.Join(c.client, ",") {
			t.Errorf("%s ClientActions=%v want=%v", c.status, got, c.client)
		}
		if got := StoreActions(o); strings.Join(got, ",") != strings.Join(c.store, ",") {
			t.Errorf("%s StoreActions=%v want=%v", c.status, got, c.store)
		}
		if got := DriverActions(o, "D1"); strings.Join(got, ",") != strings.Join(c.driver, ",") {
			t.Errorf("%s DriverActions=%v want=%v", c.status, got, c.driver)
		}
	}
	// 别人已接的单，第三人不得出现任何动作（尤其不能出现 pickup）。
	o := &order.Order{Status: order.StatusReady, RiderID: "D9"}
	if got := DriverActions(o, "D1"); len(got) != 0 {
		t.Errorf("他人已占单不应给出动作: %v", got)
	}
	o.Status = order.OrderDelivering
	if got := DriverActions(o, "D1"); len(got) != 0 {
		t.Errorf("他人配送中的单不应给出动作: %v", got)
	}
}

func TestRiderPoolDistances(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	u.riders.byID["D1"] = &rider.Rider{ID: "D1", RegionID: "RG1", Status: rider.RiderOnline}
	u.riders.byID["DF"] = &rider.Rider{ID: "DF", RegionID: "RG1", Status: rider.RiderFrozen}
	u.orders.byID["O1"] = &order.Order{ID: "O1", RegionID: "RG1", Status: order.StatusReady,
		PickupLat: 39.9042, PickupLng: 116.4074}
	u.orders.byID["O2"] = &order.Order{ID: "O2", RegionID: "RG1", Status: order.StatusReady}
	u.orders.byID["O3"] = &order.Order{ID: "O3", RegionID: "RG1", Status: order.StatusReady, RiderID: "D9"}

	if code := codeOf(t, mustPoolErr(s.Pool(ctx, "D404", 0, 0, 1, 10))); code != errcode.CodeRiderNotFound {
		t.Errorf("骑手不存在应 404, got=%d", code)
	}
	if code := codeOf(t, mustPoolErr(s.Pool(ctx, "DF", 0, 0, 1, 10))); code != errcode.CodeRiderOffline {
		t.Errorf("冻结骑手不应看到抢单池, got=%d", code)
	}

	items, total, err := s.Pool(ctx, "D1", 39.9042, 116.4074, 1, 10)
	if err != nil {
		t.Fatalf("抢单池失败: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("抢单池应只含本区未分配单: total=%d items=%d", total, len(items))
	}
	if u.orders.lastPool != "RG1" {
		t.Errorf("抢单池未按骑手区域查询: %q", u.orders.lastPool)
	}
	for _, it := range items {
		if it.Order.ID == "O1" {
			if it.DistanceKM != 0 || it.DeliveryMinutes <= 0 {
				t.Errorf("同点取货距离应为 0 且有时长: %+v", it)
			}
			continue
		}
		if it.DistanceKM != 0 || it.DeliveryMinutes != 0 {
			t.Errorf("缺坐标订单不应算出距离/时长: %+v", it)
		}
	}
	// 骑手位置无效时一律不给距离（宁可不显示，也不显示错误距离）。
	items, _, err = s.Pool(ctx, "D1", 0, 0, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.DistanceKM != 0 {
			t.Errorf("位置无效却算出了距离: %+v", it)
		}
	}
}

// --- 断言辅助：不同返回签名共用一个错误码断言 ---

func mustSetStatusErr(_ *rider.Rider, err error) error    { return err }
func mustDetailErr(_ *OrderDetail, err error) error       { return err }
func mustPoolErr(_ []*PoolItem, _ int64, err error) error { return err }
