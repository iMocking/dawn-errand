package admin

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/complaint"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/realname"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/userprofile"
	"dawn-errand/internal/pkg/errcode"
)

// UsersService 用户管理。
type UsersService struct {
	Repos  *application.RepoUnit
	Tx     application.TxRunner
	Cipher application.ConfigCipher
}

// realNameView 解密身份证号并构造脱敏视图；解密失败返回空掩码（不阻断查询）。
func (s *UsersService) realNameView(v *realname.Verification) *realname.View {
	if v == nil {
		return nil
	}
	masked := "******"
	if plain, err := s.Cipher.Decrypt(v.IDCardCipher); err == nil {
		masked = maskIDCard(plain)
	}
	return realname.ToView(v, masked)
}

func maskIDCard(s string) string {
	runes := []rune(s)
	if len(runes) <= 8 {
		return "******"
	}
	return string(runes[:6]) + "********" + string(runes[len(runes)-4:])
}

// List 用户列表。
func (s *UsersService) List(ctx context.Context, q identity.UserQuery, sc tenant.Scope) ([]*identity.User, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Users.List(q, sc)
}

// UsersOverview 用户概览。
type UsersOverview struct {
	Total       int64
	TodayNew    int64
	ActiveMonth int64
	ActiveRate  float64
	AvgOrders   float64
	Banned      int64
}

// Overview 顶部概览（scope 隔离）。
func (s *UsersService) Overview(ctx context.Context, sc tenant.Scope) (*UsersOverview, error) {
	now := time.Now().Unix()
	_, total, err := s.Repos.Users.List(identity.UserQuery{Page: 1, PageSize: 1}, sc)
	if err != nil {
		return nil, err
	}
	activeMonth, err := s.Repos.Orders.CountActiveMonth(sc, now)
	if err != nil {
		return nil, err
	}
	todayNew, err := s.Repos.Users.CountRegisteredSince(sc, startOfToday(now))
	if err != nil {
		return nil, err
	}
	banned, err := s.Repos.Users.CountStatus(sc, identity.UserStatusBanned)
	if err != nil {
		return nil, err
	}
	avg := float64(0)
	if total > 0 {
		totalOrders, err := s.Repos.Orders.CountAll(sc)
		if err != nil {
			return nil, err
		}
		avg = float64(totalOrders) / float64(total)
	}
	return &UsersOverview{
		Total: total, TodayNew: todayNew, ActiveMonth: activeMonth, Banned: banned,
		AvgOrders: avg,
	}, nil
}

func startOfToday(now int64) int64 {
	t := time.Unix(now, 0).UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix()
}

// UserDetail 用户详情。
type UserDetail struct {
	User           *identity.User
	Profile        *userprofile.Profile
	RealName       *realname.View
	TotalOrders    int64
	TotalSpent     float64
	MonthOrders    int64
	RecentOrders   []*order.Order
	ComplaintCount int64
}

// Detail 用户详情（区域站长限本区域）。
func (s *UsersService) Detail(ctx context.Context, id string, sc tenant.Scope) (*UserDetail, error) {
	u, err := s.Repos.Users.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
		}
		return nil, err
	}
	if sc.RegionOnly() && u.RegionID != sc.RegionID {
		return nil, tenantDenied()
	}
	if sc.UserOnly() && u.ID != sc.UserID {
		return nil, tenantDenied()
	}
	now := time.Now().Unix()
	prof, _ := s.Repos.Profiles.FindByUserID(id)
	rn, _ := s.Repos.RealNames.FindByUserID(ctx, id)
	totalOrders, _ := s.Repos.Orders.CountUserOrders(id)
	totalSpent, _ := s.Repos.Orders.CountUserSpent(id)
	monthOrders, _ := s.Repos.Orders.CountUserMonth(id, now)
	recent, _ := s.Repos.Orders.ListByUser(id, 10)
	complaintCount, _ := s.Repos.Complaints.CountByUserID(id)
	return &UserDetail{
		User: u, Profile: prof, RealName: s.realNameView(rn),
		TotalOrders: totalOrders, TotalSpent: totalSpent, MonthOrders: monthOrders,
		RecentOrders: recent, ComplaintCount: complaintCount,
	}, nil
}

// userAccessOK 用户对象级作用域校验（区域站长限本区域、普通用户限本人）。
func (s *UsersService) userAccessOK(u *identity.User, sc tenant.Scope) bool {
	if sc.RegionOnly() && u.RegionID != sc.RegionID {
		return false
	}
	if sc.UserOnly() && u.ID != sc.UserID {
		return false
	}
	return true
}

// OrdersByUser 用户订单记录（分页，scope 隔离）。
func (s *UsersService) OrdersByUser(ctx context.Context, id string, page, size int, sc tenant.Scope) ([]*order.Order, int64, error) {
	u, err := s.Repos.Users.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, 0, errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
		}
		return nil, 0, err
	}
	if !s.userAccessOK(u, sc) {
		return nil, 0, tenantDenied()
	}
	return s.Repos.Orders.List(order.Query{Page: page, PageSize: size, UserID: id}, sc)
}

// ComplaintsByUser 用户投诉售后（分页，scope 隔离）。
func (s *UsersService) ComplaintsByUser(ctx context.Context, id string, page, size int, sc tenant.Scope) ([]*complaint.Complaint, int64, error) {
	u, err := s.Repos.Users.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, 0, errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
		}
		return nil, 0, err
	}
	if !s.userAccessOK(u, sc) {
		return nil, 0, tenantDenied()
	}
	return s.Repos.Complaints.List(complaint.Query{Page: page, PageSize: size, UserID: id}, sc)
}

// Ban 封禁/解封用户（撤销会话）。
func (s *UsersService) Ban(ctx context.Context, id string, ban bool, operator string, sc tenant.Scope) error {
	u, err := s.Repos.Users.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
		}
		return err
	}
	if sc.RegionOnly() && u.RegionID != sc.RegionID {
		return tenantDenied()
	}
	if sc.UserOnly() && u.ID != sc.UserID {
		return tenantDenied()
	}
	status := identity.UserStatusActive
	if ban {
		status = identity.UserStatusBanned
	}
	if err := s.Repos.Users.UpdateStatus(id, status); err != nil {
		return err
	}
	if _, err := s.Repos.Credentials.BumpSessionVersion(id); err != nil {
		return err
	}
	action := "user.unban"
	if ban {
		action = "user.ban"
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: action, Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// SetRole 修改用户后台角色。
// 约束：仅 super_admin 可操作（防 cs_manager 等持 user:role 权限者垂直提权）；
// 不能改自己；商家/司机绑定角色请走账号开通流程；系统级身份（role=system）禁止修改。
func (s *UsersService) SetRole(ctx context.Context, id, role, operator string, sc tenant.Scope) error {
	if sc.Role != identity.RoleSuperAdmin {
		return tenantDenied()
	}
	if id == operator {
		return errcode.New(errcode.CodeSettingInvalid, "不能修改自己的角色")
	}
	u, err := s.Repos.Users.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
		}
		return err
	}
	if u.Role == identity.RoleSystem {
		return errcode.New(errcode.CodeSettingInvalid, "系统服务身份不可修改角色")
	}
	if u.Role == identity.RoleMerchant || u.Role == identity.RoleDriver {
		return errcode.New(errcode.CodeSettingInvalid, "商家/司机角色由入驻与账号开通流程管理，不可在此修改")
	}
	valid := false
	for _, r := range identity.AllRoles {
		if r == role {
			valid = true
			break
		}
	}
	if !valid || role == identity.RoleSystem || role == identity.RoleMerchant || role == identity.RoleDriver {
		return errcode.New(errcode.CodeSettingInvalid, "不支持的岗位角色: "+role)
	}
	if err := s.Repos.Users.UpdateRole(id, role); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "user.role.update", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// RealNames 实名认证申请列表（scope 隔离；返回脱敏视图）。
func (s *UsersService) RealNames(ctx context.Context, q realname.Query, sc tenant.Scope) ([]*realname.View, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 20
	}
	if sc.RegionOnly() || sc.UserOnly() {
		return nil, 0, tenantDenied()
	}
	rows, total, err := s.Repos.RealNames.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	views := make([]*realname.View, 0, len(rows))
	for _, r := range rows {
		views = append(views, s.realNameView(r))
	}
	return views, total, nil
}

// AuditRealName 审核实名认证申请；通过时同步 UserProfile（real_name + verified=已实名）。
func (s *UsersService) AuditRealName(ctx context.Context, id, action, reason, operator string, sc tenant.Scope) error {
	if sc.RegionOnly() || sc.UserOnly() {
		return tenantDenied()
	}
	v, err := s.Repos.RealNames.FindByID(ctx, id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeUserNotFound, "认证申请不存在")
		}
		return err
	}
	if v.Status != realname.StatusPending {
		return errcode.NewStatus(409, errcode.CodeConflict, "该申请已处理")
	}
	now := time.Now().Unix()
	if action == "approve" {
		// 事务：更新申请状态 + 同步用户资料。
		err := s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
			if err := tx.RealNames.UpdateReview(ctx, id, realname.StatusApproved, "", operator, now); err != nil {
				return err
			}
			prof, err := tx.Profiles.FindByUserID(v.UserID)
			if err != nil && err != application.ErrNotFound {
				return err
			}
			p := &userprofile.Profile{
				UserID: v.UserID, RealName: v.RealName, Verified: realname.VerifiedDone,
				UpdatedAt: now,
			}
			if prof != nil {
				p.Level = prof.Level
				p.Nickname = prof.Nickname
				p.Region = prof.Region
				p.Address = prof.Address
				p.RegisterChannel = prof.RegisterChannel
				p.BizTypes = prof.BizTypes
				p.Balance = prof.Balance
				p.RedPacketBalance = prof.RedPacketBalance
				p.RecentActiveAt = prof.RecentActiveAt
			}
			return tx.Profiles.Save(p)
		})
		if err != nil {
			return err
		}
	} else if action == "reject" {
		if err := s.Repos.RealNames.UpdateReview(ctx, id, realname.StatusRejected, reason, operator, now); err != nil {
			return err
		}
	} else {
		return errcode.New(errcode.CodeBadRequest, "action 仅支持 approve/reject")
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "user.realname." + action, Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}
