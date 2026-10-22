// Package system 系统配置与初始化应用服务。
package system

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/config"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/system"
	"dawn-errand/internal/infrastructure/casbin"
	"dawn-errand/internal/infrastructure/mailer"
	"dawn-errand/internal/infrastructure/storage"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/metrics"
	"dawn-errand/internal/pkg/ulidgen"
)

// Service 系统配置应用服务。
type Service struct {
	Repos    *application.RepoUnit
	Tx       application.TxRunner
	Hasher   application.PasswordHasher
	Cipher   application.ConfigCipher
	Policies application.PolicyManager
	Cfg      *config.Config
	Outbox   application.OutboxProvider
}

const mask = "******"

// EnsureSystemUsers 幂等创建内置系统级身份（邮件/短信/站内信/基础系统服务）。
// 系统身份复用 users 表基础信息、role=system；不创建登录凭证（禁登录）；
// 邮箱已被真实用户占用时跳过，绝不覆盖人工账号。
func (s *Service) EnsureSystemUsers(ctx context.Context) error {
	return s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		now := time.Now().Unix()
		for _, si := range identity.SystemIdentities {
			email := strings.ToLower(strings.TrimSpace(si.Email))
			existing, err := tx.Users.FindByEmail(email)
			if err == nil {
				if existing.Role == identity.RoleSystem {
					continue // 已存在系统身份，幂等跳过
				}
				// 邮箱被真实用户占用：不覆盖
				continue
			}
			if !errors.Is(err, application.ErrNotFound) {
				return err
			}
			if err := tx.Users.Save(&identity.User{
				Email: email, Name: si.Name, Status: identity.UserStatusActive,
				EmailVerified: true, Role: identity.RoleSystem,
				RegisteredAt: now, UpdatedAt: now,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// BootstrapStatus 初始化状态。
func (s *Service) BootstrapStatus(ctx context.Context) (*system.BootstrapState, error) {
	return s.Repos.Settings.GetBootstrap()
}

// BootstrapResult 初始化结果（一次性返回超管初始密码）。
type BootstrapResult struct {
	AdminUserID     string
	AdminEmail      string
	AdminName       string
	OneTimePassword string
}

// Bootstrap 幂等初始化：创建超级管理员、默认策略、系统配置、初始化标记。
func (s *Service) Bootstrap(ctx context.Context, adminEmail, initToken string) (*BootstrapResult, error) {
	st, err := s.Repos.Settings.GetBootstrap()
	if err != nil {
		return nil, err
	}
	if st.Initialized {
		return nil, errcode.New(errcode.CodeAlreadyInit, "系统已完成初始化，禁止重复初始化")
	}
	if initToken == "" || initToken != s.Cfg.BootstrapToken {
		s.Repos.Audit.Record(ctx, application.AuditEntry{
			Action: "system.bootstrap", Object: "bootstrap", Result: "fail",
			RequestID: reqID(ctx), IP: ipFrom(ctx),
		})
		return nil, errcode.NewStatus(403, errcode.CodeBootstrapInvalid, "初始化令牌无效")
	}

	adminEmail = strings.ToLower(strings.TrimSpace(adminEmail))
	if !authValidateEmail(adminEmail) {
		return nil, errcode.New(errcode.CodeBadRequest, "管理员邮箱格式不正确")
	}
	oneTime := randomPassword()
	hash, err := s.Hasher.Hash(oneTime)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	adminID := ulidgen.New()

	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		u := &identity.User{
			ID: adminID, Email: adminEmail, Name: "系统管理员",
			Status: identity.UserStatusActive, EmailVerified: true,
			Role: identity.RoleSuperAdmin, RegisteredAt: now, UpdatedAt: now,
		}
		if err := tx.Users.Save(u); err != nil {
			return err
		}
		if err := tx.Credentials.Save(&identity.Credential{
			UserID: adminID, PasswordHash: hash, SessionVersion: 1, UpdatedAt: now,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 写入默认策略（策略表独立，事务外；失败则返回错误，不标记初始化）
	if err := s.Policies.SavePolicies(casbin.DefaultPolicies()); err != nil {
		return nil, errcode.Wrap(err, errcode.CodeInternal, "默认策略写入失败")
	}
	// 默认系统配置（幂等）
	if err := s.seedDefaultSettings(ctx, adminID, now); err != nil {
		return nil, err
	}
	// 内置系统级身份（邮件/短信/站内信/基础服务，幂等创建）
	if err := s.EnsureSystemUsers(ctx); err != nil {
		return nil, errcode.Wrap(err, errcode.CodeInternal, "系统级身份创建失败")
	}
	if err := s.Repos.Settings.MarkInitialized(adminID, now); err != nil {
		return nil, err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: adminID, Action: "system.bootstrap", Object: "bootstrap", Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx), Operator: "system",
	})
	return &BootstrapResult{
		AdminUserID: adminID, AdminEmail: adminEmail, AdminName: "系统管理员",
		OneTimePassword: oneTime,
	}, nil
}

func (s *Service) seedDefaultSettings(ctx context.Context, adminID string, now int64) error {
	// SMTP：默认禁用
	if _, err := s.Repos.Settings.Get(system.KeySMTP); err == application.ErrNotFound {
		b, _ := json.Marshal(system.SMTPConfig{Host: "", Port: 587, Encryption: "ssl", TimeoutSec: 10})
		if err := s.Repos.Settings.Upsert(&system.SystemSetting{
			Key: system.KeySMTP, Status: system.SettingDisabled,
			ValueJSON: string(b), UpdatedBy: adminID, UpdatedAt: now,
		}); err != nil {
			return err
		}
	}
	// SMS：默认禁用
	if _, err := s.Repos.Settings.Get(system.KeySMS); err == application.ErrNotFound {
		b, _ := json.Marshal(system.SMSConfig{Provider: "aliyun"})
		if err := s.Repos.Settings.Upsert(&system.SystemSetting{
			Key: system.KeySMS, Status: system.SettingDisabled,
			ValueJSON: string(b), UpdatedBy: adminID, UpdatedAt: now,
		}); err != nil {
			return err
		}
	}
	// 提现规则：默认启用（手工审核，T+1 到账，0.6% 手续费）
	if _, err := s.Repos.Settings.Get(system.KeyWithdrawRule); err == application.ErrNotFound {
		b, _ := json.Marshal(system.WithdrawRuleConfig{
			Enabled: true, FeeRate: 0.006, MinAmount: 10, MaxAmount: 5000,
			DailyLimit: 5, Mode: "manual", SettleDays: 1,
		})
		if err := s.Repos.Settings.Upsert(&system.SystemSetting{
			Key: system.KeyWithdrawRule, Status: system.SettingEnabled,
			ValueJSON: string(b), UpdatedBy: adminID, UpdatedAt: now,
		}); err != nil {
			return err
		}
	}
	// OSS：默认禁用
	if _, err := s.Repos.Settings.Get(system.KeyOSS); err == application.ErrNotFound {
		b, _ := json.Marshal(system.OSSConfig{Provider: "oss"})
		if err := s.Repos.Settings.Upsert(&system.SystemSetting{
			Key: system.KeyOSS, Status: system.SettingDisabled,
			ValueJSON: string(b), UpdatedBy: adminID, UpdatedAt: now,
		}); err != nil {
			return err
		}
	}
	// 本地存储：默认启用（根目录来自环境变量）
	if _, err := s.Repos.Settings.Get(system.KeyLocal); err == application.ErrNotFound {
		b, _ := json.Marshal(system.LocalStorageConfig{
			RootDir: s.Cfg.StorageRootDir, PublicBaseURL: s.Cfg.StoragePublicURL,
			MaxSizeMB:   s.Cfg.StorageMaxSizeMB,
			AllowedMIME: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
			PathPrefix:  "uploads",
		})
		if err := s.Repos.Settings.Upsert(&system.SystemSetting{
			Key: system.KeyLocal, Status: system.SettingEnabled,
			ValueJSON: string(b), UpdatedBy: adminID, UpdatedAt: now,
		}); err != nil {
			return err
		}
	}
	// Casbin 配置（model + 默认策略）
	if _, err := s.Repos.Settings.Get(system.KeyCasbin); err == application.ErrNotFound {
		b, _ := json.Marshal(system.CasbinConfig{
			Model: casbin.ModelText(), Policy: casbin.PolicyText(casbin.DefaultPolicies()),
			PolicyStore: "db",
		})
		if err := s.Repos.Settings.Upsert(&system.SystemSetting{
			Key: system.KeyCasbin, Status: system.SettingEnabled,
			ValueJSON: string(b), UpdatedBy: adminID, UpdatedAt: now,
		}); err != nil {
			return err
		}
	}
	return nil
}

// ---- SMTP ----

// GetSMTP 读取 SMTP 配置（脱敏返回）。
func (s *Service) GetSMTP(ctx context.Context) (*system.SMTPConfig, error) {
	cfg, err := s.loadSMTP(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.Password != "" {
		cfg.Password = mask
	}
	return cfg, nil
}

func (s *Service) loadSMTP(ctx context.Context) (*system.SMTPConfig, error) {
	rec, err := s.Repos.Settings.Get(system.KeySMTP)
	if err != nil {
		if err == application.ErrNotFound {
			return &system.SMTPConfig{}, nil
		}
		return nil, err
	}
	var cfg system.SMTPConfig
	if err := json.Unmarshal([]byte(rec.ValueJSON), &cfg); err != nil {
		return nil, err
	}
	if cfg.Password != "" && cfg.Password != mask {
		if plain, err := s.Cipher.Decrypt(cfg.Password); err == nil {
			cfg.Password = plain
		}
	}
	return &cfg, nil
}

// UpdateSMTP 更新 SMTP 配置（密码留空/掩码则保持不变）。
func (s *Service) UpdateSMTP(ctx context.Context, in system.SMTPConfig, operator string) error {
	if strings.TrimSpace(in.Host) == "" || in.Port == 0 {
		return errcode.New(errcode.CodeSettingInvalid, "SMTP host 与 port 必填")
	}
	cur, err := s.loadSMTP(ctx)
	if err != nil {
		return err
	}
	if in.Password == "" || in.Password == mask {
		in.Password = cur.Password
	} else if in.Password != "" {
		enc, err := s.Cipher.Encrypt(in.Password)
		if err != nil {
			return err
		}
		in.Password = enc
	}
	now := time.Now().Unix()
	status := system.SettingEnabled
	if in.Host == "" {
		status = system.SettingDisabled
	}
	b, _ := json.Marshal(in)
	if err := s.Repos.Settings.Upsert(&system.SystemSetting{
		Key: system.KeySMTP, Status: status, ValueJSON: string(b),
		UpdatedBy: operator, UpdatedAt: now,
	}); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "config.smtp.update", Object: "smtp", Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx),
	})
	metricsConfigChange("smtp")
	return nil
}

// TestSMTP 测试 SMTP 连接（不落库）。
func (s *Service) TestSMTP(ctx context.Context, in system.SMTPConfig) error {
	return mailer.TestSMTP(ctx, &in)
}

// ---- OSS ----

// GetOSS 读取 OSS 配置（脱敏）。
func (s *Service) GetOSS(ctx context.Context) (*system.OSSConfig, error) {
	rec, err := s.Repos.Settings.Get(system.KeyOSS)
	if err != nil {
		if err == application.ErrNotFound {
			return &system.OSSConfig{}, nil
		}
		return nil, err
	}
	var cfg system.OSSConfig
	if err := json.Unmarshal([]byte(rec.ValueJSON), &cfg); err != nil {
		return nil, err
	}
	if cfg.SecretKey != "" && cfg.SecretKey != mask {
		if plain, err := s.Cipher.Decrypt(cfg.SecretKey); err == nil {
			cfg.SecretKey = plain
		}
	}
	if cfg.SecretKey != "" {
		cfg.SecretKey = mask
	}
	return &cfg, nil
}

// UpdateOSS 更新 OSS 配置。
func (s *Service) UpdateOSS(ctx context.Context, in system.OSSConfig, operator string) error {
	cur, err := s.GetOSS(ctx)
	if err != nil {
		return err
	}
	if in.SecretKey == "" || in.SecretKey == mask {
		in.SecretKey = cur.SecretKey
	} else {
		enc, err := s.Cipher.Encrypt(in.SecretKey)
		if err != nil {
			return err
		}
		in.SecretKey = enc
	}
	now := time.Now().Unix()
	status := system.SettingEnabled
	if in.Endpoint == "" || in.Bucket == "" {
		status = system.SettingDisabled
	}
	b, _ := json.Marshal(in)
	if err := s.Repos.Settings.Upsert(&system.SystemSetting{
		Key: system.KeyOSS, Status: status, ValueJSON: string(b),
		UpdatedBy: operator, UpdatedAt: now,
	}); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "config.oss.update", Object: "oss", Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx),
	})
	metricsConfigChange("oss")
	return nil
}

// ---- 本地存储 ----

// GetLocalStorage 读取本地存储配置。
func (s *Service) GetLocalStorage(ctx context.Context) (*system.LocalStorageConfig, error) {
	rec, err := s.Repos.Settings.Get(system.KeyLocal)
	if err != nil {
		if err == application.ErrNotFound {
			return &system.LocalStorageConfig{}, nil
		}
		return nil, err
	}
	var cfg system.LocalStorageConfig
	if err := json.Unmarshal([]byte(rec.ValueJSON), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpdateLocalStorage 更新本地存储配置。
func (s *Service) UpdateLocalStorage(ctx context.Context, in system.LocalStorageConfig, operator string) error {
	if strings.TrimSpace(in.RootDir) == "" {
		return errcode.New(errcode.CodeSettingInvalid, "root_dir 必填")
	}
	now := time.Now().Unix()
	b, _ := json.Marshal(in)
	if err := s.Repos.Settings.Upsert(&system.SystemSetting{
		Key: system.KeyLocal, Status: system.SettingEnabled, ValueJSON: string(b),
		UpdatedBy: operator, UpdatedAt: now,
	}); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "config.storage.local.update", Object: "storage.local", Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx),
	})
	metricsConfigChange("storage.local")
	return nil
}

// TestLocalStorage 测试本地存储。
func (s *Service) TestLocalStorage(ctx context.Context, in system.LocalStorageConfig) error {
	return storage.TestLocal(ctx, &in)
}

// ---- 短信服务 ----

// GetSMS 读取短信服务配置（脱敏）。
func (s *Service) GetSMS(ctx context.Context) (*system.SMSConfig, error) {
	cfg, err := s.loadSMS(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.SecretKey != "" {
		cfg.SecretKey = mask
	}
	return cfg, nil
}

func (s *Service) loadSMS(ctx context.Context) (*system.SMSConfig, error) {
	rec, err := s.Repos.Settings.Get(system.KeySMS)
	if err != nil {
		if err == application.ErrNotFound {
			return &system.SMSConfig{Provider: "aliyun"}, nil
		}
		return nil, err
	}
	var cfg system.SMSConfig
	if err := json.Unmarshal([]byte(rec.ValueJSON), &cfg); err != nil {
		return nil, err
	}
	if cfg.SecretKey != "" && cfg.SecretKey != mask {
		if plain, err := s.Cipher.Decrypt(cfg.SecretKey); err == nil {
			cfg.SecretKey = plain
		}
	}
	return &cfg, nil
}

// UpdateSMS 更新短信服务配置（SecretKey 留空/掩码则保持不变）。
func (s *Service) UpdateSMS(ctx context.Context, in system.SMSConfig, operator string) error {
	if strings.TrimSpace(in.AccessKey) == "" || strings.TrimSpace(in.SignName) == "" {
		return errcode.New(errcode.CodeSettingInvalid, "access_key 与 sign_name 必填")
	}
	if in.Provider == "" {
		in.Provider = "aliyun"
	}
	cur, err := s.loadSMS(ctx)
	if err != nil {
		return err
	}
	if in.SecretKey == "" || in.SecretKey == mask {
		in.SecretKey = cur.SecretKey
	} else {
		enc, err := s.Cipher.Encrypt(in.SecretKey)
		if err != nil {
			return err
		}
		in.SecretKey = enc
	}
	now := time.Now().Unix()
	status := system.SettingEnabled
	if in.AccessKey == "" || in.SignName == "" {
		status = system.SettingDisabled
	}
	b, _ := json.Marshal(in)
	if err := s.Repos.Settings.Upsert(&system.SystemSetting{
		Key: system.KeySMS, Status: status, ValueJSON: string(b),
		UpdatedBy: operator, UpdatedAt: now,
	}); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "config.sms.update", Object: "sms", Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx),
	})
	metricsConfigChange("sms")
	return nil
}

// TestSMS 校验短信服务配置（不落库）。
// 正式发送通道依赖服务商 SDK 与已备案签名，配置阶段仅做完整性校验。
func (s *Service) TestSMS(ctx context.Context, in system.SMSConfig) error {
	if strings.TrimSpace(in.AccessKey) == "" || strings.TrimSpace(in.SignName) == "" {
		return errcode.New(errcode.CodeSettingInvalid, "access_key 与 sign_name 必填")
	}
	switch in.Provider {
	case "aliyun", "tencent", "other":
	default:
		return errcode.New(errcode.CodeSMSTestFailed, "不支持的服务商: "+in.Provider)
	}
	return nil
}

// ---- 提现规则 ----

// GetWithdrawRule 读取提现规则配置（未初始化时返回默认值，不落库）。
func (s *Service) GetWithdrawRule(ctx context.Context) (*system.WithdrawRuleConfig, error) {
	rec, err := s.Repos.Settings.Get(system.KeyWithdrawRule)
	if err != nil {
		if err == application.ErrNotFound {
			return &system.WithdrawRuleConfig{
				Enabled: true, FeeRate: 0.006, MinAmount: 10, MaxAmount: 5000,
				DailyLimit: 5, Mode: "manual", SettleDays: 1,
			}, nil
		}
		return nil, err
	}
	var cfg system.WithdrawRuleConfig
	if err := json.Unmarshal([]byte(rec.ValueJSON), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpdateWithdrawRule 更新提现规则配置。
func (s *Service) UpdateWithdrawRule(ctx context.Context, in system.WithdrawRuleConfig, operator string) error {
	if in.FeeRate < 0 || in.MinAmount < 0 || in.MaxAmount < 0 {
		return errcode.New(errcode.CodeSettingInvalid, "费率与金额不能为负数")
	}
	if in.MaxAmount > 0 && in.MinAmount > 0 && in.MinAmount > in.MaxAmount {
		return errcode.New(errcode.CodeSettingInvalid, "最低提现金额不能高于最高金额")
	}
	if in.DailyLimit < 0 || in.SettleDays < 0 {
		return errcode.New(errcode.CodeSettingInvalid, "次数与到账周期不能为负数")
	}
	if in.Mode == "" {
		in.Mode = "manual"
	}
	if in.Mode != "auto" && in.Mode != "manual" {
		return errcode.New(errcode.CodeSettingInvalid, "审核模式仅支持 auto/manual")
	}
	now := time.Now().Unix()
	status := system.SettingEnabled
	if !in.Enabled {
		status = system.SettingDisabled
	}
	b, _ := json.Marshal(in)
	if err := s.Repos.Settings.Upsert(&system.SystemSetting{
		Key: system.KeyWithdrawRule, Status: status, ValueJSON: string(b),
		UpdatedBy: operator, UpdatedAt: now,
	}); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "config.withdraw_rule.update", Object: "withdraw.rule", Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx),
	})
	metricsConfigChange("withdraw.rule")
	return nil
}

// ---- Casbin 配置 ----

// GetCasbinConfig 读取 Casbin 配置。
func (s *Service) GetCasbinConfig(ctx context.Context) (*system.CasbinConfig, error) {
	rec, err := s.Repos.Settings.Get(system.KeyCasbin)
	if err != nil {
		if err == application.ErrNotFound {
			return &system.CasbinConfig{Model: casbin.ModelText(), PolicyStore: "db"}, nil
		}
		return nil, err
	}
	var cfg system.CasbinConfig
	if err := json.Unmarshal([]byte(rec.ValueJSON), &cfg); err != nil {
		return nil, err
	}
	// 始终以当前生效策略为准
	cfg.Policy = casbin.PolicyText(s.Policies.Policies())
	cfg.Model = casbin.ModelText()
	return &cfg, nil
}

// UpdateCasbinConfig 更新策略（替换并重载，审计）。
func (s *Service) UpdateCasbinConfig(ctx context.Context, policies [][]string, operator string) error {
	if len(policies) == 0 {
		return errcode.New(errcode.CodeSettingInvalid, "策略不能为空")
	}
	for _, p := range policies {
		if len(p) < 3 {
			return errcode.New(errcode.CodeSettingInvalid, "非法策略行："+strings.Join(p, ", "))
		}
	}
	if err := s.Policies.SavePolicies(policies); err != nil {
		return errcode.Wrap(err, errcode.CodeSettingInvalid, "策略校验或写入失败")
	}
	now := time.Now().Unix()
	b, _ := json.Marshal(system.CasbinConfig{
		Model: casbin.ModelText(), Policy: casbin.PolicyText(policies), PolicyStore: "db",
	})
	if err := s.Repos.Settings.Upsert(&system.SystemSetting{
		Key: system.KeyCasbin, Status: system.SettingEnabled, ValueJSON: string(b),
		UpdatedBy: operator, UpdatedAt: now,
	}); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "permission.policy.update", Object: "casbin", Result: "success",
		RequestID: reqID(ctx), IP: ipFrom(ctx),
	})
	metricsConfigChange("casbin")
	return nil
}

// DevOutbox 查询 dev 邮件发件箱。
func (s *Service) DevOutbox(ctx context.Context, limit int) ([]application.MailEntry, error) {
	if s.Outbox == nil {
		return nil, nil
	}
	return s.Outbox(ctx, limit)
}

// ---- 工具 ----

func randomPassword() string {
	const letters = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789"
	b := make([]byte, 16)
	idx := make([]byte, 16)
	if _, err := crand.Read(idx); err != nil {
		for i := range b {
			b[i] = letters[(ulidgen.New()[0]+byte(i))%uint8(len(letters))]
		}
		return string(b)
	}
	for i := range b {
		b[i] = letters[int(idx[i])%len(letters)]
	}
	return string(b)
}

func authValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func reqID(ctx context.Context) string  { return application.RequestID(ctx) }
func ipFrom(ctx context.Context) string { return application.ClientIP(ctx) }

func metricsConfigChange(key string) {
	metrics.ConfigChangesTotal.WithLabelValues(key).Inc()
}
