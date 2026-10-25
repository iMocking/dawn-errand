package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/system"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/metrics"
	"dawn-errand/internal/pkg/ulidgen"
)

// FinanceAppService 商家/骑手资金用例。
type FinanceAppService struct {
	Repos  *application.RepoUnit
	Tx     application.TxRunner
	Wallet *WalletService
	Payout application.PayoutGateway
}

// WithdrawRequest 提现申请。
type WithdrawRequest struct {
	Amount      float64
	Channel     string
	AccountName string
	BankCard    string
}

// FinanceSummary 端内资金汇总。
type FinanceSummary struct {
	Balance         float64
	MonthIncome     float64
	TodayIncome     float64
	PendingWithdraw float64
	PendingSettle   float64
}

// Summary 资金汇总。
func (s *FinanceAppService) Summary(ctx context.Context, partyType, partyID string) (*FinanceSummary, error) {
	ws, err := s.Wallet.Summary(ctx, partyType, partyID)
	if err != nil {
		return nil, err
	}
	pendingSettle, err := s.Repos.Settlements.SumPendingByParty(partyType, partyID)
	if err != nil {
		return nil, err
	}
	return &FinanceSummary{
		Balance: ws.Balance, MonthIncome: ws.MonthIncome, TodayIncome: ws.TodayIncome,
		PendingWithdraw: ws.PendingWithdraw, PendingSettle: pendingSettle,
	}, nil
}

// Records 收入流水。
func (s *FinanceAppService) Records(ctx context.Context, partyType, partyID string, page, size int) ([]*FinanceRecord, int64, error) {
	txns, total, err := s.Wallet.Records(ctx, partyType, partyID, page, size)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*FinanceRecord, 0, len(txns))
	for _, t := range txns {
		out = append(out, &FinanceRecord{
			ID: t.ID, Type: t.Type, Amount: t.Amount, BalanceAfter: t.BalanceAfter,
			OrderID: t.OrderID, Note: t.Note, CreatedAt: t.CreatedAt,
		})
	}
	return out, total, nil
}

// FinanceRecord 资金流水行。
type FinanceRecord struct {
	ID           string
	Type         string
	Amount       float64
	BalanceAfter float64
	OrderID      string
	Note         string
	CreatedAt    int64
}

// Settlements 结算单。
func (s *FinanceAppService) Settlements(ctx context.Context, partyType, partyID string, page, size int) ([]*finance.Settlement, int64, error) {
	return s.Repos.Settlements.List(finance.SettlementQuery{
		Page: page, PageSize: size, PartyType: partyType, PartyID: partyID,
	}, tenant.Scope{})
}

// Withdrawals 提现记录。
func (s *FinanceAppService) Withdrawals(ctx context.Context, partyType, partyID string, page, size int) ([]*finance.Withdrawal, int64, error) {
	return s.Repos.Withdrawals.List(finance.WithdrawQuery{
		Page: page, PageSize: size, PartyType: partyType, PartyID: partyID,
	}, tenant.Scope{})
}

// CreateWithdrawal 申请提现：校验规则、扣减余额、生成提现单。
func (s *FinanceAppService) CreateWithdrawal(ctx context.Context, partyType, partyID string, in WithdrawRequest) (*finance.Withdrawal, error) {
	if in.Amount <= 0 {
		return nil, errcode.New(errcode.CodeBadRequest, "提现金额必须大于 0")
	}
	rule, err := s.withdrawRule()
	if err != nil {
		return nil, err
	}
	if !rule.Enabled {
		return nil, errcode.New(errcode.CodeWithdrawStatus, "提现暂未开放")
	}
	if in.Amount < rule.MinAmount || (rule.MaxAmount > 0 && in.Amount > rule.MaxAmount) {
		return nil, errcode.New(errcode.CodeWithdrawStatus,
			fmt.Sprintf("单笔提现金额需在 ¥%.0f-¥%.0f 之间", rule.MinAmount, rule.MaxAmount))
	}
	now := time.Now().Unix()
	count, err := s.Repos.Withdrawals.CountByPartySince(partyType, partyID, now-86400)
	if err != nil {
		return nil, err
	}
	if rule.DailyLimit > 0 && count >= int64(rule.DailyLimit) {
		return nil, errcode.New(errcode.CodeWithdrawStatus, "今日提现次数已达上限")
	}

	name, regionID, defaultAccount := s.partyInfo(partyType, partyID)
	if name == "" {
		return nil, errcode.NewStatus(404, errcode.CodeMerchantNotFound, "主体不存在")
	}
	fee := round2(in.Amount * rule.FeeRate)
	if fee >= in.Amount {
		return nil, errcode.New(errcode.CodeWithdrawStatus, "提现金额不足以覆盖手续费")
	}
	channel := strings.TrimSpace(in.Channel)
	if channel == "" {
		channel = "bank"
	}
	accountName := strings.TrimSpace(in.AccountName)
	if accountName == "" {
		accountName = name
	}
	bankInfo := strings.TrimSpace(in.BankCard)
	if bankInfo == "" {
		bankInfo = defaultAccount
	}
	if len(bankInfo) > 128 {
		return nil, errcode.New(errcode.CodeBadRequest, "收款账户信息过长")
	}

	status := finance.WithdrawPending
	expect := fmt.Sprintf("预计 T+%d 个工作日到账", rule.SettleDays)
	if rule.Mode == "auto" {
		status = finance.WithdrawApproved
		expect = "自动审核通过，预计 T+" + fmt.Sprint(rule.SettleDays) + " 到账"
	}
	kind := "商家结算"
	if partyType == "rider" {
		kind = "骑手提现"
	}
	if status == finance.WithdrawApproved && s.Payout == nil {
		return nil, errcode.New(errcode.CodePaymentFailed, "打款通道未配置，暂不支持自动提现")
	}

	var w *finance.Withdrawal
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		ok, err := debitWallet(tx, partyType, partyID, in.Amount, false, "withdraw", "", "提现申请", now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeInsufficientFund, "可提现余额不足")
		}
		acc, err := tx.Wallets.Find(partyType, partyID)
		if err != nil {
			return err
		}
		balanceAfter := acc.Balance
		w = &finance.Withdrawal{
			ID: ulidgen.New(), RegionID: regionID,
			PartyType: partyType, PartyID: partyID, PartyName: name, PartyNo: partyID,
			BankInfo: maskAccount(bankInfo), Amount: in.Amount, Fee: fee,
			FeeNote: fmt.Sprintf("手续费 %.2f%%", rule.FeeRate*100),
			Status:  status, Kind: kind, Applicant: name, AccountHolder: accountName,
			WithdrawType: channel, ArriveExpect: expect, Purpose: kind,
			Balance: round2(balanceAfter + in.Amount), RemainBalance: balanceAfter,
			CreatedAt: now,
			Timeline: []finance.Event{
				{Time: now, Text: fmt.Sprintf("%s 申请提现 ¥%.2f", name, in.Amount), Operator: name},
				{Time: now, Text: "系统校验提现规则通过", Operator: "system"},
			},
		}
		if err := tx.Withdrawals.Save(w); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if status == finance.WithdrawApproved && s.Payout != nil {
		ref, perr := s.Payout.Payout(ctx, w.ID, partyType, partyID, in.Amount)
		if perr != nil {
			return nil, errcode.Wrap(perr, errcode.CodePaymentFailed, "自动打款失败，请联系平台")
		}
		if ref != "" {
			_ = s.Repos.Withdrawals.MarkPayout(w.ID, ref, now)
		}
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: partyID, Action: "withdraw.apply", Object: w.ID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx), Operator: name,
	})
	metrics.WithdrawalsTotal.Inc()
	return w, nil
}

func (s *FinanceAppService) withdrawRule() (*system.WithdrawRuleConfig, error) {
	rec, err := s.Repos.Settings.Get(system.KeyWithdrawRule)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
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

func (s *FinanceAppService) partyInfo(partyType, partyID string) (name, regionID, account string) {
	switch partyType {
	case "merchant":
		if m, err := s.Repos.Merchants.FindByID(partyID); err == nil {
			return m.Name, m.RegionID, m.BankAccount
		}
	case "rider":
		if r, err := s.Repos.Riders.FindByID(partyID); err == nil {
			return r.Name, r.RegionID, r.Account
		}
	}
	return "", "", ""
}

func maskAccount(s string) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= 4 {
		return string(r)
	}
	return string(r[:2]) + "****" + string(r[len(r)-2:])
}
