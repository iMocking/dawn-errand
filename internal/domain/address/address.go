// Package address 用户收货地址限界上下文。
package address

import "context"

// Address 收货地址。
type Address struct {
	ID          string
	UserID      string
	ContactName string
	Phone       string
	RegionID    string
	Detail      string
	Tag         string
	Lat         float64
	Lng         float64
	IsDefault   bool
	CreatedAt   int64
	UpdatedAt   int64
}

// Repository 地址仓储端口。
type Repository interface {
	Save(ctx context.Context, a *Address) error
	FindByID(ctx context.Context, id string) (*Address, error)
	ListByUser(ctx context.Context, userID string) ([]*Address, error)
	Delete(ctx context.Context, id string) error
	// ClearDefault 清除该用户所有默认标记。
	ClearDefault(ctx context.Context, userID string) error
	CountByUser(ctx context.Context, userID string) (int64, error)
}
