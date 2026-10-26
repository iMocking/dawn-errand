package order

import "testing"

// R5-5：状态机是全链路唯一的写入口径（UpdateStatus 的 from→to 条件更新以此为准）。
// 这里冻结允许与禁止两类边，任何新增状态若漏配 transitions，非法流转用例会先红。
func TestCanTransition(t *testing.T) {
	allowed := [][2]OrderStatus{
		{StatusPendingPay, OrderPending},
		{StatusPendingPay, OrderCanceled},
		{OrderPending, StatusPreparing},
		{OrderPending, OrderCanceled},
		{StatusPreparing, StatusReady},
		{StatusPreparing, OrderCanceled},
		{StatusReady, OrderDelivering},
		{StatusReady, OrderCanceled}, // 骑手取货前仍可取消
		{OrderDelivering, StatusDelivered},
		{OrderDelivering, OrderAbnormal},
		{StatusDelivered, StatusPendingReview},
		{StatusPendingReview, OrderDone},
		{StatusPendingReview, StatusAfterSale},
		{OrderDone, StatusAfterSale},
	}
	for _, e := range allowed {
		if !CanTransition(e[0], e[1]) {
			t.Errorf("允许流转被拒: %s -> %s", e[0], e[1])
		}
	}

	denied := [][2]OrderStatus{
		{StatusPendingPay, StatusPreparing},    // 未支付不能直接备餐
		{StatusPendingPay, OrderDelivering},    // 跳步
		{OrderPending, StatusReady},            // 跳步
		{OrderPending, OrderDelivering},        // 跳步
		{StatusPreparing, OrderDelivering},     // 商家未出餐不可取货
		{StatusPreparing, OrderDone},           // 直达完成
		{OrderDelivering, OrderCanceled},       // 配送中只能走异常
		{OrderDelivering, StatusPendingReview}, // 未送达不可评价
		{StatusDelivered, OrderDelivering},     // 不可回退
		{StatusDelivered, OrderCanceled},       // 已送达不可取消
		{OrderDone, StatusDelivered},           // 终态后不可回退
		{OrderDone, OrderCanceled},             // 完成后只能走售后
		{OrderCanceled, OrderPending},          // 取消是终态
		{StatusAfterSale, OrderDone},           // 售后是终态（退款由 refund 域推进）
		{StatusRefunded, OrderPending},         // 退款是终态
		{OrderPicked, StatusDelivered},         // 外卖链路不使用「已取件待送达」
		{OrderCookingOot, StatusReady},         // 出餐超时不参与流转
		{"", OrderPending},                     // 空状态
		{OrderPending, ""},                     // 空目标
	}
	for _, e := range denied {
		if CanTransition(e[0], e[1]) {
			t.Errorf("非法流转被放行: %s -> %s", e[0], e[1])
		}
	}
}

// R5-5：正向主流程逐格走通，保证 transitions 没有断点（断点会让订单卡在中间态）。
func TestHappyPathChain(t *testing.T) {
	path := []OrderStatus{
		StatusPendingPay, OrderPending, StatusPreparing, StatusReady,
		OrderDelivering, StatusDelivered, StatusPendingReview, OrderDone,
	}
	for i := 0; i+1 < len(path); i++ {
		if !CanTransition(path[i], path[i+1]) {
			t.Fatalf("主流程断裂: %s -> %s", path[i], path[i+1])
		}
	}
	if !CanTransition(OrderDone, StatusAfterSale) {
		t.Fatal("完成后应可进入售后")
	}
}

// R5-5：App 机器码是前端渲染依据，多端共用；同码合并（取件/配送中同为 delivering）需固定。
func TestCode(t *testing.T) {
	cases := []struct {
		status OrderStatus
		want   string
	}{
		{StatusPendingPay, "pending_pay"},
		{OrderPending, "pending_accept"},
		{StatusPreparing, "preparing"},
		{StatusReady, "ready"},
		{OrderDelivering, "delivering"},
		{OrderPicked, "delivering"},
		{StatusDelivered, "delivered"},
		{StatusPendingReview, "pending_review"},
		{OrderDone, "completed"},
		{OrderCanceled, "canceled"},
		{StatusAfterSale, "after_sale"},
		{StatusRefunded, "refunded"},
		{OrderAbnormal, "abnormal"},
		{OrderCookingOot, "abnormal"},
		{OrderStatus("未知状态"), "未知状态"}, // 兜底：原样透出，不静默变空
	}
	for _, c := range cases {
		if got := Code(c.status); got != c.want {
			t.Errorf("Code(%q)=%q want=%q", c.status, got, c.want)
		}
	}
}

// R5-5：Tab 过滤直接决定「谁的单在谁的列表里」可见范围，且 all/未知 tab 必须放行全部（nil）。
func TestTabStatuses(t *testing.T) {
	clientCases := map[string][]string{
		"ongoing":        {string(StatusPendingPay), string(OrderPending), string(StatusPreparing), string(StatusReady), string(OrderDelivering), string(StatusDelivered)},
		"pending_review": {string(StatusPendingReview)},
		"completed":      {string(OrderDone)},
		"canceled":       {string(OrderCanceled)},
		"after_sale":     {string(StatusAfterSale), string(StatusRefunded)},
		"all":            nil,
		"":               nil,
		"bogus":          nil,
	}
	for tab, want := range clientCases {
		got := ClientTabStatuses(tab)
		if !sameStrings(got, want) {
			t.Errorf("ClientTabStatuses(%q)=%v want=%v", tab, got, want)
		}
	}

	storeCases := map[string][]string{
		"new":        {string(OrderPending)},
		"preparing":  {string(StatusPreparing)},
		"ready":      {string(StatusReady)},
		"delivering": {string(OrderDelivering)},
		"done":       {string(StatusDelivered), string(StatusPendingReview), string(OrderDone)},
		"canceled":   {string(OrderCanceled)},
		"after_sale": {string(StatusAfterSale), string(StatusRefunded)},
		"":           nil,
	}
	for tab, want := range storeCases {
		if got := StoreTabStatuses(tab); !sameStrings(got, want) {
			t.Errorf("StoreTabStatuses(%q)=%v want=%v", tab, got, want)
		}
	}

	driverCases := map[string][]string{
		"pending":    {string(StatusReady)},
		"delivering": {string(OrderDelivering)},
		"done":       {string(StatusDelivered), string(StatusPendingReview), string(OrderDone)},
		"canceled":   {string(OrderCanceled), string(StatusRefunded), string(OrderAbnormal)},
		"":           nil,
	}
	for tab, want := range driverCases {
		if got := DriverTabStatuses(tab); !sameStrings(got, want) {
			t.Errorf("DriverTabStatuses(%q)=%v want=%v", tab, got, want)
		}
	}
}

// R5-5：处理中/完成类 Tab 不得混入终态（取消/退款单仍出现在待处理列表会导致重复操作）；
// after_sale Tab 例外——它按口径同时展示「退款售后」与「已退款」。
func TestTabsExcludeTerminalEverywhere(t *testing.T) {
	terminal := map[string]bool{string(OrderCanceled): true, string(StatusRefunded): true}
	for _, tab := range []string{"ongoing", "pending_review", "completed"} {
		for _, s := range ClientTabStatuses(tab) {
			if terminal[s] {
				t.Errorf("客户端 Tab %q 含终态 %s", tab, s)
			}
		}
	}
	for _, tab := range []string{"new", "preparing", "ready", "delivering", "done"} {
		for _, s := range StoreTabStatuses(tab) {
			if terminal[s] {
				t.Errorf("商家端 Tab %q 含终态 %s", tab, s)
			}
		}
	}
	// 抢单池口径由 ListPool 负责，Tab 侧只保证「待取货」单一状态，不含已分配后的配送中。
	if got := DriverTabStatuses("pending"); len(got) != 1 || got[0] != string(StatusReady) {
		t.Errorf("骑手端 pending Tab 应仅为待取货, got=%v", got)
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
