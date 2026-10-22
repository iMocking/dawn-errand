package admin

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// MerchantsService 商家管理。
type MerchantsService struct {
	Repos *application.RepoUnit
}

// List 商家列表。
func (s *MerchantsService) List(ctx context.Context, q merchant.MerchantQuery, sc tenant.Scope) ([]*merchant.Merchant, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Merchants.List(q, sc)
}

// Stats 商家聚合。
func (s *MerchantsService) Stats(ctx context.Context, sc tenant.Scope) (*merchant.Stats, error) {
	return s.Repos.Merchants.Stats(sc)
}

// AuditList 入驻审核列表。
func (s *MerchantsService) AuditList(ctx context.Context, q merchant.AuditQuery, sc tenant.Scope) ([]*merchant.Merchant, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Merchants.AuditList(q, sc)
}

// Detail 商家详情（对象级隔离）。
func (s *MerchantsService) Detail(ctx context.Context, id string, sc tenant.Scope) (*merchant.Merchant, error) {
	m, err := s.Repos.Merchants.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
		}
		return nil, err
	}
	if !merchantAccessOK(sc, m) {
		return nil, tenantDenied()
	}
	return m, nil
}

// AuditAction 审核入驻（approve/reject）。
func (s *MerchantsService) AuditAction(ctx context.Context, id, action, reason, operator string, sc tenant.Scope) error {
	m, err := s.Repos.Merchants.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeMerchantNotFound, "申请不存在")
		}
		return err
	}
	if !merchantAccessOK(sc, m) {
		return tenantDenied()
	}
	if m.AuditStatus != merchant.AuditPending {
		return errcode.New(errcode.CodeMerchantStatus, "该申请已处理")
	}
	approved := action == "approve"
	if err := s.Repos.Merchants.UpdateAudit(id, approved, reason, operator, time.Now().Unix()); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "merchant.audit", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// UpdateStatus 变更营业状态。
func (s *MerchantsService) UpdateStatus(ctx context.Context, id string, status merchant.MerchantStatus, operator string, sc tenant.Scope) error {
	m, err := s.Repos.Merchants.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
		}
		return err
	}
	if !merchantAccessOK(sc, m) {
		return tenantDenied()
	}
	if err := s.Repos.Merchants.UpdateStatus(id, status, time.Now().Unix()); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "merchant.status", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// Create 创建商家（含入驻申请）。
func (s *MerchantsService) Create(ctx context.Context, m *merchant.Merchant, sc tenant.Scope) error {
	if sc.RegionOnly() && m.RegionID != sc.RegionID {
		return tenantDenied()
	}
	now := time.Now().Unix()
	if m.ID == "" {
		m.ID = ulidgen.New()
	}
	if m.ApplyNo == "" {
		m.ApplyNo = "SQ" + time.Now().Format("20060102") + m.ID[:4]
	}
	m.AppliedAt = now
	if m.AuditStatus == "" {
		m.AuditStatus = merchant.AuditPending
	}
	m.LastUpdatedAt = now
	if m.JoinedAt == 0 {
		m.JoinedAt = now
	}
	return s.Repos.Merchants.Save(m)
}
