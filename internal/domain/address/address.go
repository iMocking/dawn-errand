// Package address 用户收货地址限界上下文。
package address

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
	Save(a *Address) error
	FindByID(id string) (*Address, error)
	ListByUser(userID string) ([]*Address, error)
	Delete(id string) error
	// ClearDefault 清除该用户所有默认标记。
	ClearDefault(userID string) error
	CountByUser(userID string) (int64, error)
}
