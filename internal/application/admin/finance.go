package admin

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
)

// FinanceService 财务/资金。
type FinanceService struct {
	Repos *application.RepoUnit
}

// Overview 财务页汇总 + 待审核提现列表。
type Overview struct {
	Stats           *finance.Stats
	SettlementStats *finance.SettlementStats
	Withdraws       []*finance.Withdrawal
	WithdrawTotal   int64
}

// Get 财务页数据（scope 隔离）。
func (s *FinanceService) Get(ctx context.Context, sc tenant.Scope) (*Overview, error) {
	now := time.Now().Unix()
	st, err := s.Repos.Withdrawals.Stats(sc, now)
	if err != nil {
		return nil, err
	}
	ss, err := s.Repos.Settlements.Stats(sc, now)
	if err != nil {
		return nil, err
	}
	list, total, err := s.Repos.Withdrawals.List(finance.WithdrawQuery{
		Page: 1, PageSize: 10, Status: "待审核",
	}, sc)
	if err != nil {
		return nil, err
	}
	return &Overview{Stats: st, SettlementStats: ss, Withdraws: list, WithdrawTotal: total}, nil
}

// SettlementList 结算流水。
func (s *FinanceService) SettlementList(ctx context.Context, q finance.SettlementQuery, sc tenant.Scope) ([]*finance.Settlement, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Settlements.List(q, sc)
}

// WithdrawList 提现列表。
func (s *FinanceService) WithdrawList(ctx context.Context, q finance.WithdrawQuery, sc tenant.Scope) ([]*finance.Withdrawal, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Withdrawals.List(q, sc)
}

// WithdrawDetail 提现单详情（对象级隔离）。
func (s *FinanceService) WithdrawDetail(ctx context.Context, id string, sc tenant.Scope) (*finance.Withdrawal, error) {
	w, err := s.Repos.Withdrawals.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, errcode.NewStatus(404, errcode.CodeWithdrawNotFound, "提现单不存在")
		}
		return nil, err
	}
	if sc.RegionOnly() && w.RegionID != sc.RegionID {
		return nil, tenantDenied()
	}
	return w, nil
}

// AuditWithdraw 审核提现（pass/reject）。
func (s *FinanceService) AuditWithdraw(ctx context.Context, id, action, remark, operator string, sc tenant.Scope) error {
	w, err := s.Repos.Withdrawals.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeWithdrawNotFound, "提现单不存在")
		}
		return err
	}
	if sc.RegionOnly() && w.RegionID != sc.RegionID {
		return tenantDenied()
	}
	if w.Status != finance.WithdrawPending {
		return errcode.New(errcode.CodeWithdrawStatus, "该提现单已审核")
	}
	if err := s.Repos.Withdrawals.Audit(id, action == "pass", remark, operator, time.Now().Unix()); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "withdraw.audit", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}
