package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dawn-errand/internal/infrastructure/ent/credential"
	"dawn-errand/internal/infrastructure/ent/user"
	entuserprofile "dawn-errand/internal/infrastructure/ent/userprofile"
)

// M0 集成用例：短信会话、App 角色边界、账号开通、区域/资料、上传。
// 前置：`go run ./cmd/seed` 已创建 client.test / store.owner / rider.owner 三个联调账号。

func randomPhone() string {
	return fmt.Sprintf("139%08d", rand.Intn(100000000))
}

func cleanupUserByPhone(t *testing.T, phone string) {
	t.Helper()
	ctx := context.Background()
	u, err := client.User.Query().Where(user.Phone(phone)).Only(ctx)
	if err != nil {
		return
	}
	_, _ = client.Credential.Delete().Where(credential.UserID(u.ID)).Exec(ctx)
	_, _ = client.UserProfile.Delete().Where(entuserprofile.UserID(u.ID)).Exec(ctx)
	_ = client.User.DeleteOneID(u.ID).Exec(ctx)
}

func cleanupUserByEmail(t *testing.T, email string) {
	t.Helper()
	ctx := context.Background()
	u, err := client.User.Query().Where(user.EmailEqualFold(email)).Only(ctx)
	if err != nil {
		return
	}
	_, _ = client.Credential.Delete().Where(credential.UserID(u.ID)).Exec(ctx)
	_, _ = client.UserProfile.Delete().Where(entuserprofile.UserID(u.ID)).Exec(ctx)
	_ = client.User.DeleteOneID(u.ID).Exec(ctx)
}

// sessionCache 缓存联调账号 token，避免测试间重复登录触发限流（30/min/IP）。
var sessionCache = map[string]struct{ Token, Role string }{}

func loginAs(t *testing.T, account, password string) (token string, role string) {
	t.Helper()
	key := account + "\x00" + password
	if s, ok := sessionCache[key]; ok {
		return s.Token, s.Role
	}
	r, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account": account, "password": password,
	})
	if r.Code != 0 {
		t.Fatalf("login %s code=%d msg=%s", account, r.Code, r.Message)
	}
	var body struct {
		Token string `json:"token"`
		User  struct {
			Role string `json:"role"`
		} `json:"user"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	sessionCache[key] = struct{ Token, Role string }{body.Token, body.User.Role}
	return body.Token, body.User.Role
}

// TestSMSLoginCreatesSession 回归：短信自动注册必须建立 credential，登录后鉴权/刷新可用。
func TestSMSLoginCreatesSession(t *testing.T) {
	phone := randomPhone()
	defer cleanupUserByPhone(t, phone)

	send, _ := do(t, http.MethodPost, "/api/auth/sms/send", "", map[string]string{"phone": phone})
	if send.Code != 0 {
		t.Fatalf("sms send code=%d msg=%s", send.Code, send.Message)
	}
	var sent struct {
		DevCode string `json:"dev_code"`
	}
	if err := json.Unmarshal(send.Body, &sent); err != nil {
		t.Fatal(err)
	}
	if sent.DevCode == "" {
		t.Fatal("dev 环境未回显 dev_code")
	}

	login, _ := do(t, http.MethodPost, "/api/auth/sms/login", "", map[string]string{
		"phone": phone, "code": sent.DevCode,
	})
	if login.Code != 0 {
		t.Fatalf("sms login code=%d msg=%s", login.Code, login.Message)
	}
	var session struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		} `json:"user"`
	}
	if err := json.Unmarshal(login.Body, &session); err != nil {
		t.Fatal(err)
	}
	if session.Token == "" || session.RefreshToken == "" || session.User.Role != "user" {
		t.Fatalf("sms login 会话不完整: %+v", session)
	}

	// 关键断言：登录后鉴权可用（修复前必然 401/1006）。
	if me, _ := do(t, http.MethodGet, "/api/auth/me", session.Token, nil); me.Code != 0 {
		t.Fatalf("sms 登录后 /auth/me 失败: code=%d msg=%s", me.Code, me.Message)
	}
	if orders, _ := do(t, http.MethodGet, "/api/orders", session.Token, nil); orders.Code != 0 {
		t.Fatalf("sms 登录后 /orders 失败: code=%d msg=%s", orders.Code, orders.Message)
	}
	ref, _ := do(t, http.MethodPost, "/api/auth/refresh", "", map[string]string{
		"refresh_token": session.RefreshToken,
	})
	if ref.Code != 0 {
		t.Fatalf("sms 用户 refresh 失败: code=%d msg=%s", ref.Code, ref.Message)
	}

	// 凭证行必须存在。
	ctx := context.Background()
	u, err := client.User.Query().Where(user.Phone(phone)).Only(ctx)
	if err != nil {
		t.Fatalf("query sms user: %v", err)
	}
	ok, err := client.Credential.Query().Where(credential.UserID(u.ID)).Exist(ctx)
	if err != nil || !ok {
		t.Fatalf("短信注册用户缺少 credential: ok=%v err=%v", ok, err)
	}
}

// TestAppRoleBoundaries 三端角色互斥与主体绑定。
func TestAppRoleBoundaries(t *testing.T) {
	userTok, userRole := loginAs(t, "client.test@dawnerrand.local", "Client123456")
	if userRole != "user" {
		t.Fatalf("client.test role=%s", userRole)
	}
	if r, _ := do(t, http.MethodGet, "/api/app/client/me", userTok, nil); r.Code != 0 {
		t.Fatalf("client me code=%d msg=%s", r.Code, r.Message)
	}
	if r, _ := do(t, http.MethodGet, "/api/app/store/me", userTok, nil); r.Code != 1203 {
		t.Fatalf("user 访问 store 应 1203, got %d", r.Code)
	}
	if r, _ := do(t, http.MethodGet, "/api/app/driver/me", userTok, nil); r.Code != 1203 {
		t.Fatalf("user 访问 driver 应 1203, got %d", r.Code)
	}

	storeTok, storeRole := loginAs(t, "store.owner@dawnerrand.local", "Store123456")
	if storeRole != "merchant" {
		t.Fatalf("store.owner role=%s", storeRole)
	}
	if r, _ := do(t, http.MethodGet, "/api/app/store/me", storeTok, nil); r.Code != 0 {
		t.Fatalf("store me code=%d msg=%s", r.Code, r.Message)
	}
	if r, _ := do(t, http.MethodGet, "/api/app/client/me", storeTok, nil); r.Code != 1203 {
		t.Fatalf("merchant 访问 client 应 1203, got %d", r.Code)
	}

	driverTok, driverRole := loginAs(t, "rider.owner@dawnerrand.local", "Rider123456")
	if driverRole != "driver" {
		t.Fatalf("rider.owner role=%s", driverRole)
	}
	if r, _ := do(t, http.MethodGet, "/api/app/driver/me", driverTok, nil); r.Code != 0 {
		t.Fatalf("driver me code=%d msg=%s", r.Code, r.Message)
	}
	if r, _ := do(t, http.MethodGet, "/api/app/store/me", driverTok, nil); r.Code != 1203 {
		t.Fatalf("driver 访问 store 应 1203, got %d", r.Code)
	}
}

// TestProvisionAccountAndAppContext 后台开通商家/骑手账号并验证端内上下文与重复开通。
func TestProvisionAccountAndAppContext(t *testing.T) {
	admin := loginAdmin(t)
	suffix := time.Now().UnixNano()
	ctx := context.Background()

	// --- 商家 ---
	mr, _ := do(t, http.MethodPost, "/api/merchants", admin, map[string]string{
		"name": fmt.Sprintf("M0 开通商家 %d", suffix), "phone": "13900003333",
		"contact": "开通测试", "category": "其他", "region_id": "area-001",
	})
	if mr.Code != 0 {
		t.Fatalf("create merchant code=%d msg=%s", mr.Code, mr.Message)
	}
	var mCreated struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(mr.Body, &mCreated); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Merchant.DeleteOneID(mCreated.ID).Exec(ctx) }()

	storeEmail := fmt.Sprintf("m0store%d@dawnerrand.local", suffix)
	defer cleanupUserByEmail(t, storeEmail)

	// 参数校验：邮箱/手机号都为空。
	if bad, _ := do(t, http.MethodPost, "/api/merchants/"+mCreated.ID+"/account", admin,
		map[string]string{"password": "Prov123456"}); bad.Code != 1301 {
		t.Fatalf("缺邮箱/手机号应 1301, got %d", bad.Code)
	}

	prov, _ := do(t, http.MethodPost, "/api/merchants/"+mCreated.ID+"/account", admin,
		map[string]string{"email": storeEmail, "password": "Prov123456", "name": "开通店长"})
	if prov.Code != 0 {
		t.Fatalf("provision merchant code=%d msg=%s", prov.Code, prov.Message)
	}
	var provBody struct {
		UserID     string `json:"user_id"`
		MerchantID string `json:"merchant_id"`
		InitialPwd string `json:"initial_password"`
	}
	if err := json.Unmarshal(prov.Body, &provBody); err != nil {
		t.Fatal(err)
	}
	if provBody.UserID == "" || provBody.MerchantID != mCreated.ID {
		t.Fatalf("provision 返回不符: %+v", provBody)
	}
	if provBody.InitialPwd != "" {
		t.Fatalf("显式密码不应回显初始密码")
	}

	// 重复开通必须 409/1402。
	if dup, _ := do(t, http.MethodPost, "/api/merchants/"+mCreated.ID+"/account", admin,
		map[string]string{"email": "other" + storeEmail, "password": "Prov123456"}); dup.Code != 1402 {
		t.Fatalf("重复开通应 1402, got %d", dup.Code)
	}

	storeTok, storeRole := loginAs(t, storeEmail, "Prov123456")
	if storeRole != "merchant" {
		t.Fatalf("开通账号 role=%s", storeRole)
	}
	me, _ := do(t, http.MethodGet, "/api/app/store/me", storeTok, nil)
	if me.Code != 0 {
		t.Fatalf("store me code=%d msg=%s", me.Code, me.Message)
	}
	var storeMe struct {
		Shop struct {
			ID string `json:"id"`
		} `json:"shop"`
		Account struct {
			UserID string `json:"user_id"`
		} `json:"account"`
	}
	if err := json.Unmarshal(me.Body, &storeMe); err != nil {
		t.Fatal(err)
	}
	if storeMe.Shop.ID != mCreated.ID || storeMe.Account.UserID != provBody.UserID {
		t.Fatalf("store me 绑定不符: %+v", storeMe)
	}

	// 主体反向绑定。
	m, err := client.Merchant.Get(ctx, mCreated.ID)
	if err != nil || m.OwnerUserID != provBody.UserID {
		t.Fatalf("merchant.OwnerUserID 未回写: %v owner=%s", err, m.OwnerUserID)
	}

	// --- 骑手 ---
	rr, _ := do(t, http.MethodPost, "/api/riders", admin, map[string]string{
		"name": "M0 开通骑手", "phone": "13900004444", "area": "王家湾镇", "tool": "电动车",
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

	riderEmail := fmt.Sprintf("m0rider%d@dawnerrand.local", suffix)
	defer cleanupUserByEmail(t, riderEmail)
	rp, _ := do(t, http.MethodPost, "/api/riders/"+rCreated.ID+"/account", admin,
		map[string]string{"email": riderEmail, "password": "Prov123456"})
	if rp.Code != 0 {
		t.Fatalf("provision rider code=%d msg=%s", rp.Code, rp.Message)
	}
	driverTok, driverRole := loginAs(t, riderEmail, "Prov123456")
	if driverRole != "driver" {
		t.Fatalf("开通骑手 role=%s", driverRole)
	}
	dm, _ := do(t, http.MethodGet, "/api/app/driver/me", driverTok, nil)
	if dm.Code != 0 {
		t.Fatalf("driver me code=%d msg=%s", dm.Code, dm.Message)
	}
	var driverMe struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(dm.Body, &driverMe); err != nil {
		t.Fatal(err)
	}
	if driverMe.ID != rCreated.ID {
		t.Fatalf("driver me id=%s want=%s", driverMe.ID, rCreated.ID)
	}
	r, err := client.Rider.Get(ctx, rCreated.ID)
	if err != nil || r.UserID == "" {
		t.Fatalf("rider.UserID 未回写: %v user=%s", err, r.UserID)
	}
}

// TestAppClientProfileRegionAndUpload 用户端资料/区域/上传闭环。
func TestAppClientProfileRegionAndUpload(t *testing.T) {
	tok, _ := loginAs(t, "client.test@dawnerrand.local", "Client123456")

	me, _ := do(t, http.MethodGet, "/api/app/client/me", tok, nil)
	if me.Code != 0 {
		t.Fatalf("client me code=%d msg=%s", me.Code, me.Message)
	}
	var before struct {
		Name   string `json:"name"`
		Region struct {
			ID string `json:"id"`
		} `json:"region"`
	}
	if err := json.Unmarshal(me.Body, &before); err != nil {
		t.Fatal(err)
	}
	if before.Region.ID == "" {
		t.Fatal("client.test 未绑定区域")
	}

	// 区域列表 + 切换（切到自身区域，幂等）。
	regions, _ := do(t, http.MethodGet, "/api/app/client/regions", tok, nil)
	if regions.Code != 0 {
		t.Fatalf("regions code=%d msg=%s", regions.Code, regions.Message)
	}
	var regionBody struct {
		List []struct {
			ID string `json:"id"`
		} `json:"list"`
	}
	if err := json.Unmarshal(regions.Body, &regionBody); err != nil {
		t.Fatal(err)
	}
	if len(regionBody.List) == 0 {
		t.Fatal("区域列表为空")
	}
	if set, _ := do(t, http.MethodPut, "/api/app/client/region", tok, map[string]string{
		"region_id": regionBody.List[0].ID,
	}); set.Code != 0 {
		t.Fatalf("set region code=%d msg=%s", set.Code, set.Message)
	}

	// 资料修改并还原。
	updated, _ := do(t, http.MethodPut, "/api/app/client/me", tok, map[string]string{"name": "联调用户M0"})
	if updated.Code != 0 {
		t.Fatalf("update profile code=%d msg=%s", updated.Code, updated.Message)
	}
	var after struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(updated.Body, &after); err != nil {
		t.Fatal(err)
	}
	if after.Name != "联调用户M0" {
		t.Fatalf("profile name=%s", after.Name)
	}
	if restore, _ := do(t, http.MethodPut, "/api/app/client/me", tok, map[string]string{"name": before.Name}); restore.Code != 0 {
		t.Fatalf("restore profile code=%d", restore.Code)
	}

	// 上传：1x1 PNG。
	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
		0x0D, 0x49, 0x44, 0x41, 0x54, 0x78, 0xDA, 0x63, 0xFC, 0xCF, 0xC0, 0x50,
		0x0F, 0x00, 0x05, 0x83, 0x02, 0x7F, 0x95, 0x8D, 0xA6, 0x00, 0x00, 0x00,
		0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	up := doMultipart(t, "/api/app/common/uploads", tok, "file", "m0-test.png", png)
	if up.Code != 0 {
		t.Fatalf("upload code=%d msg=%s", up.Code, up.Message)
	}
	var upBody struct {
		URL  string `json:"url"`
		Mime string `json:"mime"`
		Size int    `json:"size"`
	}
	if err := json.Unmarshal(up.Body, &upBody); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(upBody.URL, "/uploads/") || upBody.Mime != "image/png" || upBody.Size != len(png) {
		t.Fatalf("upload 返回不符: %+v", upBody)
	}

	// 静态访问（经同一路由）。
	path := strings.TrimPrefix(upBody.URL, "http://localhost:8080")
	if !strings.HasPrefix(path, "/uploads/") {
		t.Fatalf("上传 URL 非预期: %s", upBody.URL)
	}
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.Len() != len(png) {
		t.Fatalf("静态访问失败: status=%d len=%d", w.Code, w.Body.Len())
	}
	_ = os.Remove(filepath.Join(handler.Cfg.StorageRootDir, filepath.FromSlash(strings.TrimPrefix(path, "/uploads/"))))

	// 非图片类型应 1304。
	bad := doMultipart(t, "/api/app/common/uploads", tok, "file", "m0-test.txt", []byte("hello"))
	if bad.Code != 1304 {
		t.Fatalf("非图片上传应 1304, got %d", bad.Code)
	}

	// 后缀伪装：PNG 内容 + .html 文件名必须落盘为 .png，防止存储型 XSS。
	masked := doMultipart(t, "/api/app/common/uploads", tok, "file", "evil.html", png)
	if masked.Code != 0 {
		t.Fatalf("后缀伪装上传 code=%d msg=%s", masked.Code, masked.Message)
	}
	var maskedBody struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(masked.Body, &maskedBody); err != nil {
		t.Fatal(err)
	}
	if strings.HasSuffix(strings.ToLower(maskedBody.URL), ".html") || !strings.HasSuffix(strings.ToLower(maskedBody.URL), ".png") {
		t.Fatalf("后缀未按 MIME 重写: %s", maskedBody.URL)
	}
	_ = os.Remove(filepath.Join(handler.Cfg.StorageRootDir,
		filepath.FromSlash(strings.TrimPrefix(strings.TrimPrefix(maskedBody.URL, "http://localhost:8080"), "/uploads/"))))
}

// doMultipart 发送 multipart/form-data 请求并解析统一响应。
func doMultipart(t *testing.T, path, token, field, filename string, data []byte) *resp {
	t.Helper()
	setup()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-Request-ID", "itest-req-001")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	var r resp
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		t.Fatalf("decode multipart %s: %v", path, err)
	}
	return &r
}
