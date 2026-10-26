package admin

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

var accountEmailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// AccountsService 商家/骑手账号开通。
// 账号开通是「主体 + 用户 + 凭证 + 双向绑定」的跨表事务，必须由 application 层编排。
type AccountsService struct {
	Repos  *application.RepoUnit
	Tx     application.TxRunner
	Hasher application.PasswordHasher
}

// ProvisionInput 开通入参。Email / Phone 至少一个；Password 为空时生成一次性密码。
type ProvisionInput struct {
	Email    string
	Phone    string
	Password string
	Name     string
}

// ProvisionResult 开通结果。InitialPassword 仅在系统生成密码时返回一次。
type ProvisionResult struct {
	UserID          string
	Email           string
	Phone           string
	Name            string
	Role            string
	MerchantID      string
	RiderID         string
	InitialPassword string
}

// ResetMerchantPassword 为已开通商家重置登录密码（旧会话立即失效）。
// sc 为操作者数据范围：非全局角色只能操作归属范围内的商家。
func (s *AccountsService) ResetMerchantPassword(ctx context.Context, merchantID, operator string, sc tenant.Scope) (*ProvisionResult, error) {
	m, err := s.Repos.Merchants.FindByID(ctx, merchantID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
	}
	if !merchantAccessOK(sc, m) {
		return nil, tenantDenied()
	}
	if m.OwnerUserID == "" {
		return nil, errcode.New(errcode.CodeMerchantStatus, "该商家尚未开通账号")
	}
	u, err := s.Repos.Users.FindByID(ctx, m.OwnerUserID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeUserNotFound, "商家账号不存在")
	}
	password, err := randomPassword(10)
	if err != nil {
		return nil, err
	}
	hash, err := s.Hasher.Hash(password)
	if err != nil {
		return nil, errcode.New(errcode.CodePasswordPolicy, "密码生成失败")
	}
	cred, err := s.Repos.Credentials.FindByUserID(ctx, u.ID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeUserNotFound, "账号凭证不存在")
	}
	if err := s.Repos.Credentials.Save(ctx, &identity.Credential{
		UserID: u.ID, PasswordHash: hash,
		SessionVersion: cred.SessionVersion + 1, UpdatedAt: time.Now().Unix(),
	}); err != nil {
		return nil, err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "merchant.account.reset", Object: merchantID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx), Operator: operator,
	})
	return &ProvisionResult{
		UserID: u.ID, Email: u.Email, Phone: u.Phone, Name: u.Name,
		Role: identity.RoleMerchant, MerchantID: merchantID, InitialPassword: password,
	}, nil
}

// ResetRiderPassword 为已开通骑手重置登录密码（旧会话立即失效）。
func (s *AccountsService) ResetRiderPassword(ctx context.Context, riderID, operator string, sc tenant.Scope) (*ProvisionResult, error) {
	r, err := s.Repos.Riders.FindByID(ctx, riderID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeRiderNotFound, "骑手不存在")
	}
	if !riderAccessOK(sc, r) {
		return nil, tenantDenied()
	}
	if r.UserID == "" {
		return nil, errcode.New(errcode.CodeRiderStatus, "该骑手尚未开通账号")
	}
	u, err := s.Repos.Users.FindByID(ctx, r.UserID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeUserNotFound, "骑手账号不存在")
	}
	password, err := randomPassword(10)
	if err != nil {
		return nil, err
	}
	hash, err := s.Hasher.Hash(password)
	if err != nil {
		return nil, errcode.New(errcode.CodePasswordPolicy, "密码生成失败")
	}
	cred, err := s.Repos.Credentials.FindByUserID(ctx, u.ID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeUserNotFound, "账号凭证不存在")
	}
	if err := s.Repos.Credentials.Save(ctx, &identity.Credential{
		UserID: u.ID, PasswordHash: hash,
		SessionVersion: cred.SessionVersion + 1, UpdatedAt: time.Now().Unix(),
	}); err != nil {
		return nil, err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "rider.account.reset", Object: riderID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx), Operator: operator,
	})
	return &ProvisionResult{
		UserID: u.ID, Email: u.Email, Phone: u.Phone, Name: u.Name,
		Role: identity.RoleDriver, RiderID: riderID, InitialPassword: password,
	}, nil
}

// ProvisionMerchant 为商家开通账号。
func (s *AccountsService) ProvisionMerchant(ctx context.Context, merchantID string, in ProvisionInput, operator string, sc tenant.Scope) (*ProvisionResult, error) {
	return s.provision(ctx, identity.RoleMerchant, merchantID, in, operator, sc)
}

// ProvisionRider 为骑手开通账号。
func (s *AccountsService) ProvisionRider(ctx context.Context, riderID string, in ProvisionInput, operator string, sc tenant.Scope) (*ProvisionResult, error) {
	return s.provision(ctx, identity.RoleDriver, riderID, in, operator, sc)
}

func (s *AccountsService) provision(ctx context.Context, role, subjectID string, in ProvisionInput, operator string, sc tenant.Scope) (*ProvisionResult, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Phone = strings.TrimSpace(in.Phone)
	in.Name = strings.TrimSpace(in.Name)
	if in.Email == "" && in.Phone == "" {
		return nil, errcode.New(errcode.CodeBadRequest, "邮箱与手机号至少填写一项")
	}
	if in.Email != "" && !accountEmailRe.MatchString(in.Email) {
		return nil, errcode.New(errcode.CodeBadRequest, "邮箱格式不正确")
	}
	if in.Phone != "" && !identity.ValidatePhone(in.Phone) {
		return nil, errcode.New(errcode.CodeBadRequest, "手机号格式不正确")
	}

	// 读取主体，确定区域、默认姓名、占位邮箱与绑定动作。
	regionID, defaultName, placeholderEmail := "", "", ""
	switch role {
	case identity.RoleMerchant:
		m, err := s.Repos.Merchants.FindByID(ctx, subjectID)
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				return nil, errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
			}
			return nil, err
		}
		if m.OwnerUserID != "" {
			return nil, errcode.NewStatus(409, errcode.CodeConflict, "该商家已开通账号")
		}
		if !merchantAccessOK(sc, m) {
			return nil, tenantDenied()
		}
		regionID = m.RegionID
		defaultName = firstNonEmpty(m.Contact, m.Name)
		placeholderEmail = "merchant." + m.ID + "@shuguang.local"
	case identity.RoleDriver:
		r, err := s.Repos.Riders.FindByID(ctx, subjectID)
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				return nil, errcode.NewStatus(404, errcode.CodeRiderNotFound, "骑手不存在")
			}
			return nil, err
		}
		if r.UserID != "" {
			return nil, errcode.NewStatus(409, errcode.CodeConflict, "该骑手已开通账号")
		}
		if !riderAccessOK(sc, r) {
			return nil, tenantDenied()
		}
		regionID = r.RegionID
		defaultName = r.Name
		placeholderEmail = "rider." + r.ID + "@shuguang.local"
	default:
		return nil, errcode.New(errcode.CodeBadRequest, "不支持的角色")
	}

	password := in.Password
	generated := false
	if password == "" {
		p, err := randomPassword(10)
		if err != nil {
			return nil, err
		}
		password = p
		generated = true
	} else if err := validateAccountPassword(password); err != nil {
		return nil, err
	}
	hash, err := s.Hasher.Hash(password)
	if err != nil {
		return nil, errcode.New(errcode.CodePasswordPolicy, "密码不符合策略")
	}

	name := in.Name
	if name == "" {
		name = defaultName
	}
	if name == "" {
		name = "账号" + subjectID
	}
	if len([]rune(name)) > 32 {
		name = string([]rune(name)[:32])
	}

	var result *ProvisionResult
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		now := time.Now().Unix()

		// 事务内二次校验，避免并发重复开通。
		switch role {
		case identity.RoleMerchant:
			m, err := tx.Merchants.FindByID(ctx, subjectID)
			if err != nil {
				return err
			}
			if m.OwnerUserID != "" {
				return errcode.NewStatus(409, errcode.CodeConflict, "该商家已开通账号")
			}
		case identity.RoleDriver:
			r, err := tx.Riders.FindByID(ctx, subjectID)
			if err != nil {
				return err
			}
			if r.UserID != "" {
				return errcode.NewStatus(409, errcode.CodeConflict, "该骑手已开通账号")
			}
		}

		var u *identity.User
		found := false
		if in.Email != "" {
			if v, err := tx.Users.FindByEmail(ctx, in.Email); err == nil {
				u, found = v, true
			} else if !errors.Is(err, application.ErrNotFound) {
				return err
			}
		} else if in.Phone != "" {
			if v, err := tx.Users.FindByPhone(ctx, in.Phone); err == nil {
				u, found = v, true
			} else if !errors.Is(err, application.ErrNotFound) {
				return err
			}
		}

		initialPassword := ""
		if !found {
			email := in.Email
			if email == "" {
				email = placeholderEmail
			}
			u = &identity.User{
				ID: ulidgen.New(), Email: email, Phone: in.Phone, Name: name,
				Status: identity.UserStatusActive, EmailVerified: true,
				Role: role, RegionID: regionID,
				RegisteredAt: now, UpdatedAt: now,
			}
			applySubjectBinding(u, role, subjectID)
			if err := tx.Users.Save(ctx, u); err != nil {
				return err
			}
			if err := tx.Credentials.Save(ctx, &identity.Credential{
				UserID: u.ID, PasswordHash: hash, SessionVersion: 0, UpdatedAt: now,
			}); err != nil {
				return err
			}
			if generated {
				initialPassword = password
			}
		} else {
			// 已存在的普通用户可升级为商家/骑手；已有的其他角色或已绑定其他主体则拒绝。
			if u.Role != identity.RoleUser && u.Role != role {
				return errcode.NewStatus(409, errcode.CodeConflict, "该账号已属于其他角色，请更换邮箱/手机号")
			}
			if role == identity.RoleMerchant && u.MerchantID != "" && u.MerchantID != subjectID {
				return errcode.NewStatus(409, errcode.CodeConflict, "该账号已绑定其他商家")
			}
			if role == identity.RoleDriver && u.RiderID != "" && u.RiderID != subjectID {
				return errcode.NewStatus(409, errcode.CodeConflict, "该账号已绑定其他骑手")
			}
			u.Role = role
			applySubjectBinding(u, role, subjectID)
			if u.RegionID == "" {
				u.RegionID = regionID
			}
			u.EmailVerified = true
			u.UpdatedAt = now
			if err := tx.Users.Save(ctx, u); err != nil {
				return err
			}
			cred, err := tx.Credentials.FindByUserID(ctx, u.ID)
			switch {
			case errors.Is(err, application.ErrNotFound):
				if err := tx.Credentials.Save(ctx, &identity.Credential{
					UserID: u.ID, PasswordHash: hash, SessionVersion: 0, UpdatedAt: now,
				}); err != nil {
					return err
				}
				if generated {
					initialPassword = password
				}
			case err != nil:
				return err
			case in.Password != "":
				// 管理员显式设置密码：更新哈希并递增会话版本，使旧会话失效。
				if err := tx.Credentials.Save(ctx, &identity.Credential{
					UserID: u.ID, PasswordHash: hash,
					SessionVersion: cred.SessionVersion + 1, UpdatedAt: now,
				}); err != nil {
					return err
				}
			}
		}

		if err := bindSubject(ctx, tx, role, subjectID, u.ID, now); err != nil {
			return err
		}
		result = &ProvisionResult{
			UserID: u.ID, Email: u.Email, Phone: u.Phone, Name: u.Name,
			Role: role, InitialPassword: initialPassword,
		}
		if role == identity.RoleMerchant {
			result.MerchantID = subjectID
		} else {
			result.RiderID = subjectID
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: role + ".account.provision", Object: subjectID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
		Operator: operator,
	})
	return result, nil
}

func applySubjectBinding(u *identity.User, role, subjectID string) {
	if role == identity.RoleMerchant {
		u.MerchantID = subjectID
		u.RiderID = ""
		return
	}
	u.RiderID = subjectID
	u.MerchantID = ""
}

func bindSubject(ctx context.Context, tx *application.RepoUnit, role, subjectID, userID string, now int64) error {
	if role == identity.RoleMerchant {
		m, err := tx.Merchants.FindByID(ctx, subjectID)
		if err != nil {
			return err
		}
		m.OwnerUserID = userID
		m.LastUpdatedAt = now
		return tx.Merchants.Save(ctx, m)
	}
	r, err := tx.Riders.FindByID(ctx, subjectID)
	if err != nil {
		return err
	}
	r.UserID = userID
	return tx.Riders.Save(ctx, r)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// validateAccountPassword 与 auth.ValidatePassword 保持同一策略（8-72 位，含字母与数字）。
func validateAccountPassword(pw string) error {
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

// randomPassword 生成满足策略的一次性密码（避免易混淆字符）。
func randomPassword(n int) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	for attempt := 0; attempt < 20; attempt++ {
		buf := make([]byte, n)
		for i := range buf {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			buf[i] = charset[idx.Int64()]
		}
		pw := string(buf)
		if validateAccountPassword(pw) == nil {
			return pw, nil
		}
	}
	return "", errors.New("生成初始密码失败")
}
