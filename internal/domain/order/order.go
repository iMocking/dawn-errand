// Package order 订单限界上下文（外卖/帮买/帮送/帮取/帮办）。
package order

import "dawn-errand/internal/domain/tenant"

// OrderType 订单类型。
type OrderType string

const (
	OrderTypeFood   OrderType = "外卖"
	OrderTypeBuy    OrderType = "帮买"
	OrderTypeSend   OrderType = "帮送"
	OrderTypePick   OrderType = "帮取"
	OrderTypeErrand OrderType = "帮办"
)

// OrderStatus 订单状态。
type OrderStatus string

const (
	OrderPending    OrderStatus = "待接单"
	OrderDelivering OrderStatus = "配送中"
	OrderPicked     OrderStatus = "已取件待送达"
	OrderDone       OrderStatus = "已完成"
	OrderCanceled   OrderStatus = "已取消"
	OrderCookingOot OrderStatus = "出餐超时"
	OrderAbnormal   OrderStatus = "异常"
)

// OrderGoods 商品明细。
type OrderGoods struct {
	DishID string  `json:"dish_id"`
	Name   string  `json:"name"`
	Spec   string  `json:"spec"`
	Price  float64 `json:"price"`
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

// OrderFee 费用明细。
type OrderFee struct {
	Label    string  `json:"label"`
	Value    float64 `json:"value"`
	Negative bool    `json:"negative"`
}

// OrderStep 状态步骤。
type OrderStep struct {
	Title string `json:"title"`
	Time  int64  `json:"time"`  // 完成时间，0=未完成
	State string `json:"state"` // done / active / pending
	Sub   string `json:"sub"`
}

// MerchantSnapshot 商家快照。
type MerchantSnapshot struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Status  string `json:"status"`
}

// UserSnapshot 用户快照。
type UserSnapshot struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

// RiderSnapshot 骑手快照。
type RiderSnapshot struct {
	Name    string `json:"name"`
	Vehicle string `json:"vehicle"`
	Phone   string `json:"phone"`
	Service string `json:"service"`
	Score   string `json:"score"`
}

// Order 订单聚合根。
type Order struct {
	ID         string // 平台内部单号 SG...
	OrderNo    string // 用户可见单号（不含 #）
	RegionID   string
	MerchantID string
	UserID     string
	RiderID    string

	Type      OrderType
	Status    OrderStatus
	Route     string
	UserName  string // 下单用户姓名（快照）
	RiderName string

	Amount           float64 // 商品金额
	CreatedAt        int64
	Channel          string
	EstimatedArrival int64
	Remark           string

	Goods []OrderGoods
	Fees  []OrderFee
	Steps []OrderStep

	Merchant MerchantSnapshot
	User     UserSnapshot
	Rider    RiderSnapshot

	PayAmount float64
	PayMethod string
	FundFlow  string

	// 交易闭环字段（M2）。
	ClientOrderNo      string
	AddressID          string
	Address            AddressSnapshot
	PickupLat          float64
	PickupLng          float64
	DeliveryLat        float64
	DeliveryLng        float64
	DeliveryDistanceKM float64
	DeliveryFee        float64
	PackingFee         float64
	DiscountAmount     float64
	CouponAmount       float64
	CouponID           string
	PaymentStatus      string
	AcceptedAt         int64
	ReadyAt            int64
	PickedAt           int64
	DeliveredAt        int64
	CanceledAt         int64
	ConfirmedAt        int64
	ReviewedAt         int64
	CancelReason       string
	RiderEarning       float64
	PickupCode         string
	DeliveryCode       string
}

// AddressSnapshot 下单时地址快照（避免地址后续修改影响历史订单）。
type AddressSnapshot struct {
	ContactName string
	Phone       string
	Detail      string
	Tag         string
	Lat         float64
	Lng         float64
}

// Query 订单列表查询。
type Query struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
	Statuses []string // 多状态筛选（Tab），优先于 Status
	UserID   string   // 用户详情近期订单
	Limit    int
}

// Stats 订单模块聚合统计（dashboard 使用）。
type Stats struct {
	TodayOrders     int64
	TodayGMV        float64
	YesterdayOrders int64
	YesterdayGMV    float64
	OnTimeRate      float64
}

// DayTrend 单日趋势点。
type DayTrend struct {
	Day    string // 10/16
	Orders int64
	GMV    float64
}

// OrderRepository 订单仓储端口。
type OrderRepository interface {
	Save(o *Order) error
	FindByID(id string) (*Order, error)
	// FindByClientOrderNo 幂等查询：同一用户同一客户端单号。
	FindByClientOrderNo(userID, clientOrderNo string) (*Order, error)
	// List 订单列表（scope 强制隔离）。
	List(q Query, scope tenant.Scope) ([]*Order, int64, error)
	// ListByUser 用户近期订单。
	ListByUser(userID string, limit int) ([]*Order, error)
	// ListByMerchant 商家近期订单（管理端商家详情）。
	ListByMerchant(merchantID string, limit int) ([]*Order, error)
	// ListByRider 骑手近期订单（管理端骑手详情）。
	ListByRider(riderID string, limit int) ([]*Order, error)
	// Stats 当日/昨日统计（scope 隔离）。
	Stats(scope tenant.Scope, now int64) (*Stats, error)
	// Trend7d 近 7 日订单/交易额趋势（scope 隔离）。
	Trend7d(scope tenant.Scope, now int64) ([]DayTrend, error)
	// CountByRegion 区域今日订单数。
	CountByRegion(regionID string, now int64) (int64, error)
	// CountToday 全局今日订单数。
	CountToday(now int64) (int64, error)
	// SumTodayGMV 全局今日交易额。
	SumTodayGMV(now int64) (float64, error)
	// CountUserOrders 用户累计订单数。
	CountUserOrders(userID string) (int64, error)
	// CountUserMonth 用户本月订单数。
	CountUserMonth(userID string, now int64) (int64, error)
	// CountUserSpent 用户累计消费。
	CountUserSpent(userID string) (float64, error)
	// CountActiveMonth 本月活跃下单用户数（scope 隔离）。
	CountActiveMonth(scope tenant.Scope, now int64) (int64, error)
	// CountRecent 最近 7 天活跃用户数（用于活跃率）。
	CountRecentActiveUsers(now int64) (int64, error)
	// CountByType 按订单类型计数（scope 隔离）。
	CountByType(scope tenant.Scope) (map[OrderType]int64, error)
	// CountAll 订单总数（scope 隔离）。
	CountAll(scope tenant.Scope) (int64, error)
	// UpdateStatus 条件流转（from→to）；返回是否更新成功（0 行表示状态已变化/并发冲突）。
	UpdateStatus(id string, from, to OrderStatus, patch StatusPatch, now int64) (bool, error)
	// AssignRider 抢单：仅当订单待取货且未分配骑手时成功（并发唯一）。
	AssignRider(id, riderID, riderName string, now int64) (bool, error)
	// ListPool 抢单池：本区域待取货且未分配骑手的订单。
	ListPool(regionID string, page, pageSize int) ([]*Order, int64, error)
	// ListPendingPayBefore 超时未支付订单。
	ListPendingPayBefore(before int64, limit int) ([]*Order, error)
	// ListDeliveredBefore 超时未确认收货订单。
	ListDeliveredBefore(before int64, limit int) ([]*Order, error)
}
