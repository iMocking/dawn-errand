package admin

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// RidersService 骑手管理。
type RidersService struct {
	Repos *application.RepoUnit
}

// List 骑手列表。
func (s *RidersService) List(ctx context.Context, q rider.RiderQuery, sc tenant.Scope) ([]*rider.Rider, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Riders.List(q, sc)
}

// Stats 骑手聚合。
func (s *RidersService) Stats(ctx context.Context, sc tenant.Scope) (*rider.Stats, error) {
	return s.Repos.Riders.Stats(sc)
}

// AuditList 入驻审核列表。
func (s *RidersService) AuditList(ctx context.Context, q rider.AuditQuery, sc tenant.Scope) ([]*rider.Rider, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Riders.AuditList(q, sc)
}

// Detail 骑手详情（对象级隔离）。
func (s *RidersService) Detail(ctx context.Context, id string, sc tenant.Scope) (*rider.Rider, error) {
	r, err := s.Repos.Riders.FindByID(id)
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

// AuditAction 审核入驻。
func (s *RidersService) AuditAction(ctx context.Context, id, action, reason, operator string, sc tenant.Scope) error {
	r, err := s.Repos.Riders.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeRiderNotFound, "申请不存在")
		}
		return err
	}
	if !riderAccessOK(sc, r) {
		return tenantDenied()
	}
	if r.AuditStatus != rider.AuditPending {
		return errcode.New(errcode.CodeRiderStatus, "该申请已处理")
	}
	if err := s.Repos.Riders.UpdateAudit(id, action == "approve", reason, operator, time.Now().Unix()); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "rider.audit", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// Freeze 冻结/解冻骑手。
func (s *RidersService) Freeze(ctx context.Context, id, action, operator string, sc tenant.Scope) error {
	r, err := s.Repos.Riders.FindByID(id)
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
	if err := s.Repos.Riders.UpdateStatus(id, status, time.Now().Unix()); err != nil {
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
		regions, err := s.Repos.Regions.List()
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
	return s.Repos.Riders.Save(r)
}

func normalizeQ(page, size *int) {
	if *page < 1 {
		*page = 1
	}
	if *size < 1 || *size > 200 {
		*size = 10
	}
}
