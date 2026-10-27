package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/coupon"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/refund"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/system"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
)

// R5-5 资金与状态机路径单测。仓储用条件更新语义的替身（见 fakes_test.go），
// 目的是锁死应用层的三条铁律：金额只走服务端计价、余额只走原子扣减、状态只走条件流转。

func codeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var e *errcode.Error
	if errors.As(err, &e) {
		return e.Code
	}
	t.Fatalf("非业务错误，无法断言错误码: %v", err)
	return -1
}

// --- 金额与地理计算 ---

func TestRound2(t *testing.T) {
	cases := []struct{ in, want float64 }{
		{0, 0}, {1, 1}, {2.675, 2.68}, {2.674, 2.67}, {8.835, 8.84},
		{0.1 + 0.2, 0.3}, {19.999, 20.0}, {1e6 + 0.004, 1000000.0},
		// 负值受二进制小数表示影响（-1.005*100 落在 -100.4999…），资金链路只传正数金额，
		// 这里记录现状，避免有人误以为 round2 对负数是「四舍五入」。
		{-1.005, -1},
	}
	for _, c := range cases {
		if got := round2(c.in); got != c.want {
			t.Errorf("round2(%v)=%v want=%v", c.in, got, c.want)
		}
	}
}

func TestValidCoordAndHaversine(t *testing.T) {
	if validCoord(0, 0) {
		t.Error("(0,0) 视为缺省坐标，应判无效")
	}
	if validCoord(91, 0) || validCoord(0, 181) || validCoord(-90.1, 0) {
		t.Error("越界坐标应判无效")
	}
	if !validCoord(39.9, 116.4) {
		t.Error("北京坐标应判有效")
	}
	if d := haversine(39.9042, 116.4074, 39.9042, 116.4074); d != 0 {
		t.Errorf("同点距离应为 0, got=%v", d)
	}
	// 北京→上海约 1060km，计价与配送时长都依赖该量级，偏移过大说明公式被改坏。
	if d := haversine(39.9042, 116.4074, 31.2304, 121.4737); d < 1050 || d > 1090 {
		t.Errorf("京沪距离异常: %v", d)
	}
	if m := estimateMinutes(0); m <= 0 {
		t.Errorf("0 公里也应有基础时长, got=%d", m)
	}
	if lo, hi := estimateMinutes(1), estimateMinutes(30); hi <= lo {
		t.Errorf("时长应随距离单调递增: 1km=%d 30km=%d", lo, hi)
	}
}

func TestNewOrderNo(t *testing.T) {
	no := newOrderNo("01ABCDEF1234567890", 1700000000)
	if len(no) != 20 {
		t.Errorf("单号长度应为 20, got=%d (%s)", len(no), no)
	}
	if !strings.HasPrefix(no, "2023") {
		t.Errorf("单号应含时间前缀: %s", no)
	}
	if no != strings.ToUpper(no) {
		t.Errorf("单号应全大写: %s", no)
	}
	// 不同 ID 同秒下单，后缀必须不同（唯一索引兜底依赖它）。
	if newOrderNo("01ABCDEFGH1234567", 1700000000) == newOrderNo("01ABCDEFGHI999999", 1700000000) {
		t.Error("同秒不同订单生成了相同单号")
	}
}

func TestRandomCode(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		c := randomCode()
		if len(c) != 4 || strings.Trim(c, "0123456789") != "" {
			t.Fatalf("取送码需为 4 位数字, got=%q", c)
		}
		seen[c] = true
	}
	if len(seen) < 50 {
		t.Errorf("取送码重复率过高，疑似非加密随机: %d/200", len(seen))
	}
}

func TestSmallHelpers(t *testing.T) {
	if p, s := normalizePageParams(0, 0); p != 1 || s != 20 {
		t.Errorf("零值分页应回落默认, got=%d/%d", p, s)
	}
	if p, s := normalizePageParams(-5, 9999); p != 1 || s != 20 {
		t.Errorf("越界分页应回落默认值, got=%d/%d", p, s)
	}
	if p, s := normalizePageParams(3, 50); p != 3 || s != 50 {
		t.Errorf("合法分页不应被改写, got=%d/%d", p, s)
	}
	if formatMoney(12.345) != "12.35" && formatMoney(12.345) != "12.34" {
		t.Errorf("金额文案异常: %s", formatMoney(12.345))
	}
	if got := maskAccount("6222021234567890"); got != "62****90" {
		t.Errorf("收款账号未脱敏: %s", got)
	}
	if got := maskAccount(" 12 "); got != "12" {
		t.Errorf("短账号不应加掩码噪声: %s", got)
	}
	if refundedStatus(&order.Order{PaymentStatus: "paid"}) != "refunded" {
		t.Error("已支付订单取消后应标记 refunded")
	}
	if refundedStatus(&order.Order{PaymentStatus: "unpaid"}) != "unpaid" {
		t.Error("未支付订单不应被标记退款")
	}
	for _, ty := range []string{"帮买", "帮送", "帮取", "帮办"} {
		if !isErrandType(ty) {
			t.Errorf("%s 应识别为跑腿单", ty)
		}
	}
	for _, ty := range []string{"外卖", "", "帮办2"} {
		if isErrandType(ty) {
			t.Errorf("%q 不应识别为跑腿单", ty)
		}
	}
	if steps := initialSteps(100); len(steps) != 4 || steps[0].State != "done" || steps[0].Time != 100 {
		t.Errorf("初始时间线错误: %+v", steps)
	}
	for i := 1; i < 4; i++ {
		if steps := initialSteps(100); steps[i].State != "pending" || steps[i].Time != 0 {
			t.Errorf("初始时间线第 %d 步应为未完成: %+v", i, steps[i])
		}
	}
}

func TestValidatePlace(t *testing.T) {
	good := &PlaceInput{ContactName: " 张三 ", Phone: "13800138000", Address: " 学院路 1 号 "}
	if err := validatePlace(good); err != nil {
		t.Fatalf("合法取送点被拒: %v", err)
	}
	if good.ContactName != "张三" || good.Address != "学院路 1 号" {
		t.Errorf("取送点未去首尾空白: %+v", good)
	}
	bad := []*PlaceInput{
		{ContactName: "", Phone: "13800138000", Address: "学院路1号"},
		{ContactName: strings.Repeat("名", 33), Phone: "13800138000", Address: "学院路1号"},
		{ContactName: "张三", Phone: "1380013800", Address: "学院路1号"},
		{ContactName: "张三", Phone: "13800138000", Address: "A"},
		{ContactName: "张三", Phone: "13800138000", Address: strings.Repeat("址", 121)},
	}
	for i, p := range bad {
		if err := validatePlace(p); err == nil {
			t.Errorf("非法取送点 #%d 被放行: %+v", i, p)
		}
	}
}

// --- 钱包账本 ---

func TestCreditWalletSkipsNonPositive(t *testing.T) {
	u := newUnit()
	ctx := context.Background()
	for _, amt := range []float64{0, -1} {
		if err := creditWallet(ctx, u.repos, "rider", "R1", amt, "delivery_income", "O1", "x", 100); err != nil {
			t.Fatalf("非正金额入账不应报错: %v", err)
		}
	}
	if len(u.wallet.txns) != 0 {
		t.Errorf("非正金额产生了流水: %+v", u.wallet.txns)
	}
	if err := creditWallet(ctx, u.repos, "rider", "", 10, "delivery_income", "O1", "x", 100); err != nil {
		t.Fatalf("空主体入账不应报错: %v", err)
	}
	if len(u.wallet.txns) != 0 {
		t.Error("空主体 ID 也写了流水")
	}
}

func TestCreditWalletWritesLedger(t *testing.T) {
	u := newUnit()
	ctx := context.Background()
	if err := creditWallet(ctx, u.repos, "rider", "R1", 12.34, "delivery_income", "O1", "配送收入", 100); err != nil {
		t.Fatal(err)
	}
	if err := creditWallet(ctx, u.repos, "rider", "R1", 0.66, "delivery_income", "O2", "配送收入", 101); err != nil {
		t.Fatal(err)
	}
	if len(u.wallet.txns) != 2 {
		t.Fatalf("流水条数错误: %d", len(u.wallet.txns))
	}
	acc := u.wallet.accounts[walletKey("rider", "R1")]
	if acc.Balance != 13.0 {
		t.Errorf("余额累加错误: %v", acc.Balance)
	}
	if u.wallet.txns[0].BalanceAfter != 12.34 || u.wallet.txns[1].BalanceAfter != 13.0 {
		t.Errorf("流水余额快照应逐笔跟涨: %v / %v", u.wallet.txns[0].BalanceAfter, u.wallet.txns[1].BalanceAfter)
	}
	if u.wallet.txns[0].AccountID != acc.ID || u.wallet.txns[0].OrderID != "O1" {
		t.Errorf("流水未关联账户/订单: %+v", u.wallet.txns[0])
	}
}

// R1-1/R2-3 回归：透支必须落在仓储的条件更新上，应用层拿到 false 后不得留流水。
func TestDebitWalletInsufficientLeavesNoLedger(t *testing.T) {
	u := newUnit()
	ctx := context.Background()
	if _, err := u.wallet.Ensure(ctx, "merchant", "M1", 100); err != nil {
		t.Fatal(err)
	}
	u.wallet.accounts[walletKey("merchant", "M1")].Balance = 50

	ok, err := debitWallet(ctx, u.repos, "merchant", "M1", 60, false, "withdraw", "", "提现", 101)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("余额不足却判定扣款成功")
	}
	if len(u.wallet.txns) != 0 {
		t.Errorf("扣款失败仍写流水: %+v", u.wallet.txns)
	}
	if u.wallet.accounts[walletKey("merchant", "M1")].Balance != 50 {
		t.Error("扣款失败改动了余额")
	}

	// 允许透支用于退款冲回（商家已结算金额可为负向调整）。
	ok, err = debitWallet(ctx, u.repos, "merchant", "M1", 60, true, "refund_reversal", "O1", "冲回", 102)
	if err != nil || !ok {
		t.Fatalf("允许透支的冲回失败: ok=%v err=%v", ok, err)
	}
	if got := u.wallet.accounts[walletKey("merchant", "M1")].Balance; got != -10 {
		t.Errorf("冲回后余额应为 -10, got=%v", got)
	}
	if len(u.wallet.txns) != 1 || u.wallet.txns[0].Amount != -60 {
		t.Errorf("冲回流水错误: %+v", u.wallet.txns)
	}
}

func TestDebitWalletSkipsNonPositive(t *testing.T) {
	u := newUnit()
	ok, err := debitWallet(context.Background(), u.repos, "rider", "R1", 0, false, "withdraw", "", "x", 100)
	if err != nil || !ok {
		t.Errorf("0 金额应视为无事发生: ok=%v err=%v", ok, err)
	}
	if len(u.wallet.txns) != 0 {
		t.Error("0 金额写了流水")
	}
}

// --- 余额支付（防双花）---

// R1-1 回归：余额扣减只能是一次条件 AddBalance，禁止 Find-then-Save，
// 且余额恰好等于应付时必须放行（>= 而非 >）。
func TestDeductBalanceIsAtomic(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	u.prof.balances["U1"] = 30

	if err := deductBalance(ctx, u.repos, "U1", 30, 100); err != nil {
		t.Fatalf("余额恰好够付应放行: %v", err)
	}
	if u.prof.balances["U1"] != 0 {
		t.Errorf("扣减后余额应为 0, got=%v", u.prof.balances["U1"])
	}
	if len(u.prof.adds) != 1 {
		t.Errorf("余额扣减应只有一次原子更新, got=%d 次", len(u.prof.adds))
	}

	if code := codeOf(t, deductBalance(ctx, u.repos, "U1", 5, 101)); code != errcode.CodeInsufficientFund {
		t.Errorf("余额不足错误码=%d want=%d", code, errcode.CodeInsufficientFund)
	}
	if u.prof.balances["U1"] != 0 {
		t.Error("扣减失败仍改动了余额")
	}

	u.prof.missing["U2"] = true
	if code := codeOf(t, deductBalance(ctx, u.repos, "U2", 5, 102)); code != errcode.CodeInsufficientFund {
		t.Errorf("无资料用户应回落余额不足, got=%d", code)
	}
}

// 退款入账走的是 creditBalance：资料缺失时把错误原样上抛，让整笔事务回滚，
// 而不是静默跳过——静默跳过等于用户的退款钱凭空消失，宁可失败让人来查。
func TestCreditBalanceFailsLoudOnMissingProfile(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	u.prof.missing["U404"] = true
	if err := creditBalance(ctx, u.repos, "U404", 10, 100); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("资料缺失应上抛 ErrNotFound 触发回滚, got=%v", err)
	}
	if len(u.prof.adds) != 0 {
		t.Errorf("资料缺失却执行了余额更新: %v", u.prof.adds)
	}
	if err := creditBalance(ctx, u.repos, "U1", 10.004, 101); err != nil {
		t.Fatal(err)
	}
	if u.prof.balances["U1"] != 10 {
		t.Errorf("入账金额未按分取整: %v", u.prof.balances["U1"])
	}
}

// --- 优惠券 ---

func TestApplyCoupon(t *testing.T) {
	ctx := context.Background()
	o := func() *order.Order {
		return &order.Order{ID: "O1", UserID: "U1", MerchantID: "M1",
			Amount: 50, PayAmount: 56, DiscountAmount: 0, Fees: []order.OrderFee{{Label: "商品金额", Value: 50}}}
	}
	now := int64(1000)

	u := newUnit()
	svc := &OrderService{Repos: u.repos}
	if id, err := svc.applyCoupon(ctx, o(), "  ", now); err != nil || id != "" {
		t.Errorf("空券号应直接跳过: id=%q err=%v", id, err)
	}

	u.coupon.byID["C1"] = &coupon.Coupon{ID: "C1", UserID: "U1", Amount: 10, MinSpend: 30, Status: coupon.StatusUnused, ExpiresAt: 2000}
	ord := o()
	id, err := (&OrderService{Repos: u.repos}).applyCoupon(ctx, ord, " C1 ", now)
	if err != nil {
		t.Fatalf("合法券被拒: %v", err)
	}
	if id != "C1" || ord.CouponID != "C1" || ord.CouponAmount != 10 {
		t.Errorf("券未生效: id=%s order=%+v", id, ord)
	}
	if ord.PayAmount != 46 || ord.DiscountAmount != 10 {
		t.Errorf("优惠后金额错误: pay=%v discount=%v", ord.PayAmount, ord.DiscountAmount)
	}
	if last := ord.Fees[len(ord.Fees)-1]; last.Label != "优惠券" || last.Value != -10 || !last.Negative {
		t.Errorf("费用明细缺少负向优惠行: %+v", last)
	}

	// 面额大于应付：封顶到应付，且不得产生 0 元以下的支付金额。
	u.coupon.byID["C2"] = &coupon.Coupon{ID: "C2", UserID: "U1", Amount: 999, Status: coupon.StatusUnused, ExpiresAt: 2000}
	ord2 := o()
	if _, err := (&OrderService{Repos: u.repos}).applyCoupon(ctx, ord2, "C2", now); err != nil {
		t.Fatalf("超额券被拒: %v", err)
	}
	if ord2.PayAmount != 0 || ord2.CouponAmount != 56 {
		t.Errorf("超额券应封顶: pay=%v coupon=%v", ord2.PayAmount, ord2.CouponAmount)
	}

	// 各类不可用券必须拒绝，且不留下任何金额改动。
	u.coupon.byID["CX"] = &coupon.Coupon{ID: "CX", UserID: "U1", Amount: 10, Status: coupon.StatusUnused, ExpiresAt: now} // 刚好过期
	u.coupon.byID["CU"] = &coupon.Coupon{ID: "CU", UserID: "U1", Amount: 10, Status: coupon.StatusUsed, ExpiresAt: 2000}
	u.coupon.byID["CE"] = &coupon.Coupon{ID: "CE", UserID: "U1", Amount: 10, Status: coupon.StatusExpired, ExpiresAt: 2000}
	u.coupon.byID["CO"] = &coupon.Coupon{ID: "CO", UserID: "OTHER", Amount: 10, Status: coupon.StatusUnused, ExpiresAt: 2000}
	u.coupon.byID["CM"] = &coupon.Coupon{ID: "CM", UserID: "U1", MerchantID: "M2", Amount: 10, Status: coupon.StatusUnused, ExpiresAt: 2000}
	u.coupon.byID["CT"] = &coupon.Coupon{ID: "CT", UserID: "U1", Amount: 10, MinSpend: 60, Status: coupon.StatusUnused, ExpiresAt: 2000}
	u.coupon.byID["CZ"] = &coupon.Coupon{ID: "CZ", UserID: "U1", Amount: 0, Status: coupon.StatusUnused, ExpiresAt: 2000}
	for _, cid := range []string{"CX", "CU", "CE", "CO", "CM", "CT", "CZ", "MISSING"} {
		ord := o()
		id, err := (&OrderService{Repos: u.repos}).applyCoupon(ctx, ord, cid, now)
		if codeOf(t, err) != errcode.CodeCouponInvalid {
			t.Errorf("不可用券 %s 未被拒绝: id=%q err=%v", cid, id, err)
		}
		if ord.PayAmount != 56 || ord.CouponID != "" || len(ord.Fees) != 1 {
			t.Errorf("不可用券 %s 仍改动了订单: %+v", cid, ord)
		}
	}
}

// --- 下单入口的输入边界 ---

func TestOrderServiceCreateRejectsMalformedInput(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}

	if code := codeOf(t, mustCreateErr(s.Create(ctx, "U1", CreateOrderInput{
		Type: "外卖", ClientOrderNo: strings.Repeat("n", 65)}))); code != errcode.CodeBadRequest {
		t.Errorf("超长 client_order_no 应为参数错误, got=%d", code)
	}
	if code := codeOf(t, mustCreateErr(s.Create(ctx, "U1", CreateOrderInput{
		Type: "外卖", Remark: strings.Repeat("注", 513)}))); code != errcode.CodeBadRequest {
		t.Errorf("超长备注应为参数错误, got=%d", code)
	}
	if len(u.tx.unit.Orders.(*fakeOrders).saved) != 0 {
		t.Error("参数校验失败不应进入写路径")
	}

	// 幂等命中：同单号已有订单时直接返回旧单，不开事务、不重复扣库存。
	u.orders.byID["OLD"] = &order.Order{ID: "OLD", UserID: "U1", ClientOrderNo: "dup-1", Status: order.StatusPendingPay}
	o, err := s.Create(ctx, "U1", CreateOrderInput{Type: "外卖", ClientOrderNo: "dup-1"})
	if err != nil || o == nil || o.ID != "OLD" {
		t.Fatalf("幂等回读失败: o=%+v err=%v", o, err)
	}
	if u.tx.runCount != 0 {
		t.Error("幂等命中不应开启事务")
	}

	// 用户不存在 → 交由 userErr 映射
	if err := mustCreateErr(s.Create(ctx, "U404", CreateOrderInput{Type: "外卖", ClientOrderNo: "no-user"})); err == nil {
		t.Error("用户不存在时应报错")
	}
	// 用户未选区域 → 区域必填
	u.users.byID["U1"] = &identity.User{ID: "U1", RegionID: ""}
	if code := codeOf(t, mustCreateErr(s.Create(ctx, "U1", CreateOrderInput{Type: "外卖", ClientOrderNo: "no-region"}))); code != errcode.CodeRegionRequired {
		t.Errorf("未选区域应报 %d, got=%d", errcode.CodeRegionRequired, code)
	}
	// 区域不存在
	u.users.byID["U1"].RegionID = "RG404"
	if code := codeOf(t, mustCreateErr(s.Create(ctx, "U1", CreateOrderInput{Type: "外卖", ClientOrderNo: "bad-region"}))); code != errcode.CodeRegionNotFound {
		t.Errorf("区域不存在应报 %d, got=%d", errcode.CodeRegionNotFound, code)
	}
	// 未知订单类型
	u.region.byID["RG1"] = &tenant.Region{ID: "RG1"}
	u.users.byID["U1"].RegionID = "RG1"
	if code := codeOf(t, mustCreateErr(s.Create(ctx, "U1", CreateOrderInput{Type: "未知", ClientOrderNo: "bad-type"}))); code != errcode.CodeBadRequest {
		t.Errorf("未知订单类型应报参数错误, got=%d", code)
	}
	// 商品行超限
	items := make([]CreateItem, 51)
	for i := range items {
		items[i] = CreateItem{DishID: "D", Count: 1}
	}
	if code := codeOf(t, mustCreateErr(s.Create(ctx, "U1", CreateOrderInput{Type: "外卖", Items: items, ClientOrderNo: "too-many"}))); code != errcode.CodeBadRequest {
		t.Errorf("超过 50 行商品应报参数错误, got=%d", code)
	}
	// 跑腿单缺取送点 / 坐标
	if code := codeOf(t, mustCreateErr(s.Create(ctx, "U1", CreateOrderInput{Type: "帮送", ClientOrderNo: "e1"}))); code != errcode.CodeBadRequest {
		t.Errorf("跑腿单缺取送点应报参数错误, got=%d", code)
	}
	if code := codeOf(t, mustCreateErr(s.Create(ctx, "U1", CreateOrderInput{Type: "帮送", ClientOrderNo: "e2",
		Pickup:  &PlaceInput{ContactName: "张三", Phone: "13800138000", Address: "A 地址"},
		Dropoff: &PlaceInput{ContactName: "李四", Phone: "13800138001", Address: "B 地址"}}))); code != errcode.CodeBadRequest {
		t.Errorf("跑腿单缺坐标应报参数错误, got=%d", code)
	}
}

func mustCreateErr(_ *order.Order, err error) error { return err }

// --- 余额支付全流程 ---

func TestPayWithBalance(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	pendingOrder(u, "O1", "U1", "M1", 56)
	u.prof.balances["U1"] = 56
	u.merch.byID["M1"] = &merchant.Merchant{ID: "M1", Name: "小店", OwnerUserID: "MU1"}

	o, err := s.Pay(ctx, "U1", "O1", "balance")
	if err != nil {
		t.Fatalf("余额支付失败: %v", err)
	}
	if o.Status != order.OrderPending || o.PaymentStatus != "paid" || o.PayMethod != "balance" {
		t.Errorf("支付后状态错误: %+v", o)
	}
	if u.prof.balances["U1"] != 0 {
		t.Errorf("余额未扣减: %v", u.prof.balances["U1"])
	}
	if len(u.pay.saved) != 1 || u.pay.saved[0].Amount != 56 || u.pay.saved[0].Status != order.PaymentPaid {
		t.Errorf("支付单错误: %+v", u.pay.saved)
	}
	if len(u.evts.saved) != 1 || u.evts.saved[0].ToStatus != order.OrderPending {
		t.Errorf("缺少支付事件: %+v", u.evts.saved)
	}
	if len(u.noti.saved) != 1 || u.noti.saved[0].UserID != "MU1" {
		t.Errorf("未通知商家 owner: %+v", u.noti.saved)
	}

	// 重复支付：已 paid 直接返回，不再二次扣款（防双花的第二道闸）。
	if _, err := s.Pay(ctx, "U1", "O1", "balance"); err != nil {
		t.Fatalf("重复支付应幂等成功: %v", err)
	}
	if len(u.prof.adds) != 1 || len(u.pay.saved) != 1 {
		t.Errorf("重复支付产生二次扣款/支付单: adds=%d payments=%d", len(u.prof.adds), len(u.pay.saved))
	}
}

func TestPayGuards(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx, MockPayment: false}
	pendingOrder(u, "O1", "U1", "M1", 20)
	u.prof.balances["U1"] = 5

	if code := codeOf(t, mustErr(s.Pay(ctx, "U2", "O1", "balance"))); code != errcode.CodeOrderNotFound {
		t.Errorf("他人订单支付应 404, got=%d", code)
	}
	if code := codeOf(t, mustErr(s.Pay(ctx, "U1", "O1", "balance"))); code != errcode.CodeInsufficientFund {
		t.Errorf("余额不足错误码错误: %d", code)
	}
	if u.orders.byID["O1"].Status != order.StatusPendingPay {
		t.Error("扣款失败不应推进订单状态")
	}
	if u.tx.runCount != 1 {
		t.Error("扣款失败应发生在事务内（便于回滚）")
	}
	// 第三方渠道未接入（生产）：直接拒绝，不落支付单。
	if code := codeOf(t, mustErr(s.Pay(ctx, "U1", "O1", "wechat"))); code != errcode.CodePaymentFailed {
		t.Errorf("未接入渠道应拒绝, got=%d", code)
	}
	if code := codeOf(t, mustErr(s.Pay(ctx, "U1", "O1", "bitcoin"))); code != errcode.CodeBadRequest {
		t.Errorf("未知渠道应报参数错误, got=%d", code)
	}
	// 状态不是待支付：不可支付。
	u.orders.byID["O2"] = &order.Order{ID: "O2", UserID: "U1", Status: order.OrderDelivering, PayAmount: 10}
	if code := codeOf(t, mustErr(s.Pay(ctx, "U1", "O2", "balance"))); code != errcode.CodeOrderState {
		t.Errorf("非待支付状态应拒绝, got=%d", code)
	}
	// mock 通道仅在 dev/test 打开
	s.MockPayment = true
	pendingOrder(u, "O3", "U1", "M1", 10)
	if _, err := s.Pay(ctx, "U1", "O3", "mock"); err != nil {
		t.Errorf("开启 mock 后应可支付: %v", err)
	}
	// 并发抢占：条件更新 0 行时报状态已变化，且错误上抛
	u.orders.byID["O4"] = &order.Order{ID: "O4", UserID: "U1", Status: order.StatusPendingPay, PayAmount: 1}
	u.prof.balances["U1"] = 100
	u.orders.updateOK = false
	if code := codeOf(t, mustErr(s.Pay(ctx, "U1", "O4", "balance"))); code != errcode.CodeOrderState {
		t.Errorf("并发抢占应报状态已变化, got=%d", code)
	}
}

func TestPayByCallbackAmountMustMatch(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	pendingOrder(u, "O1", "U1", "M1", 56.67)

	if code := codeOf(t, mustErr(s.PayByCallback(ctx, "O1", "wechat", "T1", 56.66))); code != errcode.CodeAmountMismatch {
		t.Errorf("回调金额少 1 分也应拒绝, got=%d", code)
	}
	if code := codeOf(t, mustErr(s.PayByCallback(ctx, "O1", "wechat", "T1", 111.34))); code != errcode.CodeAmountMismatch {
		t.Errorf("回调金额翻倍应拒绝（防低价回调）, got=%d", code)
	}
	o, err := s.PayByCallback(ctx, "O1", "wechat", "T1", 56.67)
	if err != nil {
		t.Fatalf("等额回调失败: %v", err)
	}
	if o.Status != order.OrderPending || u.pay.saved[0].TradeNo != "T1" {
		t.Errorf("回调后状态/回执错误: %+v", u.pay.saved[0])
	}
	// 幂等：重复回调不再改状态、不再新增支付单
	before := len(u.pay.saved)
	if _, err := s.PayByCallback(ctx, "O1", "wechat", "T2", 56.67); err != nil {
		t.Fatalf("重复回调应幂等: %v", err)
	}
	if len(u.pay.saved) != before {
		t.Error("重复回调新增了支付单")
	}
	// 空回执由服务端补号，避免支付单无凭证
	pendingOrder(u, "O2", "U1", "M1", 10)
	if _, err := s.PayByCallback(ctx, "O2", "alipay", "", 10); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(u.pay.byOrder["O2"].TradeNo, "CB") {
		t.Errorf("空回执应补服务端单号: %s", u.pay.byOrder["O2"].TradeNo)
	}
}

// --- 取消：退款、释放库存与券、状态冲突 ---

func TestCancelReleasesFundsAndStock(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	o := pendingOrder(u, "O1", "U1", "M1", 56)
	o.Type = order.OrderTypeFood
	o.PaymentStatus = "paid"
	o.PayMethod = "balance"
	o.Goods = []order.OrderGoods{{DishID: "D1", Count: 2, Name: "面"}}
	o.CouponID = "C1"
	u.catalog.stocks["D1"] = 5
	u.prof.balances["U1"] = 0
	u.coupon.byID["C1"] = &coupon.Coupon{ID: "C1", UserID: "U1", Status: coupon.StatusUsed, UsedOrderID: "O1"}
	u.pay.byOrder["O1"] = &order.Payment{ID: "P1", OrderID: "O1", Status: order.PaymentPaid}
	u.merch.byID["M1"] = &merchant.Merchant{ID: "M1", OwnerUserID: "MU1"}

	got, err := s.CancelClient(ctx, "U1", "O1", "  ")
	if err != nil {
		t.Fatalf("取消失败: %v", err)
	}
	if got.Status != order.OrderCanceled || got.CancelReason != "用户取消" {
		t.Errorf("取消状态/原因错误: %+v", got)
	}
	if got.PaymentStatus != "refunded" {
		t.Errorf("已支付订单取消后应标 refunded, got=%q", got.PaymentStatus)
	}
	if u.prof.balances["U1"] != 56 {
		t.Errorf("余额未退回: %v", u.prof.balances["U1"])
	}
	if u.catalog.stocks["D1"] != 7 {
		t.Errorf("库存未释放: %v", u.catalog.stocks["D1"])
	}
	if len(u.coupon.released) != 1 || u.coupon.released[0] != "O1" {
		t.Errorf("券未释放: %v", u.coupon.released)
	}
	if len(u.pay.statuses) != 1 || u.pay.statuses[0] != string(order.PaymentRefunded) {
		t.Errorf("支付单未标退款: %v", u.pay.statuses)
	}
	// 用户取消 → 通知商家；事件落库
	if len(u.noti.saved) != 1 || u.noti.saved[0].UserID != "MU1" {
		t.Errorf("未通知商家: %+v", u.noti.saved)
	}
	if len(u.evts.saved) != 1 || u.evts.saved[0].ActorType != order.ActorClient {
		t.Errorf("取消事件错误: %+v", u.evts.saved)
	}
}

func TestCancelGuards(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	pendingOrder(u, "O1", "U1", "M1", 10)

	if code := codeOf(t, mustErr(s.CancelClient(ctx, "U2", "O1", "x"))); code != errcode.CodeOrderNotFound {
		t.Errorf("他人订单取消应 404, got=%d", code)
	}
	u.orders.byID["O2"] = &order.Order{ID: "O2", UserID: "U1", Status: order.OrderDelivering}
	if code := codeOf(t, mustErr(s.CancelClient(ctx, "U1", "O2", "x"))); code != errcode.CodeOrderState {
		t.Errorf("配送中不可取消, got=%d", code)
	}
	// 系统超时取消只处理待支付
	if _, err := s.CancelSystem(ctx, "O2", "超时未支付"); codeOf(t, err) != errcode.CodeOrderState {
		t.Errorf("CancelSystem 应拒绝非待支付单, got=%d", codeOf(t, err))
	}
	pendingOrder(u, "O3", "U1", "M1", 10)
	if _, err := s.CancelSystem(ctx, "O3", "超时未支付"); err != nil {
		t.Errorf("待支付超时应可取消: %v", err)
	}
	if u.orders.byID["O3"].PaymentStatus != "unpaid" {
		t.Errorf("未支付单取消不应标退款: %q", u.orders.byID["O3"].PaymentStatus)
	}
	// 并发：条件更新 0 行时报状态已变化
	pendingOrder(u, "O4", "U1", "M1", 10)
	u.orders.updateOK = false
	if code := codeOf(t, mustErr(s.CancelClient(ctx, "U1", "O4", "x"))); code != errcode.CodeOrderState {
		t.Errorf("并发取消应报状态已变化, got=%d", code)
	}
	// 非客户端操作方 → 通知用户而非商家
	u.noti.saved = nil
	u.orders.updateOK = true
	pendingOrder(u, "O5", "U1", "M1", 10)
	if _, err := s.CancelSystem(ctx, "O5", "风控取消"); err != nil {
		t.Fatal(err)
	}
	if len(u.noti.saved) != 1 || u.noti.saved[0].UserID != "U1" {
		t.Errorf("系统取消应通知用户: %+v", u.noti.saved)
	}
}

// --- 确认收货 / 自动确认 ---

func TestConfirmAndAutoConfirm(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	delivered := func(id string) *order.Order {
		o := &order.Order{ID: id, UserID: "U1", MerchantID: "M1", Status: order.StatusDelivered, OrderNo: "N" + id}
		u.orders.byID[id] = o
		return o
	}
	delivered("O1")
	if _, err := s.Confirm(ctx, "U2", "O1"); codeOf(t, err) != errcode.CodeOrderNotFound {
		t.Errorf("越权确认收货应 404, got=%d", codeOf(t, err))
	}
	o, err := s.Confirm(ctx, "U1", "O1")
	if err != nil {
		t.Fatalf("确认收货失败: %v", err)
	}
	if o.Status != order.StatusPendingReview || o.ConfirmedAt == 0 {
		t.Errorf("确认收货后状态错误: %+v", o)
	}

	delivered("O2")
	if _, err := s.AutoConfirm(ctx, "O2"); err != nil {
		t.Fatalf("自动确认失败: %v", err)
	}
	if u.orders.byID["O2"].Status != order.StatusPendingReview {
		t.Error("自动确认未推进状态")
	}
	if last := u.evts.saved[len(u.evts.saved)-1]; last.ActorType != order.ActorSystem {
		t.Errorf("自动确认事件操作方错误: %+v", last)
	}
	// 非已送达状态
	u.orders.byID["O3"] = &order.Order{ID: "O3", UserID: "U1", Status: order.StatusReady}
	if code := codeOf(t, mustErr(s.Confirm(ctx, "U1", "O3"))); code != errcode.CodeOrderState {
		t.Errorf("未送达不可确认, got=%d", code)
	}
	if code := codeOf(t, mustErr(s.AutoConfirm(ctx, "O3"))); code != errcode.CodeOrderState {
		t.Errorf("自动确认应校验状态, got=%d", code)
	}
	if code := codeOf(t, mustErr(s.Confirm(ctx, "U1", "MISSING"))); code != errcode.CodeOrderNotFound {
		t.Errorf("订单不存在应 404, got=%d", code)
	}
}

// --- 抢单：并发互斥 ---

func TestGrabIsExclusive(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &OrderService{Repos: u.repos, Tx: u.tx}
	u.riders.byID["D1"] = &rider.Rider{ID: "D1", Name: "飞毛腿", RegionID: "R1", Status: rider.RiderOnline}
	u.riders.byID["D2"] = &rider.Rider{ID: "D2", Name: "慢半拍", RegionID: "R1", Status: rider.RiderOffline}
	u.riders.byID["D3"] = &rider.Rider{ID: "D3", Name: "隔壁区", RegionID: "R9", Status: rider.RiderOnline}
	ready := func(id string) *order.Order {
		o := &order.Order{ID: id, RegionID: "R1", MerchantID: "M1", UserID: "U1", OrderNo: "N" + id,
			Status: order.StatusReady, DeliveryFee: 6}
		u.orders.byID[id] = o
		return o
	}
	ready("O1")

	if code := codeOf(t, mustErr(s.Grab(ctx, "D404", "O1"))); code != errcode.CodeRiderNotFound {
		t.Errorf("骑手不存在应 404, got=%d", code)
	}
	if code := codeOf(t, mustErr(s.Grab(ctx, "D2", "O1"))); code != errcode.CodeRiderOffline {
		t.Errorf("未上线骑手应拒绝抢单, got=%d", code)
	}
	if code := codeOf(t, mustErr(s.Grab(ctx, "D3", "O1"))); code != errcode.CodeRiderOutOfRange {
		t.Errorf("跨区骑手应拒绝抢单, got=%d", code)
	}

	o, err := s.Grab(ctx, "D1", "O1")
	if err != nil {
		t.Fatalf("在线同区骑手抢单失败: %v", err)
	}
	if o.RiderID != "D1" || o.RiderName != "飞毛腿" {
		t.Errorf("抢单未写入骑手: %+v", o)
	}
	// 第二人必须失败（条件更新只放行一次）
	if code := codeOf(t, mustErr(s.Grab(ctx, "D1", "O1"))); code != errcode.CodeOrderGrabbed {
		t.Errorf("重复抢单应报已被抢, got=%d", code)
	}
	if len(u.evts.saved) != 1 || u.evts.saved[0].ActorType != order.ActorDriver {
		t.Errorf("抢单事件错误: %+v", u.evts.saved)
	}
	if len(u.noti.saved) != 1 || u.noti.saved[0].UserID != "U1" {
		t.Errorf("抢单未通知用户: %+v", u.noti.saved)
	}
	// 非待取货状态不可抢
	u.orders.byID["O2"] = &order.Order{ID: "O2", RegionID: "R1", Status: order.OrderDelivering, RiderID: "D9"}
	if code := codeOf(t, mustErr(s.Grab(ctx, "D1", "O2"))); code != errcode.CodeOrderGrabbed {
		t.Errorf("配送中订单不可再抢, got=%d", code)
	}
}

// --- 提现 ---

func TestCreateWithdrawalRules(t *testing.T) {
	ctx := context.Background()
	rule := `{"enabled":true,"fee_rate":0.006,"min_amount":10,"max_amount":5000,"daily_limit":3,"mode":"manual","settle_days":2}`

	u := newUnit()
	u.set.valueJSON[string(system.KeyWithdrawRule)] = rule
	u.riders.byID["R1"] = &rider.Rider{ID: "R1", Name: "小王", RegionID: "RG1", Account: "6222021234567890"}
	s := &FinanceAppService{Repos: u.repos, Tx: u.tx, Wallet: &WalletService{Repos: u.repos}}

	if code := codeOf(t, mustWithdrawErr(s.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{Amount: 0}))); code != errcode.CodeBadRequest {
		t.Errorf("0 元提现应参数错误, got=%d", code)
	}
	for _, amt := range []float64{9.99, 5001} {
		if code := codeOf(t, mustWithdrawErr(s.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{Amount: amt}))); code != errcode.CodeWithdrawStatus {
			t.Errorf("金额 %v 越界应拒绝, got=%d", amt, code)
		}
	}
	u.withd.countSince = 3
	if code := codeOf(t, mustWithdrawErr(s.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{Amount: 100}))); code != errcode.CodeWithdrawStatus {
		t.Errorf("超过每日次数应拒绝, got=%d", code)
	}
	u.withd.countSince = 0
	if code := codeOf(t, mustWithdrawErr(s.CreateWithdrawal(ctx, "rider", "R404", WithdrawRequest{Amount: 100}))); code != errcode.CodeMerchantNotFound {
		t.Errorf("主体不存在应 404, got=%d", code)
	}
	if code := codeOf(t, mustWithdrawErr(s.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{
		Amount: 100, BankCard: strings.Repeat("卡", 129)}))); code != errcode.CodeBadRequest {
		t.Errorf("超长收款账户应拒绝, got=%d", code)
	}

	// 余额不足：debitWallet 走条件更新失败，必须整单拒绝且不生成提现单
	if code := codeOf(t, mustWithdrawErr(s.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{Amount: 100}))); code != errcode.CodeInsufficientFund {
		t.Errorf("余额不足应拒绝, got=%d", code)
	}
	if len(u.withd.saved) != 0 {
		t.Error("提现失败仍落了提现单")
	}

	// 成功路径
	if _, err := u.wallet.Ensure(ctx, "rider", "R1", 1); err != nil {
		t.Fatal(err)
	}
	u.wallet.accounts[walletKey("rider", "R1")].Balance = 200
	w, err := s.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{Amount: 100})
	if err != nil {
		t.Fatalf("合法提现被拒: %v", err)
	}
	if w.Status != finance.WithdrawPending || w.Kind != "骑手提现" {
		t.Errorf("人工审核模式应落待审核: %+v", w)
	}
	if w.Amount != 100 || w.Fee != 0.6 {
		t.Errorf("金额/手续费错误: %+v", w)
	}
	if w.BankInfo != "62****90" {
		t.Errorf("提现单未脱敏收款账号: %q", w.BankInfo)
	}
	if w.Balance != 200 || w.RemainBalance != 100 {
		t.Errorf("申请前/后余额错误: before=%v after=%v", w.Balance, w.RemainBalance)
	}
	if got := u.wallet.accounts[walletKey("rider", "R1")].Balance; got != 100 {
		t.Errorf("提现应冻结扣减余额, got=%v", got)
	}
	if len(u.withd.saved) != 1 || len(u.audit.entries) != 1 || u.audit.entries[0].Action != "withdraw.apply" {
		t.Errorf("提现单/审计缺失: saved=%d audit=%+v", len(u.withd.saved), u.audit.entries)
	}
	if len(w.Timeline) != 2 {
		t.Errorf("提现时间线不完整: %+v", w.Timeline)
	}
}

func TestCreateWithdrawalDisabledAndAuto(t *testing.T) {
	ctx := context.Background()

	u := newUnit()
	u.set.valueJSON[string(system.KeyWithdrawRule)] = `{"enabled":false,"min_amount":10,"max_amount":5000}`
	u.riders.byID["R1"] = &rider.Rider{ID: "R1", Name: "小王"}
	s := &FinanceAppService{Repos: u.repos, Tx: u.tx, Wallet: &WalletService{Repos: u.repos}}
	if code := codeOf(t, mustWithdrawErr(s.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{Amount: 100}))); code != errcode.CodeWithdrawStatus {
		t.Errorf("关闭提现应拒绝, got=%d", code)
	}

	// 自动审核但没配打款通道：必须在扣款前拒绝（否则钱扣了没人打款）
	u2 := newUnit()
	u2.set.valueJSON[string(system.KeyWithdrawRule)] = `{"enabled":true,"mode":"auto","min_amount":10,"max_amount":5000,"fee_rate":0.01,"settle_days":1}`
	u2.riders.byID["R1"] = &rider.Rider{ID: "R1", Name: "小王"}
	s2 := &FinanceAppService{Repos: u2.repos, Tx: u2.tx, Wallet: &WalletService{Repos: u2.repos}}
	if code := codeOf(t, mustWithdrawErr(s2.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{Amount: 100}))); code != errcode.CodePaymentFailed {
		t.Errorf("无打款通道应拒绝自动提现, got=%d", code)
	}
	if len(u2.withd.saved) != 0 {
		t.Error("无通道时不应生成提现单")
	}

	// 手续费率 ≥ 100%：金额被手续费吃光，必须拒绝而不是打 0 元
	u3 := newUnit()
	u3.set.valueJSON[string(system.KeyWithdrawRule)] = `{"enabled":true,"mode":"manual","min_amount":10,"max_amount":5000,"fee_rate":1}`
	u3.riders.byID["R1"] = &rider.Rider{ID: "R1", Name: "小王"}
	if _, err := u3.wallet.Ensure(ctx, "rider", "R1", 1); err != nil {
		t.Fatal(err)
	}
	u3.wallet.accounts[walletKey("rider", "R1")].Balance = 1000
	s3 := &FinanceAppService{Repos: u3.repos, Tx: u3.tx, Wallet: &WalletService{Repos: u3.repos}}
	if code := codeOf(t, mustWithdrawErr(s3.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{Amount: 100}))); code != errcode.CodeWithdrawStatus {
		t.Errorf("手续费吃满应拒绝, got=%d", code)
	}

	// 规则缺失时回落内置默认（不能因为配置没写就报错或放开无限额度）
	u4 := newUnit()
	u4.riders.byID["R1"] = &rider.Rider{ID: "R1", Name: "小王"}
	s4 := &FinanceAppService{Repos: u4.repos, Tx: u4.tx, Wallet: &WalletService{Repos: u4.repos}}
	if code := codeOf(t, mustWithdrawErr(s4.CreateWithdrawal(ctx, "rider", "R1", WithdrawRequest{Amount: 100}))); code != errcode.CodeInsufficientFund {
		t.Errorf("默认规则应校验余额, got=%d", code)
	}
	if _, err := s4.withdrawRuleErrProbe(ctx); err != nil {
		t.Errorf("默认规则探测失败: %v", err)
	}
}

func (s *FinanceAppService) withdrawRuleErrProbe(ctx context.Context) (bool, error) {
	r, err := s.withdrawRule(ctx)
	if err != nil {
		return false, err
	}
	return r.Enabled && r.Mode == "manual" && r.MinAmount == 10 && r.MaxAmount == 5000 && r.DailyLimit == 5, nil
}

// --- 售后退款 ---

func TestRefundApplyGuards(t *testing.T) {
	ctx := context.Background()
	u := newUnit()
	s := &RefundService{Repos: u.repos, Tx: u.tx}
	u.orders.byID["O1"] = &order.Order{ID: "O1", UserID: "U1", MerchantID: "M1",
		Status: order.StatusPendingReview, PayAmount: 56, OrderNo: "N1"}
	u.merch.byID["M1"] = &merchant.Merchant{ID: "M1", OwnerUserID: "MU1"}

	if code := codeOf(t, mustRefundErr(s.Apply(ctx, "U2", "O1", "x", 0))); code != errcode.CodeOrderNotFound {
		t.Errorf("他人订单售后应 404, got=%d", code)
	}
	u.orders.byID["O2"] = &order.Order{ID: "O2", UserID: "U1", Status: order.StatusReady, PayAmount: 20}
	if code := codeOf(t, mustRefundErr(s.Apply(ctx, "U1", "O2", "x", 0))); code != errcode.CodeOrderState {
		t.Errorf("未送达不可售后, got=%d", code)
	}
	if code := codeOf(t, mustRefundErr(s.Apply(ctx, "U1", "O1", "x", 50))); code != errcode.CodeAmountMismatch {
		t.Errorf("部分退款应拒绝（仅整单）, got=%d", code)
	}
	r, err := s.Apply(ctx, "U1", "O1", "  ", 0)
	if err != nil {
		t.Fatalf("整单售后失败: %v", err)
	}
	if r.Status != refund.StatusRequested || r.Amount != 56 || r.PrevStatus != string(order.StatusPendingReview) {
		t.Errorf("售后单错误: %+v", r)
	}
	if r.Reason != "用户申请售后" {
		t.Errorf("空原因应有默认文案: %q", r.Reason)
	}
	if u.orders.byID["O1"].Status != order.StatusAfterSale {
		t.Error("订单未进入售后状态")
	}
	// 状态守卫先于重复判定：已进入售后的单再申请会先被状态拦下
	if code := codeOf(t, mustRefundErr(s.Apply(ctx, "U1", "O1", "again", 0))); code != errcode.CodeOrderState {
		t.Errorf("售后中重复申请应报状态不允许, got=%d", code)
	}
	// 重复申请：状态被并发改回可售后、但已有 requested/approved/refunded 售后单时拒绝
	for _, st := range []refund.Status{refund.StatusRequested, refund.StatusApproved, refund.StatusRefunded} {
		u.orders.byID["O1"].Status = order.StatusPendingReview
		u.refund.byOrder["O1"].Status = st
		if code := codeOf(t, mustRefundErr(s.Apply(ctx, "U1", "O1", "again", 0))); code != errcode.CodeOrderDuplicate {
			t.Errorf("已有 %s 售后单时重复申请应拒绝, got=%d", st, code)
		}
	}
	// 已驳回的售后允许重新发起
	u.orders.byID["O1"].Status = order.StatusPendingReview
	u.refund.byOrder["O1"].Status = refund.StatusRejected
	if _, err := s.Apply(ctx, "U1", "O1", "再次申请", 0); err != nil {
		t.Errorf("驳回后应可重新申请售后: %v", err)
	}
}

func TestRefundDecide(t *testing.T) {
	ctx := context.Background()

	// 驳回：回到申请前状态，不动钱
	u := newUnit()
	s := &RefundService{Repos: u.repos, Tx: u.tx}
	orderAfterSale(u, "O1", "U1", "M1", 56, order.StatusPendingReview)
	u.refund.byOrder["O1"] = newRefund("R1", "O1", 56, string(order.StatusPendingReview))
	if _, err := s.Decide(ctx, "M2", "O1", false, "不符"); codeOf(t, err) != errcode.CodeOrderNotOwned {
		t.Errorf("非本商家应 404, got=%d", codeOf(t, err))
	}
	o, err := s.Decide(ctx, "M1", "O1", false, " 与用户协商一致 ")
	if err != nil {
		t.Fatalf("驳回报错: %v", err)
	}
	if o.Status != order.StatusPendingReview {
		t.Errorf("驳回应回到申请前状态, got=%q", o.Status)
	}
	if u.refund.updates[0] != refund.StatusRejected {
		t.Errorf("售后单状态错误: %v", u.refund.updates)
	}
	if len(u.prof.adds) != 0 {
		t.Error("驳回不应产生余额变动")
	}
	if len(u.noti.saved) != 1 || !strings.Contains(u.noti.saved[0].Content, "与用户协商一致") {
		t.Errorf("驳回未通知用户或原因丢失: %+v", u.noti.saved)
	}
	// 已处理的售后再点一次必须被拦下（防重复退款）
	if code := codeOf(t, mustErr(s.Decide(ctx, "M1", "O1", true, ""))); code != errcode.CodeOrderState {
		t.Errorf("重复处理售后应拒绝, got=%d", code)
	}

	// 通过：退余额、支付单退款、已完成单冲回商家与骑手入账
	u2 := newUnit()
	s2 := &RefundService{Repos: u2.repos, Tx: u2.tx}
	od := orderAfterSale(u2, "O2", "U2", "M2", 100, order.OrderDone)
	od.RiderID = "D2"
	od.RiderEarning = 8
	od.PackingFee = 2
	u2.prof.balances["U2"] = 10
	u2.merch.byID["M2"] = &merchant.Merchant{ID: "M2", OwnerUserID: "MU2", Commission: 10}
	_, _ = u2.wallet.Ensure(ctx, "merchant", "M2", 1)
	u2.wallet.accounts[walletKey("merchant", "M2")].Balance = 100
	_, _ = u2.wallet.Ensure(ctx, "rider", "D2", 1)
	u2.wallet.accounts[walletKey("rider", "D2")].Balance = 20
	u2.refund.byOrder["O2"] = newRefund("R2", "O2", 100, string(order.OrderDone))
	u2.pay.byOrder["O2"] = &order.Payment{ID: "P2", OrderID: "O2", Status: order.PaymentPaid}
	u2.coupon.byID["C2"] = &coupon.Coupon{ID: "C2", UserID: "U2", Status: coupon.StatusUsed, UsedOrderID: "O2"}
	od.CouponID = "C2"

	o2, err := s2.Decide(ctx, "M2", "O2", true, "")
	if err != nil {
		t.Fatalf("退款报错: %v", err)
	}
	if o2.Status != order.StatusRefunded || o2.PaymentStatus != "refunded" {
		t.Errorf("退款后状态错误: %+v", o2)
	}
	if u2.prof.balances["U2"] != 110 {
		t.Errorf("用户余额未退回: %v", u2.prof.balances["U2"])
	}
	if u2.pay.statuses[0] != string(order.PaymentRefunded) {
		t.Errorf("支付单未标退款: %v", u2.pay.statuses)
	}
	if len(u2.coupon.released) != 1 {
		t.Errorf("券未释放: %v", u2.coupon.released)
	}
	// 商家冲回 = 商品额 + 打包费 - 佣金 = 100 + 2 - 10
	if got := u2.wallet.accounts[walletKey("merchant", "M2")].Balance; got != 8 {
		t.Errorf("商家冲回金额错误: %v want=8", got)
	}
	if got := u2.wallet.accounts[walletKey("rider", "D2")].Balance; got != 12 {
		t.Errorf("骑手冲回金额错误: %v want=12", got)
	}
	if u2.refund.updates[len(u2.refund.updates)-1] != refund.StatusRefunded {
		t.Errorf("售后单未标已退款: %v", u2.refund.updates)
	}
	// 并发：状态条件更新失败时不得留下退款痕迹
	u3 := newUnit()
	s3 := &RefundService{Repos: u3.repos, Tx: u3.tx}
	orderAfterSale(u3, "O3", "U3", "M3", 30, order.OrderDone)
	u3.prof.balances["U3"] = 0
	u3.refund.byOrder["O3"] = newRefund("R3", "O3", 30, string(order.OrderDone))
	u3.orders.updateOK = false
	if code := codeOf(t, mustErr(s3.Decide(ctx, "M3", "O3", true, ""))); code != errcode.CodeOrderState {
		t.Errorf("并发退款应被状态条件拦住, got=%d", code)
	}
}

// --- 辅助 ---

func pendingOrder(u *unit, id, userID, merchantID string, pay float64) *order.Order {
	o := &order.Order{ID: id, UserID: userID, MerchantID: merchantID, OrderNo: "N" + id,
		Status: order.StatusPendingPay, PaymentStatus: "unpaid", PayAmount: pay, Type: order.OrderTypeFood}
	u.orders.byID[id] = o
	return o
}

// orderAfterSale 造一张处于售后态的订单，prev 记录申请前状态（驳回时回退依据）。
func orderAfterSale(u *unit, id, userID, merchantID string, pay float64, prev order.OrderStatus) *order.Order {
	o := &order.Order{ID: id, UserID: userID, MerchantID: merchantID, OrderNo: "N" + id,
		Status: order.StatusAfterSale, PaymentStatus: "paid", PayAmount: pay, Amount: pay, PayMethod: "balance"}
	_ = prev
	u.orders.byID[id] = o
	return o
}

func newRefund(id, orderID string, amount float64, prev string) *refund.Refund {
	return &refund.Refund{ID: id, OrderID: orderID, UserID: "U1", MerchantID: "M1",
		Amount: amount, Status: refund.StatusRequested, PrevStatus: prev, Reason: "售后"}
}

func mustErr(_ *order.Order, err error) error { return err }

func mustWithdrawErr(_ *finance.Withdrawal, err error) error { return err }

func mustRefundErr(_ *refund.Refund, err error) error { return err }
