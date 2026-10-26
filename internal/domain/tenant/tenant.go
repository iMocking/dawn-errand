// Package tenant 租户与数据隔离限界上下文。
package tenant

import "context"

// RegionStatus 服务区域状态。
type RegionStatus string

const (
	RegionStatusOpen      RegionStatus = "open"      // 运营中
	RegionStatusPreparing RegionStatus = "preparing" // 筹备中
)

// Region 服务区域 = 平台租户单元（乡镇级）。
type Region struct {
	ID        string
	Name      string
	Status    RegionStatus
	Villages  int     // 下辖村组数
	RadiusKM  float64 // 服务半径 km
	StartFee  float64 // 起步价（元）
	FreeKM    float64 // 起步含里程 km
	Notice    string  // 开通提示（筹备中）
	CreatedAt int64
	UpdatedAt int64
}

// Scope 数据隔离作用域。由认证中间件从 Principal 派生，注入所有租户数据查询。
// 平台角色（super_admin/cs_manager/cs_agent）全局可见；区域站长限本区域；
// 商家限本商家；司机限本人；正式会员限本人订单。
type Scope struct {
	Role       string
	RegionID   string
	MerchantID string
	RiderID    string
	UserID     string
}

// IsGlobal 平台全局角色。
func (s Scope) IsGlobal() bool {
	return s.Role == "super_admin" || s.Role == "cs_manager" || s.Role == "cs_agent"
}

// RegionOnly 是否仅限本区域。
func (s Scope) RegionOnly() bool { return s.Role == "region_station" }

// MerchantOnly 是否仅限本商家。
func (s Scope) MerchantOnly() bool { return s.Role == "merchant" }

// RiderOnly 是否仅限本人（司机）。
func (s Scope) RiderOnly() bool { return s.Role == "driver" }

// UserOnly 是否仅限本人（正式会员）。
func (s Scope) UserOnly() bool { return s.Role == "user" }

// Recognized 角色是否被数据隔离策略显式覆盖。
// 未知或新增角色返回 false，查询/鉴权侧据此默认拒绝——
// 否则「没有匹配分支」会退化成「不加过滤 = 全局可见」。
func (s Scope) Recognized() bool {
	return s.IsGlobal() || s.RegionOnly() || s.MerchantOnly() || s.RiderOnly() || s.UserOnly()
}

// Principal 认证主体：由认证中间件构建，携带数据隔离所需绑定。
type Principal struct {
	UserID     string
	Role       string
	RegionID   string
	MerchantID string
	RiderID    string
}

// Scope 从主体派生数据隔离作用域。
func (p Principal) Scope() Scope {
	return Scope{
		Role:       p.Role,
		RegionID:   p.RegionID,
		MerchantID: p.MerchantID,
		RiderID:    p.RiderID,
		UserID:     p.UserID,
	}
}

// RegionRepository 区域仓储端口。
type RegionRepository interface {
	Save(ctx context.Context, r *Region) error
	FindByID(ctx context.Context, id string) (*Region, error)
	List(ctx context.Context) ([]*Region, error)
	// ListByIDs 按 ID 批量查询。
	ListByIDs(ctx context.Context, ids []string) ([]*Region, error)
	Delete(ctx context.Context, id string) error
}
