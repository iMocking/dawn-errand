package admin

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/wallet"
	"dawn-errand/internal/pkg/errcode"
)

// FinanceService 财务/资金。
type FinanceService struct {
	Repos  *application.RepoUnit
	Tx     application.TxRunner
	Payout application.PayoutGateway
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
	st, err := s.Repos.Withdrawals.Stats(ctx, sc, now)
	if err != nil {
		return nil, err
	}
	ss, err := s.Repos.Settlements.Stats(ctx, sc, now)
	if err != nil {
		return nil, err
	}
	list, total, err := s.Repos.Withdrawals.List(ctx, finance.WithdrawQuery{
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
	return s.Repos.Settlements.List(ctx, q, sc)
}

// WithdrawList 提现列表。
func (s *FinanceService) WithdrawList(ctx context.Context, q finance.WithdrawQuery, sc tenant.Scope) ([]*finance.Withdrawal, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Withdrawals.List(ctx, q, sc)
}

// WithdrawDetail 提现单详情（对象级隔离）。
func (s *FinanceService) WithdrawDetail(ctx context.Context, id string, sc tenant.Scope) (*finance.Withdrawal, error) {
	w, err := s.Repos.Withdrawals.FindByID(ctx, id)
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

// Reconciliation 单日支付对账结果。
type Reconciliation struct {
	Date          string
	PaidOrders    int64
	OrderAmount   float64
	PaymentCount  int64
	PaymentAmount float64
	RefundCount   int64
	RefundAmount  float64
	Diff          float64
	MissingOrders int64
}

// Reconciliation 对账：支付单金额 vs 订单应付金额（按支付时间归属日期）。
// 区域站长只看本区域；商家/司机/会员等角色无权查看平台对账。
func (s *FinanceService) Reconciliation(ctx context.Context, date string, sc tenant.Scope) (*Reconciliation, error) {
	regionID := ""
	switch {
	case sc.IsGlobal():
	case sc.RegionOnly():
		regionID = sc.RegionID
	default:
		return nil, tenantDenied()
	}
	day, err := time.ParseInLocation("2006-01-02", date, time.UTC)
	if err != nil {
		return nil, errcode.New(errcode.CodeBadRequest, "日期格式应为 YYYY-MM-DD")
	}
	sum, err := s.Repos.Payments.SummarizeRecon(ctx, day.Unix(), day.AddDate(0, 0, 1).Unix(), regionID)
	if err != nil {
		return nil, err
	}
	res := &Reconciliation{
		Date:       day.Format("2006-01-02"),
		PaidOrders: sum.PaidOrderCount, OrderAmount: sum.OrderAmount,
		PaymentCount: sum.PaidCount, PaymentAmount: sum.PaidAmount,
		RefundCount: sum.RefundCount, RefundAmount: sum.RefundAmount,
		MissingOrders: sum.MissingOrders,
	}
	res.OrderAmount = round2Money(res.OrderAmount)
	res.PaymentAmount = round2Money(res.PaymentAmount)
	res.RefundAmount = round2Money(res.RefundAmount)
	res.Diff = round2Money(res.OrderAmount - res.PaymentAmount)
	return res, nil
}

func round2Money(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
func (s *FinanceService) AuditWithdraw(ctx context.Context, id, action, remark, operator string, sc tenant.Scope) error {
	w, err := s.Repos.Withdrawals.FindByID(ctx, id)
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
	if action == "pass" && s.Payout == nil {
		return errcode.New(errcode.CodePaymentFailed, "打款通道未配置")
	}
	now := time.Now().Unix()
	// 条件认领：并发双审核仅一人生效，杜绝重复打款/重复退回。
	claimed, err := s.Repos.Withdrawals.ClaimAudit(ctx, id, action == "pass", remark, operator, now)
	if err != nil {
		return err
	}
	if !claimed {
		return errcode.New(errcode.CodeWithdrawStatus, "该提现单已审核")
	}
	// 通过：发起打款并记录回执；失败则回退待审核供重试（资金仍处扣划冻结态）。
	if action == "pass" {
		ref, err := s.Payout.Payout(ctx, id, w.PartyType, w.PartyID, w.Amount)
		if err != nil {
			if verr := s.Repos.Withdrawals.RevertAuditToPending(ctx, id, string(finance.WithdrawApproved), operator, time.Now().Unix()); verr != nil {
				return errcode.Wrap(err, errcode.CodePaymentFailed, "打款失败且状态回退异常，请人工对账")
			}
			return errcode.Wrap(err, errcode.CodePaymentFailed, "打款失败，已回退待审核，请重试")
		}
		if ref != "" {
			if err := s.Repos.Withdrawals.MarkPayout(ctx, id, ref, time.Now().Unix()); err != nil {
				return err
			}
		}
		s.Repos.Audit.Record(ctx, application.AuditEntry{
			UserID: operator, Action: "withdraw.audit", Object: id, Result: "success",
			RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
		})
		return nil
	}
	// 驳回：退回已扣减的钱包余额；失败则回退认领，保证状态与资金可重试对齐。
	if w.PartyID != "" && s.Tx != nil {
		rnow := time.Now().Unix()
		if err := s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
			acc, err := tx.Wallets.Ensure(ctx, w.PartyType, w.PartyID, rnow)
			if err != nil {
				return err
			}
			ok, updated, err := tx.Wallets.Adjust(ctx, w.PartyType, w.PartyID, w.Amount, true, rnow)
			if err != nil {
				return err
			}
			if !ok || updated == nil {
				return errcode.New(errcode.CodeWithdrawStatus, "退款入账失败")
			}
			return tx.Wallets.SaveTransaction(ctx, &wallet.Transaction{
				AccountID: acc.ID, PartyType: w.PartyType, PartyID: w.PartyID,
				Type: "withdraw_refund", Amount: w.Amount,
				BalanceAfter: updated.Balance, Note: "提现驳回退回", CreatedAt: rnow,
			})
		}); err != nil {
			if verr := s.Repos.Withdrawals.RevertAuditToPending(ctx, id, string(finance.WithdrawRejected), operator, rnow); verr == nil {
				return errcode.Wrap(err, errcode.CodeWithdrawStatus, "退款失败，已回退待审核，请重试驳回")
			}
			return err
		}
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "withdraw.audit", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}
