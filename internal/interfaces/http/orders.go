package http

import (
	"github.com/gin-gonic/gin"

	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/pkg/response"
)

// ListOrders GET /orders。
func (h *Handler) ListOrders(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Orders.List(c.Request.Context(), order.Query{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: c.Query("status"),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]OrderRowDto, 0, len(rows))
	for _, o := range rows {
		list = append(list, toOrderRow(o))
	}
	pageResult(c, list, total, page, size)
}

// GetOrder GET /orders/:id。
func (h *Handler) GetOrder(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	o, err := h.Orders.Detail(c.Request.Context(), id, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	// 后台详情步骤以 order_events 为准（历史订单无事件时保留快照步骤）。
	if events, err := h.Orders.Events(c.Request.Context(), id); err == nil && len(events) > 0 {
		o.Steps = stepsFromEvents(events)
	}
	response.OK(c, toOrderDetail(o))
}

// stepsFromEvents 将事件时间线映射为后台展示步骤。
func stepsFromEvents(events []*order.Event) []order.OrderStep {
	out := make([]order.OrderStep, 0, len(events))
	for i, e := range events {
		state := "done"
		if i == len(events)-1 && !terminalEvent(e.ToStatus) {
			state = "active"
		}
		out = append(out, order.OrderStep{
			Title: eventStepTitle(e.ToStatus),
			Time:  e.CreatedAt,
			State: state,
			Sub:   e.Note,
		})
	}
	return out
}

func terminalEvent(s order.OrderStatus) bool {
	switch s {
	case order.OrderDone, order.OrderCanceled, order.StatusRefunded:
		return true
	}
	return false
}

func eventStepTitle(s order.OrderStatus) string {
	switch s {
	case order.StatusPendingPay:
		return "已下单"
	case order.OrderPending:
		return "商家接单"
	case order.StatusPreparing:
		return "商家备餐"
	case order.StatusReady:
		return "待取货"
	case order.OrderDelivering:
		return "骑手取餐配送"
	case order.StatusDelivered:
		return "送达完成"
	case order.StatusPendingReview:
		return "待评价"
	case order.OrderDone:
		return "已完成"
	case order.OrderCanceled:
		return "已取消"
	case order.StatusAfterSale:
		return "退款售后"
	case order.StatusRefunded:
		return "已退款"
	}
	return string(s)
}
