package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// R3-7 回归：经营资格守卫是「审核通过」AND「已绑定账号」，缺一不可。
// 旧写法 `!=approved && OwnerUserID==""` 只在两者同时不满足才拦截，
// 导致「已绑定账号但审核未通过」的商家可以照常进入店铺端。
func TestStoreGuardRequiresApproved(t *testing.T) {
	admin := loginAdmin(t)
	suffix := time.Now().UnixNano()
	name := fmt.Sprintf("未审核商家%d", suffix)
	r, _ := do(t, http.MethodPost, "/api/merchants", admin, map[string]any{
		"name": name, "phone": "13900008888", "category": "其他", "region_id": "area-001",
	})
	if r.Code != 0 {
		t.Fatalf("create merchant code=%d msg=%s", r.Code, r.Message)
	}
	var m struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(r.Body, &m); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Merchant.DeleteOneID(m.ID).Exec(t.Context()) }()

	email := fmt.Sprintf("pending%d@dawnerrand.local", suffix)
	defer cleanupUserByEmail(t, email)
	p, _ := do(t, http.MethodPost, "/api/merchants/"+m.ID+"/account", admin,
		map[string]string{"email": email, "password": "Prov123456"})
	if p.Code != 0 {
		t.Fatalf("provision code=%d msg=%s", p.Code, p.Message)
	}

	// 账号已绑定（OwnerUserID 非空）但审核仍 pending → 必须拒绝进入店铺端。
	tok, _ := loginAs(t, email, "Prov123456")
	if me, _ := do(t, http.MethodGet, "/api/app/store/me", tok, nil); me.Code != 1502 {
		t.Fatalf("pending 商家访问店铺端应 1502, got %d msg=%s", me.Code, me.Message)
	}

	// 审核通过后放行。
	approveMerchant(t, m.ID)
	if me, _ := do(t, http.MethodGet, "/api/app/store/me", tok, nil); me.Code != 0 {
		t.Fatalf("审核通过后应放行, got %d msg=%s", me.Code, me.Message)
	}
}
