// Package wallet 钱包账本：账户余额与流水。
package wallet

import "context"

// Account 钱包账户。PartyType: user / merchant / rider。
type Account struct {
	ID        string
	PartyType string
	PartyID   string
	Balance   float64
	Frozen    float64
	UpdatedAt int64
}

// Transaction 钱包流水（只增不改）。
type Transaction struct {
	ID           string
	AccountID    string
	PartyType    string
	PartyID      string
	OrderID      string
	Type         string // order_income / delivery_income / withdraw / refund / adjust
	Amount       float64
	BalanceAfter float64
	Note         string
	CreatedAt    int64
}

// Query 流水查询。
type Query struct {
	PartyType string
	PartyID   string
	Page      int
	PageSize  int
}

// Repository 钱包仓储端口。每个方法以 ctx 为首参，由调用方传入请求上下文
// （超时/取消/链路只到这一层才不会断）。
type Repository interface {
	// Ensure 返回账户，不存在则创建。
	Ensure(ctx context.Context, partyType, partyID string, now int64) (*Account, error)
	Find(ctx context.Context, partyType, partyID string) (*Account, error)
	// Adjust 调整余额（delta 可为负），返回调整后账户；allowNegative=false 且余额不足返回 false。
	Adjust(ctx context.Context, partyType, partyID string, delta float64, allowNegative bool, now int64) (bool, *Account, error)
	SaveTransaction(ctx context.Context, t *Transaction) error
	ListTransactions(ctx context.Context, q Query) ([]*Transaction, int64, error)
	// SumByTypeSince 统计指定类型流水金额（用于本月收入）。
	SumByTypeSince(ctx context.Context, partyType, partyID string, types []string, since int64) (float64, error)
}
