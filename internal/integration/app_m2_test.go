package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"dawn-errand/internal/infrastructure/ent/dish"
	entnotification "dawn-errand/internal/infrastructure/ent/notification"
	"dawn-errand/internal/infrastructure/ent/orderevent"
	"dawn-errand/internal/infrastructure/ent/payment"
	"dawn-errand/internal/infrastructure/ent/refund"
	"dawn-errand/internal/infrastructure/ent/review"
	entusercoupon "dawn-errand/internal/infrastructure/ent/usercoupon"
	"dawn-errand/internal/infrastructure/ent/wallettransaction"
)

// M2 集成用例：下单/支付/接单/出餐/抢单/取送/确认、幂等、取消退款、并发抢单、越权。

func orderTokens(t *testing.T) (clientTok, storeTok, driverTok string) {
	t.Helper()
	clientTok, _ = loginAs(t, "client.test@dawnerrand.local", "Client123456")
	storeTok, _ = loginAs(t, "store.owner@dawnerrand.local", "Store123456")
	driverTok, _ = loginAs(t, "rider.owner@dawnerrand.local", "Rider123456")
	return
}

func dishStock(t *testing.T, dishID string) int {
	t.Helper()
	d, err := client.Dish.Get(context.Background(), dishID)
	if err != nil {
		t.Fatalf("get dish %s: %v", dishID, err)
	}
	return d.Stock
}

// cleanupOrder 删除订单并补偿库存（已取消订单不再补偿，避免库存虚增）。
func cleanupOrder(t *testing.T, orderID string) {
	t.Helper()
	ctx := context.Background()
	o, err := client.Order.Get(ctx, orderID)
	if err == nil {
		if string(o.Status) != "已取消" {
			var goods []struct {
				DishID string `json:"DishID"`
				Count  int    `json:"Count"`
			}
			_ = json.Unmarshal(o.Goods, &goods)
			for _, g := range goods {
				if g.DishID != "" && g.Count > 0 {
					_, _ = client.Dish.Update().Where(dish.ID(g.DishID)).AddStock(g.Count).Save(ctx)
				}
			}
		}
	}
	_, _ = client.OrderEvent.Delete().Where(orderevent.OrderID(orderID)).Exec(ctx)
	_, _ = client.Payment.Delete().Where(payment.OrderID(orderID)).Exec(ctx)
	_, _ = client.Notification.Delete().Where(entnotification.RefID(orderID)).Exec(ctx)
	_, _ = client.Review.Delete().Where(review.OrderID(orderID)).Exec(ctx)
	_, _ = client.Refund.Delete().Where(refund.OrderID(orderID)).Exec(ctx)
	// 释放该订单占用的优惠券。
	_ = client.UserCoupon.Update().
		Where(entusercoupon.UsedOrderID(orderID)).
		SetStatus("unused").SetUsedOrderID("").Exec(ctx)
	// 冲回该订单产生的钱包入账/出账，避免测试污染余额。
	txns, _ := client.WalletTransaction.Query().Where(wallettransaction.OrderID(orderID)).All(ctx)
	for _, txn := range txns {
		_ = client.WalletAccount.UpdateOneID(txn.AccountID).AddBalance(-txn.Amount).Exec(ctx)
	}
	_, _ = client.WalletTransaction.Delete().Where(wallettransaction.OrderID(orderID)).Exec(ctx)
	_ = client.Order.DeleteOneID(orderID).Exec(ctx)
}

func createOrderHTTP(t *testing.T, tok, clientOrderNo string, items []map[string]any) string {
	t.Helper()
	r, _ := do(t, http.MethodPost, "/api/app/client/orders", tok, map[string]any{
		"type": "外卖", "merchant_id": "M00268", "address_id": "ADDR001",
		"items": items, "remark": "M2 itest", "client_order_no": clientOrderNo,
	})
	if r.Code != 0 {
		t.Fatalf("create order code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	return body.ID
}

func payMock(t *testing.T, tok, orderID string) *resp {
	t.Helper()
	r, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/pay", tok, map[string]string{"channel": "mock"})
	return r
}

func readyOrder(t *testing.T, clientTok, storeTok, orderID string) (pickupCode, deliveryCode string) {
	t.Helper()
	if r, _ := do(t, http.MethodPost, "/api/app/store/orders/"+orderID+"/accept", storeTok, nil); r.Code != 0 {
		t.Fatalf("accept code=%d msg=%s", r.Code, r.Message)
	}
	r, _ := do(t, http.MethodPost, "/api/app/store/orders/"+orderID+"/ready", storeTok, nil)
	if r.Code != 0 {
		t.Fatalf("ready code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		PickupCode   string `json:"pickup_code"`
		DeliveryCode string `json:"delivery_code"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	if body.PickupCode == "" || body.DeliveryCode == "" || body.PickupCode == body.DeliveryCode {
		t.Fatalf("取送码不合法: %+v", body)
	}
	return body.PickupCode, body.DeliveryCode
}

// TestOrderLifecycle 全链路 + 下单幂等 + 支付幂等 + 取送码校验。
func TestOrderLifecycle(t *testing.T) {
	clientTok, storeTok, driverTok := orderTokens(t)
	no := fmt.Sprintf("m2-life-%d", time.Now().UnixNano())

	stockBefore := dishStock(t, "D001")
	orderID := createOrderHTTP(t, clientTok, no, []map[string]any{
		{"dish_id": "D001", "count": 1}, {"dish_id": "D005", "count": 2},
	})
	// 幂等：同一 client_order_no 再下同内容订单，返回同一单且不重复扣库存。
	dup := createOrderHTTP(t, clientTok, no, []map[string]any{{"dish_id": "D001", "count": 1}})
	if dup != orderID {
		t.Fatalf("幂等失败: %s != %s", dup, orderID)
	}
	if got := dishStock(t, "D001"); got != stockBefore-1 {
		t.Fatalf("库存扣减异常: before=%d after=%d", stockBefore, got)
	}
	defer cleanupOrder(t, orderID)

	if r, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/pay", clientTok,
		map[string]string{"channel": "mock"}); r.Code != 0 {
		t.Fatalf("pay code=%d msg=%s", r.Code, r.Message)
	}
	if r := payMock(t, clientTok, orderID); r.Code != 0 {
		t.Fatalf("重复支付应幂等, got %d", r.Code)
	}

	// 商家新单列表应包含该单。
	list, _ := do(t, http.MethodGet, "/api/app/store/orders?tab=new&page=1&page_size=20", storeTok, nil)
	if list.Code != 0 {
		t.Fatalf("store orders code=%d msg=%s", list.Code, list.Message)
	}
	if !bodyContainsID(t, list.Body, "list", orderID) {
		t.Fatalf("商家新单列表未包含订单 %s", orderID)
	}

	pickupCode, deliveryCode := readyOrder(t, clientTok, storeTok, orderID)

	// 抢单前骑手池应包含该单。
	pool, _ := do(t, http.MethodGet, "/api/app/driver/order-pool?lat=30.42&lng=115.33", driverTok, nil)
	if pool.Code != 0 {
		t.Fatalf("pool code=%d msg=%s", pool.Code, pool.Message)
	}
	if !bodyContainsID(t, pool.Body, "list", orderID) {
		t.Fatalf("抢单池未包含订单 %s", orderID)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/grab", driverTok, nil); r.Code != 0 {
		t.Fatalf("grab code=%d msg=%s", r.Code, r.Message)
	}
	// 已抢后再次抢单应 1801。
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/grab", driverTok, nil); r.Code != 1801 {
		t.Fatalf("重复抢单应 1801, got %d", r.Code)
	}
	// 错误取货码。
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/pickup", driverTok,
		map[string]string{"pickup_code": wrongCode(pickupCode)}); r.Code != 1804 {
		t.Fatalf("错误取货码应 1804, got %d", r.Code)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/pickup", driverTok,
		map[string]string{"pickup_code": pickupCode}); r.Code != 0 {
		t.Fatalf("pickup code=%d msg=%s", r.Code, r.Message)
	}
	// 错误送达码。
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/deliver", driverTok,
		map[string]string{"delivery_code": wrongCode(deliveryCode)}); r.Code != 1804 {
		t.Fatalf("错误送达码应 1804, got %d", r.Code)
	}
	deliver, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/deliver", driverTok,
		map[string]string{"delivery_code": deliveryCode})
	if deliver.Code != 0 {
		t.Fatalf("deliver code=%d msg=%s", deliver.Code, deliver.Message)
	}
	var delivered struct {
		RiderEarning float64 `json:"rider_earning"`
	}
	if err := json.Unmarshal(deliver.Body, &delivered); err != nil {
		t.Fatal(err)
	}
	if delivered.RiderEarning != 4 {
		t.Fatalf("骑手收入应为配送费 80%%=4, got %v", delivered.RiderEarning)
	}

	if r, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/confirm", clientTok, nil); r.Code != 0 {
		t.Fatalf("confirm code=%d msg=%s", r.Code, r.Message)
	}
	detail, _ := do(t, http.MethodGet, "/api/app/client/orders/"+orderID, clientTok, nil)
	var d struct {
		StatusCode string `json:"status_code"`
		Actions    []string
		Timeline   []map[string]any `json:"timeline"`
	}
	if err := json.Unmarshal(detail.Body, &d); err != nil {
		t.Fatal(err)
	}
	if d.StatusCode != "pending_review" {
		t.Fatalf("最终状态应为 pending_review, got %s", d.StatusCode)
	}
	if len(d.Timeline) < 8 {
		t.Fatalf("时间线事件不足: %d", len(d.Timeline))
	}
	if len(d.Actions) != 0 {
		t.Fatalf("待评价阶段暂无可执行动作（评价在 M4）, got %v", d.Actions)
	}
}

// TestOrderCancelAndRefund 取消释放库存、已支付退款、余额不足。
func TestOrderCancelAndRefund(t *testing.T) {
	clientTok, _, _ := orderTokens(t)
	stockBefore := dishStock(t, "D001")

	// 未支付取消：库存释放。
	no1 := fmt.Sprintf("m2-cancel-%d", time.Now().UnixNano())
	id1 := createOrderHTTP(t, clientTok, no1, []map[string]any{{"dish_id": "D001", "count": 2}})
	if got := dishStock(t, "D001"); got != stockBefore-2 {
		t.Fatalf("下单后库存应为 %d, got %d", stockBefore-2, got)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/client/orders/"+id1+"/cancel", clientTok,
		map[string]string{"reason": "计划有变"}); r.Code != 0 {
		t.Fatalf("cancel code=%d msg=%s", r.Code, r.Message)
	}
	if got := dishStock(t, "D001"); got != stockBefore {
		t.Fatalf("取消后库存应恢复为 %d, got %d", stockBefore, got)
	}
	defer cleanupOrder(t, id1)

	// 已支付取消：退款 + 支付单置 refunded。
	no2 := fmt.Sprintf("m2-refund-%d", time.Now().UnixNano())
	id2 := createOrderHTTP(t, clientTok, no2, []map[string]any{{"dish_id": "D001", "count": 1}})
	if r := payMock(t, clientTok, id2); r.Code != 0 {
		t.Fatalf("pay code=%d msg=%s", r.Code, r.Message)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/client/orders/"+id2+"/cancel", clientTok,
		map[string]string{"reason": "不想要了"}); r.Code != 0 {
		t.Fatalf("cancel paid code=%d msg=%s", r.Code, r.Message)
	}
	ctx := context.Background()
	o, err := client.Order.Get(ctx, id2)
	if err != nil {
		t.Fatal(err)
	}
	if string(o.Status) != "已取消" || o.PaymentStatus != "refunded" {
		t.Fatalf("退款状态异常: status=%s payment=%s", o.Status, o.PaymentStatus)
	}
	p, err := client.Payment.Query().Where(payment.OrderID(id2)).First(ctx)
	if err != nil || p.Status != "refunded" {
		t.Fatalf("支付单未退款: %v %+v", err, p)
	}
	if got := dishStock(t, "D001"); got != stockBefore {
		t.Fatalf("退款取消后库存应恢复为 %d, got %d", stockBefore, got)
	}
	defer cleanupOrder(t, id2)

	// 余额不足：client.test 余额为 0。
	no3 := fmt.Sprintf("m2-balance-%d", time.Now().UnixNano())
	id3 := createOrderHTTP(t, clientTok, no3, []map[string]any{{"dish_id": "D001", "count": 1}})
	r, _ := do(t, http.MethodPost, "/api/app/client/orders/"+id3+"/pay", clientTok,
		map[string]string{"channel": "balance"})
	if r.Code != 1902 {
		t.Fatalf("余额不足应 1902, got %d", r.Code)
	}
	o3, _ := client.Order.Get(ctx, id3)
	if string(o3.Status) != "待支付" {
		t.Fatalf("失败支付不应改变状态, got %s", o3.Status)
	}
	_, _ = do(t, http.MethodPost, "/api/app/client/orders/"+id3+"/cancel", clientTok, nil)
	defer cleanupOrder(t, id3)
}

// TestGrabConcurrency 两名骑手并发抢同一单，仅一人成功。
func TestGrabConcurrency(t *testing.T) {
	clientTok, storeTok, _ := orderTokens(t)
	admin := loginAdmin(t)
	ctx := context.Background()
	suffix := time.Now().UnixNano()

	// 第二名骑手 + 账号。
	rr, _ := do(t, http.MethodPost, "/api/riders", admin, map[string]any{
		"name": fmt.Sprintf("并发骑手%d", suffix), "phone": "13900006666",
		"area": "王家湾镇", "tool": "电动车",
	})
	if rr.Code != 0 {
		t.Fatalf("create rider code=%d msg=%s", rr.Code, rr.Message)
	}
	var rCreated struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rr.Body, &rCreated); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Rider.DeleteOneID(rCreated.ID).Exec(ctx) }()
	email := fmt.Sprintf("grab%d@dawnerrand.local", suffix)
	defer cleanupUserByEmail(t, email)
	if p, _ := do(t, http.MethodPost, "/api/riders/"+rCreated.ID+"/account", admin,
		map[string]string{"email": email, "password": "Prov123456"}); p.Code != 0 {
		t.Fatalf("provision rider code=%d msg=%s", p.Code, p.Message)
	}
	rider2, _ := loginAs(t, email, "Prov123456")
	// 新骑手需先通过审核。
	if a, _ := do(t, http.MethodPost, "/api/riders/audit", admin,
		map[string]any{"id": rCreated.ID, "action": "approve"}); a.Code != 0 {
		t.Fatalf("approve rider code=%d msg=%s", a.Code, a.Message)
	}
	if r, _ := do(t, http.MethodPut, "/api/app/driver/status", rider2, map[string]bool{"online": true}); r.Code != 0 {
		t.Fatalf("rider2 online code=%d msg=%s", r.Code, r.Message)
	}
	driver1, _ := loginAs(t, "rider.owner@dawnerrand.local", "Rider123456")

	no := fmt.Sprintf("m2-grab-%d", suffix)
	orderID := createOrderHTTP(t, clientTok, no, []map[string]any{{"dish_id": "D001", "count": 1}})
	defer cleanupOrder(t, orderID)
	if r := payMock(t, clientTok, orderID); r.Code != 0 {
		t.Fatalf("pay code=%d", r.Code)
	}
	readyOrder(t, clientTok, storeTok, orderID)

	type result struct {
		rider string
		code  int
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for name, tok := range map[string]string{"driver1": driver1, "driver2": rider2} {
		wg.Add(1)
		go func(name, tok string) {
			defer wg.Done()
			results <- result{rider: name, code: doCode(http.MethodPost, "/api/app/driver/orders/"+orderID+"/grab", tok)}
		}(name, tok)
	}
	wg.Wait()
	close(results)

	success, conflict := 0, 0
	winner := ""
	for res := range results {
		switch res.code {
		case 0:
			success++
			winner = res.rider
		case 1801:
			conflict++
		default:
			t.Fatalf("并发抢单返回异常 code=%d", res.code)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("并发抢单应 1 成功 1 冲突, got success=%d conflict=%d", success, conflict)
	}
	o, err := client.Order.Get(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}
	if o.RiderID == "" {
		t.Fatalf("抢单成功后 rider_id 为空（winner=%s）", winner)
	}
}

// TestOrderAccess 跨用户/跨端访问订单被拒绝。
func TestOrderAccess(t *testing.T) {
	clientTok, storeTok, driverTok := orderTokens(t)
	orderID := createOrderHTTP(t, clientTok, fmt.Sprintf("m2-access-%d", time.Now().UnixNano()),
		[]map[string]any{{"dish_id": "D001", "count": 1}})
	defer cleanupOrder(t, orderID)

	// 另一个用户（短信自动注册）无权查看。
	phone := randomPhone()
	defer cleanupUserByPhone(t, phone)
	send, _ := do(t, http.MethodPost, "/api/auth/sms/send", "", map[string]string{"phone": phone})
	var sent struct {
		DevCode string `json:"dev_code"`
	}
	_ = json.Unmarshal(send.Body, &sent)
	login, _ := do(t, http.MethodPost, "/api/auth/sms/login", "", map[string]string{"phone": phone, "code": sent.DevCode})
	var other struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(login.Body, &other)
	if r, _ := do(t, http.MethodGet, "/api/app/client/orders/"+orderID, other.Token, nil); r.Code != 1507 {
		t.Fatalf("他人订单应 1507, got %d", r.Code)
	}
	// 商家/骑手不能调用用户端订单详情。
	if r, _ := do(t, http.MethodGet, "/api/app/client/orders/"+orderID, storeTok, nil); r.Code != 1203 {
		t.Fatalf("商家调用户端应 1203, got %d", r.Code)
	}
	if r, _ := do(t, http.MethodGet, "/api/app/client/orders/"+orderID, driverTok, nil); r.Code != 1203 {
		t.Fatalf("骑手调用户端应 1203, got %d", r.Code)
	}
	// 商家不能处理非本店订单：新建一个商家账号并尝试接单。
	admin := loginAdmin(t)
	suffix := time.Now().UnixNano()
	mr, _ := do(t, http.MethodPost, "/api/merchants", admin, map[string]any{
		"name": fmt.Sprintf("越权商家%d", suffix), "phone": "13900007777", "category": "其他", "region_id": "area-001",
	})
	var m struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(mr.Body, &m)
	defer func() { _ = client.Merchant.DeleteOneID(m.ID).Exec(context.Background()) }()
	email := fmt.Sprintf("otherstore%d@dawnerrand.local", suffix)
	defer cleanupUserByEmail(t, email)
	_, _ = do(t, http.MethodPost, "/api/merchants/"+m.ID+"/account", admin,
		map[string]string{"email": email, "password": "Prov123456"})
	// 先过入驻审核，让请求推进到「本店归属」校验而不是停在经营资格守卫上。
	approveMerchant(t, m.ID)
	otherStore, _ := loginAs(t, email, "Prov123456")
	if r, _ := do(t, http.MethodPost, "/api/app/store/orders/"+orderID+"/accept", otherStore, nil); r.Code != 1705 {
		t.Fatalf("跨店接单应 1705, got %d", r.Code)
	}
}

// doCode 发送请求并返回业务码（用于并发断言，不使用 t.Fatal）。
func doCode(method, path, token string) int {
	setup()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("X-Request-ID", "itest-req-001")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var r resp
	if err := json.NewDecoder(w.Body).Decode(&r); err != nil {
		return -1
	}
	return r.Code
}

func wrongCode(code string) string {
	if code == "0000" {
		return "1111"
	}
	return "0000"
}

func bodyContainsID(t *testing.T, body json.RawMessage, field, id string) bool {
	t.Helper()
	var wrapper struct {
		List []map[string]any `json:"list"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		t.Fatal(err)
	}
	for _, row := range wrapper.List {
		if row["id"] == id || row["order_id"] == id {
			return true
		}
	}
	return false
}
