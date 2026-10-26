package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func loginClient(t *testing.T) string {
	t.Helper()
	tok, role := loginAs(t, "client.test@dawnerrand.local", "Client123456")
	if role != "user" {
		t.Fatalf("client.test role=%s", role)
	}
	return tok
}

func loginStore(t *testing.T) string {
	t.Helper()
	tok, role := loginAs(t, "store.owner@dawnerrand.local", "Store123456")
	if role != "merchant" {
		t.Fatalf("store.owner role=%s", role)
	}
	return tok
}

// TestClientCatalog 用户端商家列表/详情/菜单（含售罄与下架过滤）。
func TestClientCatalog(t *testing.T) {
	tok := loginClient(t)

	list, _ := do(t, http.MethodGet, "/api/app/client/merchants?page=1&page_size=10&lat=30.42&lng=115.33", tok, nil)
	if list.Code != 0 {
		t.Fatalf("merchants code=%d msg=%s", list.Code, list.Message)
	}
	var lb struct {
		List []struct {
			ID              string  `json:"id"`
			Name            string  `json:"name"`
			DistanceKM      float64 `json:"distance_km"`
			DeliveryMinutes int     `json:"delivery_minutes"`
			Status          string  `json:"status"`
		} `json:"list"`
		Total    int64 `json:"total"`
		Page     int   `json:"page"`
		PageSize int   `json:"page_size"`
	}
	if err := json.Unmarshal(list.Body, &lb); err != nil {
		t.Fatal(err)
	}
	if lb.Total < 1 || len(lb.List) == 0 {
		t.Fatalf("商家列表为空: %+v", lb)
	}
	found := false
	for _, m := range lb.List {
		if m.ID == "M00268" {
			found = true
			if m.DistanceKM <= 0 || m.DeliveryMinutes <= 0 {
				t.Fatalf("距离/预计时长未计算: %+v", m)
			}
			if m.Status == "" {
				t.Fatalf("缺少展示状态")
			}
		}
	}
	if !found {
		t.Fatalf("列表未返回 M00268: %+v", lb.List)
	}

	det, _ := do(t, http.MethodGet, "/api/app/client/merchants/M00268?lat=30.42&lng=115.33", tok, nil)
	if det.Code != 0 {
		t.Fatalf("merchant detail code=%d msg=%s", det.Code, det.Message)
	}
	var d struct {
		Categories []struct {
			ID string `json:"id"`
		} `json:"categories"`
		Dishes []struct {
			ID      string `json:"id"`
			OnSale  bool   `json:"on_sale"`
			SoldOut bool   `json:"sold_out"`
			Price   float64
		} `json:"dishes"`
		DistanceKM float64 `json:"distance_km"`
	}
	if err := json.Unmarshal(det.Body, &d); err != nil {
		t.Fatal(err)
	}
	if len(d.Categories) < 2 || len(d.Dishes) < 5 {
		t.Fatalf("菜单数据不足: cats=%d dishes=%d", len(d.Categories), len(d.Dishes))
	}
	seen := map[string]bool{}
	for _, dish := range d.Dishes {
		seen[dish.ID] = true
		if dish.ID == "D003" && !dish.SoldOut {
			t.Fatalf("D003 应为售罄")
		}
	}
	if seen["D006"] {
		t.Fatalf("下架菜品 D006 不应出现在用户端菜单")
	}
	if d.DistanceKM <= 0 {
		t.Fatalf("详情距离未计算")
	}
}

// TestStoreMenuLifecycle 商家菜单 CRUD、上下架、分类占用校验、营业设置与状态。
func TestStoreMenuLifecycle(t *testing.T) {
	tok := loginStore(t)
	ctx := context.Background()

	// 保存原营业设置并在结束后还原。
	biz, _ := do(t, http.MethodGet, "/api/app/store/business", tok, nil)
	if biz.Code != 0 {
		t.Fatalf("get business code=%d msg=%s", biz.Code, biz.Message)
	}
	var origin map[string]any
	if err := json.Unmarshal(biz.Body, &origin); err != nil {
		t.Fatal(err)
	}
	defer func() {
		restore := map[string]any{
			"open_hours": origin["open_hours"], "delivery_range": origin["delivery_range"],
			"auto_accept": origin["auto_accept"], "min_price": origin["min_price"],
			"delivery_fee": origin["delivery_fee"], "packing_fee": origin["packing_fee"],
			"announcement": origin["announcement"], "service_phone": origin["service_phone"],
		}
		_, _ = do(t, http.MethodPut, "/api/app/store/business", tok, restore)
	}()

	updated, _ := do(t, http.MethodPut, "/api/app/store/business", tok, map[string]any{
		"open_hours": "08:30-22:00", "delivery_range": "全镇 6km",
		"auto_accept": true, "min_price": 20.5, "delivery_fee": 6,
		"packing_fee": 2.5, "announcement": "集成测试公告", "service_phone": "13900002680",
	})
	if updated.Code != 0 {
		t.Fatalf("update business code=%d msg=%s", updated.Code, updated.Message)
	}
	var ub struct {
		OpenHours  string  `json:"open_hours"`
		AutoAccept bool    `json:"auto_accept"`
		MinPrice   float64 `json:"min_price"`
	}
	if err := json.Unmarshal(updated.Body, &ub); err != nil {
		t.Fatal(err)
	}
	if ub.OpenHours != "08:30-22:00" || !ub.AutoAccept || ub.MinPrice != 20.5 {
		t.Fatalf("营业设置回读不符: %+v", ub)
	}

	off, _ := do(t, http.MethodPut, "/api/app/store/status", tok, map[string]bool{"open": false})
	if off.Code != 0 {
		t.Fatalf("status off code=%d msg=%s", off.Code, off.Message)
	}
	on, _ := do(t, http.MethodPut, "/api/app/store/status", tok, map[string]bool{"open": true})
	if on.Code != 0 {
		t.Fatalf("status on code=%d msg=%s", on.Code, on.Message)
	}

	cat, _ := do(t, http.MethodPost, "/api/app/store/categories", tok, map[string]any{
		"name": "集成测试分类", "sort": 99,
	})
	if cat.Code != 0 {
		t.Fatalf("create category code=%d msg=%s", cat.Code, cat.Message)
	}
	var c struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(cat.Body, &c); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DishCategory.DeleteOneID(c.ID).Exec(ctx) }()

	dishResp, _ := do(t, http.MethodPost, "/api/app/store/dishes", tok, map[string]any{
		"category_id": c.ID, "name": "集成测试菜品", "description": "测试",
		"price": 9.9, "stock": 5, "on_sale": true,
	})
	if dishResp.Code != 0 {
		t.Fatalf("create dish code=%d msg=%s", dishResp.Code, dishResp.Message)
	}
	var d struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(dishResp.Body, &d); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Dish.DeleteOneID(d.ID).Exec(ctx) }()

	upd, _ := do(t, http.MethodPut, "/api/app/store/dishes/"+d.ID, tok, map[string]any{
		"category_id": c.ID, "name": "集成测试菜品改", "price": 10.5, "stock": 8, "on_sale": true,
	})
	if upd.Code != 0 {
		t.Fatalf("update dish code=%d msg=%s", upd.Code, upd.Message)
	}
	sale, _ := do(t, http.MethodPost, "/api/app/store/dishes/"+d.ID+"/on-sale", tok, map[string]any{
		"on_sale": false, "stock": 8,
	})
	if sale.Code != 0 {
		t.Fatalf("set sale code=%d msg=%s", sale.Code, sale.Message)
	}
	var sd struct {
		OnSale bool `json:"on_sale"`
		Stock  int  `json:"stock"`
	}
	if err := json.Unmarshal(sale.Body, &sd); err != nil {
		t.Fatal(err)
	}
	if sd.OnSale || sd.Stock != 8 {
		t.Fatalf("上下架结果不符: %+v", sd)
	}

	// 分类下仍有菜品，删除应 1704。
	if del, _ := do(t, http.MethodDelete, "/api/app/store/categories/"+c.ID, tok, nil); del.Code != 1704 {
		t.Fatalf("分类占用应 1704, got %d", del.Code)
	}
	if del, _ := do(t, http.MethodDelete, "/api/app/store/dishes/"+d.ID, tok, nil); del.Code != 0 {
		t.Fatalf("delete dish code=%d msg=%s", del.Code, del.Message)
	}
	if del, _ := do(t, http.MethodDelete, "/api/app/store/categories/"+c.ID, tok, nil); del.Code != 0 {
		t.Fatalf("delete category code=%d msg=%s", del.Code, del.Message)
	}
}

// TestAddressLifecycle 地址 CRUD、默认地址切换与删除回退。
func TestAddressLifecycle(t *testing.T) {
	tok := loginClient(t)
	ctx := context.Background()

	before, _ := do(t, http.MethodGet, "/api/app/client/addresses", tok, nil)
	var list struct {
		List  []map[string]any `json:"list"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(before.Body, &list); err != nil {
		t.Fatal(err)
	}
	if list.Total < 1 {
		t.Fatalf("地址种子缺失")
	}

	created, _ := do(t, http.MethodPost, "/api/app/client/addresses", tok, map[string]any{
		"contact_name": "集成测试", "phone": "13800001111",
		"detail": "王家湾镇集成测试地址 1 号", "tag": "测试", "is_default": false,
	})
	if created.Code != 0 {
		t.Fatalf("create address code=%d msg=%s", created.Code, created.Message)
	}
	var a struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body, &a); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.UserAddress.DeleteOneID(a.ID).Exec(ctx) }()

	def, _ := do(t, http.MethodPost, "/api/app/client/addresses/"+a.ID+"/default", tok, nil)
	if def.Code != 0 {
		t.Fatalf("set default code=%d msg=%s", def.Code, def.Message)
	}
	after, _ := do(t, http.MethodGet, "/api/app/client/addresses", tok, nil)
	var list2 struct {
		List []struct {
			ID        string `json:"id"`
			IsDefault bool   `json:"is_default"`
		} `json:"list"`
	}
	if err := json.Unmarshal(after.Body, &list2); err != nil {
		t.Fatal(err)
	}
	defaults := 0
	for _, row := range list2.List {
		if row.IsDefault {
			defaults++
			if row.ID != a.ID {
				t.Fatalf("默认地址应为新建地址, got %s", row.ID)
			}
		}
	}
	if defaults != 1 {
		t.Fatalf("默认地址数量应为 1, got %d", defaults)
	}

	upd, _ := do(t, http.MethodPut, "/api/app/client/addresses/"+a.ID, tok, map[string]any{
		"contact_name": "集成测试", "phone": "13800001111",
		"detail": "王家湾镇集成测试地址 2 号", "tag": "测试",
	})
	if upd.Code != 0 {
		t.Fatalf("update address code=%d msg=%s", upd.Code, upd.Message)
	}

	// 非法手机号应 1301。
	if bad, _ := do(t, http.MethodPost, "/api/app/client/addresses", tok, map[string]any{
		"contact_name": "测试", "phone": "12345", "detail": "地址地址",
	}); bad.Code != 1301 {
		t.Fatalf("非法手机号应 1301, got %d", bad.Code)
	}

	if del, _ := do(t, http.MethodDelete, "/api/app/client/addresses/"+a.ID, tok, nil); del.Code != 0 {
		t.Fatalf("delete address code=%d msg=%s", del.Code, del.Message)
	}
	// 删除默认地址后应自动回退到剩余地址，且仍只有一个默认。
	final, _ := do(t, http.MethodGet, "/api/app/client/addresses", tok, nil)
	var list3 struct {
		List []struct {
			IsDefault bool `json:"is_default"`
		} `json:"list"`
	}
	if err := json.Unmarshal(final.Body, &list3); err != nil {
		t.Fatal(err)
	}
	defCount := 0
	for _, row := range list3.List {
		if row.IsDefault {
			defCount++
		}
	}
	if len(list3.List) > 0 && defCount != 1 {
		t.Fatalf("删除默认后默认数量应为 1, got %d", defCount)
	}
}

// TestStoreResourceOwnership 商家之间不能修改对方分类/菜品。
func TestStoreResourceOwnership(t *testing.T) {
	admin := loginAdmin(t)
	ctx := context.Background()
	suffix := time.Now().UnixNano()

	createMerchant := func(name string) (merchantID, email string) {
		r, _ := do(t, http.MethodPost, "/api/merchants", admin, map[string]any{
			"name": name, "phone": "13900005555", "category": "其他", "region_id": "area-001",
		})
		if r.Code != 0 {
			t.Fatalf("create merchant %s code=%d msg=%s", name, r.Code, r.Message)
		}
		var m struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(r.Body, &m); err != nil {
			t.Fatal(err)
		}
		email = fmt.Sprintf("own%d%s@dawnerrand.local", suffix, m.ID[len(m.ID)-4:])
		approveMerchant(t, m.ID)
		p, _ := do(t, http.MethodPost, "/api/merchants/"+m.ID+"/account", admin,
			map[string]string{"email": email, "password": "Prov123456"})
		if p.Code != 0 {
			t.Fatalf("provision %s code=%d msg=%s", name, p.Code, p.Message)
		}
		return m.ID, email
	}

	aID, aEmail := createMerchant(fmt.Sprintf("归属测试A%d", suffix))
	bID, bEmail := createMerchant(fmt.Sprintf("归属测试B%d", suffix))
	defer func() {
		_ = client.Merchant.DeleteOneID(aID).Exec(ctx)
		_ = client.Merchant.DeleteOneID(bID).Exec(ctx)
	}()
	defer cleanupUserByEmail(t, aEmail)
	defer cleanupUserByEmail(t, bEmail)

	aTok, _ := loginAs(t, aEmail, "Prov123456")
	cat, _ := do(t, http.MethodPost, "/api/app/store/categories", aTok, map[string]any{"name": "归属分类", "sort": 1})
	if cat.Code != 0 {
		t.Fatalf("A create category code=%d msg=%s", cat.Code, cat.Message)
	}
	var c struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(cat.Body, &c); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.DishCategory.DeleteOneID(c.ID).Exec(ctx) }()

	d, _ := do(t, http.MethodPost, "/api/app/store/dishes", aTok, map[string]any{
		"category_id": c.ID, "name": "归属菜品", "price": 12, "stock": 3, "on_sale": true,
	})
	if d.Code != 0 {
		t.Fatalf("A create dish code=%d msg=%s", d.Code, d.Message)
	}
	var dishRow struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(d.Body, &dishRow); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Dish.DeleteOneID(dishRow.ID).Exec(ctx) }()

	bTok, _ := loginAs(t, bEmail, "Prov123456")
	if r, _ := do(t, http.MethodPut, "/api/app/store/categories/"+c.ID, bTok, map[string]any{
		"name": "越权改名", "sort": 2,
	}); r.Code != 1702 {
		t.Fatalf("B 改 A 分类应 1702, got %d", r.Code)
	}
	if r, _ := do(t, http.MethodPut, "/api/app/store/dishes/"+dishRow.ID, bTok, map[string]any{
		"category_id": c.ID, "name": "越权改名", "price": 1, "stock": 1,
	}); r.Code != 1701 {
		t.Fatalf("B 改 A 菜品应 1701, got %d", r.Code)
	}
	if r, _ := do(t, http.MethodDelete, "/api/app/store/dishes/"+dishRow.ID, bTok, nil); r.Code != 1701 {
		t.Fatalf("B 删 A 菜品应 1701, got %d", r.Code)
	}
	if r, _ := do(t, http.MethodGet, "/api/app/store/menu", bTok, nil); r.Code != 0 {
		t.Fatalf("B 菜单读取失败: %d", r.Code)
	}

	// 清理未删除的菜品/分类由 defer 完成。
}
