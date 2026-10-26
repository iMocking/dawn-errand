package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

// postSettings 发一条区域写入请求，返回统一响应。
func postSettings(t *testing.T, token string, payload map[string]any) *resp {
	t.Helper()
	r, _ := do(t, http.MethodPost, "/api/settings", token, payload)
	return r
}

// TestSettingsAreaInputGuard 区域写入的判空、长度、枚举与不回显入参。
func TestSettingsAreaInputGuard(t *testing.T) {
	setup()
	admin := loginAdmin(t)

	if r := postSettings(t, admin, map[string]any{"action": "disable"}); r.Code != 1301 {
		t.Fatalf("disable 缺 area 应 1301（修复前是空指针 500）: code=%d msg=%s", r.Code, r.Message)
	}
	if r := postSettings(t, admin, map[string]any{"action": "delete"}); r.Code != 1301 {
		t.Fatalf("delete 缺 area 应 1301: code=%d msg=%s", r.Code, r.Message)
	}

	// 未知操作：入参原文不得出现在响应里（此前是「未知操作：purge」）。
	r := postSettings(t, admin, map[string]any{"action": "purge-all-regions"})
	if r.Code != 1301 {
		t.Fatalf("未知 action 应 1301: code=%d", r.Code)
	}
	if strings.Contains(r.Message, "purge") {
		t.Fatalf("响应不得回显入参 action: %s", r.Message)
	}

	// 长度与枚举上限：与库表 varchar 宽度一致，超限应在应用层拦成 400。
	name := strings.Repeat("区", 65)
	if r := postSettings(t, admin, map[string]any{"action": "create", "area": map[string]any{"name": name}}); r.Code != 1301 {
		t.Fatalf("65 字区域名应 1301: code=%d msg=%s", r.Code, r.Message)
	}
	if r := postSettings(t, admin, map[string]any{"action": "create", "area": map[string]any{"name": "越权状态区", "status": "closed"}}); r.Code != 1301 {
		t.Fatalf("非法区域状态应 1301: code=%d msg=%s", r.Code, r.Message)
	}
	if r := postSettings(t, admin, map[string]any{"action": "create", "area": map[string]any{"name": "负半径区", "radiusKM": -5}}); r.Code != 1301 {
		t.Fatalf("负数半径应 1301: code=%d msg=%s", r.Code, r.Message)
	}

	// 时间戳由服务端生成：请求体里的 created_at/updatedAt 不应写库。
	stamp := "itest-area-" + time.Now().Format("150405.000000")
	created := postSettings(t, admin, map[string]any{"action": "create", "area": map[string]any{
		"name": stamp, "status": "preparing", "villages": 3,
		"createdAt": 1, "updatedAt": 2, "created_at": 3,
	}})
	if created.Code != 0 {
		t.Fatalf("创建区域失败: code=%d msg=%s", created.Code, created.Message)
	}
	var cb struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body, &cb); err != nil || cb.ID == "" {
		t.Fatalf("创建应返回 id: %s", created.Body)
	}
	defer func() {
		_ = client.Region.DeleteOneID(cb.ID).Exec(context.Background())
	}()
	row, err := client.Region.Get(context.Background(), cb.ID)
	if err != nil {
		t.Fatal(err)
	}
	if row.CreatedAt == 0 || row.CreatedAt == 1 || row.CreatedAt == 3 || row.UpdatedAt == 2 {
		t.Fatalf("时间戳应由服务端生成: created=%d updated=%d", row.CreatedAt, row.UpdatedAt)
	}
	if row.Villages != 3 || string(row.Status) != "preparing" {
		t.Fatalf("白名单字段应正常写入: %+v", row)
	}
}

// TestFeeRuleInputGuard 费率规则写入白名单：enabled 由 status 派生、状态枚举受限。
func TestFeeRuleInputGuard(t *testing.T) {
	setup()
	admin := loginAdmin(t)

	if r, _ := do(t, http.MethodPost, "/api/settings/fee-rules", admin, map[string]any{
		"name": "缺分类", "status": "已启用",
	}); r.Code != 1301 {
		t.Fatalf("缺分类应 1301: code=%d msg=%s", r.Code, r.Message)
	}
	if r, _ := do(t, http.MethodPost, "/api/settings/fee-rules", admin, map[string]any{
		"name": "非法状态", "category": "basic", "status": "草稿",
	}); r.Code != 1301 {
		t.Fatalf("非法状态应 1301: code=%d", r.Code)
	}
	if r, _ := do(t, http.MethodPost, "/api/settings/fee-rules", admin, map[string]any{
		"name": "超长说明", "category": "basic", "desc": strings.Repeat("说", 129),
	}); r.Code != 1301 {
		t.Fatalf("超长 desc 应 1301: code=%d", r.Code)
	}

	// 矛盾行（status=已停用 但 enabled=true）不可能再出现：enabled 不进白名单。
	name := "itest-fee-" + time.Now().Format("150405.000000")
	r, _ := do(t, http.MethodPost, "/api/settings/fee-rules", admin, map[string]any{
		"name": name, "category": "ladder", "status": "已停用", "enabled": true,
		"createdAt": 1, "rider_share": "snake 键不再被当作 RiderShare",
	})
	if r.Code != 0 {
		t.Fatalf("保存费率失败: code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil || body.ID == "" {
		t.Fatalf("应返回 id: %s", r.Body)
	}
	defer func() {
		_ = client.FeeRule.DeleteOneID(body.ID).Exec(context.Background())
	}()
	row, err := client.FeeRule.Get(context.Background(), body.ID)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != "已停用" || row.Enabled {
		t.Fatalf("enabled 应由 status 派生: status=%s enabled=%v", row.Status, row.Enabled)
	}
}

// TestAnnouncementWriteWhitelistKeepsServerFields 公告编辑不得覆盖服务端维护的字段。
func TestAnnouncementWriteWhitelistKeepsServerFields(t *testing.T) {
	setup()
	admin := loginAdmin(t)
	operator := adminUserID(t, admin)

	title := "itest-公告-" + time.Now().Format("150405.000000")
	r, _ := do(t, http.MethodPost, "/api/settings/announcements", admin, map[string]any{
		"action": "save",
		"announcement": map[string]any{
			"title": title, "type": "活动", "typeTone": "blue",
			"audience": "全部用户", "status": "审核中", "content": "初稿",
		},
	})
	if r.Code != 0 {
		t.Fatalf("公告创建失败: code=%d msg=%s", r.Code, r.Message)
	}
	var cb struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(r.Body, &cb); err != nil || cb.ID == "" {
		t.Fatalf("公告创建应返回 id: %s", r.Body)
	}
	defer func() {
		_ = client.Announcement.DeleteOneID(cb.ID).Exec(context.Background())
	}()

	first, err := client.Announcement.Get(context.Background(), cb.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.ActionName != "发布" || first.CreatedBy != operator {
		t.Fatalf("新建公告应由服务端补默认按钮文案与归属: action=%q createdBy=%q want=%q", first.ActionName, first.CreatedBy, operator)
	}

	// 编辑：注入 created_by/created_at/action_name/stats/超长 status，白名单外字段一律忽略。
	if inj, _ := do(t, http.MethodPost, "/api/settings/announcements", admin, map[string]any{
		"action": "save",
		"announcement": map[string]any{
			"id": cb.ID, "title": title + "-改", "content": "二稿",
			"createdBy": "attacker", "createdAt": 1, "actionName": "跳转到钓鱼页",
			"stats":  []map[string]any{{"label": "伪造", "value": 999}},
			"status": strings.Repeat("非", 20),
		},
	}); inj.Code != 1301 {
		t.Fatalf("非法 status 应 1301: code=%d msg=%s", inj.Code, inj.Message)
	}
	if ok, _ := do(t, http.MethodPost, "/api/settings/announcements", admin, map[string]any{
		"action": "save",
		"announcement": map[string]any{
			"id": cb.ID, "title": title + "-改", "content": "二稿",
			"createdBy": "attacker", "createdAt": 1, "actionName": "跳转到钓鱼页",
			"stats": []map[string]any{{"label": "伪造", "value": 999}},
		},
	}); ok.Code != 0 {
		t.Fatalf("编辑公告失败: code=%d msg=%s", ok.Code, ok.Message)
	}
	row, err := client.Announcement.Get(context.Background(), cb.ID)
	if err != nil {
		t.Fatal(err)
	}
	if row.CreatedBy != operator || row.CreatedAt != first.CreatedAt {
		t.Fatalf("created_by/created_at 不可被请求体覆盖: by=%q created=%d origin=%d/%q", row.CreatedBy, row.CreatedAt, first.CreatedAt, operator)
	}
	if row.ActionName != "发布" {
		t.Fatalf("action_name 应保持原行值（白名单不含该字段）: %q", row.ActionName)
	}
	if row.Title != title+"-改" || row.Content != "二稿" {
		t.Fatalf("白名单字段应更新: %+v", row)
	}
}

// adminUserID 由超管令牌解析出用户 id，用于断言 created_by 归属。
func adminUserID(t *testing.T, token string) string {
	t.Helper()
	r, _ := do(t, http.MethodGet, "/api/auth/me", token, nil)
	var body struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil || body.User.ID == "" {
		t.Fatalf("/auth/me 未返回用户 id: err=%v body=%s", err, r.Body)
	}
	return body.User.ID
}
