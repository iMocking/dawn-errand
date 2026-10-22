package admin

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/userprofile"
	"dawn-errand/internal/pkg/errcode"
)

// UsersService 用户管理。
type UsersService struct {
	Repos *application.RepoUnit
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
	totalOrders, _ := s.Repos.Orders.CountUserOrders(id)
	totalSpent, _ := s.Repos.Orders.CountUserSpent(id)
	monthOrders, _ := s.Repos.Orders.CountUserMonth(id, now)
	recent, _ := s.Repos.Orders.ListByUser(id, 10)
	complaintCount, _ := s.Repos.Complaints.CountByUserID(id)
	return &UserDetail{
		User: u, Profile: prof,
		TotalOrders: totalOrders, TotalSpent: totalSpent, MonthOrders: monthOrders,
		RecentOrders: recent, ComplaintCount: complaintCount,
	}, nil
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
