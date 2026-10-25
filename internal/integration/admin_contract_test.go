package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestMerchantWorkspace 商家详情子页数据（证照/经营/菜单/评价/异常）。
func TestMerchantWorkspace(t *testing.T) {
	admin := loginAdmin(t)
	r, _ := do(t, http.MethodGet, "/api/merchants/M00268/workspace", admin, nil)
	if r.Code != 0 {
		t.Fatalf("workspace code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		Certs []struct {
			Label string `json:"label"`
		} `json:"certs"`
		Trend []struct {
			Date string `json:"date"`
		} `json:"trend"`
		Menu struct {
			Categories []any `json:"categories"`
			Dishes     []any `json:"dishes"`
		} `json:"menu"`
		Reviews struct {
			Total int `json:"total"`
		} `json:"reviews"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Certs) < 1 || len(body.Trend) != 7 || len(body.Menu.Categories) < 1 || len(body.Menu.Dishes) < 1 {
		t.Fatalf("workspace 聚合数据不完整: certs=%d trend=%d cats=%d dishes=%d",
			len(body.Certs), len(body.Trend), len(body.Menu.Categories), len(body.Menu.Dishes))
	}
	if body.Reviews.Total < 1 {
		t.Fatalf("评价数据为空: %s", r.Body)
	}
}

// TestWorkspaceMessages 管理端消息中心与左侧待办角标使用真实数据。
func TestWorkspaceMessages(t *testing.T) {
	admin := loginAdmin(t)
	r, _ := do(t, http.MethodGet, "/api/messages", admin, nil)
	if r.Code != 0 {
		t.Fatalf("messages code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		Badges struct {
			Complaints     int64 `json:"complaints"`
			MerchantAudits int64 `json:"merchantAudits"`
			RiderAudits    int64 `json:"riderAudits"`
			Withdrawals    int64 `json:"withdrawals"`
		} `json:"badges"`
		List []struct {
			Type string `json:"type"`
			Link string `json:"link"`
		} `json:"list"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	if body.Badges.Complaints < 1 {
		t.Fatalf("待处理客诉角标应 >=1: %+v", body.Badges)
	}
	found := false
	for _, m := range body.List {
		if m.Type == "complaint" && m.Link != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("消息中心缺少客诉待办: %s", r.Body)
	}
}

// TestAdminContract 管理端关键 DTO/写接口契约：详情字段小写、提现时间线、公告 ID/发布、费率开关。
func TestAdminContract(t *testing.T) {
	admin := loginAdmin(t)

	// 订单详情：商家/商品快照必须是前端约定的 snake_case。
	od, _ := do(t, http.MethodGet, "/api/orders/SG202610220086", admin, nil)
	if od.Code != 0 {
		t.Fatalf("order detail code=%d msg=%s", od.Code, od.Message)
	}
	var orderBody struct {
		Merchant map[string]any   `json:"merchant"`
		Goods    []map[string]any `json:"goods"`
	}
	if err := json.Unmarshal(od.Body, &orderBody); err != nil {
		t.Fatal(err)
	}
	if orderBody.Merchant["name"] == nil || orderBody.Merchant["Name"] != nil {
		t.Fatalf("merchant 快照应为小写字段: %s", od.Body)
	}
	if len(orderBody.Goods) == 0 || orderBody.Goods[0]["name"] == nil {
		t.Fatalf("goods 快照应为小写字段: %s", od.Body)
	}

	// 提现详情时间线：time/text/operator 小写。
	wd, _ := do(t, http.MethodGet, "/api/finance/withdraw/TX20261022016", admin, nil)
	var wdBody struct {
		Timeline []map[string]any `json:"timeline"`
	}
	if err := json.Unmarshal(wd.Body, &wdBody); err != nil {
		t.Fatal(err)
	}
	if len(wdBody.Timeline) == 0 || wdBody.Timeline[0]["time"] == nil || wdBody.Timeline[0]["text"] == nil {
		t.Fatalf("提现时间线应为小写字段: %s", wd.Body)
	}

	// 公告创建返回 ID，中文 action“发布”可被后端识别。
	title := fmt.Sprintf("对接契约测试-%d", time.Now().UnixNano())
	created, _ := do(t, http.MethodPost, "/api/settings/announcements", admin, map[string]any{
		"action": "save",
		"announcement": map[string]any{
			"title": title, "type": "活动", "typeTone": "blue",
			"audience": "全部用户", "status": "审核中", "content": "契约测试",
		},
	})
	var createdBody struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(created.Body, &createdBody)
	if created.Code != 0 || createdBody.ID == "" {
		t.Fatalf("公告创建应返回 id: code=%d body=%s", created.Code, created.Body)
	}
	defer func() {
		_ = client.Announcement.DeleteOneID(createdBody.ID).Exec(context.Background())
	}()
	pub, _ := do(t, http.MethodPost, "/api/settings/announcements", admin, map[string]any{
		"action": "发布", "announcement": map[string]any{"id": createdBody.ID},
	})
	if pub.Code != 0 {
		t.Fatalf("公告发布 code=%d msg=%s", pub.Code, pub.Message)
	}
	got, err := client.Announcement.Get(context.Background(), createdBody.ID)
	if err != nil || got.Status != "已发布" {
		t.Fatalf("公告发布后状态异常: %v %+v", err, got)
	}

	// 费率开关：enabled=false 必须持久化为已停用，再恢复。
	rules, _ := do(t, http.MethodGet, "/api/settings/fee-rules", admin, nil)
	var ruleBody struct {
		List []map[string]any `json:"list"`
	}
	_ = json.Unmarshal(rules.Body, &ruleBody)
	var summary map[string]any
	for _, row := range ruleBody.List {
		if row["category"] == "summary" {
			summary = row
			break
		}
	}
	if summary == nil {
		t.Fatalf("缺少 summary 费率行")
	}
	summary["status"] = "已停用"
	summary["enabled"] = false
	if off, _ := do(t, http.MethodPost, "/api/settings/fee-rules", admin, summary); off.Code != 0 {
		t.Fatalf("停用费率 code=%d", off.Code)
	}
	rules2, _ := do(t, http.MethodGet, "/api/settings/fee-rules", admin, nil)
	var ruleBody2 struct {
		List []map[string]any `json:"list"`
	}
	_ = json.Unmarshal(rules2.Body, &ruleBody2)
	for _, row := range ruleBody2.List {
		if row["id"] == summary["id"] && row["status"] != "已停用" {
			t.Fatalf("费率停用未生效: %+v", row)
		}
	}
	summary["status"] = "已启用"
	summary["enabled"] = true
	if on, _ := do(t, http.MethodPost, "/api/settings/fee-rules", admin, summary); on.Code != 0 {
		t.Fatalf("恢复费率 code=%d", on.Code)
	}
}
