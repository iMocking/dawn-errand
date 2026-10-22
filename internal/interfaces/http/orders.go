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
	response.OK(c, toOrderDetail(o))
}
