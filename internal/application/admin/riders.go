package admin

import (
	"context"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// RidersService 骑手管理。
type RidersService struct {
	Repos    *application.RepoUnit
	Accounts *AccountsService
}

// List 骑手列表。
func (s *RidersService) List(ctx context.Context, q rider.RiderQuery, sc tenant.Scope) ([]*rider.Rider, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Riders.List(ctx, q, sc)
}

// Stats 骑手聚合。
func (s *RidersService) Stats(ctx context.Context, sc tenant.Scope) (*rider.Stats, error) {
	return s.Repos.Riders.Stats(ctx, sc)
}

// AuditList 入驻审核列表。
func (s *RidersService) AuditList(ctx context.Context, q rider.AuditQuery, sc tenant.Scope) ([]*rider.Rider, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Riders.AuditList(ctx, q, sc)
}

// Detail 骑手详情（对象级隔离）。
func (s *RidersService) Detail(ctx context.Context, id string, sc tenant.Scope) (*rider.Rider, error) {
	r, err := s.Repos.Riders.FindByID(ctx, id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, errcode.NewStatus(404, errcode.CodeRiderNotFound, "骑手不存在")
		}
		return nil, err
	}
	if !riderAccessOK(sc, r) {
		return nil, tenantDenied()
	}
	return r, nil
}

// AuditAction 审核入驻：通过时自动开通骑手端账号并返回一次性初始密码。
func (s *RidersService) AuditAction(ctx context.Context, id, action, reason, operator string, sc tenant.Scope) (*ProvisionResult, error) {
	r, err := s.Repos.Riders.FindByID(ctx, id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, errcode.NewStatus(404, errcode.CodeRiderNotFound, "申请不存在")
		}
		return nil, err
	}
	if !riderAccessOK(sc, r) {
		return nil, tenantDenied()
	}
	if r.AuditStatus != rider.AuditPending {
		return nil, errcode.New(errcode.CodeRiderStatus, "该申请已处理")
	}
	approved := action == "approve"
	if !approved {
		if strings.TrimSpace(reason) == "" {
			return nil, errcode.New(errcode.CodeBadRequest, "驳回必须填写原因，便于补件")
		}
		if err := s.Repos.Riders.UpdateAudit(ctx, id, false, strings.TrimSpace(reason), operator, time.Now().Unix()); err != nil {
			return nil, err
		}
		s.audit(ctx, id, operator)
		return nil, nil
	}
	if strings.TrimSpace(r.Phone) == "" {
		return nil, errcode.New(errcode.CodeBadRequest, "骑手手机号必填，用于登录骑手端")
	}
	if err := s.Repos.Riders.UpdateAudit(ctx, id, true, "", operator, time.Now().Unix()); err != nil {
		return nil, err
	}
	var account *ProvisionResult
	if r.UserID == "" {
		if s.Accounts == nil {
			return nil, errcode.New(errcode.CodeInternal, "账号开通服务未配置")
		}
		account, err = s.Accounts.ProvisionRider(ctx, id, ProvisionInput{
			Phone: strings.TrimSpace(r.Phone), Name: r.Name,
		}, operator, sc)
		if err != nil {
			return nil, err
		}
	}
	s.audit(ctx, id, operator)
	return account, nil
}

func (s *RidersService) audit(ctx context.Context, id, operator string) {
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "rider.audit", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
}

// Freeze 冻结/解冻骑手。
func (s *RidersService) Freeze(ctx context.Context, id, action, operator string, sc tenant.Scope) error {
	r, err := s.Repos.Riders.FindByID(ctx, id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeRiderNotFound, "骑手不存在")
		}
		return err
	}
	if !riderAccessOK(sc, r) {
		return tenantDenied()
	}
	status := rider.RiderFrozen
	if action == "unfreeze" {
		status = rider.RiderOffline
	}
	if err := s.Repos.Riders.UpdateStatus(ctx, id, status, time.Now().Unix()); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "rider.freeze", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// Create 创建骑手（含入驻申请）。
func (s *RidersService) Create(ctx context.Context, r *rider.Rider, sc tenant.Scope) error {
	if sc.RegionOnly() {
		if r.RegionID == "" {
			r.RegionID = sc.RegionID
		}
		if r.RegionID != sc.RegionID {
			return tenantDenied()
		}
	} else if r.RegionID == "" {
		// 超管未指定区域：使用第一个服务区域（ent schema region_id NotEmpty）。
		regions, err := s.Repos.Regions.List(ctx)
		if err != nil || len(regions) == 0 {
			return errcode.New(errcode.CodeBadRequest, "未配置服务区域，无法创建骑手")
		}
		r.RegionID = regions[0].ID
	}
	now := time.Now().Unix()
	if r.ID == "" {
		r.ID = ulidgen.New()
	}
	if r.ApplyNo == "" {
		r.ApplyNo = "SQ" + time.Now().Format("20060102") + r.ID[:4]
	}
	r.SubmittedAt = now
	if r.AuditStatus == "" {
		r.AuditStatus = rider.AuditPending
	}
	return s.Repos.Riders.Save(ctx, r)
}

func normalizeQ(page, size *int) {
	if *page < 1 {
		*page = 1
	}
	if *size < 1 || *size > 200 {
		*size = 10
	}
}
