package http

import (
	"strings"

	"github.com/gin-gonic/gin"

	appsvc "dawn-errand/internal/application/app"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/response"
)

// ===== 用户端订单 =====

type orderPlaceBody struct {
	ContactName string  `json:"contact_name"`
	Phone       string  `json:"phone"`
	Address     string  `json:"address"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

type createOrderBody struct {
	Type       string `json:"type"`
	MerchantID string `json:"merchant_id"`
	Items      []struct {
		DishID string `json:"dish_id"`
		Count  int    `json:"count"`
	} `json:"items"`
	AddressID      string          `json:"address_id"`
	CouponID       string          `json:"coupon_id"`
	Remark         string          `json:"remark"`
	ClientOrderNo  string          `json:"client_order_no"`
	Pickup         *orderPlaceBody `json:"pickup"`
	Dropoff        *orderPlaceBody `json:"dropoff"`
	GoodsDesc      string          `json:"goods_desc"`
	WeightKG       float64         `json:"weight_kg"`
	ExpectedAmount float64         `json:"expected_amount"`
}

// AppCreateOrder POST /api/app/client/orders。
func (h *Handler) AppCreateOrder(c *gin.Context) {
	var in createOrderBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	items := make([]appsvc.CreateItem, 0, len(in.Items))
	for _, it := range in.Items {
		items = append(items, appsvc.CreateItem{DishID: it.DishID, Count: it.Count})
	}
	p := middleware.Principal(c)
	o, err := h.OrderApp.Create(c.Request.Context(), p.UserID, appsvc.CreateOrderInput{
		Type: in.Type, MerchantID: in.MerchantID, Items: items,
		AddressID: in.AddressID, CouponID: in.CouponID, Remark: in.Remark,
		ClientOrderNo: in.ClientOrderNo,
		Pickup:        toPlaceInput(in.Pickup), Dropoff: toPlaceInput(in.Dropoff),
		GoodsDesc: in.GoodsDesc, WeightKG: in.WeightKG, ExpectedAmount: in.ExpectedAmount,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "订单已创建", toClientOrderDetail(&appsvc.OrderDetail{Order: o}))
}

func toPlaceInput(p *orderPlaceBody) *appsvc.PlaceInput {
	if p == nil {
		return nil
	}
	return &appsvc.PlaceInput{
		ContactName: p.ContactName, Phone: p.Phone, Address: p.Address, Lat: p.Lat, Lng: p.Lng,
	}
}

// AppListClientOrders GET /api/app/client/orders。
func (h *Handler) AppListClientOrders(c *gin.Context) {
	p := middleware.Principal(c)
	page, size := appPage(c)
	rows, total, err := h.OrderApp.ClientList(c.Request.Context(), p.UserID,
		c.DefaultQuery("tab", "all"), c.Query("keyword"), page, size)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, o := range rows {
		list = append(list, toClientOrderRow(o))
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": size,
		"has_more": int64(page*size) < total,
	})
}

// AppGetClientOrder GET /api/app/client/orders/:id。
func (h *Handler) AppGetClientOrder(c *gin.Context) {
	p := middleware.Principal(c)
	d, err := h.OrderApp.ClientDetail(c.Request.Context(), p.UserID, c.Param("id"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toClientOrderDetail(d))
}

// AppCancelClientOrder POST /api/app/client/orders/:id/cancel。
func (h *Handler) AppCancelClientOrder(c *gin.Context) {
	var in struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&in)
	p := middleware.Principal(c)
	o, err := h.OrderApp.CancelClient(c.Request.Context(), p.UserID, c.Param("id"), in.Reason)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "订单已取消", toClientOrderRow(o))
}

// AppPayOrder POST /api/app/client/orders/:id/pay。
func (h *Handler) AppPayOrder(c *gin.Context) {
	var in struct {
		Channel string `json:"channel"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if strings.TrimSpace(in.Channel) == "" {
		response.ErrStatus(c, 400, errcode.CodeBadRequest, "支付方式必填")
		return
	}
	p := middleware.Principal(c)
	o, err := h.OrderApp.Pay(c.Request.Context(), p.UserID, c.Param("id"), in.Channel)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "支付成功", gin.H{
		"id": o.ID, "order_no": o.OrderNo, "status": string(o.Status),
		"status_code": order.Code(o.Status), "payment_status": o.PaymentStatus,
		"pay_amount": o.PayAmount, "pay_method": o.PayMethod,
	})
}

// AppConfirmOrder POST /api/app/client/orders/:id/confirm。
func (h *Handler) AppConfirmOrder(c *gin.Context) {
	p := middleware.Principal(c)
	o, err := h.OrderApp.Confirm(c.Request.Context(), p.UserID, c.Param("id"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已确认收货", toClientOrderRow(o))
}

// ===== 商家端订单 =====

// AppListStoreOrders GET /api/app/store/orders。
func (h *Handler) AppListStoreOrders(c *gin.Context) {
	p := middleware.Principal(c)
	page, size := appPage(c)
	rows, total, err := h.OrderApp.StoreList(c.Request.Context(), p.MerchantID,
		c.DefaultQuery("tab", "new"), c.Query("keyword"), page, size)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, o := range rows {
		list = append(list, toStoreOrderRow(o))
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": size,
		"has_more": int64(page*size) < total,
	})
}

// AppGetStoreOrder GET /api/app/store/orders/:id。
func (h *Handler) AppGetStoreOrder(c *gin.Context) {
	p := middleware.Principal(c)
	d, err := h.OrderApp.StoreDetail(c.Request.Context(), p.MerchantID, c.Param("id"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toStoreOrderDetail(d))
}

// AppAcceptOrder POST /api/app/store/orders/:id/accept。
func (h *Handler) AppAcceptOrder(c *gin.Context) {
	p := middleware.Principal(c)
	o, err := h.OrderApp.Accept(c.Request.Context(), p.MerchantID, c.Param("id"), middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已接单", toStoreOrderRow(o))
}

// AppRejectOrder POST /api/app/store/orders/:id/reject。
func (h *Handler) AppRejectOrder(c *gin.Context) {
	var in struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&in)
	p := middleware.Principal(c)
	o, err := h.OrderApp.Reject(c.Request.Context(), p.MerchantID, c.Param("id"), in.Reason, middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已拒单", toStoreOrderRow(o))
}

// AppReadyOrder POST /api/app/store/orders/:id/ready。
func (h *Handler) AppReadyOrder(c *gin.Context) {
	p := middleware.Principal(c)
	o, err := h.OrderApp.Ready(c.Request.Context(), p.MerchantID, c.Param("id"), middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已出餐", gin.H{
		"id": o.ID, "status": string(o.Status), "status_code": order.Code(o.Status),
		"pickup_code": o.PickupCode, "delivery_code": o.DeliveryCode,
	})
}

// ===== 骑手端订单 =====

// AppSetDriverStatus PUT /api/app/driver/status。
func (h *Handler) AppSetDriverStatus(c *gin.Context) {
	var in struct {
		Online bool `json:"online"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	r, err := h.OrderApp.SetDriverStatus(c.Request.Context(), p.RiderID, in.Online, middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "状态已更新", gin.H{
		"online": r.Status == "online", "status": string(r.Status), "status_code": string(r.Status),
	})
}

// AppDriverPool GET /api/app/driver/order-pool。
func (h *Handler) AppDriverPool(c *gin.Context) {
	p := middleware.Principal(c)
	page, size := appPage(c)
	lat, lng := appLatLng(c)
	items, total, err := h.OrderApp.Pool(c.Request.Context(), p.RiderID, lat, lng, page, size)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(items))
	for _, item := range items {
		list = append(list, toPoolRow(item))
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": size,
		"has_more": int64(page*size) < total,
	})
}

// AppListDriverOrders GET /api/app/driver/orders。
func (h *Handler) AppListDriverOrders(c *gin.Context) {
	p := middleware.Principal(c)
	page, size := appPage(c)
	rows, total, err := h.OrderApp.DriverList(c.Request.Context(), p.RiderID,
		c.DefaultQuery("tab", "pending"), page, size)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, o := range rows {
		list = append(list, toDriverOrderRow(o, p.RiderID))
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": size,
		"has_more": int64(page*size) < total,
	})
}

// AppGetDriverOrder GET /api/app/driver/orders/:id。
func (h *Handler) AppGetDriverOrder(c *gin.Context) {
	p := middleware.Principal(c)
	d, err := h.OrderApp.DriverDetail(c.Request.Context(), p.RiderID, c.Param("id"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toDriverOrderDetail(d, p.RiderID))
}

// AppGrabOrder POST /api/app/driver/orders/:id/grab。
func (h *Handler) AppGrabOrder(c *gin.Context) {
	p := middleware.Principal(c)
	o, err := h.OrderApp.Grab(c.Request.Context(), p.RiderID, c.Param("id"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "抢单成功", gin.H{
		"id": o.ID, "order_no": o.OrderNo, "status": string(o.Status),
		"status_code": order.Code(o.Status),
		"pickup_code": o.PickupCode, "delivery_code": o.DeliveryCode,
	})
}

// AppPickupOrder POST /api/app/driver/orders/:id/pickup。
func (h *Handler) AppPickupOrder(c *gin.Context) {
	var in struct {
		PickupCode string `json:"pickup_code"`
		PhotoURL   string `json:"photo_url"`
	}
	_ = c.ShouldBindJSON(&in)
	p := middleware.Principal(c)
	o, err := h.OrderApp.Pickup(c.Request.Context(), p.RiderID, c.Param("id"), in.PickupCode)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已确认取货", toDriverOrderRow(o, p.RiderID))
}

// AppDeliverOrder POST /api/app/driver/orders/:id/deliver。
func (h *Handler) AppDeliverOrder(c *gin.Context) {
	var in struct {
		DeliveryCode string `json:"delivery_code"`
		PhotoURL     string `json:"photo_url"`
	}
	_ = c.ShouldBindJSON(&in)
	p := middleware.Principal(c)
	o, err := h.OrderApp.Deliver(c.Request.Context(), p.RiderID, c.Param("id"), in.DeliveryCode)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已送达", toDriverOrderRow(o, p.RiderID))
}

// ===== 订单 DTO =====

func orderStatusDto(o *order.Order) (string, string) {
	return string(o.Status), order.Code(o.Status)
}

func orderItemsSummary(o *order.Order) string {
	if len(o.Goods) == 0 {
		return ""
	}
	parts := make([]string, 0, 2)
	for i, g := range o.Goods {
		if i >= 2 {
			break
		}
		parts = append(parts, g.Name+"×"+itoa(g.Count))
	}
	out := strings.Join(parts, "、")
	if len(o.Goods) > 2 {
		out += " 等" + itoa(len(o.Goods)) + "件"
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func orderPricingDto(o *order.Order) gin.H {
	return gin.H{
		"goods_amount": o.Amount, "packing_fee": o.PackingFee,
		"delivery_fee": o.DeliveryFee, "discount_amount": o.DiscountAmount,
		"pay_amount": o.PayAmount,
	}
}

func orderItemsDto(o *order.Order) []gin.H {
	out := make([]gin.H, 0, len(o.Goods))
	for _, g := range o.Goods {
		out = append(out, gin.H{
			"dish_id": g.DishID, "name": g.Name, "spec": g.Spec,
			"price": g.Price, "count": g.Count, "amount": g.Amount,
		})
	}
	return out
}

func orderTimelineDto(events []*order.Event) []gin.H {
	out := make([]gin.H, 0, len(events))
	for i, e := range events {
		state := "done"
		if i == len(events)-1 && !terminalStatus(e.ToStatus) {
			state = "active"
		}
		out = append(out, gin.H{
			"status": string(e.ToStatus), "status_code": order.Code(e.ToStatus),
			"title": orderEventTitle(e.ToStatus), "at": e.CreatedAt,
			"state": state, "sub": e.Note,
		})
	}
	return out
}

func terminalStatus(s order.OrderStatus) bool {
	switch s {
	case order.OrderDone, order.OrderCanceled, order.StatusRefunded:
		return true
	}
	return false
}

func orderEventTitle(s order.OrderStatus) string {
	switch s {
	case order.StatusPendingPay:
		return "已下单"
	case order.OrderPending:
		return "待商家接单"
	case order.StatusPreparing:
		return "商家备餐"
	case order.StatusReady:
		return "待骑手取货"
	case order.OrderDelivering:
		return "骑手配送中"
	case order.StatusDelivered:
		return "已送达"
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

func toClientOrderRow(o *order.Order) gin.H {
	status, code := orderStatusDto(o)
	return gin.H{
		"id": o.ID, "order_no": o.OrderNo, "type": string(o.Type),
		"status": status, "status_code": code,
		"shop_name": o.Merchant.Name, "items_summary": orderItemsSummary(o),
		"pay_amount": o.PayAmount, "created_at": o.CreatedAt,
		"actions": appsvc.ClientActions(o),
	}
}

func toClientOrderDetail(d *appsvc.OrderDetail) gin.H {
	o := d.Order
	status, code := orderStatusDto(o)
	body := gin.H{
		"id": o.ID, "order_no": o.OrderNo, "type": string(o.Type),
		"status": status, "status_code": code,
		"created_at": o.CreatedAt, "estimated_arrival": o.EstimatedArrival,
		"merchant": gin.H{
			"id": o.MerchantID, "name": o.Merchant.Name,
			"phone": maskPhone(o.Merchant.Phone), "address": o.Merchant.Address,
		},
		"address": gin.H{
			"contact_name": o.Address.ContactName, "phone": maskPhone(o.Address.Phone),
			"detail": o.Address.Detail, "tag": o.Address.Tag,
			"lat": o.Address.Lat, "lng": o.Address.Lng,
		},
		"items":    orderItemsDto(o),
		"pricing":  orderPricingDto(o),
		"payment":  gin.H{"channel": o.PayMethod, "status": o.PaymentStatus},
		"timeline": orderTimelineDto(d.Events),
		"remark":   o.Remark, "reviewed": o.ReviewedAt > 0,
		"cancel_reason":        o.CancelReason,
		"actions":              appsvc.ClientActions(o),
		"delivery_distance_km": o.DeliveryDistanceKM,
	}
	if o.RiderID != "" {
		body["rider"] = gin.H{
			"id": o.RiderID, "name": o.Rider.Name,
			"phone": maskPhone(o.Rider.Phone), "vehicle": o.Rider.Vehicle,
			"rating": o.Rider.Score,
		}
	}
	return body
}

func toStoreOrderRow(o *order.Order) gin.H {
	status, code := orderStatusDto(o)
	return gin.H{
		"id": o.ID, "order_no": o.OrderNo,
		"status": status, "status_code": code,
		"created_at":    o.CreatedAt,
		"customer":      gin.H{"name": o.Address.ContactName, "phone": o.Address.Phone},
		"address":       o.Address.Detail,
		"items_summary": orderItemsSummary(o),
		"goods_amount":  o.Amount, "pay_amount": o.PayAmount,
		"paid": o.PaymentStatus == "paid", "remark": o.Remark,
		"actions": appsvc.StoreActions(o),
	}
}

func toStoreOrderDetail(d *appsvc.OrderDetail) gin.H {
	o := d.Order
	status, code := orderStatusDto(o)
	body := gin.H{
		"id": o.ID, "order_no": o.OrderNo,
		"status": status, "status_code": code,
		"created_at": o.CreatedAt, "estimated_arrival": o.EstimatedArrival,
		"customer": gin.H{
			"name": o.Address.ContactName, "phone": o.Address.Phone,
		},
		"address":  o.Address.Detail,
		"items":    orderItemsDto(o),
		"pricing":  orderPricingDto(o),
		"payment":  gin.H{"channel": o.PayMethod, "status": o.PaymentStatus},
		"timeline": orderTimelineDto(d.Events),
		"remark":   o.Remark,
		"actions":  appsvc.StoreActions(o),
		"rider": gin.H{
			"id": o.RiderID, "name": o.Rider.Name, "phone": o.Rider.Phone,
		},
	}
	return body
}

func toDriverOrderRow(o *order.Order, riderID string) gin.H {
	status, code := orderStatusDto(o)
	body := gin.H{
		"id": o.ID, "order_no": o.OrderNo, "type": string(o.Type),
		"status": status, "status_code": code,
		"created_at": o.CreatedAt,
		"pickup": gin.H{
			"name": o.Merchant.Name, "address": o.Merchant.Address,
		},
		"dropoff": gin.H{
			"name": o.Address.ContactName, "address": o.Address.Detail,
		},
		"distance_km":   o.DeliveryDistanceKM,
		"reward":        o.DeliveryFee,
		"rider_earning": o.RiderEarning,
		"actions":       appsvc.DriverActions(o, riderID),
	}
	return body
}

func toDriverOrderDetail(d *appsvc.OrderDetail, riderID string) gin.H {
	o := d.Order
	status, code := orderStatusDto(o)
	body := gin.H{
		"id": o.ID, "order_no": o.OrderNo, "type": string(o.Type),
		"status": status, "status_code": code,
		"created_at": o.CreatedAt, "estimated_arrival": o.EstimatedArrival,
		"pickup": gin.H{
			"name": o.Merchant.Name, "address": o.Merchant.Address,
			"phone": o.Merchant.Phone, "lat": o.PickupLat, "lng": o.PickupLng,
		},
		"dropoff": gin.H{
			"name": o.Address.ContactName, "address": o.Address.Detail,
			"phone": o.Address.Phone, "lat": o.DeliveryLat, "lng": o.DeliveryLng,
		},
		"items":         orderItemsDto(o),
		"distance_km":   o.DeliveryDistanceKM,
		"reward":        o.DeliveryFee,
		"rider_earning": o.RiderEarning,
		"pickup_code":   o.PickupCode, "delivery_code": o.DeliveryCode,
		"remark":   o.Remark,
		"timeline": orderTimelineDto(d.Events),
		"actions":  appsvc.DriverActions(o, riderID),
	}
	return body
}

func toPoolRow(item *appsvc.PoolItem) gin.H {
	o := item.Order
	status, code := orderStatusDto(o)
	return gin.H{
		"order_id": o.ID, "order_no": o.OrderNo, "type": string(o.Type),
		"status": status, "status_code": code,
		"distance_km": item.DistanceKM, "expected_minutes": item.DeliveryMinutes,
		"pickup": gin.H{
			"name": o.Merchant.Name, "address": o.Merchant.Address,
			"lat": o.PickupLat, "lng": o.PickupLng,
		},
		"dropoff": gin.H{
			"name": o.Address.ContactName, "address": maskAddress(o.Address.Detail),
			"lat": o.DeliveryLat, "lng": o.DeliveryLng,
			"phone_masked": maskPhone(o.Address.Phone),
		},
		"reward":       o.DeliveryFee,
		"reward_label": "配送费",
		"created_at":   o.CreatedAt,
	}
}
