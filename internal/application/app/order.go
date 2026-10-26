package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/coupon"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/notification"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/metrics"
	"dawn-errand/internal/pkg/ulidgen"
)

const (
	// extraKMFee 超出区域免费里程后的每公里配送费（元）；后续由费率规则配置化。
	extraKMFee = 1.5
	// riderShareRate 骑手配送费分成比例；M4 钱包账本落地后由配置驱动。
	riderShareRate = 0.8
)

// OrderService 三端订单用例：下单、支付、接单、取送、取消、确认。
type OrderService struct {
	Repos *application.RepoUnit
	Tx    application.TxRunner
	// MockPayment 是否允许模拟第三方支付（仅 dev/test；生产必须走真实渠道回调）。
	MockPayment bool
}

// CreateItem 下单商品行。
type CreateItem struct {
	DishID string
	Count  int
}

// PlaceInput 跑腿取/送点。
type PlaceInput struct {
	ContactName string
	Phone       string
	Address     string
	Lat         float64
	Lng         float64
}

// CreateOrderInput 下单入参（外卖与跑腿共用一个入口）。
type CreateOrderInput struct {
	Type           string
	MerchantID     string
	Items          []CreateItem
	AddressID      string
	CouponID       string
	Remark         string
	ClientOrderNo  string
	Pickup         *PlaceInput
	Dropoff        *PlaceInput
	GoodsDesc      string
	WeightKG       float64
	ExpectedAmount float64
}

// OrderDetail 订单 + 事件时间线。
type OrderDetail struct {
	Order  *order.Order
	Events []*order.Event
}

// PoolItem 抢单池条目。
type PoolItem struct {
	Order           *order.Order
	DistanceKM      float64
	DeliveryMinutes int
}

// pricing 服务端计价结果。
type pricing struct {
	Goods    float64
	Packing  float64
	Delivery float64
	Discount float64
	Pay      float64
}

// Create 创建订单（外卖/帮买/帮送/帮取/帮办）。
func (s *OrderService) Create(ctx context.Context, userID string, in CreateOrderInput) (*order.Order, error) {
	in.Type = strings.TrimSpace(in.Type)
	in.ClientOrderNo = strings.TrimSpace(in.ClientOrderNo)
	in.Remark = strings.TrimSpace(in.Remark)
	if len(in.ClientOrderNo) > 64 {
		return nil, errcode.New(errcode.CodeBadRequest, "client_order_no 过长")
	}
	if utf8.RuneCountInString(in.Remark) > 512 {
		return nil, errcode.New(errcode.CodeBadRequest, "备注不能超过 512 字")
	}
	if in.ClientOrderNo != "" {
		if existing, err := s.Repos.Orders.FindByClientOrderNo(ctx, userID, in.ClientOrderNo); err == nil {
			return existing, nil
		} else if !errors.Is(err, application.ErrNotFound) {
			return nil, err
		}
	}
	if len(in.Items) > 50 {
		return nil, errcode.New(errcode.CodeBadRequest, "单笔订单最多 50 个商品")
	}
	u, err := s.Repos.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, userErr(err)
	}
	if u.RegionID == "" {
		return nil, errcode.NewStatus(400, errcode.CodeRegionRequired, "请先选择服务区域")
	}
	region, err := s.Repos.Regions.FindByID(ctx, u.RegionID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeRegionNotFound, "服务区域不存在")
	}

	now := time.Now().Unix()
	o := &order.Order{
		ID: ulidgen.New(), RegionID: u.RegionID, UserID: u.ID, UserName: u.Name,
		Type: order.OrderType(in.Type), Status: order.StatusPendingPay,
		CreatedAt: now, Channel: "app", ClientOrderNo: in.ClientOrderNo,
		PaymentStatus: "unpaid", Remark: in.Remark,
		User: order.UserSnapshot{Name: u.Name, Phone: u.Phone},
	}
	o.OrderNo = newOrderNo(o.ID, now)
	if in.Type == string(order.OrderTypeFood) {
		if err := s.buildFoodOrder(ctx, o, in, u, region, now); err != nil {
			return nil, err
		}
	} else if isErrandType(in.Type) {
		if err := s.buildErrandOrder(o, in, u, region, now); err != nil {
			return nil, err
		}
	} else {
		return nil, errcode.New(errcode.CodeBadRequest, "不支持的订单类型")
	}
	couponID, err := s.applyCoupon(ctx, o, in.CouponID, now)
	if err != nil {
		return nil, err
	}
	o.Steps = initialSteps(now)

	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		// 外卖：下单即锁库存；取消/超时释放。
		if o.Type == order.OrderTypeFood {
			for _, it := range o.Goods {
				ok, err := tx.Catalog.AdjustStock(ctx, it.DishID, -it.Count)
				if err != nil {
					return err
				}
				if !ok {
					return errcode.New(errcode.CodeDishUnavailable, "商品库存不足："+it.Name)
				}
			}
		}
		if err := tx.Orders.Save(ctx, o); err != nil {
			return err
		}
		if couponID != "" {
			ok, err := tx.Coupons.MarkUsed(ctx, couponID, userID, o.ID, now)
			if err != nil {
				return err
			}
			if !ok {
				return errcode.New(errcode.CodeCouponInvalid, "优惠券不可用或已被使用")
			}
		}
		return s.record(ctx, tx, o, "", order.StatusPendingPay, order.ActorClient, userID, "用户下单", now)
	})
	if err != nil {
		// 并发重复提交：唯一索引冲突时回读首次订单。
		if in.ClientOrderNo != "" {
			if existing, e := s.Repos.Orders.FindByClientOrderNo(ctx, userID, in.ClientOrderNo); e == nil {
				return existing, nil
			}
		}
		return nil, err
	}
	metrics.OrdersCreatedTotal.Inc()
	return o, nil
}

func (s *OrderService) buildFoodOrder(ctx context.Context, o *order.Order, in CreateOrderInput, u *identity.User, region *tenant.Region, now int64) error {
	if len(in.Items) == 0 {
		return errcode.New(errcode.CodeBadRequest, "请选择商品")
	}
	m, err := s.Repos.Merchants.FindByID(ctx, in.MerchantID)
	if err != nil {
		return merchantErr(err)
	}
	if m.RegionID != region.ID {
		return errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
	}
	if m.AuditStatus != merchant.AuditApproved || m.Status == merchant.MerchantStatusBanned {
		return errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
	}
	if m.Status != merchant.MerchantStatusOpen {
		return errcode.New(errcode.CodeMerchantClosed, "商家休息中，暂不接单")
	}
	a, err := s.Repos.Addresses.FindByID(ctx, in.AddressID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return errcode.NewStatus(404, errcode.CodeAddressNotFound, "收货地址不存在")
		}
		return err
	}
	if a.UserID != u.ID {
		return errcode.NewStatus(404, errcode.CodeAddressNotFound, "收货地址不存在")
	}
	if a.RegionID != "" && a.RegionID != m.RegionID {
		return errcode.New(errcode.CodeDeliveryOutRange, "收货地址不在商家配送区域")
	}

	var goods []order.OrderGoods
	seen := map[string]bool{}
	for _, it := range in.Items {
		if it.Count < 1 || it.Count > 99 {
			return errcode.New(errcode.CodeBadRequest, "商品数量需在 1-99 之间")
		}
		d, err := s.Repos.Catalog.FindDish(ctx, it.DishID)
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				return errcode.New(errcode.CodeDishUnavailable, "商品不存在或已下架")
			}
			return err
		}
		if d.MerchantID != m.ID || d.DeletedAt > 0 || !d.OnSale {
			return errcode.New(errcode.CodeDishUnavailable, "商品不存在或已下架："+d.Name)
		}
		if d.Stock < it.Count {
			return errcode.New(errcode.CodeDishUnavailable, "商品库存不足："+d.Name)
		}
		if seen[d.ID] {
			return errcode.New(errcode.CodeBadRequest, "商品重复："+d.Name)
		}
		seen[d.ID] = true
		amount := round2(d.Price * float64(it.Count))
		goods = append(goods, order.OrderGoods{
			DishID: d.ID, Name: d.Name, Price: d.Price, Count: it.Count, Amount: amount,
		})
	}

	goodsAmount := 0.0
	for _, g := range goods {
		goodsAmount += g.Amount
	}
	goodsAmount = round2(goodsAmount)
	if m.MinPrice > 0 && goodsAmount < m.MinPrice {
		return errcode.New(errcode.CodeAmountMismatch, fmt.Sprintf("未达到起送价 ¥%.2f", m.MinPrice))
	}
	deliveryFee := m.DeliveryFee
	if deliveryFee <= 0 {
		deliveryFee = region.StartFee
	}
	distance := 0.0
	if validCoord(m.Lat, m.Lng) && validCoord(a.Lat, a.Lng) {
		distance = haversine(m.Lat, m.Lng, a.Lat, a.Lng)
		if region.RadiusKM > 0 && distance > region.RadiusKM {
			return errcode.New(errcode.CodeDeliveryOutRange, "超出商家配送范围")
		}
	}
	pr := pricing{
		Goods: goodsAmount, Packing: round2(m.PackingFee),
		Delivery: round2(deliveryFee),
	}
	pr.Discount = 0
	pr.Pay = round2(math.Max(0, pr.Goods+pr.Packing+pr.Delivery-pr.Discount))

	o.MerchantID = m.ID
	o.Merchant = order.MerchantSnapshot{Name: m.Name, Address: m.Address, Phone: m.Phone, Status: "营业中"}
	o.Amount = pr.Goods
	o.PayAmount = pr.Pay
	o.PayMethod = ""
	o.DeliveryFee = pr.Delivery
	o.PackingFee = pr.Packing
	o.DiscountAmount = pr.Discount
	o.CouponAmount = 0
	o.CouponID = in.CouponID
	o.DeliveryDistanceKM = round2(distance)
	o.PickupLat, o.PickupLng = m.Lat, m.Lng
	o.DeliveryLat, o.DeliveryLng = a.Lat, a.Lng
	o.AddressID = a.ID
	o.Address = order.AddressSnapshot{
		ContactName: a.ContactName, Phone: a.Phone, Detail: a.Detail, Tag: a.Tag,
		Lat: a.Lat, Lng: a.Lng,
	}
	o.User = order.UserSnapshot{Name: a.ContactName, Phone: a.Phone, Address: a.Detail}
	o.Route = m.Name + " → " + a.Detail
	o.Goods = goods
	o.Fees = []order.OrderFee{
		{Label: "商品金额", Value: pr.Goods},
		{Label: "打包费", Value: pr.Packing},
		{Label: "配送费", Value: pr.Delivery},
	}
	if pr.Discount > 0 {
		o.Fees = append(o.Fees, order.OrderFee{Label: "优惠", Value: -pr.Discount, Negative: true})
	}
	o.EstimatedArrival = now + int64(estimateMinutes(distance))*60
	return nil
}

func (s *OrderService) buildErrandOrder(o *order.Order, in CreateOrderInput, u *identity.User, region *tenant.Region, now int64) error {
	if in.Pickup == nil || in.Dropoff == nil {
		return errcode.New(errcode.CodeBadRequest, "跑腿订单需填写取件点与送达点")
	}
	for name, p := range map[string]*PlaceInput{"取件点": in.Pickup, "送达点": in.Dropoff} {
		if err := validatePlace(p); err != nil {
			return errcode.New(errcode.CodeBadRequest, name+err.Error())
		}
	}
	if !validCoord(in.Pickup.Lat, in.Pickup.Lng) || !validCoord(in.Dropoff.Lat, in.Dropoff.Lng) {
		return errcode.New(errcode.CodeBadRequest, "跑腿订单需提供取件/送达坐标")
	}
	distance := haversine(in.Pickup.Lat, in.Pickup.Lng, in.Dropoff.Lat, in.Dropoff.Lng)
	if region.RadiusKM > 0 && distance > region.RadiusKM*2 {
		return errcode.New(errcode.CodeDeliveryOutRange, "超出跑腿服务范围")
	}
	fee := region.StartFee
	if distance > region.FreeKM {
		fee += (distance - region.FreeKM) * extraKMFee
	}
	pr := pricing{Delivery: round2(fee), Pay: round2(fee)}

	o.MerchantID = ""
	o.Merchant = order.MerchantSnapshot{
		Name: in.Pickup.ContactName, Address: in.Pickup.Address, Phone: in.Pickup.Phone,
	}
	o.Amount = pr.Pay
	o.PayAmount = pr.Pay
	o.DeliveryFee = pr.Delivery
	o.PackingFee = 0
	o.DiscountAmount = 0
	o.DeliveryDistanceKM = round2(distance)
	o.PickupLat, o.PickupLng = in.Pickup.Lat, in.Pickup.Lng
	o.DeliveryLat, o.DeliveryLng = in.Dropoff.Lat, in.Dropoff.Lng
	o.Address = order.AddressSnapshot{
		ContactName: in.Dropoff.ContactName, Phone: in.Dropoff.Phone,
		Detail: in.Dropoff.Address, Lat: in.Dropoff.Lat, Lng: in.Dropoff.Lng,
	}
	o.User = order.UserSnapshot{Name: in.Dropoff.ContactName, Phone: in.Dropoff.Phone, Address: in.Dropoff.Address}
	o.Route = in.Pickup.Address + " → " + in.Dropoff.Address
	desc := strings.TrimSpace(in.GoodsDesc)
	if desc == "" {
		desc = "跑腿物品"
	}
	o.Goods = []order.OrderGoods{{Name: desc, Price: pr.Pay, Count: 1, Amount: pr.Pay}}
	o.Fees = []order.OrderFee{{Label: "跑腿费", Value: pr.Delivery}}
	o.EstimatedArrival = now + int64(estimateMinutes(distance)+10)*60
	_ = u
	return nil
}

// applyCoupon 校验并应用优惠券（下单时锁定，取消/退款释放）。
func (s *OrderService) applyCoupon(ctx context.Context, o *order.Order, raw string, now int64) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	c, err := s.Repos.Coupons.FindByID(ctx, raw)
	if err != nil {
		return "", errcode.New(errcode.CodeCouponInvalid, "优惠券不存在")
	}
	if c.UserID != o.UserID || c.Status != coupon.StatusUnused || c.ExpiresAt <= now {
		return "", errcode.New(errcode.CodeCouponInvalid, "优惠券不可用或已过期")
	}
	if c.MerchantID != "" && c.MerchantID != o.MerchantID {
		return "", errcode.New(errcode.CodeCouponInvalid, "该优惠券不适用于此商家")
	}
	if c.MinSpend > 0 && o.Amount < c.MinSpend {
		return "", errcode.New(errcode.CodeCouponInvalid, "未达到优惠券使用门槛")
	}
	discount := c.Amount
	if discount > o.PayAmount {
		discount = o.PayAmount
	}
	discount = round2(discount)
	if discount <= 0 {
		return "", errcode.New(errcode.CodeCouponInvalid, "优惠券金额无效")
	}
	o.CouponID = c.ID
	o.CouponAmount = discount
	o.DiscountAmount = round2(o.DiscountAmount + discount)
	o.PayAmount = round2(math.Max(0, o.PayAmount-discount))
	o.Fees = append(o.Fees, order.OrderFee{Label: "优惠券", Value: -discount, Negative: true})
	return c.ID, nil
}

func isErrandType(t string) bool {
	switch order.OrderType(t) {
	case order.OrderTypeBuy, order.OrderTypeSend, order.OrderTypePick, order.OrderTypeErrand:
		return true
	}
	return false
}

func validatePlace(p *PlaceInput) error {
	p.ContactName = strings.TrimSpace(p.ContactName)
	p.Phone = strings.TrimSpace(p.Phone)
	p.Address = strings.TrimSpace(p.Address)
	if n := utf8.RuneCountInString(p.ContactName); n < 1 || n > 32 {
		return fmt.Errorf("联系人需为 1-32 字")
	}
	if !identity.ValidatePhone(p.Phone) {
		return fmt.Errorf("手机号格式不正确")
	}
	if n := utf8.RuneCountInString(p.Address); n < 2 || n > 120 {
		return fmt.Errorf("地址需为 2-120 字")
	}
	return nil
}

// newOrderNo 生成用户可见单号：时间戳 + ULID 后缀（20 位，唯一索引兜底）。
func newOrderNo(id string, now int64) string {
	suffix := id
	if len(suffix) > 6 {
		suffix = suffix[len(suffix)-6:]
	}
	return time.Unix(now, 0).Format("20060102150405") + strings.ToUpper(suffix)
}

func initialSteps(now int64) []order.OrderStep {
	return []order.OrderStep{
		{Title: "已下单", Time: now, State: "done", Sub: "等待支付"},
		{Title: "商家接单", Time: 0, State: "pending"},
		{Title: "骑手取餐", Time: 0, State: "pending"},
		{Title: "送达完成", Time: 0, State: "pending"},
	}
}

func (s *OrderService) record(ctx context.Context, tx *application.RepoUnit, o *order.Order, from, to order.OrderStatus, actor order.ActorType, actorID, note string, now int64) error {
	return tx.OrderEvents.Save(ctx, &order.Event{
		OrderID: o.ID, FromStatus: from, ToStatus: to,
		ActorType: actor, ActorID: actorID, Note: note, CreatedAt: now,
	})
}

// notify 写站内通知（失败不影响主流程）。
func (s *OrderService) notify(ctx context.Context, tx *application.RepoUnit, userID, title, content, typ, refID string, now int64) {
	if userID == "" {
		return
	}
	_ = tx.Notifications.Save(ctx, &notification.Notification{
		UserID: userID, Title: title, Content: content, Type: typ, RefID: refID, CreatedAt: now,
	})
}

// notifyMerchant 给订单所属商家账号发通知。
func (s *OrderService) notifyMerchant(ctx context.Context, tx *application.RepoUnit, o *order.Order, title, content string, now int64) {
	if o.MerchantID == "" {
		return
	}
	m, err := tx.Merchants.FindByID(ctx, o.MerchantID)
	if err != nil {
		return
	}
	s.notify(ctx, tx, m.OwnerUserID, title, content, "order", o.ID, now)
}

func (s *OrderService) load(ctx context.Context, orderID string) (*order.Order, error) {
	o, err := s.Repos.Orders.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
		}
		return nil, err
	}
	return o, nil
}

// CancelSystem 系统超时取消（仅待支付）。
func (s *OrderService) CancelSystem(ctx context.Context, orderID, reason string) (*order.Order, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != order.StatusPendingPay {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可超时取消")
	}
	return s.cancel(ctx, o, order.ActorSystem, "system", reason)
}

// AutoConfirm 系统自动确认收货（已送达 → 待评价）。
func (s *OrderService) AutoConfirm(ctx context.Context, orderID string) (*order.Order, error) {
	o, err := s.load(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != order.StatusDelivered {
		return nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可自动确认")
	}
	now := time.Now().Unix()
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.StatusDelivered, order.StatusPendingReview,
			order.StatusPatch{ConfirmedAt: now}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单状态已变化")
		}
		return s.record(ctx, tx, o, order.StatusDelivered, order.StatusPendingReview, order.ActorSystem, "system", "超时自动确认", now)
	})
	if err != nil {
		return nil, err
	}
	return s.load(ctx, orderID)
}

// cancel 取消订单：释放库存、余额退款、状态流转、事件。
func (s *OrderService) cancel(ctx context.Context, o *order.Order, actor order.ActorType, actorID, reason string) (*order.Order, error) {
	now := time.Now().Unix()
	err := s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if o.CouponID != "" {
			if err := tx.Coupons.ReleaseByOrder(ctx, o.ID, now); err != nil {
				return err
			}
		}
		if o.Type == order.OrderTypeFood {
			for _, it := range o.Goods {
				if it.DishID == "" {
					continue
				}
				if _, err := tx.Catalog.AdjustStock(ctx, it.DishID, it.Count); err != nil {
					return err
				}
			}
		}
		if o.PaymentStatus == "paid" {
			if o.PayMethod == "balance" {
				if err := creditBalance(ctx, tx, o.UserID, o.PayAmount, now); err != nil {
					return err
				}
			}
			if p, err := tx.Payments.FindByOrder(ctx, o.ID); err == nil && p.Status == order.PaymentPaid {
				if err := tx.Payments.UpdateStatus(ctx, p.ID, order.PaymentRefunded, "", now); err != nil {
					return err
				}
			}
		}
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, o.Status, order.OrderCanceled, order.StatusPatch{
			CanceledAt: now, CancelReason: reason, PaymentStatus: refundedStatus(o),
		}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单状态已变化，请刷新后重试")
		}
		if err := s.record(ctx, tx, o, o.Status, order.OrderCanceled, actor, actorID, reason, now); err != nil {
			return err
		}
		if actor == order.ActorClient {
			s.notifyMerchant(ctx, tx, o, "用户取消订单", "订单 "+o.OrderNo+" 已被用户取消", now)
		} else {
			s.notify(ctx, tx, o.UserID, "订单已取消", "订单 "+o.OrderNo+" 已取消："+reason, "order", o.ID, now)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.load(ctx, o.ID)
}

func refundedStatus(o *order.Order) string {
	if o.PaymentStatus == "paid" {
		return "refunded"
	}
	return o.PaymentStatus
}

// deductBalance 原子扣减余额：单条件 UPDATE，禁止 Find-then-Save（并发双花）。
func deductBalance(ctx context.Context, tx *application.RepoUnit, userID string, amount float64, now int64) error {
	if _, err := tx.Profiles.FindByUserID(ctx, userID); err != nil {
		return errcode.New(errcode.CodeInsufficientFund, "余额不足，请更换支付方式")
	}
	affected, err := tx.Profiles.AddBalance(ctx, userID, -round2(amount), round2(amount), now)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errcode.New(errcode.CodeInsufficientFund, "余额不足，请更换支付方式")
	}
	return nil
}

// creditBalance 原子入账余额。
func creditBalance(ctx context.Context, tx *application.RepoUnit, userID string, amount float64, now int64) error {
	if _, err := tx.Profiles.FindByUserID(ctx, userID); err != nil {
		return err
	}
	_, err := tx.Profiles.AddBalance(ctx, userID, round2(amount), -1, now)
	return err
}

// randomCode 生成 4 位取送校验码（加密随机）。
func randomCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "0000"
	}
	return fmt.Sprintf("%04d", n.Int64())
}
