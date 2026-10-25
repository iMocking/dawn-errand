package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	entreview "dawn-errand/internal/infrastructure/ent/review"
	entsettlement "dawn-errand/internal/infrastructure/ent/settlement"
	entuser "dawn-errand/internal/infrastructure/ent/user"
	entuserprofile "dawn-errand/internal/infrastructure/ent/userprofile"
	entwallet "dawn-errand/internal/infrastructure/ent/walletaccount"
	entwallettx "dawn-errand/internal/infrastructure/ent/wallettransaction"
)

// M4 集成用例：评价与商家入账、提现与驳回退回、售后退款与冲回。

func completeOrder(t *testing.T, clientTok, storeTok, driverTok string) string {
	t.Helper()
	orderID := createOrderHTTP(t, clientTok, fmt.Sprintf("m4-%d", time.Now().UnixNano()),
		[]map[string]any{{"dish_id": "D001", "count": 1}})
	if r := payMock(t, clientTok, orderID); r.Code != 0 {
		t.Fatalf("pay code=%d msg=%s", r.Code, r.Message)
	}
	pickupCode, deliveryCode := readyOrder(t, clientTok, storeTok, orderID)
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/grab", driverTok, nil); r.Code != 0 {
		t.Fatalf("grab code=%d msg=%s", r.Code, r.Message)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/pickup", driverTok,
		map[string]string{"pickup_code": pickupCode}); r.Code != 0 {
		t.Fatalf("pickup code=%d msg=%s", r.Code, r.Message)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/deliver", driverTok,
		map[string]string{"delivery_code": deliveryCode}); r.Code != 0 {
		t.Fatalf("deliver code=%d msg=%s", r.Code, r.Message)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/confirm", clientTok, nil); r.Code != 0 {
		t.Fatalf("confirm code=%d msg=%s", r.Code, r.Message)
	}
	return orderID
}

func cleanupSettlements(t *testing.T, since int64) {
	t.Helper()
	_, _ = client.Settlement.Delete().
		Where(entsettlement.SettledAtGTE(since), entsettlement.PartyIDIn("M00268", "R00086")).
		Exec(context.Background())
}

func clientUserID(t *testing.T) string {
	t.Helper()
	u, err := client.User.Query().Where(entuser.EmailEqualFold("client.test@dawnerrand.local")).Only(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return u.ID
}

func setClientBalance(t *testing.T, balance float64) {
	t.Helper()
	_, err := client.UserProfile.Update().
		Where(entuserprofile.UserID(clientUserID(t))).
		SetBalance(balance).Save(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

// TestReviewAndFinance 评价完成订单、商家入账与结算、评价回复。
func TestReviewAndFinance(t *testing.T) {
	clientTok, storeTok, driverTok := orderTokens(t)
	start := time.Now().Unix()
	orderID := completeOrder(t, clientTok, storeTok, driverTok)
	defer cleanupOrder(t, orderID)
	defer cleanupSettlements(t, start)

	// 骑手送达入账：5 元配送费 × 80% = 4。
	riderFin, _ := do(t, http.MethodGet, "/api/app/driver/finance/summary", driverTok, nil)
	var rf struct {
		Balance     float64 `json:"balance"`
		TodayIncome float64 `json:"today_income"`
	}
	if err := json.Unmarshal(riderFin.Body, &rf); err != nil {
		t.Fatal(err)
	}
	if rf.Balance < 4 || rf.TodayIncome < 4 {
		t.Fatalf("骑手入账异常: %+v", rf)
	}
	records, _ := do(t, http.MethodGet, "/api/app/driver/finance/records", driverTok, nil)
	if !listContains(t, records.Body, "title", "配送收入") {
		t.Fatalf("骑手流水缺少配送收入: %s", records.Body)
	}

	// 评价。
	reviewResp, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/review", clientTok,
		map[string]any{"rating": 5, "tags": []string{"味道好", "送得快"}, "content": "很好"})
	if reviewResp.Code != 0 {
		t.Fatalf("review code=%d msg=%s", reviewResp.Code, reviewResp.Message)
	}
	var rv struct {
		Status string `json:"status"`
		Review struct {
			ID string `json:"id"`
		} `json:"review"`
	}
	if err := json.Unmarshal(reviewResp.Body, &rv); err != nil {
		t.Fatal(err)
	}
	if rv.Status != "已完成" || rv.Review.ID == "" {
		t.Fatalf("评价后订单/评价异常: %+v", rv)
	}
	// 重复评价 1608。
	if dup, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/review", clientTok,
		map[string]any{"rating": 4}); dup.Code != 1608 {
		t.Fatalf("重复评价应 1608, got %d", dup.Code)
	}

	// 商家评价列表 + 回复。
	list, _ := do(t, http.MethodGet, "/api/app/store/reviews?page=1&page_size=10", storeTok, nil)
	if list.Code != 0 || !listContains(t, list.Body, "content", "很好") {
		t.Fatalf("商家评价列表异常: code=%d body=%s", list.Code, list.Body)
	}
	if reply, _ := do(t, http.MethodPost, "/api/app/store/reviews/"+rv.Review.ID+"/reply", storeTok,
		map[string]string{"content": "感谢支持"}); reply.Code != 0 {
		t.Fatalf("reply code=%d msg=%s", reply.Code, reply.Message)
	}

	// 商家入账：商品 22 + 打包 2 - 佣金 22×12%=2.64 → 21.36。
	storeFin, _ := do(t, http.MethodGet, "/api/app/store/finance/summary", storeTok, nil)
	var sf struct {
		Balance     float64 `json:"balance"`
		TodayIncome float64 `json:"today_income"`
	}
	if err := json.Unmarshal(storeFin.Body, &sf); err != nil {
		t.Fatal(err)
	}
	if sf.Balance < 21.36 || sf.TodayIncome < 21.36 {
		t.Fatalf("商家入账异常: %+v", sf)
	}
	settles, _ := do(t, http.MethodGet, "/api/app/store/settlements", storeTok, nil)
	var st struct {
		Total int64 `json:"total"`
	}
	_ = json.Unmarshal(settles.Body, &st)
	if st.Total < 1 {
		t.Fatalf("商家结算单缺失: %s", settles.Body)
	}
}

// TestWithdrawalFlow 提现申请、管理端驳回与余额退回。
func TestWithdrawalFlow(t *testing.T) {
	_, storeTok, _ := orderTokens(t)
	admin := loginAdmin(t)
	ctx := context.Background()

	// 准备 100 元商家余额。
	_, _ = client.WalletAccount.Delete().
		Where(entwallet.PartyType("merchant"), entwallet.PartyID("M00268")).Exec(ctx)
	_, _ = client.WalletTransaction.Delete().
		Where(entwallettx.PartyType("merchant"), entwallettx.PartyID("M00268")).Exec(ctx)
	acc, err := client.WalletAccount.Create().
		SetPartyType("merchant").SetPartyID("M00268").SetBalance(100).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var withdrawID string
	var withdrawID2 string
	defer func() {
		_, _ = client.WalletTransaction.Delete().
			Where(entwallettx.PartyType("merchant"), entwallettx.PartyID("M00268")).Exec(ctx)
		_ = client.WalletAccount.DeleteOneID(acc.ID).Exec(ctx)
		if withdrawID != "" {
			_ = client.Withdrawal.DeleteOneID(withdrawID).Exec(ctx)
		}
		if withdrawID2 != "" {
			_ = client.Withdrawal.DeleteOneID(withdrawID2).Exec(ctx)
		}
	}()

	resp, _ := do(t, http.MethodPost, "/api/app/store/withdrawals", storeTok, map[string]any{
		"amount": 50, "channel": "bank", "account_name": "老陈", "bank_card": "6222021012345678",
	})
	if resp.Code != 0 {
		t.Fatalf("withdraw code=%d msg=%s", resp.Code, resp.Message)
	}
	var w struct {
		ID            string  `json:"id"`
		Fee           float64 `json:"fee"`
		Status        string  `json:"status"`
		RemainBalance float64 `json:"remain_balance"`
	}
	if err := json.Unmarshal(resp.Body, &w); err != nil {
		t.Fatal(err)
	}
	withdrawID = w.ID
	if w.Fee != 0.30 || w.RemainBalance != 50 || w.Status != "待审核" {
		t.Fatalf("提现返回异常: %+v", w)
	}
	// 余额被扣减。
	fin, _ := do(t, http.MethodGet, "/api/app/store/finance/summary", storeTok, nil)
	var fs struct {
		Balance         float64 `json:"balance"`
		PendingWithdraw float64 `json:"pending_withdraw"`
	}
	_ = json.Unmarshal(fin.Body, &fs)
	if fs.Balance != 50 || fs.PendingWithdraw != 50 {
		t.Fatalf("提现后余额异常: %+v", fs)
	}

	// 管理端驳回 → 余额退回。
	reject, _ := do(t, http.MethodPost, "/api/finance/withdraw/"+w.ID+"/audit", admin,
		map[string]string{"action": "reject", "remark": "测试驳回"})
	if reject.Code != 0 {
		t.Fatalf("audit code=%d msg=%s", reject.Code, reject.Message)
	}
	fin2, _ := do(t, http.MethodGet, "/api/app/store/finance/summary", storeTok, nil)
	var fs2 struct {
		Balance         float64 `json:"balance"`
		PendingWithdraw float64 `json:"pending_withdraw"`
	}
	_ = json.Unmarshal(fin2.Body, &fs2)
	if fs2.Balance != 100 || fs2.PendingWithdraw != 0 {
		t.Fatalf("驳回后余额应恢复 100: %+v", fs2)
	}
	got, err := client.Withdrawal.Get(ctx, w.ID)
	if err != nil || string(got.Status) != "已驳回" {
		t.Fatalf("提现单状态异常: %v %+v", err, got)
	}

	// 再次申请并通过 → Mock 打款回执。
	resp2, _ := do(t, http.MethodPost, "/api/app/store/withdrawals", storeTok, map[string]any{
		"amount": 20, "channel": "bank", "bank_card": "6222021012345678",
	})
	if resp2.Code != 0 {
		t.Fatalf("withdraw2 code=%d msg=%s", resp2.Code, resp2.Message)
	}
	var w2 struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(resp2.Body, &w2)
	withdrawID2 = w2.ID
	if approved, _ := do(t, http.MethodPost, "/api/finance/withdraw/"+w2.ID+"/audit", admin,
		map[string]string{"action": "pass"}); approved.Code != 0 {
		t.Fatalf("approve code=%d msg=%s", approved.Code, approved.Message)
	}
	got2, err := client.Withdrawal.Get(ctx, w2.ID)
	if err != nil || string(got2.Status) != "已通过" {
		t.Fatalf("通过后状态异常: %v %+v", err, got2)
	}
	w2Detail, _ := client.Withdrawal.Get(ctx, w2.ID)
	var timeline []map[string]any
	_ = json.Unmarshal(w2Detail.Timeline, &timeline)
	foundPayout := false
	for _, ev := range timeline {
		if ev["text"] == "打款成功" && ev["operator"] != "" {
			foundPayout = true
		}
	}
	if !foundPayout {
		t.Fatalf("通过后缺少打款回执: %s", w2Detail.Timeline)
	}
}

// TestAfterSaleRefund 售后退款：用户退款、商家/骑手入账冲回。
func TestAfterSaleRefund(t *testing.T) {
	clientTok, storeTok, driverTok := orderTokens(t)
	start := time.Now().Unix()
	baseline := 0.0
	if me, _ := do(t, http.MethodGet, "/api/app/client/me", clientTok, nil); me.Code == 0 {
		var body struct {
			Balance float64 `json:"balance"`
		}
		_ = json.Unmarshal(me.Body, &body)
		baseline = body.Balance
	}
	orderID := completeOrder(t, clientTok, storeTok, driverTok)
	defer cleanupOrder(t, orderID)
	defer cleanupSettlements(t, start)
	defer setClientBalance(t, baseline)

	rv, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/review", clientTok,
		map[string]any{"rating": 4, "content": "售后测试"})
	if rv.Code != 0 {
		t.Fatalf("review code=%d msg=%s", rv.Code, rv.Message)
	}
	// 申请售后。
	as, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/after-sale", clientTok,
		map[string]string{"reason": "餐品问题"})
	if as.Code != 0 {
		t.Fatalf("after-sale code=%d msg=%s", as.Code, as.Message)
	}
	detail, _ := do(t, http.MethodGet, "/api/app/client/orders/"+orderID, clientTok, nil)
	var od struct {
		StatusCode string `json:"status_code"`
	}
	_ = json.Unmarshal(detail.Body, &od)
	if od.StatusCode != "after_sale" {
		t.Fatalf("售后中状态应为 after_sale, got %s", od.StatusCode)
	}

	// 商家同意 → 退款完成。
	dec, _ := do(t, http.MethodPost, "/api/app/store/orders/"+orderID+"/refund/decision", storeTok,
		map[string]any{"approve": true})
	if dec.Code != 0 {
		t.Fatalf("refund decision code=%d msg=%s", dec.Code, dec.Message)
	}
	var decBody struct {
		Status        string `json:"status"`
		PaymentStatus string `json:"payment_status"`
	}
	_ = json.Unmarshal(dec.Body, &decBody)
	if decBody.Status != "已退款" || decBody.PaymentStatus != "refunded" {
		t.Fatalf("退款结果异常: %+v", decBody)
	}
	// 用户余额退回订单实付。
	me, _ := do(t, http.MethodGet, "/api/app/client/me", clientTok, nil)
	var meBody struct {
		Balance float64 `json:"balance"`
	}
	_ = json.Unmarshal(me.Body, &meBody)
	if meBody.Balance != baseline+29 {
		t.Fatalf("退款后余额应为 %.2f, got %.2f", baseline+29, meBody.Balance)
	}
	// 商家/骑手入账被冲回。
	sf, _ := do(t, http.MethodGet, "/api/app/store/finance/summary", storeTok, nil)
	var sfs struct {
		Balance float64 `json:"balance"`
	}
	_ = json.Unmarshal(sf.Body, &sfs)
	if sfs.Balance > 0.001 {
		t.Fatalf("商家余额应被冲回为 0, got %.2f", sfs.Balance)
	}
	df, _ := do(t, http.MethodGet, "/api/app/driver/finance/summary", driverTok, nil)
	var dfs struct {
		Balance float64 `json:"balance"`
	}
	_ = json.Unmarshal(df.Body, &dfs)
	if dfs.Balance > 0.001 {
		t.Fatalf("骑手余额应被冲回为 0, got %.2f", dfs.Balance)
	}
	// 评价仍存在（售后不回滚评价）。
	if _, err := client.Review.Query().Where(entreview.OrderID(orderID)).Only(context.Background()); err != nil {
		t.Fatalf("评价记录应保留: %v", err)
	}
}

func listContains(t *testing.T, body json.RawMessage, field, want string) bool {
	t.Helper()
	var wrapper struct {
		List []map[string]any `json:"list"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		t.Fatal(err)
	}
	for _, row := range wrapper.List {
		if fmt.Sprint(row[field]) == want {
			return true
		}
	}
	return false
}
