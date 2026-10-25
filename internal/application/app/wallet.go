package app

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/wallet"
)

// WalletService 钱包查询。
type WalletService struct {
	Repos *application.RepoUnit
}

// WalletSummary 钱包汇总。
type WalletSummary struct {
	Balance         float64
	MonthIncome     float64
	TodayIncome     float64
	PendingWithdraw float64
}

// Summary 读取钱包汇总（自动创建账户）。
func (s *WalletService) Summary(ctx context.Context, partyType, partyID string) (*WalletSummary, error) {
	acc, err := s.Repos.Wallets.Ensure(partyType, partyID, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Unix()
	incomeTypes := []string{"delivery_income", "order_income"}
	monthIncome, err := s.Repos.Wallets.SumByTypeSince(partyType, partyID, incomeTypes, monthStart)
	if err != nil {
		return nil, err
	}
	todayIncome, err := s.Repos.Wallets.SumByTypeSince(partyType, partyID, incomeTypes, dayStart)
	if err != nil {
		return nil, err
	}
	pending, err := s.Repos.Withdrawals.SumPendingByParty(partyType, partyID)
	if err != nil {
		return nil, err
	}
	return &WalletSummary{
		Balance: acc.Balance, MonthIncome: monthIncome,
		TodayIncome: todayIncome, PendingWithdraw: pending,
	}, nil
}

// Records 钱包流水。
func (s *WalletService) Records(ctx context.Context, partyType, partyID string, page, size int) ([]*wallet.Transaction, int64, error) {
	page, size = normalizePageParams(page, size)
	return s.Repos.Wallets.ListTransactions(wallet.Query{
		PartyType: partyType, PartyID: partyID, Page: page, PageSize: size,
	})
}

// creditWallet 入账并写流水（供订单/评价/退款在事务内调用）。
func creditWallet(tx *application.RepoUnit, partyType, partyID string, amount float64, txnType, orderID, note string, now int64) error {
	if amount <= 0 || partyID == "" {
		return nil
	}
	acc, err := tx.Wallets.Ensure(partyType, partyID, now)
	if err != nil {
		return err
	}
	ok, updated, err := tx.Wallets.Adjust(partyType, partyID, amount, true, now)
	if err != nil {
		return err
	}
	if !ok || updated == nil {
		return nil
	}
	return tx.Wallets.SaveTransaction(&wallet.Transaction{
		AccountID: acc.ID, PartyType: partyType, PartyID: partyID,
		OrderID: orderID, Type: txnType, Amount: amount,
		BalanceAfter: updated.Balance, Note: note, CreatedAt: now,
	})
}

// debitWallet 出账并写流水；allowNegative=false 时余额不足返回 false。
func debitWallet(tx *application.RepoUnit, partyType, partyID string, amount float64, allowNegative bool, txnType, orderID, note string, now int64) (bool, error) {
	if amount <= 0 || partyID == "" {
		return true, nil
	}
	acc, err := tx.Wallets.Ensure(partyType, partyID, now)
	if err != nil {
		return false, err
	}
	ok, updated, err := tx.Wallets.Adjust(partyType, partyID, -amount, allowNegative, now)
	if err != nil || !ok || updated == nil {
		return false, err
	}
	if err := tx.Wallets.SaveTransaction(&wallet.Transaction{
		AccountID: acc.ID, PartyType: partyType, PartyID: partyID,
		OrderID: orderID, Type: txnType, Amount: -amount,
		BalanceAfter: updated.Balance, Note: note, CreatedAt: now,
	}); err != nil {
		return false, err
	}
	return true, nil
}
