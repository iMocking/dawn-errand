package order

import "context"

// StatusPatch 状态流转时一并写入的字段（零值表示不修改）。
type StatusPatch struct {
	RiderID          string
	RiderName        string
	AcceptedAt       int64
	ReadyAt          int64
	PickedAt         int64
	DeliveredAt      int64
	CanceledAt       int64
	ConfirmedAt      int64
	ReviewedAt       int64
	CancelReason     string
	PaymentStatus    string
	PayMethod        string
	FundFlow         string
	PickupCode       string
	DeliveryCode     string
	RiderEarning     *float64
	DeliveryDistance *float64
}

// ActorType 事件操作方。
type ActorType string

const (
	ActorClient ActorType = "client"
	ActorStore  ActorType = "store"
	ActorDriver ActorType = "driver"
	ActorSystem ActorType = "system"
)

// Event 订单事件。
type Event struct {
	ID         string
	OrderID    string
	FromStatus OrderStatus
	ToStatus   OrderStatus
	ActorType  ActorType
	ActorID    string
	Note       string
	CreatedAt  int64
}

// EventRepository 订单事件仓储端口。
type EventRepository interface {
	Save(ctx context.Context, e *Event) error
	ListByOrder(ctx context.Context, orderID string) ([]*Event, error)
}

// PaymentStatus 支付状态。
type PaymentStatus string

const (
	PaymentCreated  PaymentStatus = "created"
	PaymentPaid     PaymentStatus = "paid"
	PaymentFailed   PaymentStatus = "failed"
	PaymentRefunded PaymentStatus = "refunded"
)

// Payment 支付单。
type Payment struct {
	ID        string
	OrderID   string
	UserID    string
	Channel   string
	Amount    float64
	Status    PaymentStatus
	TradeNo   string
	PaidAt    int64
	CreatedAt int64
	UpdatedAt int64
}

// ReconSummary 单日对账聚合结果（按支付时间归属）。
type ReconSummary struct {
	PaidCount      int64   // 支付成功笔数
	PaidAmount     float64 // 支付成功金额
	PaidOrderCount int64   // 能对到订单的支付笔数
	OrderAmount    float64 // 上述订单的应付金额合计
	MissingOrders  int64   // 找不到订单（或不在可见区域）的支付笔数
	RefundCount    int64
	RefundAmount   float64
}

// PaymentRepository 支付单仓储端口。
type PaymentRepository interface {
	Save(ctx context.Context, p *Payment) error
	FindByOrder(ctx context.Context, orderID string) (*Payment, error)
	UpdateStatus(ctx context.Context, id string, status PaymentStatus, tradeNo string, paidAt int64) error
	// SummarizeRecon 聚合指定支付时间窗口的对账口径；regionID 为空表示全平台。
	// 聚合在仓储层完成（只取轻量列 + 订单批量匹配），禁止逐单 FindByID 造成 N+1。
	SummarizeRecon(ctx context.Context, start, end int64, regionID string) (*ReconSummary, error)
}

// Code 返回订单状态的 App 机器码（后台展示仍用中文 status）。
func Code(s OrderStatus) string {
	switch s {
	case OrderPending:
		return "pending_accept"
	case OrderCookingOot:
		return "abnormal"
	case OrderDelivering:
		return "delivering"
	case OrderPicked:
		return "delivering"
	case OrderDone:
		return "completed"
	case OrderCanceled:
		return "canceled"
	case OrderAbnormal:
		return "abnormal"
	case StatusPendingPay:
		return "pending_pay"
	case StatusPreparing:
		return "preparing"
	case StatusReady:
		return "ready"
	case StatusDelivered:
		return "delivered"
	case StatusPendingReview:
		return "pending_review"
	case StatusAfterSale:
		return "after_sale"
	case StatusRefunded:
		return "refunded"
	default:
		return string(s)
	}
}

// 交易闭环新增状态（中文入库，兼容后台展示）。
const (
	StatusPendingPay    OrderStatus = "待支付"
	StatusPreparing     OrderStatus = "备餐中"
	StatusReady         OrderStatus = "待取货"
	StatusDelivered     OrderStatus = "已送达"
	StatusPendingReview OrderStatus = "待评价"
	StatusAfterSale     OrderStatus = "退款售后"
	StatusRefunded      OrderStatus = "已退款"
)

// transitions 允许的状态流转。
var transitions = map[OrderStatus][]OrderStatus{
	StatusPendingPay:    {OrderPending, OrderCanceled},
	OrderPending:        {StatusPreparing, OrderCanceled},
	StatusPreparing:     {StatusReady, OrderCanceled},
	StatusReady:         {OrderDelivering, OrderCanceled},
	OrderDelivering:     {StatusDelivered, OrderAbnormal},
	StatusDelivered:     {StatusPendingReview},
	StatusPendingReview: {OrderDone, StatusAfterSale},
	OrderDone:           {StatusAfterSale},
}

// CanTransition 是否允许 from→to。
func CanTransition(from, to OrderStatus) bool {
	for _, next := range transitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// ClientTabStatuses 用户端 Tab 对应状态。
func ClientTabStatuses(tab string) []string {
	switch tab {
	case "ongoing":
		return statuses(StatusPendingPay, OrderPending, StatusPreparing, StatusReady, OrderDelivering, StatusDelivered)
	case "pending_review":
		return statuses(StatusPendingReview)
	case "completed":
		return statuses(OrderDone)
	case "canceled":
		return statuses(OrderCanceled)
	case "after_sale":
		return statuses(StatusAfterSale, StatusRefunded)
	case "all", "":
		return nil
	default:
		return nil
	}
}

// StoreTabStatuses 商家端 Tab 对应状态。
func StoreTabStatuses(tab string) []string {
	switch tab {
	case "new":
		return statuses(OrderPending)
	case "preparing":
		return statuses(StatusPreparing)
	case "ready":
		return statuses(StatusReady)
	case "delivering":
		return statuses(OrderDelivering)
	case "done":
		return statuses(StatusDelivered, StatusPendingReview, OrderDone)
	case "canceled":
		return statuses(OrderCanceled)
	case "after_sale":
		return statuses(StatusAfterSale, StatusRefunded)
	default:
		return nil
	}
}

// DriverTabStatuses 骑手端 Tab 对应状态。
func DriverTabStatuses(tab string) []string {
	switch tab {
	case "pending":
		return statuses(StatusReady)
	case "delivering":
		return statuses(OrderDelivering)
	case "done":
		return statuses(StatusDelivered, StatusPendingReview, OrderDone)
	case "canceled":
		return statuses(OrderCanceled, StatusRefunded, OrderAbnormal)
	default:
		return nil
	}
}

func statuses(list ...OrderStatus) []string {
	out := make([]string, 0, len(list))
	for _, s := range list {
		out = append(out, string(s))
	}
	return out
}
