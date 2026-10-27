package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
)

// R5-5 认证用例单测：注册/验证/登录/验证码/重置/刷新/登出/主体。
//
// integration 用例已经覆盖「HTTP 全链路能跑通」，这一层锁的是各分支的**判定顺序**
// 与**副作用**：先限流再查库（防枚举与爆破）、密码校验通过才谈封禁/未验证、
// 改密必须撤销全部会话、错误码必须与三端约定一致。替身见 service_fakes_test.go。

func authCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var e *errcode.Error
	if errors.As(err, &e) {
		return e.Code
	}
	t.Fatalf("非业务错误，无法断言错误码: %v", err)
	return -1
}

func TestRegisterValidatesAndProvisions(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()

	if code := authCodeOf(t, mustAuthErr(u.svc.Register(ctx, RegisterInput{Email: "bad", Password: "abcd1234"}))); code != errcode.CodeBadRequest {
		t.Errorf("非法邮箱应 1301, got=%d", code)
	}
	if code := authCodeOf(t, mustAuthErr(u.svc.Register(ctx, RegisterInput{Email: "a@b.co", Password: "short"}))); code != errcode.CodePasswordPolicy {
		t.Errorf("弱密码应报密码策略, got=%d", code)
	}
	u.addUser("U1", "taken@example.com", "abcd1234")
	if code := authCodeOf(t, mustAuthErr(u.svc.Register(ctx, RegisterInput{
		Email: " TAKEN@example.com ", Password: "abcd1234"}))); code != errcode.CodeEmailAlreadyExists {
		t.Errorf("重复邮箱应 1004, got=%d", code)
	}

	res, err := u.svc.Register(ctx, RegisterInput{Email: " New.User@Example.COM ", Password: "abcd1234", Name: "  "})
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	if res.User.Email != "new.user@example.com" {
		t.Errorf("邮箱未归一化: %q", res.User.Email)
	}
	// 空昵称回落到邮箱前缀。
	if res.User.Name != "new.user" {
		t.Errorf("空昵称应回落为邮箱前缀: %q", res.User.Name)
	}
	stored := u.users.byEmail["new.user@example.com"]
	if stored == nil || stored.Role != identity.RoleUser || stored.EmailVerified {
		t.Errorf("新用户应未验证邮箱且角色为 user: %+v", stored)
	}
	if c := u.creds.byUser[stored.ID]; c == nil || c.PasswordHash != "argon2:abcd1234" {
		t.Errorf("凭证未落库或未哈希: %+v", c)
	}
	if len(u.emails.created) != 1 || len(u.emails.created[0].TokenHash) != 64 {
		t.Errorf("邮箱验证令牌应为 sha256 十六进制: %+v", u.emails.created)
	}
	if len(u.mailer.sent) != 1 || u.mailer.sent[0].to != "new.user@example.com" {
		t.Errorf("验证邮件未发送: %+v", u.mailer.sent)
	}
	// 令牌明文不得出现在库里（只存哈希）。
	for _, tok := range u.emails.created {
		if strings.Contains(tok.TokenHash, "new.user") || len(tok.TokenHash) != 64 {
			t.Errorf("令牌哈希形态异常: %+v", tok)
		}
	}

	// 超长昵称按 32 个字符截断（按 rune 计，不是字节）。
	long := strings.Repeat("名", 40)
	if _, err := u.svc.Register(ctx, RegisterInput{Email: "long@example.com", Password: "abcd1234", Name: long}); err != nil {
		t.Fatal(err)
	}
	if got := u.users.byEmail["long@example.com"].Name; len([]rune(got)) != 32 {
		t.Errorf("超长昵称未按字符截断: %d", len([]rune(got)))
	}

	// 发信失败不阻断注册（邮件可重发，账号已建）。
	u2 := newAuthUnit()
	u2.mailer.sendErr = errors.New("smtp down")
	if _, err := u2.svc.Register(ctx, RegisterInput{Email: "m@example.com", Password: "abcd1234"}); err != nil {
		t.Errorf("发信失败不应阻断注册: %v", err)
	}
}

func TestVerifyEmailBranches(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()
	u.addUser("U1", "a@example.com", "abcd1234")
	u.users.byID["U1"].EmailVerified = false

	if code := authCodeOf(t, u.svc.VerifyEmail(ctx, "nope")); code != errcode.CodeTokenInvalid {
		t.Errorf("未知令牌应 1006, got=%d", code)
	}

	expired := "expired-token"
	u.emails.byHash[hashToken(expired)] = &identity.EmailVerificationToken{
		ID: "T1", UserID: "U1", TokenHash: hashToken(expired), ExpiresAt: time.Now().Unix() - 1,
	}
	if code := authCodeOf(t, u.svc.VerifyEmail(ctx, expired)); code != errcode.CodeTokenInvalid {
		t.Errorf("过期令牌应 1006, got=%d", code)
	}
	if u.users.byID["U1"].EmailVerified {
		t.Error("过期令牌却完成了验证")
	}

	ok := "ok-token"
	u.emails.byHash[hashToken(ok)] = &identity.EmailVerificationToken{
		ID: "T2", UserID: "U1", TokenHash: hashToken(ok), ExpiresAt: time.Now().Unix() + 60,
	}
	if err := u.svc.VerifyEmail(ctx, "  "+ok+"  "); err != nil {
		t.Fatalf("验证失败: %v", err)
	}
	if !u.users.byID["U1"].EmailVerified {
		t.Error("验证成功后未标记邮箱已验证")
	}
	if len(u.emails.markUsed) != 1 || u.emails.markUsed[0] != "T2" {
		t.Errorf("令牌未标记已使用: %v", u.emails.markUsed)
	}
	if len(u.audit.entries) != 1 || u.audit.entries[0].Action != "auth.email.verify" {
		t.Errorf("验证未留审计: %+v", u.audit.entries)
	}
}

func TestResendVerificationBranches(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()
	u.addUser("U1", "a@example.com", "abcd1234")
	u.users.byID["U1"].EmailVerified = false

	u.limiter.allow = false
	if code := authCodeOf(t, u.svc.ResendVerification(ctx, "a@example.com")); code != errcode.CodeTooManyRequest {
		t.Errorf("限流应 1007, got=%d", code)
	}
	u.limiter.allow = true

	// 防枚举：账号不存在返回 nil，且不发信。
	if err := u.svc.ResendVerification(ctx, "ghost@example.com"); err != nil {
		t.Errorf("未知邮箱应静默成功: %v", err)
	}
	if len(u.mailer.sent) != 0 {
		t.Error("未知邮箱不应发信")
	}
	// 已验证不再重发。
	u.users.byID["U1"].EmailVerified = true
	if code := authCodeOf(t, u.svc.ResendVerification(ctx, "a@example.com")); code != errcode.CodeConflict {
		t.Errorf("已验证邮箱应 1402, got=%d", code)
	}

	u.users.byID["U1"].EmailVerified = false
	if err := u.svc.ResendVerification(ctx, " A@Example.com "); err != nil {
		t.Fatalf("重发失败: %v", err)
	}
	if len(u.emails.created) != 1 || len(u.mailer.sent) != 1 {
		t.Errorf("重发应生成新令牌并发送: tokens=%d mails=%d", len(u.emails.created), len(u.mailer.sent))
	}
	if ttl := u.emails.created[0].ExpiresAt - time.Now().Unix(); ttl < 23*3600 || ttl > 24*3600+5 {
		t.Errorf("验证令牌有效期应约 24h, got=%ds", ttl)
	}
}

func TestLoginGuardsAndAudits(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()
	u.addUser("U1", "a@example.com", "Good1234")
	u.addUser("U2", "b@example.com", "Good1234", func(x *identity.User) { x.Status = identity.UserStatusBanned })
	u.addUser("U3", "c@example.com", "Good1234", func(x *identity.User) { x.EmailVerified = false })
	u.addUser("U4", "13800138000", "", func(x *identity.User) { x.Phone = "13800138000" })

	u.limiter.allow = false
	if code := authCodeOf(t, mustLoginErr(u.svc.Login(ctx, "a@example.com", "Good1234"))); code != errcode.CodeTooManyRequest {
		t.Errorf("登录限流应 1007, got=%d", code)
	}
	u.limiter.allow = true

	// 账号不存在与密码错误必须同码同文案（防枚举）。
	notFound := authCodeOf(t, mustLoginErr(u.svc.Login(ctx, "ghost@example.com", "Good1234")))
	wrongPw := authCodeOf(t, mustLoginErr(u.svc.Login(ctx, "a@example.com", "Bad1234")))
	if notFound != errcode.CodeInvalidCredentials || wrongPw != errcode.CodeInvalidCredentials {
		t.Errorf("账号不存在/密码错误应同为 1002, got=%d/%d", notFound, wrongPw)
	}
	// 有用户无凭证：同样落到 1002，而不是暴露「无凭证」。
	u.addUser("U5", "nocred@example.com", "")
	if code := authCodeOf(t, mustLoginErr(u.svc.Login(ctx, "nocred@example.com", "Good1234"))); code != errcode.CodeInvalidCredentials {
		t.Errorf("无凭证账号应 1002, got=%d", code)
	}
	// 封禁 / 未验证：密码正确之后才判定。
	if code := authCodeOf(t, mustLoginErr(u.svc.Login(ctx, "b@example.com", "Good1234"))); code != errcode.CodeUserBanned {
		t.Errorf("封禁账号应 1005, got=%d", code)
	}
	if code := authCodeOf(t, mustLoginErr(u.svc.Login(ctx, "c@example.com", "Good1234"))); code != errcode.CodeEmailNotVerified {
		t.Errorf("未验证邮箱应 1003, got=%d", code)
	}

	// 成功：签发一对令牌 + 权限。
	u.enforcer.policies[identity.RoleUser] = []string{"user:read", "order:create"}
	res, err := u.svc.Login(ctx, " A@Example.com ", "Good1234")
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if res.AccessToken != "at-for-U1" || res.RefreshToken != "rt-U1" {
		t.Errorf("令牌未签发: %+v", res)
	}
	if len(res.Permissions) != 2 {
		t.Errorf("权限未随登录下发: %v", res.Permissions)
	}
	if claims := u.tokens.accessIssued[0]; claims.UserID != "U1" || claims.SessionVersion != 1 {
		t.Errorf("access token 应带会话版本: %+v", claims)
	}
	if last := u.audit.entries[len(u.audit.entries)-1]; last.Action != "auth.login" || last.Result != "success" {
		t.Errorf("登录成功未留审计: %+v", last)
	}
	// 手机号登录路径（账号不含 @ → FindByPhone）；U4 无凭证，应落 1002。
	if code := authCodeOf(t, mustLoginErr(u.svc.Login(ctx, "13800138000", "Good1234"))); code != errcode.CodeInvalidCredentials {
		t.Errorf("手机号无凭证应 1002, got=%d", code)
	}
	u.addUser("U6", "phone@example.com", "Good1234", func(x *identity.User) { x.Phone = "13600136000" })
	if res, err := u.svc.Login(ctx, "13600136000", "Good1234"); err != nil || res.User.ID != "U6" {
		t.Errorf("手机号登录应命中 FindByPhone: res=%+v err=%v", res, err)
	}
}

func TestSendSMSCodeBranches(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()

	if code := authCodeOf(t, mustCodeErr(u.svc.SendSMSCode(ctx, "123"))); code != errcode.CodeBadRequest {
		t.Errorf("非法手机号应 1301, got=%d", code)
	}
	u.limiter.allow = false
	if code := authCodeOf(t, mustCodeErr(u.svc.SendSMSCode(ctx, "13800138000"))); code != errcode.CodeTooManyRequest {
		t.Errorf("发送限流应 1007, got=%d", code)
	}
	u.limiter.allow = true
	u.codes.setErr = errors.New("redis down")
	if code := authCodeOf(t, mustCodeErr(u.svc.SendSMSCode(ctx, "13800138000"))); code != errcode.CodeSMSSendFailed {
		t.Errorf("验证码写入失败应 1109, got=%d", code)
	}
	u.codes.setErr = nil

	// dev 环境回显验证码，且与存储一致（联调用，不落日志）。
	res, err := u.svc.SendSMSCode(ctx, " 13800138000 ")
	if err != nil {
		t.Fatalf("发送失败: %v", err)
	}
	if len(res.DevCode) != 6 || u.codes.values["dawn:sms:code:13800138000"] != res.DevCode {
		t.Errorf("dev 验证码未回显或未存储: %+v", res)
	}
	// 生产环境不回显。
	u.cfg.AppEnv = "prod"
	res, err = u.svc.SendSMSCode(ctx, "13800138001")
	if err != nil {
		t.Fatal(err)
	}
	if res.DevCode != "" {
		t.Errorf("生产环境不应回显验证码: %+v", res)
	}
}

func TestLoginBySMSBranches(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()
	u.addUser("U1", "13800138000@x.com", "", func(x *identity.User) {
		x.Phone, x.EmailVerified = "13800138000", true
	})

	if code := authCodeOf(t, mustLoginErr(u.svc.LoginBySMS(ctx, "123", "123456"))); code != errcode.CodeBadRequest {
		t.Errorf("非法手机号应 1301, got=%d", code)
	}
	// 无验证码。
	if code := authCodeOf(t, mustLoginErr(u.svc.LoginBySMS(ctx, "13800138000", "123456"))); code != errcode.CodeCaptchaInvalid {
		t.Errorf("无验证码应 1008, got=%d", code)
	}
	// 错误验证码 + 达到失败上限 → 立即作废验证码。
	u.codes.values["dawn:sms:code:13800138000"] = "111111"
	u.limiter.allow = false
	if code := authCodeOf(t, mustLoginErr(u.svc.LoginBySMS(ctx, "13800138000", "222222"))); code != errcode.CodeCaptchaInvalid {
		t.Errorf("错误验证码应 1008, got=%d", code)
	}
	if len(u.codes.deleted) != 1 {
		t.Errorf("连续失败应作废验证码: %v", u.codes.deleted)
	}
	// 正确验证码：一次性，消费成功后删除。
	u.limiter.allow = true
	u.codes.values["dawn:sms:code:13800138000"] = "111111"
	res, err := u.svc.LoginBySMS(ctx, " 13800138000 ", "111111")
	if err != nil {
		t.Fatalf("验证码登录失败: %v", err)
	}
	if res.User.ID != "U1" || res.AccessToken != "at-for-U1" {
		t.Errorf("验证码登录结果错误: %+v", res)
	}
	if _, ok := u.codes.values["dawn:sms:code:13800138000"]; ok {
		t.Error("验证码应为一次性，成功后必须删除")
	}

	// 新手机号：自动注册（用户 + 无密码凭证 + 资料行），再登录。
	u.codes.values["dawn:sms:code:13900139000"] = "333333"
	res, err = u.svc.LoginBySMS(ctx, "13900139000", "333333")
	if err != nil {
		t.Fatalf("新手机号自动注册失败: %v", err)
	}
	nu := res.User
	if nu.Phone != "13900139000" || !nu.EmailVerified || nu.Role != identity.RoleUser {
		t.Errorf("自动注册用户字段错误: %+v", nu)
	}
	if c := u.creds.byUser[nu.ID]; c == nil || c.PasswordHash != passwordlessHash {
		t.Errorf("自动注册应写占位凭证: %+v", c)
	}
	if len(u.profiles.saved) != 1 || u.profiles.saved[0].UserID != nu.ID {
		t.Errorf("自动注册应初始化资料行: %+v", u.profiles.saved)
	}

	// 封禁用户：验证码正确也不放行。
	u.addUser("U9", "13700137000@x.com", "", func(x *identity.User) {
		x.Phone, x.EmailVerified, x.Status = "13700137000", true, identity.UserStatusBanned
	})
	u.codes.values["dawn:sms:code:13700137000"] = "444444"
	if code := authCodeOf(t, mustLoginErr(u.svc.LoginBySMS(ctx, "13700137000", "444444"))); code != errcode.CodeUserBanned {
		t.Errorf("封禁用户验证码登录应 1005, got=%d", code)
	}
}

func TestPasswordResetFlow(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()
	u.addUser("U1", "a@example.com", "Old12345")

	u.limiter.allow = false
	if code := authCodeOf(t, u.svc.ForgotPassword(ctx, "a@example.com")); code != errcode.CodeTooManyRequest {
		t.Errorf("忘记密码限流应 1007, got=%d", code)
	}
	u.limiter.allow = true
	if err := u.svc.ForgotPassword(ctx, "ghost@example.com"); err != nil {
		t.Errorf("未知邮箱应静默成功: %v", err)
	}
	if len(u.mailer.sent) != 0 {
		t.Error("未知邮箱不应发信")
	}
	if err := u.svc.ForgotPassword(ctx, " A@Example.com "); err != nil {
		t.Fatalf("忘记密码失败: %v", err)
	}
	if len(u.pwTokens.created) != 1 || len(u.mailer.sent) != 1 {
		t.Errorf("重置令牌/邮件缺失: tokens=%d mails=%d", len(u.pwTokens.created), len(u.mailer.sent))
	}
	if ttl := u.pwTokens.created[0].ExpiresAt - time.Now().Unix(); ttl < 29*60 || ttl > 30*60+5 {
		t.Errorf("重置令牌有效期应约 30min, got=%ds", ttl)
	}

	if code := authCodeOf(t, u.svc.ResetPassword(ctx, "whatever", "short")); code != errcode.CodePasswordPolicy {
		t.Errorf("弱密码应 1009, got=%d", code)
	}
	if code := authCodeOf(t, u.svc.ResetPassword(ctx, "unknown", "New12345")); code != errcode.CodeTokenInvalid {
		t.Errorf("未知重置令牌应 1006, got=%d", code)
	}
	expired := "expired-reset"
	u.pwTokens.byHash[hashToken(expired)] = &identity.PasswordResetToken{
		ID: "R1", UserID: "U1", TokenHash: hashToken(expired), ExpiresAt: time.Now().Unix() - 1,
	}
	if code := authCodeOf(t, u.svc.ResetPassword(ctx, expired, "New12345")); code != errcode.CodeTokenInvalid {
		t.Errorf("过期重置令牌应 1006, got=%d", code)
	}

	// 成功：换哈希、会话版本 +1（踢掉旧 access token）、令牌一次性、撤销全部刷新令牌。
	live := "live-reset"
	u.pwTokens.byHash[hashToken(live)] = &identity.PasswordResetToken{
		ID: "R2", UserID: "U1", TokenHash: hashToken(live), ExpiresAt: time.Now().Unix() + 60,
	}
	if err := u.svc.ResetPassword(ctx, " "+live+" ", "New12345"); err != nil {
		t.Fatalf("重置密码失败: %v", err)
	}
	cred := u.creds.byUser["U1"]
	if cred.PasswordHash != "argon2:New12345" || cred.SessionVersion != 2 {
		t.Errorf("重置后凭证错误: %+v", cred)
	}
	if len(u.pwTokens.markUsed) != 1 || u.pwTokens.markUsed[0] != "R2" {
		t.Errorf("重置令牌未标记已使用: %v", u.pwTokens.markUsed)
	}
	if len(u.pwTokens.deletedBy) != 1 || u.pwTokens.deletedBy[0] != "U1" {
		t.Errorf("改密后应清空该用户重置令牌: %v", u.pwTokens.deletedBy)
	}
	if len(u.tokens.revokedAll) != 1 || u.tokens.revokedAll[0] != "U1" {
		t.Errorf("改密后应撤销全部会话: %v", u.tokens.revokedAll)
	}
	if last := u.audit.entries[len(u.audit.entries)-1]; last.Action != "auth.password.reset" {
		t.Errorf("重置未留审计: %+v", last)
	}
	// 旧密码立即失效（哈希已换）。
	if code := authCodeOf(t, mustLoginErr(u.svc.Login(ctx, "a@example.com", "Old12345"))); code != errcode.CodeInvalidCredentials {
		t.Errorf("旧密码应失效, got=%d", code)
	}
}

func TestChangePasswordBranches(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()
	u.addUser("U1", "a@example.com", "Old12345")

	if code := authCodeOf(t, u.svc.ChangePassword(ctx, "U1", "Old12345", "short")); code != errcode.CodePasswordPolicy {
		t.Errorf("弱密码应 1009, got=%d", code)
	}
	if code := authCodeOf(t, u.svc.ChangePassword(ctx, "U1", "Wrong1234", "New12345")); code != errcode.CodeOldPasswordWrong {
		t.Errorf("原密码错误应 1010, got=%d", code)
	}
	// 无凭证用户改密：先要求原密码，无凭证即原密码错误（不能凭空凭证越权设密）。
	u.addUser("U2", "b@example.com", "")
	if code := authCodeOf(t, u.svc.ChangePassword(ctx, "U2", "", "New12345")); code != errcode.CodeOldPasswordWrong {
		t.Errorf("无凭证改密应 1010, got=%d", code)
	}
	if err := u.svc.ChangePassword(ctx, "U1", "Old12345", "New12345"); err != nil {
		t.Fatalf("改密失败: %v", err)
	}
	if cred := u.creds.byUser["U1"]; cred.PasswordHash != "argon2:New12345" || cred.SessionVersion != 2 {
		t.Errorf("改密后凭证错误: %+v", cred)
	}
	if len(u.pwTokens.deletedBy) != 1 {
		t.Error("改密后应清空重置令牌")
	}
}

func TestRefreshLogoutMeAndPrincipal(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()
	u.addUser("U1", "a@example.com", "Good1234")
	u.enforcer.policies[identity.RoleUser] = []string{"user:read"}

	// Refresh：校验失败 / 用户不存在 / 封禁 / 成功轮换。
	u.tokens.verifyErr = errors.New("bad")
	if code := authCodeOf(t, mustLoginErr(u.svc.Refresh(ctx, "U1", "rt-x"))); code != errcode.CodeTokenInvalid {
		t.Errorf("无效刷新令牌应 1006, got=%d", code)
	}
	u.tokens.verifyErr = nil
	if code := authCodeOf(t, mustLoginErr(u.svc.Refresh(ctx, "U404", "rt-x"))); code != errcode.CodeTokenInvalid {
		t.Errorf("用户不存在应 1006, got=%d", code)
	}
	u.users.byID["U1"].Status = identity.UserStatusBanned
	if code := authCodeOf(t, mustLoginErr(u.svc.Refresh(ctx, "U1", "rt-x"))); code != errcode.CodeUserBanned {
		t.Errorf("封禁用户刷新应 1005, got=%d", code)
	}
	u.users.byID["U1"].Status = identity.UserStatusActive
	res, err := u.svc.Refresh(ctx, "U1", "rt-old")
	if err != nil {
		t.Fatalf("刷新失败: %v", err)
	}
	if res.AccessToken != "at-for-U1" || len(u.tokens.revokedOne) != 1 || u.tokens.revokedOne[0][1] != "rt-old" {
		t.Errorf("刷新应撤销旧刷新令牌并签发新的: %+v", u.tokens.revokedOne)
	}

	// RefreshAll：先 bump 会话版本，再清空全部刷新令牌。
	res, err = u.svc.RefreshAll(ctx, "U1", "rt-old")
	if err != nil {
		t.Fatalf("全端刷新失败: %v", err)
	}
	if u.creds.bumps != 1 {
		t.Errorf("全端刷新应提升会话版本, bumps=%d", u.creds.bumps)
	}
	if len(u.tokens.revokedAll) != 1 {
		t.Errorf("全端刷新应清空全部刷新令牌: %v", u.tokens.revokedAll)
	}
	if claims := u.tokens.accessIssued[len(u.tokens.accessIssued)-1]; claims.SessionVersion != 2 {
		t.Errorf("新 access token 应带提升后的会话版本: %+v", claims)
	}

	// Logout：当前设备只撤销当前令牌；全端撤销 + bump。
	if err := u.svc.Logout(ctx, "U1", "", false); err != nil {
		t.Errorf("无刷新令牌的当前设备登出应无事发生: %v", err)
	}
	beforeOne := len(u.tokens.revokedOne)
	if err := u.svc.Logout(ctx, "U1", "rt-cur", false); err != nil {
		t.Fatal(err)
	}
	if len(u.tokens.revokedOne) != beforeOne+1 || u.creds.bumps != 1 {
		t.Error("当前设备登出不应提升会话版本，也不应撤销其他端")
	}
	if err := u.svc.Logout(ctx, "U1", "rt-cur", true); err != nil {
		t.Fatal(err)
	}
	if len(u.tokens.revokedAll) != 2 || u.creds.bumps != 2 {
		t.Error("全端登出应撤销全部刷新令牌并提升会话版本")
	}

	// Me。
	if code := authCodeOf(t, mustLoginErr(u.svc.Me(ctx, "U404"))); code != errcode.CodeTokenInvalid {
		t.Errorf("用户不存在应 1006, got=%d", code)
	}
	me, err := u.svc.Me(ctx, "U1")
	if err != nil {
		t.Fatal(err)
	}
	if me.User.ID != "U1" || len(me.Permissions) != 1 {
		t.Errorf("Me 返回错误: %+v", me)
	}

	// GetPrincipal：把用户上的租户绑定搬进数据隔离主体。
	u.users.byID["U1"].RegionID = "RG1"
	u.users.byID["U1"].MerchantID = "M1"
	p, err := u.svc.GetPrincipal(ctx, "U1")
	if err != nil {
		t.Fatalf("取主体失败: %v", err)
	}
	if p.UserID != "U1" || p.RegionID != "RG1" || p.MerchantID != "M1" || p.Role != identity.RoleUser {
		t.Errorf("主体字段错误: %+v", p)
	}
	u.users.byID["U1"].Status = identity.UserStatusBanned
	if code := authCodeOf(t, mustPrincipalErr(u.svc.GetPrincipal(ctx, "U1"))); code != errcode.CodeUserBanned {
		t.Errorf("封禁用户不应拿到主体, got=%d", code)
	}
}

func TestEnsureCredentialBackfillsPasswordlessRow(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()
	u.addUser("U1", "a@example.com", "")

	cred, err := u.svc.ensureCredential(ctx, "U1", 100)
	if err != nil {
		t.Fatalf("补凭证失败: %v", err)
	}
	if cred.PasswordHash != passwordlessHash || cred.SessionVersion != 0 {
		t.Errorf("补出的凭证应为无密码占位: %+v", cred)
	}
	// 并发：Save 冲突但另一请求已写入 → 回读成功，不把登录打死。
	u.addUser("U2", "b@example.com", "")
	u.creds.conflictSave = true
	cred2, err := u.svc.ensureCredential(ctx, "U2", 100)
	if err != nil {
		t.Fatalf("并发补凭证应回读成功: %v", err)
	}
	if cred2 == nil || cred2.UserID != "U2" {
		t.Errorf("回读结果错误: %+v", cred2)
	}
	// 已存在时直接返回，不重复写。
	before := len(u.creds.byUser)
	if _, err := u.svc.ensureCredential(ctx, "U1", 200); err != nil {
		t.Fatal(err)
	}
	if len(u.creds.byUser) != before {
		t.Error("已有凭证不应重复写入")
	}
}

func TestIssueTokensPropagatesFailures(t *testing.T) {
	ctx := context.Background()
	u := newAuthUnit()
	usr := u.addUser("U1", "a@example.com", "Good1234")

	u.tokens.issueErr = errors.New("token backend down")
	if _, err := u.svc.issueTokens(ctx, usr, "auth.login"); !errors.Is(err, u.tokens.issueErr) {
		t.Errorf("签发失败应上抛: %v", err)
	}
	u.tokens.issueErr = nil
	res, err := u.svc.issueTokens(ctx, usr, "auth.login")
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}
	if len(res.Permissions) == 0 {
		t.Error("权限应随签发返回")
	}
	if len(u.audit.entries) != 1 || u.audit.entries[0].Action != "auth.login" {
		t.Errorf("签发应留审计: %+v", u.audit.entries)
	}
}

// --- 断言辅助 ---

func mustAuthErr(_ *UserResult, err error) error            { return err }
func mustLoginErr(_ *LoginResult, err error) error          { return err }
func mustCodeErr(_ *SendSMSCodeResult, err error) error     { return err }
func mustPrincipalErr(_ *tenant.Principal, err error) error { return err }
