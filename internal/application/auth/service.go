// Package auth 认证应用服务：注册/邮箱验证/登录/登出/改密/重置/令牌。
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/config"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/userprofile"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// Service 认证应用服务。
type Service struct {
	Repos    *application.RepoUnit
	Tx       application.TxRunner
	Hasher   application.PasswordHasher
	Tokens   application.TokenManager
	Mailer   application.MailerPort
	Cipher   application.ConfigCipher
	Cfg      *config.Config
	Codes    application.CodeStore
	Enforcer interface {
		PermissionsForRole(role string) []string
	}
	Limiter application.RateLimiter
}

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// ValidateEmail 校验邮箱格式。
func ValidateEmail(email string) bool { return emailRe.MatchString(email) }

// ValidatePassword 密码策略：8-72 位，须含字母与数字。
func ValidatePassword(pw string) error {
	if len(pw) < 8 || len(pw) > 72 {
		return errcode.New(errcode.CodePasswordPolicy, "密码长度需为 8-72 位")
	}
	hasLetter, hasDigit := false, false
	for _, r := range pw {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errcode.New(errcode.CodePasswordPolicy, "密码需同时包含字母和数字")
	}
	return nil
}

// RegisterInput 注册入参。
type RegisterInput struct {
	Email    string
	Password string
	Name     string
	Phone    string
}

// UserResult 注册/查询返回。
type UserResult struct {
	User *identity.User
}

// Register 注册：创建用户+凭证+邮箱验证令牌（事务），随后发送验证邮件。
func (s *Service) Register(ctx context.Context, in RegisterInput) (*UserResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if !ValidateEmail(email) {
		return nil, errcode.New(errcode.CodeBadRequest, "邮箱格式不正确")
	}
	if err := ValidatePassword(in.Password); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = email[:strings.Index(email, "@")]
	}
	if len([]rune(name)) > 32 {
		name = string([]rune(name)[:32])
	}

	var userID string
	var plainToken string
	err := s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if _, err := tx.Users.FindByEmail(email); err == nil {
			return errcode.New(errcode.CodeEmailAlreadyExists, "该邮箱已注册")
		}
		hash, err := s.Hasher.Hash(in.Password)
		if err != nil {
			return err
		}
		now := time.Now().Unix()
		u := &identity.User{
			ID: ulidgen.New(), Email: email, Name: name, Phone: strings.TrimSpace(in.Phone),
			Status: identity.UserStatusActive, Role: identity.RoleUser, RegisteredAt: now, UpdatedAt: now,
		}
		if err := tx.Users.Save(u); err != nil {
			return err
		}
		if err := tx.Credentials.Save(&identity.Credential{
			UserID: u.ID, PasswordHash: hash, UpdatedAt: now,
		}); err != nil {
			return err
		}
		plainToken = newToken()
		if err := tx.EmailTokens.Create(&identity.EmailVerificationToken{
			ID: ulidgen.New(), UserID: u.ID, TokenHash: hashToken(plainToken),
			ExpiresAt: now + 24*3600,
		}); err != nil {
			return err
		}
		userID = u.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 事务外发送验证邮件（失败不阻断注册）。
	body := "点击链接完成邮箱验证：/auth/email/verify?token=" + plainToken +
		"<br/>验证码：<b>" + plainToken + "</b>（24 小时内有效）"
	_ = s.Mailer.Send(ctx, email, "【曙光跑腿】验证您的邮箱", body)
	return &UserResult{User: &identity.User{ID: userID, Email: email, Name: name}}, nil
}

// VerifyEmail 邮箱验证。
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	rec, err := s.Repos.EmailTokens.FindByHash(hashToken(strings.TrimSpace(token)))
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return errcode.New(errcode.CodeTokenInvalid, "验证链接无效或已使用")
		}
		return err
	}
	if rec.ExpiresAt < time.Now().Unix() {
		return errcode.New(errcode.CodeTokenInvalid, "验证链接已过期")
	}
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if err := tx.EmailTokens.MarkUsed(rec.ID); err != nil {
			return err
		}
		u, err := tx.Users.FindByID(rec.UserID)
		if err != nil {
			return err
		}
		u.EmailVerified = true
		u.UpdatedAt = time.Now().Unix()
		return tx.Users.Save(u)
	})
	if err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: rec.UserID, Action: "auth.email.verify", Object: rec.UserID, Result: "success",
		RequestID: reqID(ctx),
	})
	return nil
}

// ResendVerification 重发验证邮件（限流；防枚举）。
func (s *Service) ResendVerification(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	ok, err := s.Limiter.Allow(ctx, "auth:resend:"+email, 3, time.Minute)
	if err != nil {
		return err
	}
	if !ok {
		return errcode.New(errcode.CodeTooManyRequest, "请求过于频繁，请稍后再试")
	}
	u, err := s.Repos.Users.FindByEmail(email)
	if err != nil {
		return nil // 防枚举
	}
	if u.EmailVerified {
		return errcode.New(errcode.CodeConflict, "邮箱已验证")
	}
	plain := newToken()
	if err := s.Repos.EmailTokens.Create(&identity.EmailVerificationToken{
		ID: ulidgen.New(), UserID: u.ID, TokenHash: hashToken(plain),
		ExpiresAt: time.Now().Unix() + 24*3600,
	}); err != nil {
		return err
	}
	body := "点击链接完成邮箱验证：/auth/email/verify?token=" + plain +
		"<br/>验证码：<b>" + plain + "</b>（24 小时内有效）"
	_ = s.Mailer.Send(ctx, email, "【曙光跑腿】验证您的邮箱", body)
	return nil
}

// LoginResult 登录结果。
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         *identity.User
	Permissions  []string
}

// Login 登录：邮箱/手机号 + 密码（防枚举、限流、审计）。
func (s *Service) Login(ctx context.Context, account, password string) (*LoginResult, error) {
	account = strings.ToLower(strings.TrimSpace(account))
	ok, err := s.Limiter.Allow(ctx, "auth:login:"+account, 10, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errcode.New(errcode.CodeTooManyRequest, "尝试次数过多，请 5 分钟后再试")
	}

	var u *identity.User
	if ValidateEmail(account) {
		u, err = s.Repos.Users.FindByEmail(account)
	} else {
		u, err = s.Repos.Users.FindByPhone(account)
	}
	if err != nil {
		s.Repos.Audit.Record(ctx, application.AuditEntry{
			Action: "auth.login", Object: account, Result: "fail",
			RequestID: reqID(ctx), IP: ipFrom(ctx),
		})
		// 防计时侧信道：账号不存在也执行一次真实 Argon2 校验，
		// 使「不存在 / 密码错误」两条路径 CPU 耗时同量级。
		s.dummyVerify(password)
		// 防枚举：账号不存在与密码错误同响应
		return nil, errcode.NewStatus(401, errcode.CodeInvalidCredentials, "账号或密码错误")
	}
	cred, err := s.Repos.Credentials.FindByUserID(u.ID)
	if err != nil {
		s.dummyVerify(password)
		return nil, errcode.NewStatus(401, errcode.CodeInvalidCredentials, "账号或密码错误")
	}
	if !s.Hasher.Verify(cred.PasswordHash, password) {
		s.Repos.Audit.Record(ctx, application.AuditEntry{
			UserID: u.ID, Action: "auth.login", Object: u.ID, Result: "fail",
			RequestID: reqID(ctx), IP: ipFrom(ctx),
		})
		return nil, errcode.NewStatus(401, errcode.CodeInvalidCredentials, "账号或密码错误")
	}
	if !u.IsActive() {
		return nil, errcode.NewStatus(403, errcode.CodeUserBanned, "账号已被封禁")
	}
	if !u.EmailVerified {
		return nil, errcode.NewStatus(403, errcode.CodeEmailNotVerified, "请先验证邮箱后再登录")
	}

	return s.issueTokens(ctx, u, "auth.login")
}

// dummyHash 首次使用时按当前 Argon2 参数生成，之后复用，避免真实用户存在。
var (
	dummyHashOnce sync.Once
	dummyHash     string
)

// dummyVerify 对占位哈希执行一次真实校验，抹平「账号不存在」与「密码错误」的耗时差异。
func (s *Service) dummyVerify(password string) {
	dummyHashOnce.Do(func() {
		dummyHash, _ = s.Hasher.Hash("dawn-errand-timing-equalizer")
	})
	_ = s.Hasher.Verify(dummyHash, password)
}

// SendSMSCodeResult 发送验证码结果；dev 环境回显验证码便于联调（不落日志）。
type SendSMSCodeResult struct {
	DevCode string `json:"dev_code,omitempty"`
}

// SendSMSCode 发送手机验证码：60 秒限流一次，验证码有效期 5 分钟。
// 短信服务商 SDK 未接入前（配置仅校验完整性），dev 环境直接回显验证码；
// prod 环境发送通道为占位，接入服务商后替换为真实下发（见实施进度记录）。
func (s *Service) SendSMSCode(ctx context.Context, phone string) (*SendSMSCodeResult, error) {
	phone = strings.TrimSpace(phone)
	if !identity.ValidatePhone(phone) {
		return nil, errcode.New(errcode.CodeBadRequest, "手机号格式不正确")
	}
	ok, err := s.Limiter.Allow(ctx, "auth:sms:send:"+phone, 1, time.Minute)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errcode.New(errcode.CodeTooManyRequest, "发送过于频繁，请 60 秒后再试")
	}
	code := smsCode()
	if err := s.Codes.Set(ctx, "dawn:sms:code:"+phone, code, 5*time.Minute); err != nil {
		return nil, errcode.Wrap(err, errcode.CodeSMSSendFailed, "验证码发送失败，请稍后再试")
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		Action: "auth.sms.send", Object: phone, Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx),
	})
	if s.Cfg.IsDev() {
		return &SendSMSCodeResult{DevCode: code}, nil
	}
	return &SendSMSCodeResult{}, nil
}

// LoginBySMS 手机号 + 验证码登录；手机号未注册时自动创建账号并登录。
// 验证码校验成功后立即作废（一次性）。
func (s *Service) LoginBySMS(ctx context.Context, phone, code string) (*LoginResult, error) {
	phone = strings.TrimSpace(phone)
	if !identity.ValidatePhone(phone) {
		return nil, errcode.New(errcode.CodeBadRequest, "手机号格式不正确")
	}
	code = strings.TrimSpace(code)
	key := "dawn:sms:code:" + phone
	stored, err := s.Codes.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if stored == "" || len(code) != 6 || subtle.ConstantTimeCompare([]byte(stored), []byte(code)) != 1 {
		s.Repos.Audit.Record(ctx, application.AuditEntry{
			Action: "auth.sms.login", Object: phone, Result: "fail",
			RequestID: reqID(ctx), IP: ipFrom(ctx),
		})
		// 防爆破：按手机号计数，连续 5 次错误立即作废验证码（重新获取成本高于猜测收益）。
		if stored != "" {
			ok, lerr := s.Limiter.Allow(ctx, "auth:sms:fail:"+phone, 5, 10*time.Minute)
			if lerr == nil && !ok {
				_ = s.Codes.Delete(ctx, key)
			}
		}
		return nil, errcode.NewStatus(401, errcode.CodeCaptchaInvalid, "验证码错误或已过期")
	}
	if err := s.Codes.Delete(ctx, key); err != nil {
		return nil, err
	}
	u, err := s.Repos.Users.FindByPhone(phone)
	if errors.Is(err, application.ErrNotFound) {
		u, err = s.autoRegisterByPhone(ctx, phone)
	}
	if err != nil {
		return nil, err
	}
	if !u.IsActive() {
		return nil, errcode.NewStatus(403, errcode.CodeUserBanned, "账号已被封禁")
	}
	return s.issueTokens(ctx, u, "auth.sms.login")
}

// autoRegisterByPhone 手机号验证码登录时自动注册（邮箱生成占位地址，实名/密码后续可补）。
func (s *Service) autoRegisterByPhone(ctx context.Context, phone string) (*identity.User, error) {
	now := time.Now().Unix()
	u := &identity.User{
		ID:            ulidgen.New(),
		Email:         "sms." + phone + "@shuguang.local",
		Name:          "用户" + phone[len(phone)-4:],
		Phone:         phone,
		Status:        identity.UserStatusActive,
		Role:          identity.RoleUser,
		EmailVerified: true, // 手机验证即完成身份验证
		RegisteredAt:  now,
		UpdatedAt:     now,
	}
	err := s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if err := tx.Users.Save(u); err != nil {
			return err
		}
		// 无密码账号同样需要凭证行：认证中间件以 session_version 校验会话有效性。
		if err := tx.Credentials.Save(&identity.Credential{
			UserID: u.ID, PasswordHash: passwordlessHash, UpdatedAt: now,
		}); err != nil {
			return err
		}
		// 初始化用户资料行（个人中心展示余额/积分等资产）。
		return tx.Profiles.Save(&userprofile.Profile{
			UserID: u.ID, Level: "Lv.1 新用户", Verified: "未实名",
			RegisterChannel: "sms_login", UpdatedAt: now,
		})
	})
	if err != nil {
		return nil, err
	}
	return u, nil
}

// passwordlessHash 无密码账号（手机验证码/后台开通未设密码）的占位哈希。
// 该值不是合法 Argon2id 格式，Verify 必返回 false，因此不能用于密码登录。
const passwordlessHash = "!"

// ensureCredential 确保用户存在凭证行；无密码账号使用占位哈希。
// 认证中间件要求 credential 存在以校验 session_version，缺失会导致登录后全部 401/1006。
func (s *Service) ensureCredential(ctx context.Context, userID string, now int64) (*identity.Credential, error) {
	cred, err := s.Repos.Credentials.FindByUserID(userID)
	if err == nil {
		return cred, nil
	}
	if !errors.Is(err, application.ErrNotFound) {
		return nil, err
	}
	c := &identity.Credential{
		UserID: userID, PasswordHash: passwordlessHash, SessionVersion: 0, UpdatedAt: now,
	}
	if err := s.Repos.Credentials.Save(c); err != nil {
		// 并发下可能已被其他请求创建，回读一次。
		if again, e := s.Repos.Credentials.FindByUserID(userID); e == nil {
			return again, nil
		}
		return nil, err
	}
	return c, nil
}

// issueTokens 统一签发访问/刷新令牌并记录审计。
func (s *Service) issueTokens(ctx context.Context, u *identity.User, action string) (*LoginResult, error) {
	now := time.Now().Unix()
	cred, err := s.ensureCredential(ctx, u.ID, now)
	if err != nil {
		return nil, err
	}
	sv := cred.SessionVersion
	access, err := s.Tokens.IssueAccessToken(application.TokenClaims{
		UserID: u.ID, Role: u.Role, SessionVersion: sv,
	}, s.Cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}
	refresh, _, err := s.Tokens.IssueRefreshToken(ctx, u.ID, s.Cfg.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: u.ID, Action: action, Object: u.ID, Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx), Operator: u.Name,
	})
	return &LoginResult{
		AccessToken: access, RefreshToken: refresh, User: u,
		Permissions: s.Enforcer.PermissionsForRole(u.Role),
	}, nil
}

// RefreshUserID 由明文刷新令牌解析用户 ID（公开 refresh 入口）。
func (s *Service) RefreshUserID(ctx context.Context, refreshToken string) (string, error) {
	return s.Tokens.UserIDForRefreshToken(ctx, refreshToken)
}

// Refresh 刷新访问令牌（轮换刷新令牌）。
func (s *Service) Refresh(ctx context.Context, userID, refreshToken string) (*LoginResult, error) {
	if err := s.Tokens.VerifyRefreshToken(ctx, userID, refreshToken); err != nil {
		return nil, errcode.NewStatus(401, errcode.CodeTokenInvalid, "刷新令牌无效或已过期")
	}
	u, err := s.Repos.Users.FindByID(userID)
	if err != nil {
		return nil, errcode.NewStatus(401, errcode.CodeTokenInvalid, "用户不存在")
	}
	if !u.IsActive() {
		return nil, errcode.NewStatus(403, errcode.CodeUserBanned, "账号已被封禁")
	}
	cred, err := s.ensureCredential(ctx, userID, time.Now().Unix())
	if err != nil {
		return nil, errcode.NewStatus(401, errcode.CodeTokenInvalid, "凭证不存在")
	}
	// 轮换：撤销旧刷新令牌
	_ = s.Tokens.RevokeRefreshToken(ctx, userID, refreshToken)
	access, err := s.Tokens.IssueAccessToken(application.TokenClaims{
		UserID: u.ID, Role: u.Role, SessionVersion: cred.SessionVersion,
	}, s.Cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}
	refresh, _, err := s.Tokens.IssueRefreshToken(ctx, u.ID, s.Cfg.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken: access, RefreshToken: refresh, User: u,
		Permissions: s.Enforcer.PermissionsForRole(u.Role),
	}, nil
}

// RefreshAll 刷新令牌并吊销该用户其他全部会话（异地登录 / 踢出其他设备）。
// 与 Refresh 的区别：先提升会话版本使其他设备 access token 立即失效，
// 再清空全部刷新令牌并只为当前设备签发新的一对。
func (s *Service) RefreshAll(ctx context.Context, userID, refreshToken string) (*LoginResult, error) {
	if err := s.Tokens.VerifyRefreshToken(ctx, userID, refreshToken); err != nil {
		return nil, errcode.NewStatus(401, errcode.CodeTokenInvalid, "刷新令牌无效或已过期")
	}
	u, err := s.Repos.Users.FindByID(userID)
	if err != nil {
		return nil, errcode.NewStatus(401, errcode.CodeTokenInvalid, "用户不存在")
	}
	if !u.IsActive() {
		return nil, errcode.NewStatus(403, errcode.CodeUserBanned, "账号已被封禁")
	}
	if _, err := s.ensureCredential(ctx, userID, time.Now().Unix()); err != nil {
		return nil, errcode.NewStatus(401, errcode.CodeTokenInvalid, "凭证不存在")
	}
	ver, err := s.Repos.Credentials.BumpSessionVersion(userID)
	if err != nil {
		return nil, err
	}
	if err := s.Tokens.RevokeRefreshTokens(ctx, userID); err != nil {
		return nil, err
	}
	access, err := s.Tokens.IssueAccessToken(application.TokenClaims{
		UserID: u.ID, Role: u.Role, SessionVersion: ver,
	}, s.Cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}
	refresh, _, err := s.Tokens.IssueRefreshToken(ctx, u.ID, s.Cfg.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken: access, RefreshToken: refresh, User: u,
		Permissions: s.Enforcer.PermissionsForRole(u.Role),
	}, nil
}

// Logout 登出。
// allDevices=true：撤销该用户全部刷新令牌并提升会话版本，所有设备的 access token 立即失效；
// allDevices=false：仅撤销当前设备持有的刷新令牌，其他设备会话不受影响，
// 当前设备的 access token 靠客户端本地清除 + 短 TTL 自然过期（不提升会话版本，否则会误伤其他端）。
func (s *Service) Logout(ctx context.Context, userID, refreshToken string, allDevices bool) error {
	if !allDevices {
		if refreshToken != "" {
			return s.Tokens.RevokeRefreshToken(ctx, userID, refreshToken)
		}
		return nil
	}
	if err := s.Tokens.RevokeRefreshTokens(ctx, userID); err != nil {
		return err
	}
	// 提升会话版本使已签发 access token 全部失效。
	if _, err := s.Repos.Credentials.BumpSessionVersion(userID); err != nil {
		return err
	}
	return nil
}

// Me 当前用户信息。
func (s *Service) Me(ctx context.Context, userID string) (*LoginResult, error) {
	u, err := s.Repos.Users.FindByID(userID)
	if err != nil {
		return nil, errcode.NewStatus(401, errcode.CodeTokenInvalid, "用户不存在")
	}
	return &LoginResult{
		User: u, Permissions: s.Enforcer.PermissionsForRole(u.Role),
	}, nil
}

// ForgotPassword 忘记密码：发送重置邮件（防枚举）。
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	ok, err := s.Limiter.Allow(ctx, "auth:forgot:"+email, 3, time.Minute)
	if err != nil {
		return err
	}
	if !ok {
		return errcode.New(errcode.CodeTooManyRequest, "请求过于频繁，请稍后再试")
	}
	u, err := s.Repos.Users.FindByEmail(email)
	if err != nil {
		return nil // 防枚举
	}
	plain := newToken()
	if err := s.Repos.PasswordTokens.Create(&identity.PasswordResetToken{
		ID: ulidgen.New(), UserID: u.ID, TokenHash: hashToken(plain),
		ExpiresAt: time.Now().Unix() + 30*60,
	}); err != nil {
		return err
	}
	body := "点击链接重置密码：/auth/password/reset?token=" + plain +
		"<br/>重置码：<b>" + plain + "</b>（30 分钟内有效）"
	_ = s.Mailer.Send(ctx, email, "【曙光跑腿】重置您的密码", body)
	return nil
}

// ResetPassword 重置密码：新密码生效并撤销全部会话。
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}
	rec, err := s.Repos.PasswordTokens.FindByHash(hashToken(strings.TrimSpace(token)))
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return errcode.New(errcode.CodeTokenInvalid, "重置链接无效或已使用")
		}
		return err
	}
	if rec.ExpiresAt < time.Now().Unix() {
		return errcode.New(errcode.CodeTokenInvalid, "重置链接已过期")
	}
	return s.changePassword(ctx, rec.UserID, rec.ID, "", newPassword, "auth.password.reset", true)
}

// ChangePassword 修改密码：旧密码校验 + 会话撤销。
func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}
	return s.changePassword(ctx, userID, "", oldPassword, newPassword, "auth.password.change", false)
}

func (s *Service) changePassword(ctx context.Context, userID, resetTokenID, oldPassword, newPassword, action string, isReset bool) error {
	err := s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		cred, err := tx.Credentials.FindByUserID(userID)
		if err != nil && !errors.Is(err, application.ErrNotFound) {
			return err
		}
		if !isReset {
			if cred == nil || !s.Hasher.Verify(cred.PasswordHash, oldPassword) {
				return errcode.New(errcode.CodeOldPasswordWrong, "原密码错误")
			}
		}
		hash, err := s.Hasher.Hash(newPassword)
		if err != nil {
			return err
		}
		next := int64(1)
		if cred != nil {
			next = cred.SessionVersion + 1
		}
		now := time.Now().Unix()
		if err := tx.Credentials.Save(&identity.Credential{
			UserID: userID, PasswordHash: hash, SessionVersion: next, UpdatedAt: now,
		}); err != nil {
			return err
		}
		if resetTokenID != "" {
			if err := tx.PasswordTokens.MarkUsed(resetTokenID); err != nil {
				return err
			}
		}
		return tx.PasswordTokens.DeleteByUser(userID)
	})
	if err != nil {
		return err
	}
	// 事务外撤销全部会话
	_ = s.Tokens.RevokeRefreshTokens(ctx, userID)
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: userID, Action: action, Object: userID, Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx),
	})
	return nil
}

// GetPrincipal 构建数据隔离主体。
func (s *Service) GetPrincipal(ctx context.Context, userID string) (*tenant.Principal, error) {
	u, err := s.Repos.Users.FindByID(userID)
	if err != nil {
		return nil, errcode.NewStatus(401, errcode.CodeTokenInvalid, "用户不存在")
	}
	if !u.IsActive() {
		return nil, errcode.NewStatus(403, errcode.CodeUserBanned, "账号已被封禁")
	}
	return &tenant.Principal{
		UserID:     u.ID,
		Role:       u.Role,
		RegionID:   u.RegionID,
		MerchantID: u.MerchantID,
		RiderID:    u.RiderID,
	}, nil
}

// ---- 工具 ----

func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ulidgen.New() + ulidgen.New()
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// smsCode 生成 6 位数字验证码（crypto/rand 均匀取样，无取模偏差）。
func smsCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

func reqID(ctx context.Context) string { return application.RequestID(ctx) }

func ipFrom(ctx context.Context) string { return application.ClientIP(ctx) }
