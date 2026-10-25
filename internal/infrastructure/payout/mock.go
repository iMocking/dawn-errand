// Package payout 提现打款网关实现。
package payout

import (
	"context"
	"strings"

	"dawn-errand/internal/pkg/ulidgen"
)

// Mock 模拟打款网关（dev/staging；生产替换为银行/微信代付实现）。
type Mock struct{}

// NewMock 构建模拟打款网关。
func NewMock() *Mock { return &Mock{} }

// Payout 返回模拟回执号。
func (m *Mock) Payout(ctx context.Context, withdrawalID, partyType, partyID string, amount float64) (string, error) {
	return "MOCKPAY" + strings.ToUpper(ulidgen.New())[:20], nil
}
