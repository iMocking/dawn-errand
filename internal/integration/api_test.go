// Package integration 端到端集成测试：装配真实依赖（Postgres/Redis），
// 经 httptest 走完整 HTTP 栈验证统一返回、权限、数据隔离与业务流程。
// 依赖：本地 docker 容器 dawn-errand-postgres / dawn-errand-redis 健康，
// 且已执行 `go run ./cmd/migrate up` 与 `go run ./cmd/seed --email admin@shuguang.local`
// （超管一次性密码在 seed 输出中）。
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/assembly"
	"dawn-errand/internal/config"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/credential"
	"dawn-errand/internal/infrastructure/ent/emailverificationtoken"
	"dawn-errand/internal/infrastructure/ent/passwordresettoken"
	"dawn-errand/internal/infrastructure/ent/systemsetting"
	"dawn-errand/internal/infrastructure/ent/user"
	"dawn-errand/internal/infrastructure/passwd"
	apihandlers "dawn-errand/internal/interfaces/http"
)

const (
	adminEmail    = "admin@shuguang.local"
	adminPassword = "PxqHfrNFzTXupfaB"
)

var (
	router  *gin.Engine
	handler *apihandlers.Handler
	client  *ent.Client
)

type resp struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Body      json.RawMessage `json:"body"`
	RequestID string          `json:"request_id"`
}

func setup() {
	if router != nil {
		return
	}
	gin.SetMode(gin.TestMode)
	cfg := config.Load()
	app, err := assembly.Build(cfg)
	if err != nil {
		panic(err)
	}
	client = app.Client
	handler = app.Handler
	router = apihandlers.NewRouter(app.Handler, app.MW)
}

func do(t *testing.T, method, path string, token string, payload any) (*resp, http.Header) {
	t.Helper()
	setup()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "itest-req-001")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	var r resp
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		t.Fatalf("decode %s %s: %v", method, path, err)
	}
	if r.RequestID != "itest-req-001" {
		t.Errorf("%s %s: request_id 未透传, got %q", method, path, r.RequestID)
	}
	return &r, res.Header
}

var adminTokCache string

func loginAdmin(t *testing.T) string {
	t.Helper()
	if adminTokCache != "" {
		return adminTokCache
	}
	r, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account":  adminEmail,
		"password": adminPassword,
	})
	if r.Code != 0 {
		t.Fatalf("login code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			Role string `json:"role"`
		} `json:"user"`
		Permissions []string `json:"permissions"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	if body.Token == "" || body.RefreshToken == "" {
		t.Fatal("token/refresh_token 缺失")
	}
	if body.User.Role != "super_admin" {
		t.Fatalf("role=%s", body.User.Role)
	}
	hasStar := false
	for _, p := range body.Permissions {
		if p == "*" {
			hasStar = true
		}
	}
	if !hasStar {
		t.Fatal("super_admin permissions 应包含 *")
	}
	adminTokCache = body.Token
	return body.Token
}

func cleanUser(t *testing.T, email string) {
	t.Helper()
	ctx := context.Background()
	u, err := client.User.Query().Where(user.EmailEqualFold(email)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		t.Fatal(err)
	}
	if err == nil {
		_, _ = client.EmailVerificationToken.Delete().Where(emailverificationtoken.UserID(u.ID)).Exec(ctx)
		_, _ = client.PasswordResetToken.Delete().Where(passwordresettoken.UserID(u.ID)).Exec(ctx)
		_, _ = client.Credential.Delete().Where(credential.UserID(u.ID)).Exec(ctx)
		_ = client.User.DeleteOneID(u.ID).Exec(ctx)
	}
}

var mailTokenRe = regexp.MustCompile(`token=([A-Za-z0-9_\-\.]+)`)

// mailToken 从 dev 发件箱中取回指定邮箱、指定主题（subjectKeyword）的最新邮件并解析明文令牌。
func mailToken(t *testing.T, email, subjectKeyword string) string {
	t.Helper()
	token := loginAdmin(t)
	r, _ := do(t, http.MethodGet, "/api/system/dev/outbox", token, nil)
	if r.Code != 0 {
		t.Fatalf("outbox code=%d", r.Code)
	}
	var body struct {
		List []struct {
			ToEmail string `json:"to_email"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		} `json:"list"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatal(err)
	}
	for _, m := range body.List {
		if m.ToEmail != email || !strings.Contains(m.Subject, subjectKeyword) {
			continue
		}
		mm := mailTokenRe.FindStringSubmatch(m.Body)
		if len(mm) == 2 {
			return mm[1]
		}
	}
	t.Fatalf("outbox 中未找到 %s 的令牌", email)
	return ""
}

func TestHealthz(t *testing.T) {
	r, _ := do(t, http.MethodGet, "/healthz", "", nil)
	if r.Code != 0 {
		t.Fatalf("healthz code=%d", r.Code)
	}
}

func TestBootstrapStatus(t *testing.T) {
	r, _ := do(t, http.MethodGet, "/api/system/bootstrap/status", "", nil)
	if r.Code != 0 {
		t.Fatalf("code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		Initialized bool `json:"initialized"`
	}
	_ = json.Unmarshal(r.Body, &body)
	if !body.Initialized {
		t.Fatal("expected initialized=true")
	}
}

func TestBootstrapRejected(t *testing.T) {
	r, _ := do(t, http.MethodPost, "/api/system/bootstrap", "", map[string]string{
		"email":     "x@shuguang.local",
		"initToken": "dev_bootstrap_token_2026",
	})
	if r.Code != 1102 {
		t.Fatalf("bootstrap 重复应返回 1102, got code=%d msg=%s", r.Code, r.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	r, _ := do(t, http.MethodGet, "/api/orders", "", nil)
	if r.Code != 1001 {
		t.Fatalf("无 token 应返回 1001, got %d", r.Code)
	}
}

func TestAdminEndpoints(t *testing.T) {
	token := loginAdmin(t)
	cases := []struct {
		path string
		body string // 非空则校验 body 含该子串
	}{
		{"/api/auth/me", `"role":"super_admin"`},
		{"/api/dashboard", `"todayOrders"`},
		{"/api/orders", `"list"`},
		{"/api/orders/SG202610220086", `"SG202610220086"`},
		{"/api/merchants", `"list"`},
		{"/api/merchants/M00268", `"M00268"`},
		{"/api/merchants/stats", `"total"`},
		{"/api/merchants/audit", `"list"`},
		{"/api/riders", `"list"`},
		{"/api/riders/R00086", `"R00086"`},
		{"/api/riders/stats", `"total"`},
		{"/api/riders/audit", `"list"`},
		{"/api/users", `"list"`},
		{"/api/users/U86540", `"U86540"`},
		{"/api/users/overview", `"total"`},
		{"/api/finance", `"todayIncome"`},
		{"/api/finance/settlement", `"list"`},
		{"/api/finance/withdraws", `"list"`},
		{"/api/finance/withdraw/TX20261022016", `"TX20261022016"`},
		{"/api/complaints", `"list"`},
		{"/api/complaints/GD20261022006", `"GD20261022006"`},
		{"/api/settings", `"areas"`},
		{"/api/settings/fee-rules", `"list"`},
		{"/api/settings/announcements", `"list"`},
		{"/api/system/config/smtp", `"enabled"`},
		{"/api/system/config/oss", `"enabled"`},
		{"/api/system/config/storage/local", `"root_dir"`},
		{"/api/system/config/casbin", `"model"`},
		{"/api/system/dev/outbox", `"list"`},
	}
	for _, c := range cases {
		r, _ := do(t, http.MethodGet, c.path, token, nil)
		if r.Code != 0 {
			t.Errorf("%s code=%d msg=%s", c.path, r.Code, r.Message)
			continue
		}
		if c.body != "" && !bytes.Contains(r.Body, []byte(c.body)) {
			t.Errorf("%s body 缺少 %q: %s", c.path, c.body, string(r.Body))
		}
	}
}

func TestRoleDenied(t *testing.T) {
	setup()
	email := "deny-it@shuguang.local"
	cleanUser(t, email)
	r, _ := do(t, http.MethodPost, "/api/auth/register", "", map[string]string{
		"email": email, "password": "DenyPass123", "name": "权限测试用户",
	})
	if r.Code != 0 {
		t.Fatalf("register code=%d msg=%s", r.Code, r.Message)
	}
	vt := mailToken(t, email, "验证您的邮箱")
	do(t, http.MethodGet, "/api/auth/email/verify?token="+vt, "", nil)

	login, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account": email, "password": "DenyPass123",
	})
	if login.Code != 0 {
		t.Fatalf("login code=%d msg=%s", login.Code, login.Message)
	}
	var body struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(login.Body, &body)
	deny, _ := do(t, http.MethodGet, "/api/merchants", body.Token, nil)
	if deny.Code != 1201 {
		t.Fatalf("普通用户访问商家列表应 1201, got %d msg=%s", deny.Code, deny.Message)
	}
}

func TestRegisterVerifyLogin(t *testing.T) {
	setup()
	email := "reg-it@shuguang.local"
	cleanUser(t, email)
	r, _ := do(t, http.MethodPost, "/api/auth/register", "", map[string]string{
		"email": email, "password": "RegPass123", "name": "集成注册用户",
	})
	if r.Code != 0 {
		t.Fatalf("register code=%d msg=%s", r.Code, r.Message)
	}
	// 未验证登录 → 1003
	login1, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account": email, "password": "RegPass123",
	})
	if login1.Code != 1003 {
		t.Fatalf("未验证登录应 1003, got %d", login1.Code)
	}
	// 验证邮箱（从 dev 发件箱取令牌）
	vt := mailToken(t, email, "验证您的邮箱")
	verify, _ := do(t, http.MethodGet, "/api/auth/email/verify?token="+vt, "", nil)
	if verify.Code != 0 {
		t.Fatalf("verify code=%d msg=%s", verify.Code, verify.Message)
	}
	// 验证后可登录
	login2, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account": email, "password": "RegPass123",
	})
	if login2.Code != 0 {
		t.Fatalf("验证后登录 code=%d msg=%s", login2.Code, login2.Message)
	}
}

func TestForgotResetAndSessionInvalidation(t *testing.T) {
	setup()
	email := "reset-it@shuguang.local"
	cleanUser(t, email)
	r, _ := do(t, http.MethodPost, "/api/auth/register", "", map[string]string{
		"email": email, "password": "OldPass123", "name": "重置测试",
	})
	if r.Code != 0 {
		t.Fatalf("register code=%d", r.Code)
	}
	vt := mailToken(t, email, "验证您的邮箱")
	do(t, http.MethodGet, "/api/auth/email/verify?token="+vt, "", nil)

	login, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{"account": email, "password": "OldPass123"})
	if login.Code != 0 {
		t.Fatalf("login code=%d", login.Code)
	}
	var lb struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(login.Body, &lb)

	f, _ := do(t, http.MethodPost, "/api/auth/password/forgot", "", map[string]string{"email": email})
	if f.Code != 0 {
		t.Fatalf("forgot code=%d msg=%s", f.Code, f.Message)
	}
	rt := mailToken(t, email, "重置您的密码")
	rs, _ := do(t, http.MethodPost, "/api/auth/password/reset", "", map[string]string{
		"token": rt, "new_password": "NewPass456",
	})
	if rs.Code != 0 {
		t.Fatalf("reset code=%d msg=%s", rs.Code, rs.Message)
	}
	me, _ := do(t, http.MethodGet, "/api/auth/me", lb.Token, nil)
	if me.Code != 1006 {
		t.Fatalf("重置后旧 token 应失效 1006, got %d", me.Code)
	}
	login2, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{"account": email, "password": "NewPass456"})
	if login2.Code != 0 {
		t.Fatalf("新密码登录 code=%d", login2.Code)
	}
}

func TestChangePassword(t *testing.T) {
	setup()
	email := "chpw-it@shuguang.local"
	cleanUser(t, email)
	do(t, http.MethodPost, "/api/auth/register", "", map[string]string{"email": email, "password": "ChpwPass123", "name": "改密测试"})
	vt := mailToken(t, email, "验证您的邮箱")
	do(t, http.MethodGet, "/api/auth/email/verify?token="+vt, "", nil)
	login, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{"account": email, "password": "ChpwPass123"})
	var lb struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(login.Body, &lb)
	cp, _ := do(t, http.MethodPost, "/api/auth/password/change", lb.Token, map[string]string{
		"old_password": "ChpwPass123", "new_password": "ChpwPass456",
	})
	if cp.Code != 0 {
		t.Fatalf("change code=%d msg=%s", cp.Code, cp.Message)
	}
	me, _ := do(t, http.MethodGet, "/api/auth/me", lb.Token, nil)
	if me.Code != 1006 {
		t.Fatalf("改密后旧 token 应失效 1006, got %d", me.Code)
	}
}

func TestRefreshRotateAndLogout(t *testing.T) {
	setup()
	email := "refresh-it@shuguang.local"
	cleanUser(t, email)
	reg, _ := do(t, http.MethodPost, "/api/auth/register", "", map[string]string{
		"email": email, "password": "RfshPass123", "name": "刷新用户",
	})
	if reg.Code != 0 {
		t.Fatalf("register code=%d msg=%s", reg.Code, reg.Message)
	}
	vt := mailToken(t, email, "验证您的邮箱")
	vr, _ := do(t, http.MethodGet, "/api/auth/email/verify?token="+vt, "", nil)
	if vr.Code != 0 {
		t.Fatalf("verify code=%d", vr.Code)
	}
	login, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{"account": email, "password": "RfshPass123"})
	if login.Code != 0 {
		t.Fatalf("login code=%d", login.Code)
	}
	var lb struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.Unmarshal(login.Body, &lb)
	rf, _ := do(t, http.MethodPost, "/api/auth/refresh", "", map[string]string{"refresh_token": lb.RefreshToken})
	if rf.Code != 0 {
		t.Fatalf("refresh code=%d msg=%s", rf.Code, rf.Message)
	}
	var rb struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.Unmarshal(rf.Body, &rb)
	if rb.Token == "" || rb.RefreshToken == "" || rb.RefreshToken == lb.RefreshToken {
		t.Fatal("refresh 应轮换并返回新令牌")
	}
	again, _ := do(t, http.MethodPost, "/api/auth/refresh", "", map[string]string{"refresh_token": lb.RefreshToken})
	if again.Code != 1006 {
		t.Fatalf("旧 refresh 复用应 1006, got %d", again.Code)
	}
	lo, _ := do(t, http.MethodPost, "/api/auth/logout", rb.Token, nil)
	if lo.Code != 0 {
		t.Fatalf("logout code=%d", lo.Code)
	}
	me, _ := do(t, http.MethodGet, "/api/auth/me", rb.Token, nil)
	if me.Code != 1006 {
		t.Fatalf("登出后 token 应失效 1006, got %d", me.Code)
	}
}

func TestNotFound(t *testing.T) {
	r, _ := do(t, http.MethodGet, "/api/orders/DOES-NOT-EXIST", loginAdmin(t), nil)
	if r.Code != 1507 {
		t.Fatalf("不存在订单应 1507, got %d", r.Code)
	}
}

func TestMetrics(t *testing.T) {
	setup()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if !bytes.Contains(w.Body.Bytes(), []byte("dawn_http_requests_total")) {
		t.Fatal("metrics 缺少 dawn_http_requests_total")
	}
}

func TestPasswdRoundTrip(t *testing.T) {
	h := passwd.Hasher{}
	hash, err := h.Hash("RoundTrip123")
	if err != nil {
		t.Fatal(err)
	}
	if !h.Verify(hash, "RoundTrip123") {
		t.Fatal("往返验证失败")
	}
	if h.Verify(hash, "WrongPass123") {
		t.Fatal("错误密码不应通过")
	}
}

// TestSystemIdentities 内置系统级身份：初始化后自动创建、可在用户列表检索、禁登录。
func TestSystemIdentities(t *testing.T) {
	setup()
	tok := loginAdmin(t)
	r, _ := do(t, http.MethodGet, "/api/users?keyword=system&pageSize=50", tok, nil)
	if r.Code != 0 {
		t.Fatalf("users 列表 code=%d msg=%s", r.Code, r.Message)
	}
	var body struct {
		List []struct {
			Name string `json:"name"`
		} `json:"list"`
		Total int64 `json:"total"`
	}
	_ = json.Unmarshal(r.Body, &body)
	names := map[string]bool{}
	for _, u := range body.List {
		names[u.Name] = true
	}
	for _, si := range identity.SystemIdentities {
		if !names[si.Name] {
			t.Fatalf("缺少系统身份 %s (%s)", si.Email, si.Name)
		}
	}
	// 系统身份禁登录：无凭证，登录必须失败且不暴露密码比对对象
	bad, _ := do(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"account": "mailer@system.shuguang", "password": "Whatever123",
	})
	if bad.Code != 1002 {
		t.Fatalf("系统身份登录应 1002, got %d", bad.Code)
	}
}

// TestAdminCreateMerchantRiderAndUserQueries 验证商家/骑手创建接口与用户维度订单/投诉查询接口。
func TestAdminCreateMerchantRiderAndUserQueries(t *testing.T) {
	token := loginAdmin(t)

	// 必填校验：缺名称返回 400（HTTP）/ 1301（业务码）
	bad, _ := do(t, http.MethodPost, "/api/merchants", token, map[string]string{"phone": "13800000001"})
	if bad.Code != 1301 {
		t.Fatalf("merchant 缺 name 应 1301, got %d", bad.Code)
	}

	// 创建商家
	mr, _ := do(t, http.MethodPost, "/api/merchants", token, map[string]string{
		"name": "测试接入商家", "phone": "13800000001", "contact": "测试员", "category": "其他",
	})
	if mr.Code != 0 {
		t.Fatalf("create merchant code=%d msg=%s", mr.Code, mr.Message)
	}
	var mCreated struct {
		ID          string `json:"id"`
		ApplyNo     string `json:"applyNo"`
		AuditStatus string `json:"auditStatus"`
	}
	if err := json.Unmarshal(mr.Body, &mCreated); err != nil {
		t.Fatal(err)
	}
	if mCreated.ID == "" || mCreated.ApplyNo == "" || mCreated.AuditStatus != "pending" {
		t.Fatalf("merchant created 缺字段: %+v", mCreated)
	}

	// 创建骑手
	rr, _ := do(t, http.MethodPost, "/api/riders", token, map[string]string{
		"name": "测试接入骑手", "phone": "13800000002", "area": "王家湾镇", "tool": "电动车",
	})
	if rr.Code != 0 {
		t.Fatalf("create rider code=%d msg=%s", rr.Code, rr.Message)
	}
	var rCreated struct {
		ID      string `json:"id"`
		ApplyNo string `json:"applyNo"`
	}
	if err := json.Unmarshal(rr.Body, &rCreated); err != nil {
		t.Fatal(err)
	}
	if rCreated.ID == "" || rCreated.ApplyNo == "" {
		t.Fatalf("rider created 缺字段: %+v", rCreated)
	}

	// 取一个真实用户
	ur, _ := do(t, http.MethodGet, "/api/users?page=1&pageSize=1", token, nil)
	if ur.Code != 0 {
		t.Fatalf("users list code=%d msg=%s", ur.Code, ur.Message)
	}
	var uBody struct {
		List []struct {
			ID string `json:"id"`
		} `json:"list"`
	}
	if err := json.Unmarshal(ur.Body, &uBody); err != nil {
		t.Fatal(err)
	}
	if len(uBody.List) == 0 {
		t.Fatal("users 列表为空")
	}
	uid := uBody.List[0].ID

	// 用户订单记录（分页，空列表亦需 code 0 且结构完整）
	or, _ := do(t, http.MethodGet, "/api/users/"+uid+"/orders?page=1&pageSize=10", token, nil)
	if or.Code != 0 {
		t.Fatalf("user orders code=%d msg=%s", or.Code, or.Message)
	}
	var oBody struct {
		List  []json.RawMessage `json:"list"`
		Total int64             `json:"total"`
	}
	if err := json.Unmarshal(or.Body, &oBody); err != nil {
		t.Fatal(err)
	}

	// 用户投诉售后（分页）
	cr, _ := do(t, http.MethodGet, "/api/users/"+uid+"/complaints?page=1&pageSize=10", token, nil)
	if cr.Code != 0 {
		t.Fatalf("user complaints code=%d msg=%s", cr.Code, cr.Message)
	}
	var cBody struct {
		List  []json.RawMessage `json:"list"`
		Total int64             `json:"total"`
	}
	if err := json.Unmarshal(cr.Body, &cBody); err != nil {
		t.Fatal(err)
	}

	// 清理测试数据
	ctx := context.Background()
	_ = client.Merchant.DeleteOneID(mCreated.ID).Exec(ctx)
	_ = client.Rider.DeleteOneID(rCreated.ID).Exec(ctx)
}

// TestAdminSmsWithdrawRuleAndUserRole 验证短信配置/提现规则/用户角色接口。
func TestAdminSmsWithdrawRuleAndUserRole(t *testing.T) {
	token := loginAdmin(t)

	// 短信配置：GET 懒默认 → PUT 保存 → GET 回读脱敏 → test 校验
	do(t, http.MethodGet, "/api/system/config/sms", token, nil)
	putSMS := map[string]string{
		"provider": "aliyun", "access_key": "LTAI5tTest", "secret_key": "TestSecret123",
		"sign_name": "曙光跑腿", "template_code": "SMS_123456789",
	}
	sp, _ := do(t, http.MethodPut, "/api/system/config/sms", token, putSMS)
	if sp.Code != 0 {
		t.Fatalf("update sms code=%d msg=%s", sp.Code, sp.Message)
	}
	sg, _ := do(t, http.MethodGet, "/api/system/config/sms", token, nil)
	if sg.Code != 0 {
		t.Fatalf("get sms code=%d", sg.Code)
	}
	var sms struct {
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
		SignName  string `json:"sign_name"`
		Enabled   bool   `json:"enabled"`
	}
	if err := json.Unmarshal(sg.Body, &sms); err != nil {
		t.Fatal(err)
	}
	if sms.AccessKey != "LTAI5tTest" || !sms.Enabled || sms.SecretKey != "******" {
		t.Fatalf("sms 回读不符: %+v", sms)
	}
	st, _ := do(t, http.MethodPost, "/api/system/config/sms/test", token, putSMS)
	if st.Code != 0 {
		t.Fatalf("sms test code=%d msg=%s", st.Code, st.Message)
	}

	// 提现规则：GET 懒默认 → PUT → GET 回读
	wg, _ := do(t, http.MethodGet, "/api/system/config/withdraw-rule", token, nil)
	if wg.Code != 0 {
		t.Fatalf("get withdraw-rule code=%d", wg.Code)
	}
	var wdef struct {
		Enabled  bool    `json:"enabled"`
		FeeRate  float64 `json:"fee_rate"`
		Mode     string  `json:"mode"`
	}
	if err := json.Unmarshal(wg.Body, &wdef); err != nil {
		t.Fatal(err)
	}
	if !wdef.Enabled || wdef.FeeRate != 0.006 || wdef.Mode != "manual" {
		t.Fatalf("withdraw-rule 懒默认不符: %+v", wdef)
	}
	wp, _ := do(t, http.MethodPut, "/api/system/config/withdraw-rule", token, map[string]any{
		"enabled": true, "fee_rate": 0.008, "min_amount": 20, "max_amount": 8000,
		"daily_limit": 3, "mode": "auto", "settle_days": 2,
	})
	if wp.Code != 0 {
		t.Fatalf("update withdraw-rule code=%d msg=%s", wp.Code, wp.Message)
	}
	wg2, _ := do(t, http.MethodGet, "/api/system/config/withdraw-rule", token, nil)
	var w2 struct {
		FeeRate   float64 `json:"fee_rate"`
		MinAmount float64 `json:"min_amount"`
		Mode      string  `json:"mode"`
	}
	if err := json.Unmarshal(wg2.Body, &w2); err != nil {
		t.Fatal(err)
	}
	if w2.FeeRate != 0.008 || w2.MinAmount != 20 || w2.Mode != "auto" {
		t.Fatalf("withdraw-rule 回读不符: %+v", w2)
	}

	// 用户角色：列表带角色 → 改角色 → 详情回读 → 改回；改自己被拒
	ur, _ := do(t, http.MethodGet, "/api/users?page=1&pageSize=20", token, nil)
	var uBody struct {
		List []struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		} `json:"list"`
	}
	if err := json.Unmarshal(ur.Body, &uBody); err != nil {
		t.Fatal(err)
	}
	target := ""
	for _, u := range uBody.List {
		if u.Role == identity.RoleUser {
			target = u.ID
			break
		}
	}
	if target == "" {
		t.Fatal("未找到可改角色的普通用户")
	}
	rr, _ := do(t, http.MethodPost, "/api/users/"+target+"/role", token, map[string]string{"role": identity.RoleRegionStation})
	if rr.Code != 0 {
		t.Fatalf("set role code=%d msg=%s", rr.Code, rr.Message)
	}
	dd, _ := do(t, http.MethodGet, "/api/users/"+target, token, nil)
	var uDetail struct {
		Role     string `json:"role"`
		RoleName string `json:"roleName"`
	}
	if err := json.Unmarshal(dd.Body, &uDetail); err != nil {
		t.Fatal(err)
	}
	if uDetail.Role != identity.RoleRegionStation {
		t.Fatalf("角色未生效: %+v", uDetail)
	}
	if rr2, _ := do(t, http.MethodPost, "/api/users/"+target+"/role", token, map[string]string{"role": identity.RoleUser}); rr2.Code != 0 {
		t.Fatalf("还原角色 code=%d msg=%s", rr2.Code, rr2.Message)
	}
	// 改自己：超管自身 → 1105 拒绝
	me, _ := do(t, http.MethodGet, "/api/auth/me", token, nil)
	var meBody struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(me.Body, &meBody); err != nil {
		t.Fatal(err)
	}
	self, _ := do(t, http.MethodPost, "/api/users/"+meBody.User.ID+"/role", token, map[string]string{"role": identity.RoleUser})
	if self.Code != 1105 {
		t.Fatalf("改自己应 1105, got %d", self.Code)
	}

	// 清理：删除测试写入的两个配置键
	ctx := context.Background()
	_, _ = client.SystemSetting.Delete().Where(systemsetting.KeyIn("sms", "withdraw.rule")).Exec(ctx)
}
