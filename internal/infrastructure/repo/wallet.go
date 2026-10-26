package repo

import (
	"context"

	"dawn-errand/internal/domain/wallet"
	"dawn-errand/internal/infrastructure/ent"
	entwallet "dawn-errand/internal/infrastructure/ent/walletaccount"
	"dawn-errand/internal/infrastructure/ent/wallettransaction"
)

// WalletRepo 实现 wallet.Repository。
type WalletRepo struct{ client *ent.Client }

func (r WalletRepo) Ensure(ctx context.Context, partyType, partyID string, now int64) (*wallet.Account, error) {
	e, err := r.client.WalletAccount.Query().
		Where(entwallet.PartyType(partyType), entwallet.PartyID(partyID)).Only(ctx)
	if err == nil {
		return toWalletAccount(e), nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	e, err = r.client.WalletAccount.Create().
		SetPartyType(partyType).SetPartyID(partyID).SetUpdatedAt(now).Save(ctx)
	if err != nil {
		// 并发创建：回读。
		if again, e2 := r.client.WalletAccount.Query().
			Where(entwallet.PartyType(partyType), entwallet.PartyID(partyID)).Only(ctx); e2 == nil {
			return toWalletAccount(again), nil
		}
		return nil, err
	}
	return toWalletAccount(e), nil
}

func (r WalletRepo) Find(ctx context.Context, partyType, partyID string) (*wallet.Account, error) {
	e, err := r.client.WalletAccount.Query().
		Where(entwallet.PartyType(partyType), entwallet.PartyID(partyID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toWalletAccount(e), nil
}

func (r WalletRepo) Adjust(ctx context.Context, partyType, partyID string, delta float64, allowNegative bool, now int64) (bool, *wallet.Account, error) {
	if _, err := r.Ensure(ctx, partyType, partyID, now); err != nil {
		return false, nil, err
	}
	q := r.client.WalletAccount.Update().
		Where(entwallet.PartyType(partyType), entwallet.PartyID(partyID))
	if delta < 0 && !allowNegative {
		q = q.Where(entwallet.BalanceGTE(-delta))
	}
	affected, err := q.AddBalance(delta).SetUpdatedAt(now).Save(ctx)
	if err != nil {
		return false, nil, err
	}
	if affected == 0 {
		return false, nil, nil
	}
	acc, err := r.Find(ctx, partyType, partyID)
	return err == nil, acc, err
}

func (r WalletRepo) SaveTransaction(ctx context.Context, t *wallet.Transaction) error {
	if t.ID == "" {
		t.ID = newULID()
	}
	return r.client.WalletTransaction.Create().
		SetID(t.ID).SetAccountID(t.AccountID).SetPartyType(t.PartyType).SetPartyID(t.PartyID).
		SetOrderID(t.OrderID).SetType(t.Type).SetAmount(t.Amount).
		SetBalanceAfter(t.BalanceAfter).SetNote(t.Note).SetCreatedAt(t.CreatedAt).
		Exec(ctx)
}

func (r WalletRepo) SumByTypeSince(ctx context.Context, partyType, partyID string, types []string, since int64) (float64, error) {
	q := r.client.WalletTransaction.Query().
		Where(wallettransaction.PartyType(partyType), wallettransaction.PartyID(partyID),
			wallettransaction.CreatedAtGTE(since))
	if len(types) > 0 {
		q.Where(wallettransaction.TypeIn(types...))
	}
	return sumFloat64(ctx, q, wallettransaction.FieldAmount)
}

func (r WalletRepo) ListTransactions(ctx context.Context, q wallet.Query) ([]*wallet.Transaction, int64, error) {
	bq := r.client.WalletTransaction.Query().
		Where(wallettransaction.PartyType(q.PartyType), wallettransaction.PartyID(q.PartyID))
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(wallettransaction.FieldCreatedAt)).
		Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*wallet.Transaction, 0, len(rows))
	for _, e := range rows {
		out = append(out, &wallet.Transaction{
			ID: e.ID, AccountID: e.AccountID, PartyType: e.PartyType, PartyID: e.PartyID,
			OrderID: e.OrderID, Type: e.Type, Amount: e.Amount,
			BalanceAfter: e.BalanceAfter, Note: e.Note, CreatedAt: e.CreatedAt,
		})
	}
	return out, int64(total), nil
}

func toWalletAccount(e *ent.WalletAccount) *wallet.Account {
	return &wallet.Account{
		ID: e.ID, PartyType: e.PartyType, PartyID: e.PartyID,
		Balance: e.Balance, Frozen: e.Frozen, UpdatedAt: e.UpdatedAt,
	}
}
