package http

import (
	"github.com/gin-gonic/gin"

	appsvc "dawn-errand/internal/application/app"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/response"
)

// AppReportLocation POST /api/app/driver/location。
func (h *Handler) AppReportLocation(c *gin.Context) {
	var in struct {
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
		Heading  float64 `json:"heading"`
		Speed    float64 `json:"speed"`
		Accuracy float64 `json:"accuracy"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	loc, err := h.Tracking.Report(c.Request.Context(), p.RiderID, appsvc.LocationInput{
		Lat: in.Lat, Lng: in.Lng, Heading: in.Heading, Speed: in.Speed, Accuracy: in.Accuracy,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "位置已上报", gin.H{"updated_at": loc.UpdatedAt})
}

// AppTrackOrder GET /api/app/client/orders/:id/track。
func (h *Handler) AppTrackOrder(c *gin.Context) {
	p := middleware.Principal(c)
	view, err := h.Tracking.Track(c.Request.Context(), p.UserID, c.Param("id"))
	if err != nil {
		response.Err(c, err)
		return
	}
	o := view.Order
	body := gin.H{
		"order_id": o.ID, "status": string(o.Status), "status_code": order.Code(o.Status),
		"pickup": gin.H{
			"name": merchantPickupName(o), "address": o.Merchant.Address,
			"lat": o.PickupLat, "lng": o.PickupLng,
		},
		"delivery": gin.H{
			"name": "送达点", "address": o.Address.Detail,
			"lat": o.DeliveryLat, "lng": o.DeliveryLng,
		},
		"route":        trackRouteDto(view.Route),
		"traveled_km":  view.TraveledKM,
		"remaining_km": view.RemainKM,
		"eta_minutes":  view.ETAMinutes,
		"progress":     view.Progress,
		"steps":        trackStepsDto(view.Events),
	}
	if view.Rider != nil {
		body["rider"] = gin.H{
			"id": view.Rider.ID, "name": view.Rider.Name,
			"phone": maskPhone(view.Rider.Phone), "rating": view.Rider.Rating,
			"vehicle": view.Rider.Vehicle,
		}
	}
	if view.Location != nil {
		body["rider_lat"] = view.Location.Lat
		body["rider_lng"] = view.Location.Lng
		body["rider_heading"] = view.Location.Heading
		body["rider_speed"] = view.Location.Speed
		body["updated_at"] = view.Location.UpdatedAt
	}
	response.OK(c, body)
}

func merchantPickupName(o *order.Order) string {
	if o.Merchant.Name != "" {
		return o.Merchant.Name
	}
	return "取件点"
}

func trackRouteDto(points []appsvc.TrackPoint) []gin.H {
	out := make([]gin.H, 0, len(points))
	for _, p := range points {
		out = append(out, gin.H{"lat": p.Lat, "lng": p.Lng})
	}
	return out
}

func trackStepsDto(events []*order.Event) []gin.H {
	out := make([]gin.H, 0, len(events))
	for i, e := range events {
		state := "done"
		if i == len(events)-1 && !trackTerminal(e.ToStatus) {
			state = "active"
		}
		out = append(out, gin.H{
			"title": orderEventTitle(e.ToStatus), "status_code": order.Code(e.ToStatus),
			"state": state, "at": e.CreatedAt,
		})
	}
	return out
}

func trackTerminal(s order.OrderStatus) bool {
	switch s {
	case order.OrderDone, order.OrderCanceled, order.StatusRefunded:
		return true
	}
	return false
}

// AppListNotifications GET /api/app/common/notifications。
func (h *Handler) AppListNotifications(c *gin.Context) {
	p := middleware.Principal(c)
	page, size := appPage(c)
	rows, total, err := h.Notifications.List(c.Request.Context(), p.UserID, page, size)
	if err != nil {
		response.Err(c, err)
		return
	}
	unread, _ := h.Notifications.UnreadCount(c.Request.Context(), p.UserID)
	list := make([]gin.H, 0, len(rows))
	for _, n := range rows {
		list = append(list, gin.H{
			"id": n.ID, "title": n.Title, "content": n.Content,
			"type": n.Type, "ref_id": n.RefID,
			"read": n.ReadAt > 0, "created_at": n.CreatedAt,
		})
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "unread": unread,
		"page": page, "page_size": size, "has_more": int64(page*size) < total,
	})
}

// AppReadNotification POST /api/app/common/notifications/:id/read。
func (h *Handler) AppReadNotification(c *gin.Context) {
	p := middleware.Principal(c)
	if err := h.Notifications.MarkRead(c.Request.Context(), p.UserID, c.Param("id")); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已读", gin.H{})
}

// AppReadAllNotifications POST /api/app/common/notifications/read-all。
func (h *Handler) AppReadAllNotifications(c *gin.Context) {
	p := middleware.Principal(c)
	if err := h.Notifications.MarkAllRead(c.Request.Context(), p.UserID); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "全部已读", gin.H{})
}
