package http

import (
	"dawn-errand/internal/domain/order"
)

// ---- 订单 ----

var orderTypeColor = map[order.OrderType][2]string{
	order.OrderTypeFood:   {"#FFF3E8", "#FF6B2C"},
	order.OrderTypeBuy:    {"#E8F0FF", "#2B7FFF"},
	order.OrderTypeSend:   {"#E8F7F0", "#00B578"},
	order.OrderTypePick:   {"#FFF8E0", "#D48806"},
	order.OrderTypeErrand: {"#F3E8FF", "#722ED1"},
}

var orderStatusColor = map[order.OrderStatus]string{
	order.OrderPending:    "#FF4D4F",
	order.OrderDelivering: "#D48806",
	order.OrderPicked:     "#FF6B2C",
	order.OrderDone:       "#00B578",
	order.OrderCanceled:   "#999999",
	order.OrderCookingOot: "#FF4D4F",
	order.OrderAbnormal:   "#FF4D4F",
}

func orderTypeColors(t order.OrderType) (string, string) {
	if c, ok := orderTypeColor[t]; ok {
		return c[0], c[1]
	}
	return "#F0F0F0", "#666666"
}

// OrderRowDto 订单列表行（对齐前端 OrderRow）。
type OrderRowDto struct {
	ID        string  `json:"id"`
	Route     string  `json:"route"`
	UserName  string  `json:"user_name"`
	Rider     string  `json:"rider"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	CreatedAt int64   `json:"created_at"`
}

func toOrderRow(o *order.Order) OrderRowDto {
	riderText := "待分配骑手"
	if o.RiderName != "" {
		riderText = "骑手：" + o.RiderName
	}
	if o.Status == order.OrderCanceled {
		riderText = "已取消"
	}
	return OrderRowDto{
		ID: o.ID, Route: o.Route, UserName: o.UserName, Rider: riderText,
		Type: string(o.Type), Amount: o.Amount, Status: string(o.Status), CreatedAt: o.CreatedAt,
	}
}

// OrderStepDto 订单步骤（对齐前端 OrderStep）。
type OrderStepDto struct {
	Title string `json:"title"`
	Time  int64  `json:"time,omitempty"`
	State string `json:"state"`
	Sub   string `json:"sub,omitempty"`
}

type OrderGoodsDto struct {
	Name  string  `json:"name"`
	Spec  string  `json:"spec"`
	Price float64 `json:"price"`
}

type OrderFeeDto struct {
	Label    string  `json:"label"`
	Value    float64 `json:"value"`
	Negative bool    `json:"negative,omitempty"`
}

// OrderDetailDto 订单详情（对齐前端 OrderDetail）。
type OrderDetailDto struct {
	ID               string                 `json:"id"`
	OrderNo          string                 `json:"order_no"`
	Status           string                 `json:"status"`
	Type             string                 `json:"type"`
	CreatedAt        int64                  `json:"created_at"`
	Channel          string                 `json:"channel"`
	EstimatedArrival int64                  `json:"estimated_arrival"`
	Steps            []OrderStepDto         `json:"steps"`
	Merchant         order.MerchantSnapshot `json:"merchant"`
	User             order.UserSnapshot     `json:"user"`
	Rider            order.RiderSnapshot    `json:"rider"`
	Goods            []order.OrderGoods     `json:"goods"`
	Remark           string                 `json:"remark"`
	Fees             []order.OrderFee       `json:"fees"`
	PayAmount        float64                `json:"pay_amount"`
	PayMethod        string                 `json:"pay_method"`
	FundFlow         string                 `json:"fund_flow"`
}

func toOrderDetail(o *order.Order) OrderDetailDto {
	steps := make([]OrderStepDto, 0, len(o.Steps))
	for _, s := range o.Steps {
		d := OrderStepDto{Title: s.Title, State: s.State, Sub: s.Sub}
		if s.Time > 0 {
			d.Time = s.Time
		}
		steps = append(steps, d)
	}
	return OrderDetailDto{
		ID: o.ID, OrderNo: o.OrderNo, Status: string(o.Status), Type: string(o.Type),
		CreatedAt: o.CreatedAt, Channel: o.Channel, EstimatedArrival: o.EstimatedArrival,
		Steps: steps, Merchant: o.Merchant, User: o.User, Rider: o.Rider,
		Goods: o.Goods, Remark: o.Remark, Fees: o.Fees,
		PayAmount: o.PayAmount, PayMethod: o.PayMethod, FundFlow: o.FundFlow,
	}
}
