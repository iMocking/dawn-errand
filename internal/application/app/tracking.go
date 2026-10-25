package app

import (
	"context"
	"errors"
	"math"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tracking"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/metrics"
)

// TrackingService 骑手位置与配送追踪。
type TrackingService struct {
	Repos *application.RepoUnit
	Geo   application.GeoStore
}

// LocationInput 位置上报入参。
type LocationInput struct {
	Lat      float64
	Lng      float64
	Heading  float64
	Speed    float64
	Accuracy float64
}

// TrackPoint 坐标点。
type TrackPoint struct {
	Lat float64
	Lng float64
}

// TrackView 配送追踪视图。
type TrackView struct {
	Order      *order.Order
	Rider      *rider.Rider
	Location   *tracking.Location
	Pickup     TrackPoint
	Delivery   TrackPoint
	Route      []TrackPoint
	TraveledKM float64
	RemainKM   float64
	Progress   float64
	ETAMinutes int
	Events     []*order.Event
}

// Report 上报骑手位置（仅在线/忙碌骑手）。
func (s *TrackingService) Report(ctx context.Context, riderID string, in LocationInput) (*tracking.Location, error) {
	if !validCoord(in.Lat, in.Lng) {
		return nil, errcode.New(errcode.CodeBadRequest, "经纬度不合法")
	}
	if in.Heading < 0 || in.Heading > 360 {
		return nil, errcode.New(errcode.CodeBadRequest, "heading 需在 0-360")
	}
	if in.Speed < 0 || in.Speed > 120 {
		return nil, errcode.New(errcode.CodeBadRequest, "speed 不合法")
	}
	if in.Accuracy < 0 || in.Accuracy > 1000 {
		return nil, errcode.New(errcode.CodeBadRequest, "accuracy 不合法")
	}
	r, err := s.Repos.Riders.FindByID(riderID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeRiderNotFound, "骑手不存在")
		}
		return nil, err
	}
	if r.Status != rider.RiderOnline && r.Status != rider.RiderBusy {
		return nil, errcode.New(errcode.CodeRiderOffline, "请先上线再上报位置")
	}
	now := time.Now().Unix()
	loc := &tracking.Location{
		RiderID: riderID, RegionID: r.RegionID,
		Lat: in.Lat, Lng: in.Lng, Heading: in.Heading, Speed: in.Speed,
		Accuracy: in.Accuracy, UpdatedAt: now,
	}
	if err := s.Repos.Locations.Upsert(loc); err != nil {
		return nil, err
	}
	if s.Geo != nil {
		_ = s.Geo.SetRiderLocation(ctx, r.RegionID, riderID, in.Lat, in.Lng)
	}
	metrics.LocationReportsTotal.Inc()
	return loc, nil
}

// Track 用户查看订单配送轨迹（仅本人订单）。
func (s *TrackingService) Track(ctx context.Context, userID, orderID string) (*TrackView, error) {
	o, err := s.Repos.Orders.FindByID(orderID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
		}
		return nil, err
	}
	if o.UserID != userID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
	}
	metrics.TrackRequestsTotal.Inc()
	events, err := s.Repos.OrderEvents.ListByOrder(o.ID)
	if err != nil {
		return nil, err
	}
	view := &TrackView{
		Order:    o,
		Pickup:   TrackPoint{Lat: o.PickupLat, Lng: o.PickupLng},
		Delivery: TrackPoint{Lat: o.DeliveryLat, Lng: o.DeliveryLng},
		Events:   events,
	}
	if o.RiderID != "" {
		view.Rider, _ = s.Repos.Riders.FindByID(o.RiderID)
		if loc, err := s.Repos.Locations.FindByRider(o.RiderID); err == nil {
			view.Location = loc
		}
	}
	total := haversine(view.Pickup.Lat, view.Pickup.Lng, view.Delivery.Lat, view.Delivery.Lng)
	if total > 0 {
		view.Route = interpolate(view.Pickup, view.Delivery, 6)
	}
	switch {
	case o.Status == order.OrderDone, o.Status == order.StatusPendingReview, o.Status == order.StatusDelivered:
		view.Progress = 1
		view.TraveledKM = round2(total)
	case o.Status == order.OrderDelivering && view.Location != nil:
		view.TraveledKM = round2(haversine(view.Pickup.Lat, view.Pickup.Lng, view.Location.Lat, view.Location.Lng))
		view.RemainKM = round2(haversine(view.Location.Lat, view.Location.Lng, view.Delivery.Lat, view.Delivery.Lng))
		if total > 0 {
			view.Progress = math.Min(1, view.TraveledKM/total)
		}
	default:
		view.RemainKM = round2(total)
	}
	if view.Progress < 1 && view.RemainKM > 0 {
		view.ETAMinutes = estimateMinutes(view.RemainKM)
	}
	return view, nil
}

func interpolate(a, b TrackPoint, n int) []TrackPoint {
	if n < 2 {
		return []TrackPoint{a, b}
	}
	out := make([]TrackPoint, 0, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		out = append(out, TrackPoint{
			Lat: a.Lat + (b.Lat-a.Lat)*t,
			Lng: a.Lng + (b.Lng-a.Lng)*t,
		})
	}
	return out
}
