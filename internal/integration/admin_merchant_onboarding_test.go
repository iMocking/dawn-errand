package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	entcredential "dawn-errand/internal/infrastructure/ent/credential"
	entuser "dawn-errand/internal/infrastructure/ent/user"
)

func merchantApplicationPayload() map[string]any {
	return map[string]any{
		"licenseNo": "91330000TEST000099", "licenseImage": "http://example.com/license.jpg",
		"legalName": "测试法人", "legalIdCard": "330100199001010022",
		"idCardFront": "http://example.com/id-front.jpg", "idCardBack": "http://example.com/id-back.jpg",
		"storefrontImages": []string{"http://example.com/store.jpg"},
		"bankName":         "工商银行", "bankAccountName": "测试法人", "bankAccountNo": "6222000000000000099",
	}
}

func merchantPayload(name, phone string) map[string]any {
	return map[string]any{
		"name": name, "phone": phone, "contact": "测试法人", "category": "其他",
		"address": "测试镇测试路 99 号", "openHours": "09:00-21:00", "regionId": "area-001",
		"application": merchantApplicationPayload(),
	}
}

func cleanupMerchantOnboarding(t *testing.T, merchantID, phone string) {
	t.Helper()
	ctx := context.Background()
	if u, err := client.User.Query().Where(entuser.Phone(phone)).Only(ctx); err == nil {
		_, _ = client.Credential.Delete().Where(entcredential.UserID(u.ID)).Exec(ctx)
		_ = client.User.DeleteOneID(u.ID).Exec(ctx)
	}
	_ = client.Merchant.DeleteOneID(merchantID).Exec(ctx)
}

// TestRiderOnboardingApprove 骑手完整入驻 → 审核通过 → 自动账号 → 骑手端登录。
func TestRiderOnboardingApprove(t *testing.T) {
	admin := loginAdmin(t)
	phone := fmt.Sprintf("137%08d", time.Now().UnixNano()%100000000)
	name := fmt.Sprintf("地推骑手-%d", time.Now().UnixNano())
	created, _ := do(t, http.MethodPost, "/api/riders", admin, map[string]any{
		"name": name, "phone": phone, "area": "王家湾镇", "tool": "电动车",
		"orderTypes": "外卖,帮送", "idCard": "330100199001010055",
		"serviceType": "外卖/帮送", "vehicle": "测试电动车", "account": "6222000000000000055",
	})
	if created.Code != 0 {
		t.Fatalf("rider create code=%d msg=%s", created.Code, created.Message)
	}
	var cb struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(created.Body, &cb)
	if cb.ID == "" {
		t.Fatalf("rider create 缺 id: %s", created.Body)
	}
	defer func() {
		ctx := context.Background()
		if u, err := client.User.Query().Where(entuser.Phone(phone)).Only(ctx); err == nil {
			_, _ = client.Credential.Delete().Where(entcredential.UserID(u.ID)).Exec(ctx)
			_ = client.User.DeleteOneID(u.ID).Exec(ctx)
		}
		_ = client.Rider.DeleteOneID(cb.ID).Exec(ctx)
	}()

	approved, _ := do(t, http.MethodPost, "/api/riders/audit", admin,
		map[string]string{"id": cb.ID, "action": "approve"})
	var ab struct {
		Account struct {
			Phone           string `json:"phone"`
			InitialPassword string `json:"initialPassword"`
		} `json:"account"`
	}
	_ = json.Unmarshal(approved.Body, &ab)
	if approved.Code != 0 || ab.Account.InitialPassword == "" {
		t.Fatalf("rider approve code=%d body=%s", approved.Code, approved.Body)
	}
	login, _ := do(t, http.MethodPost, "/api/auth/login", "",
		map[string]string{"account": phone, "password": ab.Account.InitialPassword})
	var lb struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(login.Body, &lb)
	if login.Code != 0 || lb.Token == "" {
		t.Fatalf("rider login code=%d body=%s", login.Code, login.Body)
	}
	if me, _ := do(t, http.MethodGet, "/api/app/driver/me", lb.Token, nil); me.Code != 0 {
		t.Fatalf("driver me code=%d msg=%s", me.Code, me.Message)
	}

	// 重置密码后旧 token 失效。
	reset, _ := do(t, http.MethodPost, "/api/riders/"+cb.ID+"/account/reset", admin, nil)
	var rb struct {
		Account struct {
			InitialPassword string `json:"initialPassword"`
		} `json:"account"`
	}
	_ = json.Unmarshal(reset.Body, &rb)
	if reset.Code != 0 || rb.Account.InitialPassword == "" {
		t.Fatalf("rider reset code=%d body=%s", reset.Code, reset.Body)
	}
	if stale, _ := do(t, http.MethodGet, "/api/app/driver/me", lb.Token, nil); stale.Code == 0 {
		t.Fatalf("骑手重置密码后旧会话应失效")
	}
}

// TestRiderWorkspace 骑手详情子页数据（配送/绩效/证件/奖惩）。
func TestRiderWorkspace(t *testing.T) {
	admin := loginAdmin(t)
	r, _ := do(t, http.MethodGet, "/api/riders/R00086/workspace", admin, nil)
	if r.Code != 0 {
		t.Fatalf("rider workspace code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		Records []any `json:"records"`
		Trend   []any `json:"trend"`
		Profile struct {
			IDCard string `json:"idCard"`
		} `json:"profile"`
		Reviews struct {
			Total int `json:"total"`
		} `json:"reviews"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Trend) != 7 || len(body.Records) < 1 || body.Profile.IDCard == "" || body.Reviews.Total < 1 {
		t.Fatalf("rider workspace 数据不完整: %s", r.Body)
	}
}

// TestMerchantOnboardingApprove 地推录入完整资料 → 审核通过 → 自动开通账号 → 商家端可登录经营。
func TestMerchantOnboardingApprove(t *testing.T) {
	admin := loginAdmin(t)
	phone := fmt.Sprintf("139%08d", time.Now().UnixNano()%100000000)
	name := fmt.Sprintf("地推联调商家-%d", time.Now().UnixNano())

	created, _ := do(t, http.MethodPost, "/api/merchants", admin, merchantPayload(name, phone))
	if created.Code != 0 {
		t.Fatalf("create code=%d msg=%s", created.Code, created.Message)
	}
	var cb struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(created.Body, &cb)
	if cb.ID == "" {
		t.Fatalf("create 缺 id: %s", created.Body)
	}
	defer cleanupMerchantOnboarding(t, cb.ID, phone)

	// 资料完整，审核通过并自动开通账号。
	approved, _ := do(t, http.MethodPost, "/api/merchants/audit/"+cb.ID, admin, map[string]string{"action": "approve"})
	if approved.Code != 0 {
		t.Fatalf("approve code=%d msg=%s", approved.Code, approved.Message)
	}
	var ab struct {
		Account struct {
			Phone           string `json:"phone"`
			InitialPassword string `json:"initialPassword"`
		} `json:"account"`
	}
	if err := json.Unmarshal(approved.Body, &ab); err != nil {
		t.Fatal(err)
	}
	if ab.Account.InitialPassword == "" || ab.Account.Phone != phone {
		t.Fatalf("审核通过应返回初始账号/密码: %s", approved.Body)
	}

	// 商家用手机号 + 初始密码登录，并访问商家端 me。
	login, _ := do(t, http.MethodPost, "/api/auth/login", "",
		map[string]string{"account": phone, "password": ab.Account.InitialPassword})
	if login.Code != 0 {
		t.Fatalf("merchant login code=%d msg=%s", login.Code, login.Message)
	}
	var lb struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(login.Body, &lb)
	if lb.Token == "" {
		t.Fatalf("merchant login 缺 token: %s", login.Body)
	}
	me, _ := do(t, http.MethodGet, "/api/app/store/me", lb.Token, nil)
	if me.Code != 0 {
		t.Fatalf("store me code=%d msg=%s", me.Code, me.Message)
	}

	// 重置密码后旧会话失效，新密码可登录。
	reset, _ := do(t, http.MethodPost, "/api/merchants/"+cb.ID+"/account/reset", admin, nil)
	var rb struct {
		Account struct {
			InitialPassword string `json:"initialPassword"`
		} `json:"account"`
	}
	_ = json.Unmarshal(reset.Body, &rb)
	if reset.Code != 0 || rb.Account.InitialPassword == "" {
		t.Fatalf("reset code=%d body=%s", reset.Code, reset.Body)
	}
	if stale, _ := do(t, http.MethodGet, "/api/app/store/me", lb.Token, nil); stale.Code == 0 {
		t.Fatalf("重置密码后旧会话应失效")
	}
	login2, _ := do(t, http.MethodPost, "/api/auth/login", "",
		map[string]string{"account": phone, "password": rb.Account.InitialPassword})
	if login2.Code != 0 {
		t.Fatalf("new password login code=%d msg=%s", login2.Code, login2.Message)
	}

	// 商家详情应带认证信息与已绑定账号。
	detail, _ := do(t, http.MethodGet, "/api/merchants/"+cb.ID, admin, nil)
	var db struct {
		AuditStatus string `json:"auditStatus"`
		OwnerUserID string `json:"ownerUserId"`
		Application struct {
			LicenseNo string `json:"licenseNo"`
		} `json:"application"`
	}
	_ = json.Unmarshal(detail.Body, &db)
	if db.AuditStatus != "已通过" || db.OwnerUserID == "" || db.Application.LicenseNo == "" {
		t.Fatalf("商家详情未体现认证/账号: %s", detail.Body)
	}
}

// TestMerchantOnboardingRejectResubmit 驳回必填原因，补件后可重新提交。
func TestMerchantOnboardingRejectResubmit(t *testing.T) {
	admin := loginAdmin(t)
	phone := fmt.Sprintf("138%08d", time.Now().UnixNano()%100000000)
	name := fmt.Sprintf("地推补件商家-%d", time.Now().UnixNano())
	created, _ := do(t, http.MethodPost, "/api/merchants", admin, merchantPayload(name, phone))
	var cb struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(created.Body, &cb)
	if cb.ID == "" {
		t.Fatalf("create 缺 id: %s", created.Body)
	}
	defer cleanupMerchantOnboarding(t, cb.ID, phone)

	// 驳回不填原因 → 1301。
	if bad, _ := do(t, http.MethodPost, "/api/merchants/audit/"+cb.ID, admin,
		map[string]string{"action": "reject"}); bad.Code != 1301 {
		t.Fatalf("无原因驳回应 1301, got %d", bad.Code)
	}
	rej, _ := do(t, http.MethodPost, "/api/merchants/audit/"+cb.ID, admin,
		map[string]string{"action": "reject", "reason": "门头照不清晰"})
	if rej.Code != 0 {
		t.Fatalf("reject code=%d msg=%s", rej.Code, rej.Message)
	}

	// 补件重提。
	payload := merchantPayload(name+"-补件", phone)
	payload["application"].(map[string]any)["licenseNo"] = "91330000TEST000100"
	upd, _ := do(t, http.MethodPut, "/api/merchants/"+cb.ID+"/application", admin, payload)
	if upd.Code != 0 {
		t.Fatalf("resubmit code=%d msg=%s", upd.Code, upd.Message)
	}
	row, err := client.Merchant.Get(context.Background(), cb.ID)
	if err != nil || row.AuditStatus != "pending" || row.RejectReason != "" {
		t.Fatalf("补件后应回到待审核: %v %+v", err, row)
	}
}
