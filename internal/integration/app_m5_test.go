package integration

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"dawn-errand/internal/config"
	entusercoupon "dawn-errand/internal/infrastructure/ent/usercoupon"
)

// M5 集成用例：优惠券、单日对账、后台事件步骤。

func TestCouponFlow(t *testing.T) {
	clientTok, _, _ := orderTokens(t)
	ctx := context.Background()
	// 重置两张联调券。
	_ = client.UserCoupon.Update().
		Where(entusercoupon.IDIn("CPN001", "CPN002")).
		SetStatus("unused").SetUsedOrderID("").Exec(ctx)
	defer func() {
		_ = client.UserCoupon.Update().
			Where(entusercoupon.IDIn("CPN001", "CPN002")).
			SetStatus("unused").SetUsedOrderID("").Exec(ctx)
	}()

	list, _ := do(t, http.MethodGet, "/api/app/client/coupons?status=available", clientTok, nil)
	var lb struct {
		Total int `json:"total"`
	}
	_ = json.Unmarshal(list.Body, &lb)
	if lb.Total < 2 {
		t.Fatalf("可用券应至少 2 张: %s", list.Body)
	}

	// 满 15 减 3。
	orderID := createOrderWithCoupon(t, clientTok, "CPN001", []map[string]any{
		{"dish_id": "D001", "count": 1}, {"dish_id": "D005", "count": 2},
	})
	defer cleanupOrder(t, orderID)
	detail, _ := do(t, http.MethodGet, "/api/app/client/orders/"+orderID, clientTok, nil)
	var od struct {
		Pricing struct {
			PayAmount      float64 `json:"pay_amount"`
			DiscountAmount float64 `json:"discount_amount"`
		} `json:"pricing"`
	}
	if err := json.Unmarshal(detail.Body, &od); err != nil {
		t.Fatal(err)
	}
	if od.Pricing.PayAmount != 32 || od.Pricing.DiscountAmount != 3 {
		t.Fatalf("券后价格应为 32/优惠 3, got %+v", od.Pricing)
	}
	used, _ := do(t, http.MethodGet, "/api/app/client/coupons?status=used", clientTok, nil)
	if !listContains(t, used.Body, "id", "CPN001") {
		t.Fatalf("券未被占用: %s", used.Body)
	}
	// 取消释放券。
	if r, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/cancel", clientTok,
		map[string]string{"reason": "测试"}); r.Code != 0 {
		t.Fatalf("cancel code=%d", r.Code)
	}
	avail, _ := do(t, http.MethodGet, "/api/app/client/coupons?status=available", clientTok, nil)
	if !listContains(t, avail.Body, "id", "CPN001") {
		t.Fatalf("取消后券未释放: %s", avail.Body)
	}

	// 未达门槛 30 → 1605（不产生订单）。
	if r, _ := do(t, http.MethodPost, "/api/app/client/orders", clientTok, map[string]any{
		"type": "外卖", "merchant_id": "M00268", "address_id": "ADDR001",
		"items": []map[string]any{{"dish_id": "D001", "count": 1}}, "coupon_id": "CPN002",
		"client_order_no": fmt.Sprintf("m5-coupon-low-%d", time.Now().UnixNano()),
	}); r.Code != 1605 {
		t.Fatalf("未达门槛应 1605, got %d msg=%s", r.Code, r.Message)
	}

	// 平台券满 30 减 5。
	order2 := createOrderWithCoupon(t, clientTok, "CPN002", []map[string]any{{"dish_id": "D002", "count": 1}})
	defer cleanupOrder(t, order2)
	detail2, _ := do(t, http.MethodGet, "/api/app/client/orders/"+order2, clientTok, nil)
	var od2 struct {
		Pricing struct {
			PayAmount      float64 `json:"pay_amount"`
			DiscountAmount float64 `json:"discount_amount"`
		} `json:"pricing"`
	}
	_ = json.Unmarshal(detail2.Body, &od2)
	if od2.Pricing.PayAmount != 60 || od2.Pricing.DiscountAmount != 5 {
		t.Fatalf("平台券后价格应为 60/优惠 5, got %+v", od2.Pricing)
	}
}

func createOrderWithCoupon(t *testing.T, tok, couponID string, items []map[string]any) string {
	t.Helper()
	r, _ := do(t, http.MethodPost, "/api/app/client/orders", tok, map[string]any{
		"type": "外卖", "merchant_id": "M00268", "address_id": "ADDR001",
		"items": items, "coupon_id": couponID,
		"client_order_no": fmt.Sprintf("m5-coupon-%d", time.Now().UnixNano()),
	})
	if r.Code != 0 {
		t.Fatalf("create with coupon code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	return body.ID
}

// TestReconciliationAndAdminSteps 单日对账与后台订单步骤事件化。
func TestReconciliationAndAdminSteps(t *testing.T) {
	clientTok, storeTok, driverTok := orderTokens(t)
	admin := loginAdmin(t)

	orderID := createOrderHTTP(t, clientTok, fmt.Sprintf("m5-recon-%d", time.Now().UnixNano()),
		[]map[string]any{{"dish_id": "D001", "count": 1}})
	defer cleanupOrder(t, orderID)
	if r := payMock(t, clientTok, orderID); r.Code != 0 {
		t.Fatalf("pay code=%d", r.Code)
	}
	// 走到配送中，验证后台步骤来自事件。
	readyOrder(t, clientTok, storeTok, orderID)
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/grab", driverTok, nil); r.Code != 0 {
		t.Fatalf("grab code=%d", r.Code)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/pickup", driverTok,
		map[string]string{"pickup_code": pickupCodeFor(t, orderID)}); r.Code != 0 {
		t.Fatalf("pickup code=%d msg=%s", r.Code, r.Message)
	}

	detail, _ := do(t, http.MethodGet, "/api/orders/"+orderID, admin, nil)
	if detail.Code != 0 {
		t.Fatalf("admin order detail code=%d msg=%s", detail.Code, detail.Message)
	}
	var od struct {
		Steps []struct {
			Title string `json:"title"`
			State string `json:"state"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(detail.Body, &od); err != nil {
		t.Fatal(err)
	}
	if len(od.Steps) < 5 {
		t.Fatalf("后台步骤应来自事件（>=5）, got %d", len(od.Steps))
	}
	found := false
	for _, s := range od.Steps {
		if s.Title == "骑手取餐配送" {
			found = true
		}
	}
	if !found {
		t.Fatalf("后台步骤缺少骑手取餐: %+v", od.Steps)
	}

	// 单日对账：订单应付与支付单金额一致。
	date := time.Now().UTC().Format("2006-01-02")
	recon, _ := do(t, http.MethodGet, "/api/finance/reconciliation?date="+date, admin, nil)
	if recon.Code != 0 {
		t.Fatalf("reconciliation code=%d msg=%s", recon.Code, recon.Message)
	}
	var rc struct {
		Paid struct {
			Orders int64   `json:"orders"`
			Amount float64 `json:"amount"`
		} `json:"paid"`
		Payment struct {
			Count  int64   `json:"count"`
			Amount float64 `json:"amount"`
		} `json:"payment"`
		Diff          float64 `json:"diff"`
		MissingOrders int64   `json:"missing_orders"`
	}
	if err := json.Unmarshal(recon.Body, &rc); err != nil {
		t.Fatal(err)
	}
	if rc.Paid.Orders < 1 || rc.Payment.Count < 1 {
		t.Fatalf("对账未统计到订单/支付: %+v", rc)
	}
	if rc.Diff != 0 || rc.MissingOrders != 0 {
		t.Fatalf("对账差异应为 0: %+v", rc)
	}
}

// TestPaymentCallback 支付回调验签、幂等与金额校验。
func TestPaymentCallback(t *testing.T) {
	clientTok, _, _ := orderTokens(t)
	orderID := createOrderHTTP(t, clientTok, fmt.Sprintf("m5-callback-%d", time.Now().UnixNano()),
		[]map[string]any{{"dish_id": "D001", "count": 1}})
	defer cleanupOrder(t, orderID)

	detail, _ := do(t, http.MethodGet, "/api/app/client/orders/"+orderID, clientTok, nil)
	var od struct {
		Pricing struct {
			PayAmount float64 `json:"pay_amount"`
		} `json:"pricing"`
	}
	if err := json.Unmarshal(detail.Body, &od); err != nil {
		t.Fatal(err)
	}
	amount := od.Pricing.PayAmount
	ts := time.Now().Unix()
	secret := config.Load().PaymentCallbackSecret
	sign := func(orderID, tradeNo, status string, amount float64, ts int64) string {
		payload := fmt.Sprintf("%s|%s|%.2f|%s|%d", orderID, tradeNo, amount, status, ts)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(payload))
		return hex.EncodeToString(mac.Sum(nil))
	}
	body := map[string]any{
		"order_id": orderID, "trade_no": "MOCKCB001", "amount": amount,
		"status": "paid", "timestamp": ts, "sign": sign(orderID, "MOCKCB001", "paid", amount, ts),
	}
	cb, _ := do(t, http.MethodPost, "/api/payments/callback/wechat", "", body)
	if cb.Code != 0 {
		t.Fatalf("callback code=%d msg=%s", cb.Code, cb.Message)
	}
	after, _ := do(t, http.MethodGet, "/api/app/client/orders/"+orderID, clientTok, nil)
	var od2 struct {
		StatusCode string `json:"status_code"`
	}
	_ = json.Unmarshal(after.Body, &od2)
	if od2.StatusCode != "pending_accept" {
		t.Fatalf("回调后订单应待接单, got %s", od2.StatusCode)
	}
	// 重复回调幂等。
	dup, _ := do(t, http.MethodPost, "/api/payments/callback/wechat", "", body)
	if dup.Code != 0 {
		t.Fatalf("重复回调应幂等, got %d", dup.Code)
	}
	// 错误签名拒绝。
	bad := map[string]any{
		"order_id": orderID, "trade_no": "MOCKCB002", "amount": amount,
		"status": "paid", "timestamp": ts, "sign": "deadbeef",
	}
	if r, _ := do(t, http.MethodPost, "/api/payments/callback/wechat", "", bad); r.Code != 1201 {
		t.Fatalf("错误签名应 1201, got %d", r.Code)
	}
}

// TestReadyAndSecurityHeaders 就绪探针与安全响应头。
func TestReadyAndSecurityHeaders(t *testing.T) {
	r, h := do(t, http.MethodGet, "/readyz", "", nil)
	if r.Code != 0 {
		t.Fatalf("readyz code=%d msg=%s", r.Code, r.Message)
	}
	if h.Get("X-Content-Type-Options") != "nosniff" || h.Get("X-Frame-Options") != "DENY" {
		t.Fatalf("安全响应头缺失: %v", h)
	}
}

// TestAdminIssueCoupon 管理端发券并即时对用户可见。
func TestAdminIssueCoupon(t *testing.T) {
	admin := loginAdmin(t)
	clientTok, _, _ := orderTokens(t)
	userID := clientUserID(t)
	now := time.Now().Unix()
	r, _ := do(t, http.MethodPost, "/api/coupons", admin, map[string]any{
		"userId": userID, "title": "测试发放券", "amount": 2, "minSpend": 10,
		"expiresAt": now + 86400,
	})
	if r.Code != 0 {
		t.Fatalf("issue code=%d msg=%s", r.Code, r.Message)
	}
	var issued struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(r.Body, &issued)
	if issued.ID == "" {
		t.Fatalf("发放返回缺少 id: %s", r.Body)
	}
	defer func() { _ = client.UserCoupon.DeleteOneID(issued.ID).Exec(context.Background()) }()

	list, _ := do(t, http.MethodGet, "/api/app/client/coupons?status=available", clientTok, nil)
	if !listContains(t, list.Body, "id", issued.ID) {
		t.Fatalf("用户端未看到新券: %s", list.Body)
	}
	// 金额非法 → 1301。
	if bad, _ := do(t, http.MethodPost, "/api/coupons", admin, map[string]any{
		"userId": userID, "title": "非法券", "amount": 0,
	}); bad.Code != 1301 {
		t.Fatalf("非法金额应 1301, got %d", bad.Code)
	}
}

// pickupCodeFor 从司机端订单详情读取取货码（当前骑手已抢单）。
func pickupCodeFor(t *testing.T, orderID string) string {
	t.Helper()
	tok, _ := loginAs(t, "rider.owner@dawnerrand.local", "Rider123456")
	r, _ := do(t, http.MethodGet, "/api/app/driver/orders/"+orderID, tok, nil)
	if r.Code != 0 {
		t.Fatalf("driver detail code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		PickupCode string `json:"pickup_code"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	return body.PickupCode
}
