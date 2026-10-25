package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"dawn-errand/internal/infrastructure/ent/riderlocation"
)

// M3 集成用例：位置上报/追踪、站内通知、定时任务。

func TestTrackingAndNotifications(t *testing.T) {
	clientTok, storeTok, driverTok := orderTokens(t)
	orderID := createOrderHTTP(t, clientTok, fmt.Sprintf("m3-track-%d", time.Now().UnixNano()),
		[]map[string]any{{"dish_id": "D001", "count": 1}})
	defer cleanupOrder(t, orderID)
	defer func() {
		_, _ = client.RiderLocation.Delete().Where(riderlocation.RiderID("R00086")).Exec(context.Background())
	}()

	if r := payMock(t, clientTok, orderID); r.Code != 0 {
		t.Fatalf("pay code=%d msg=%s", r.Code, r.Message)
	}
	pickupCode, _ := readyOrder(t, clientTok, storeTok, orderID)
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/grab", driverTok, nil); r.Code != 0 {
		t.Fatalf("grab code=%d msg=%s", r.Code, r.Message)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+orderID+"/pickup", driverTok,
		map[string]string{"pickup_code": pickupCode}); r.Code != 0 {
		t.Fatalf("pickup code=%d msg=%s", r.Code, r.Message)
	}

	// 位置上报。
	if r, _ := do(t, http.MethodPost, "/api/app/driver/location", driverTok, map[string]any{
		"lat": 30.4255, "lng": 115.3320, "heading": 88.5, "speed": 18.2, "accuracy": 6.5,
	}); r.Code != 0 {
		t.Fatalf("location code=%d msg=%s", r.Code, r.Message)
	}
	// 越权上报：用户端不能调骑手接口。
	if r, _ := do(t, http.MethodPost, "/api/app/driver/location", clientTok, map[string]any{
		"lat": 30.4255, "lng": 115.3320,
	}); r.Code != 1203 {
		t.Fatalf("用户上报位置应 1203, got %d", r.Code)
	}

	track, _ := do(t, http.MethodGet, "/api/app/client/orders/"+orderID+"/track", clientTok, nil)
	if track.Code != 0 {
		t.Fatalf("track code=%d msg=%s", track.Code, track.Message)
	}
	var tr struct {
		Progress   float64 `json:"progress"`
		TraveledKM float64 `json:"traveled_km"`
		RemainKM   float64 `json:"remaining_km"`
		ETAMinutes int     `json:"eta_minutes"`
		RiderLat   float64 `json:"rider_lat"`
		RiderLng   float64 `json:"rider_lng"`
		Route      []struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"route"`
		Steps []map[string]any `json:"steps"`
		Rider map[string]any   `json:"rider"`
	}
	if err := json.Unmarshal(track.Body, &tr); err != nil {
		t.Fatal(err)
	}
	if tr.RiderLat == 0 || tr.RiderLng == 0 {
		t.Fatalf("追踪未返回骑手位置: %+v", tr)
	}
	if tr.Progress <= 0 || tr.Progress >= 1 {
		t.Fatalf("配送中进度应 (0,1), got %v", tr.Progress)
	}
	if tr.TraveledKM <= 0 || tr.RemainKM <= 0 || tr.ETAMinutes <= 0 {
		t.Fatalf("距离/ETA 异常: %+v", tr)
	}
	if len(tr.Route) < 2 || len(tr.Steps) < 4 {
		t.Fatalf("路线/步骤不足: route=%d steps=%d", len(tr.Route), len(tr.Steps))
	}
	if tr.Rider["phone"] == "13700008686" {
		t.Fatalf("骑手手机号未脱敏")
	}

	// 商家收到「新订单」通知；用户收到接单/骑手通知。
	storeNotes, _ := do(t, http.MethodGet, "/api/app/common/notifications?page=1&page_size=20", storeTok, nil)
	if storeNotes.Code != 0 || !notificationsContain(t, storeNotes.Body, "新订单待接单") {
		t.Fatalf("商家通知缺失: code=%d body=%s", storeNotes.Code, storeNotes.Body)
	}
	clientNotes, _ := do(t, http.MethodGet, "/api/app/common/notifications?page=1&page_size=20", clientTok, nil)
	if clientNotes.Code != 0 {
		t.Fatalf("client notifications code=%d", clientNotes.Code)
	}
	var notes struct {
		Unread int64 `json:"unread"`
		List   []struct {
			Title string `json:"title"`
			Read  bool   `json:"read"`
		} `json:"list"`
	}
	if err := json.Unmarshal(clientNotes.Body, &notes); err != nil {
		t.Fatal(err)
	}
	titles := map[string]bool{}
	for _, n := range notes.List {
		titles[n.Title] = true
	}
	if !titles["商家已接单"] || !titles["骑手已接单"] {
		t.Fatalf("用户通知缺失: %+v", titles)
	}
	if notes.Unread == 0 {
		t.Fatalf("未读数应 >0")
	}
	if r, _ := do(t, http.MethodPost, "/api/app/common/notifications/read-all", clientTok, nil); r.Code != 0 {
		t.Fatalf("read all code=%d msg=%s", r.Code, r.Message)
	}
	after, _ := do(t, http.MethodGet, "/api/app/common/notifications", clientTok, nil)
	var afterBody struct {
		Unread int64 `json:"unread"`
	}
	_ = json.Unmarshal(after.Body, &afterBody)
	if afterBody.Unread != 0 {
		t.Fatalf("全部已读后 unread 应为 0, got %d", afterBody.Unread)
	}
}

// TestJobsTimeoutAutoConfirmAndAnnounce 验证定时任务：超时取消、自动确认、公告发布。
func TestJobsTimeoutAutoConfirmAndAnnounce(t *testing.T) {
	setup()
	clientTok, storeTok, driverTok := orderTokens(t)
	ctx := context.Background()

	// 1) 待支付超时取消。
	unpaid := createOrderHTTP(t, clientTok, fmt.Sprintf("m3-timeout-%d", time.Now().UnixNano()),
		[]map[string]any{{"dish_id": "D001", "count": 1}})
	defer cleanupOrder(t, unpaid)
	now := time.Now().Unix()
	if _, err := client.Order.UpdateOneID(unpaid).SetCreatedAt(now - 3600).Save(ctx); err != nil {
		t.Fatal(err)
	}
	appInst.Jobs.RunOnce(ctx)
	o, err := client.Order.Get(ctx, unpaid)
	if err != nil {
		t.Fatal(err)
	}
	if string(o.Status) != "已取消" || o.CancelReason == "" {
		t.Fatalf("超时订单未取消: status=%s reason=%s", o.Status, o.CancelReason)
	}

	// 2) 送达后自动确认。
	deliveredID := createOrderHTTP(t, clientTok, fmt.Sprintf("m3-autoconfirm-%d", time.Now().UnixNano()),
		[]map[string]any{{"dish_id": "D001", "count": 1}})
	defer cleanupOrder(t, deliveredID)
	if r := payMock(t, clientTok, deliveredID); r.Code != 0 {
		t.Fatalf("pay code=%d", r.Code)
	}
	pickupCode, deliveryCode := readyOrder(t, clientTok, storeTok, deliveredID)
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+deliveredID+"/grab", driverTok, nil); r.Code != 0 {
		t.Fatalf("grab code=%d", r.Code)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+deliveredID+"/pickup", driverTok,
		map[string]string{"pickup_code": pickupCode}); r.Code != 0 {
		t.Fatalf("pickup code=%d", r.Code)
	}
	if r, _ := do(t, http.MethodPost, "/api/app/driver/orders/"+deliveredID+"/deliver", driverTok,
		map[string]string{"delivery_code": deliveryCode}); r.Code != 0 {
		t.Fatalf("deliver code=%d", r.Code)
	}
	if _, err := client.Order.UpdateOneID(deliveredID).SetDeliveredAt(now - 90000).Save(ctx); err != nil {
		t.Fatal(err)
	}
	appInst.Jobs.RunOnce(ctx)
	o2, _ := client.Order.Get(ctx, deliveredID)
	if string(o2.Status) != "待评价" {
		t.Fatalf("自动确认后应为待评价, got %s", o2.Status)
	}

	// 3) 到期公告发布。
	ann, err := client.Announcement.Create().
		SetTitle("M3 定时发布公告").
		SetStatus("待发布").
		SetPlannedAt(now - 60).
		SetCreatedAt(now - 120).
		SetUpdatedAt(now - 120).
		Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Announcement.DeleteOneID(ann.ID).Exec(ctx) }()
	appInst.Jobs.RunOnce(ctx)
	got, err := client.Announcement.Get(ctx, ann.ID)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Status) != "已发布" {
		t.Fatalf("到期公告未发布: %s", got.Status)
	}
}

func notificationsContain(t *testing.T, body json.RawMessage, title string) bool {
	t.Helper()
	var wrapper struct {
		List []struct {
			Title string `json:"title"`
		} `json:"list"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		t.Fatal(err)
	}
	for _, n := range wrapper.List {
		if n.Title == title {
			return true
		}
	}
	return false
}
