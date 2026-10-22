package http

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"dawn-errand/internal/domain/cms"
	"dawn-errand/internal/domain/complaint"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/userprofile"
)

// ---- 通用 ----

// thousandSep 千分位格式化整数。
func thousandSep(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b []byte
	for i, r := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			b = append(b, ',')
		}
		b = append(b, r)
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

func moneyStr(v float64) string {
	return "¥" + thousandSep(int64(v+0.5))
}

func pctStr(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64) + "%"
}

func timeHM(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("15:04")
}

func timeMD(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("01/02")
}

func timeYMD(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("2006-01-02")
}

func deltaStr(v float64) string {
	if v >= 0 {
		return "▲ " + strconv.FormatFloat(v, 'f', 1, 64) + "%"
	}
	return "▼ " + strconv.FormatFloat(-v, 'f', 1, 64) + "%"
}

// hashIdx 稳定哈希取模（用于颜色/图标选择）。
func hashIdx(s string, n int) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return int(h.Sum32() % uint32(n))
}

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

// ---- 商家 ----

func merchantLogoIcon(category string) string {
	switch {
	case strings.Contains(category, "药"):
		return "pharmacy"
	case strings.Contains(category, "商超"), strings.Contains(category, "超市"), strings.Contains(category, "便利"):
		return "shop"
	case strings.Contains(category, "餐"), strings.Contains(category, "菜"), strings.Contains(category, "食"):
		return "food"
	}
	return "other"
}

func categoryColors(category string) (string, string) {
	switch merchantLogoIcon(category) {
	case "pharmacy":
		return "#E8F7F0", "#00B578"
	case "shop":
		return "#FFF8E0", "#D48806"
	case "food":
		return "#FFF3E8", "#FF6B2C"
	}
	return "#E8F0FF", "#2B7FFF"
}

func merchantLogoColors(m *merchant.Merchant) (string, string, string) {
	bg, color := categoryColors(m.Category)
	return bg, color, merchantLogoIcon(m.Category)
}

func merchantStatusText(s merchant.MerchantStatus) string {
	switch s {
	case merchant.MerchantStatusOpen:
		return "营业中"
	case merchant.MerchantStatusRest:
		return "休息中"
	case merchant.MerchantStatusBanned:
		return "已封禁"
	}
	return string(s)
}

func merchantAuditText(s merchant.AuditStatus) string {
	switch s {
	case merchant.AuditApproved:
		return "已通过"
	case merchant.AuditRejected:
		return "已驳回"
	}
	return "待审核"
}

// MerchantRowDto 商家列表行（对齐前端 MerchantRow）。
type MerchantRowDto struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	LogoBg        string   `json:"logoBg"`
	LogoColor     string   `json:"logoColor"`
	LogoIcon      string   `json:"logoIcon"`
	Address       string   `json:"address"`
	AddressSub    string   `json:"addressSub"`
	Contact       string   `json:"contact"`
	Phone         string   `json:"phone"`
	Category      string   `json:"category"`
	CategoryBg    string   `json:"categoryBg"`
	CategoryColor string   `json:"categoryColor"`
	Status        string   `json:"status"`
	MonthlyOrders *int64   `json:"monthlyOrders"`
	Rating        *float64 `json:"rating"`
}

func toMerchantRow(m *merchant.Merchant) MerchantRowDto {
	bg, color, icon := merchantLogoColors(m)
	cbg, ccol := categoryColors(m.Category)
	row := MerchantRowDto{
		ID: m.ID, Name: m.Name, LogoBg: bg, LogoColor: color, LogoIcon: icon,
		Address: m.Address, AddressSub: m.AddressSub, Contact: m.Contact, Phone: m.Phone,
		Category: m.Category, CategoryBg: cbg, CategoryColor: ccol,
		Status: merchantStatusText(m.Status),
	}
	if m.Status != merchant.MerchantStatusBanned {
		mo := m.MonthlyOrders
		ra := m.Rating
		row.MonthlyOrders = &mo
		row.Rating = &ra
	}
	return row
}

// MerchantStatsDto 商家顶部聚合（对齐前端 MerchantStats）。
type MerchantStatsDto struct {
	Total        int64  `json:"total"`
	WeekNew      int64  `json:"weekNew"`
	Open         int64  `json:"open"`
	OpenRate     string `json:"openRate"`
	Pending      int64  `json:"pending"`
	Closed       int64  `json:"closed"`
	ClosedDetail string `json:"closedDetail"`
}

// MerchantAuditRowDto 入驻申请行（对齐前端 MerchantAuditRow）。
type MerchantAuditRowDto struct {
	ID        string     `json:"id"`
	ApplyNo   string     `json:"applyNo"`
	Name      string     `json:"name"`
	Operator  string     `json:"operator"`
	Type      string     `json:"type"`
	Address   string     `json:"address"`
	AppliedAt int64      `json:"appliedAt"`
	LogoBg    string     `json:"logoBg"`
	LogoColor string     `json:"logoColor"`
	LogoIcon  string     `json:"logoIcon"`
	Certs     []CertItem `json:"certs"`
	Status    string     `json:"status"`
}

type CertItem struct {
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

func merchantCerts(m *merchant.Merchant) []CertItem {
	out := []CertItem{{Label: "营业执照", Icon: "license"}}
	switch merchantLogoIcon(m.Category) {
	case "pharmacy":
		out = append(out, CertItem{Label: "药品经营许可证", Icon: "pharmacylicense"})
	case "food":
		out = append(out, CertItem{Label: "食品经营许可证", Icon: "foodlicense"})
	}
	out = append(out, CertItem{Label: "门店照片", Icon: "storefront"})
	return out
}

func toMerchantAuditRow(m *merchant.Merchant) MerchantAuditRowDto {
	bg, color, icon := merchantLogoColors(m)
	return MerchantAuditRowDto{
		ID: m.ID, ApplyNo: m.ApplyNo, Name: m.Name, Operator: m.Operator,
		Type: m.Category, Address: m.Address, AppliedAt: m.AppliedAt,
		LogoBg: bg, LogoColor: color, LogoIcon: icon,
		Certs: merchantCerts(m), Status: merchantAuditText(m.AuditStatus),
	}
}

// MerchantDetailDto 商家详情（对齐前端 MerchantDetail）。
type MerchantDetailDto struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Status         string           `json:"status"`
	Category       string           `json:"category"`
	CategoryBg     string           `json:"categoryBg"`
	CategoryColor  string           `json:"categoryColor"`
	JoinedAt       int64            `json:"joinedAt"`
	Rating         float64          `json:"rating"`
	MonthlyOrders  int64            `json:"monthlyOrders"`
	MonthlyRevenue float64          `json:"monthlyRevenue"`
	AvgOrder       float64          `json:"avgOrder"`
	GoodRate       string           `json:"goodRate"`
	LogoBg         string           `json:"logoBg"`
	LogoColor      string           `json:"logoColor"`
	Basic          MerchantBasicDto `json:"basic"`
	LastUpdatedAt  int64            `json:"lastUpdatedAt"`
	Completeness   int              `json:"completeness"`
	Today          MerchantTodayDto `json:"today"`
	ReviewCount    int              `json:"reviewCount"`
	LastOperation  string           `json:"lastOperation"`
}

type MerchantBasicDto struct {
	Name           string `json:"name"`
	Owner          string `json:"owner"`
	Phone          string `json:"phone"`
	Type           string `json:"type"`
	Address        string `json:"address"`
	OpenHours      string `json:"openHours"`
	DeliveryRange  string `json:"deliveryRange"`
	Commission     string `json:"commission"`
	BankAccount    string `json:"bankAccount"`
	DeliveryMethod string `json:"deliveryMethod"`
}

// MerchantTodayDto 商家今日速览。
type MerchantTodayDto struct {
	Orders          int64   `json:"orders"`
	OrdersDelta     string  `json:"ordersDelta"`
	Revenue         float64 `json:"revenue"`
	RevenueDelta    string  `json:"revenueDelta"`
	OnTimeRate      string  `json:"onTimeRate"`
	OnTimeTarget    string  `json:"onTimeTarget"`
	PendingAbnormal int64   `json:"pendingAbnormal"`
}

func toMerchantDetail(m *merchant.Merchant) MerchantDetailDto {
	bg, color, _ := merchantLogoColors(m)
	cbg, ccol := categoryColors(m.Category)
	commission := strconv.FormatFloat(m.Commission, 'f', 1, 64) + "%"
	return MerchantDetailDto{
		ID: m.ID, Name: m.Name, Status: merchantStatusText(m.Status),
		Category: m.Category, CategoryBg: cbg, CategoryColor: ccol,
		JoinedAt: m.JoinedAt, Rating: m.Rating, MonthlyOrders: m.MonthlyOrders,
		MonthlyRevenue: m.MonthlyRevenue, AvgOrder: m.AvgOrder,
		GoodRate: pctStr(m.GoodRate), LogoBg: bg, LogoColor: color,
		Basic: MerchantBasicDto{
			Name: m.Name, Owner: m.Contact, Phone: m.Phone, Type: m.Category,
			Address: m.Address, OpenHours: m.OpenHours, DeliveryRange: m.DeliveryRange,
			Commission: commission, BankAccount: m.BankAccount, DeliveryMethod: m.DeliveryMethod,
		},
		LastUpdatedAt: m.LastUpdatedAt, Completeness: m.Completeness,
		Today: MerchantTodayDto{
			Orders: m.TodayOrders, Revenue: m.TodayRevenue,
			OnTimeRate: pctStr(m.OnTimeRate), OnTimeTarget: "95%",
			PendingAbnormal: int64(m.PendingAbnormal),
		},
		ReviewCount: m.ReviewCount, LastOperation: m.LastOperation,
	}
}

// ---- 骑手 ----

func riderAvatarColors(r *rider.Rider) (string, string) {
	palette := [][2]string{
		{"#E8F0FF", "#2B7FFF"}, {"#FFF3E8", "#FF6B2C"}, {"#E8F7F0", "#00B578"},
		{"#FFF8E0", "#D48806"}, {"#F3E8FF", "#722ED1"}, {"#FFE8E8", "#FF4D4F"},
	}
	return palette[hashIdx(r.Name, len(palette))][0], palette[hashIdx(r.Name, len(palette))][1]
}

func riderStatusText(s rider.RiderStatus) string {
	switch s {
	case rider.RiderOnline:
		return "在线"
	case rider.RiderBusy:
		return "忙碌"
	case rider.RiderFrozen:
		return "已冻结"
	}
	return "离线"
}

func riderAuditText(s rider.AuditStatus) string {
	switch s {
	case rider.AuditApproved:
		return "已通过"
	case rider.AuditRejected:
		return "已驳回"
	}
	return "待审核"
}

// RiderRowDto 骑手列表行（对齐前端 RiderRow）。
type RiderRowDto struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Phone       string   `json:"phone"`
	Area        string   `json:"area"`
	Tool        string   `json:"tool"`
	Level       string   `json:"level"`
	Status      string   `json:"status"`
	TodayOrders *int64   `json:"todayOrders"`
	Rating      *float64 `json:"rating"`
	AvatarBg    string   `json:"avatarBg"`
	AvatarColor string   `json:"avatarColor"`
}

func toRiderRow(r *rider.Rider) RiderRowDto {
	bg, color := riderAvatarColors(r)
	row := RiderRowDto{
		ID: r.ID, Name: r.Name, Phone: r.Phone, Area: r.Area, Tool: r.Tool,
		Level: r.Level, Status: riderStatusText(r.Status),
		AvatarBg: bg, AvatarColor: color,
	}
	if r.Status != rider.RiderFrozen {
		to, ra := r.TodayOrders, r.Rating
		row.TodayOrders = &to
		row.Rating = &ra
	}
	return row
}

// RiderStatsDto 骑手顶部聚合（对齐前端 RiderStatsDto）。
type RiderStatsDto struct {
	Total        int64              `json:"total"`
	Online       int64              `json:"online"`
	OnlineRate   string             `json:"onlineRate"`
	PendingAudit int64              `json:"pendingAudit"`
	Frozen       int64              `json:"frozen"`
	FrozenNote   string             `json:"frozenNote"`
	WeekDelta    string             `json:"weekDelta"`
	Audit        RiderAuditStatsDto `json:"audit"`
}

type RiderAuditStatsDto struct {
	TodayNew      int64  `json:"todayNew"`
	Approved      int64  `json:"approved"`
	Rejected      int64  `json:"rejected"`
	MonthApproved int64  `json:"monthApproved"`
	MomRate       string `json:"momRate"`
}

// RiderAuditRowDto 入驻申请行（对齐前端 RiderAuditRow）。
type RiderAuditRowDto struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	AvatarBg    string `json:"avatarBg"`
	AvatarColor string `json:"avatarColor"`
	ApplyNo     string `json:"applyNo"`
	SubmittedAt int64  `json:"submittedAt"`
	Area        string `json:"area"`
	Phone       string `json:"phone"`
	Tool        string `json:"tool"`
	OrderTypes  string `json:"orderTypes"`
	IDCheck     string `json:"idCheck"`
	NoticeType  string `json:"noticeType"`
	Notice      string `json:"notice"`
	DocCount    int    `json:"docCount"`
	Docs        string `json:"docs"`
}

func toRiderAuditRow(r *rider.Rider) RiderAuditRowDto {
	bg, color := riderAvatarColors(r)
	noticeType := "none"
	if r.Notice != "" {
		noticeType = "warning"
	}
	return RiderAuditRowDto{
		ID: r.ID, Status: riderAuditText(r.AuditStatus), Name: r.Name,
		AvatarBg: bg, AvatarColor: color, ApplyNo: r.ApplyNo, SubmittedAt: r.SubmittedAt,
		Area: r.Area, Phone: r.Phone, Tool: r.Tool, OrderTypes: r.OrderTypes,
		IDCheck: r.IDCheck, NoticeType: noticeType, Notice: r.Notice,
		DocCount: r.DocCount, Docs: r.Docs,
	}
}

// RiderDetailDto 骑手详情（对齐前端 RiderDetailDto）。
type RiderDetailDto struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Status        string           `json:"status"`
	ServiceType   string           `json:"serviceType"`
	JoinedAt      int64            `json:"joinedAt"`
	Rating        float64          `json:"rating"`
	TodayOrders   int64            `json:"todayOrders"`
	OnlineHours   float64          `json:"onlineHours"`
	MonthIncome   float64          `json:"monthIncome"`
	OnTimeRate    string           `json:"onTimeRate"`
	AvatarBg      string           `json:"avatarBg"`
	Profile       RiderProfileDto  `json:"profile"`
	TodayOverview RiderOverviewDto `json:"todayOverview"`
	LastAction    string           `json:"lastAction"`
	LastActionAt  int64            `json:"lastActionAt"`
}

type RiderProfileDto struct {
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	IDCard        string `json:"idCard"`
	Area          string `json:"area"`
	Vehicle       string `json:"vehicle"`
	OrderTypes    string `json:"orderTypes"`
	Account       string `json:"account"`
	JoinedAt      string `json:"joinedAt"`
	ServiceMonths int    `json:"serviceMonths"`
}

type RiderOverviewDto struct {
	TodayOrders      int64  `json:"todayOrders"`
	TodayOrdersDelta string `json:"todayOrdersDelta"`
	OnTimeRate       string `json:"onTimeRate"`
	OnTimeTarget     string `json:"onTimeTarget"`
	AvgMinutes       string `json:"avgMinutes"`
	PlatformAvg      string `json:"platformAvg"`
	Complaints       int    `json:"complaints"`
	ComplaintNote    string `json:"complaintNote"`
}

func toRiderDetail(r *rider.Rider) RiderDetailDto {
	bg, _ := riderAvatarColors(r)
	return RiderDetailDto{
		ID: r.ID, Name: r.Name, Status: riderStatusText(r.Status),
		ServiceType: r.ServiceType, JoinedAt: r.JoinedAt, Rating: r.Rating,
		TodayOrders: r.TodayOrders, OnlineHours: r.OnlineHours, MonthIncome: r.MonthIncome,
		OnTimeRate: pctStr(r.OnTimeRate), AvatarBg: bg,
		Profile: RiderProfileDto{
			Name: r.Name, Phone: r.Phone, IDCard: r.IDCard, Area: r.Area,
			Vehicle: r.Vehicle, OrderTypes: r.OrderTypes, Account: r.Account,
			JoinedAt: timeYMD(r.JoinedAt), ServiceMonths: r.ServiceMonths,
		},
		TodayOverview: RiderOverviewDto{
			TodayOrders: r.TodayOrders, TodayOrdersDelta: r.TodayOrdersDelta,
			OnTimeRate: pctStr(r.OnTimeRate), OnTimeTarget: "95%",
			AvgMinutes:  strconv.FormatFloat(r.AvgMinutes, 'f', 1, 64),
			PlatformAvg: "32.0", Complaints: r.ComplaintsToday,
			ComplaintNote: "今日无投诉",
		},
		LastAction: r.LastAction, LastActionAt: r.LastActionAt,
	}
}

// ---- 用户 ----

func userAvatarColors(u *identity.User) (string, string) {
	palette := [][2]string{
		{"#E8F0FF", "#2B7FFF"}, {"#FFF3E8", "#FF6B2C"}, {"#E8F7F0", "#00B578"},
		{"#FFF8E0", "#D48806"}, {"#F3E8FF", "#722ED1"}, {"#FFE8E8", "#FF4D4F"},
	}
	return palette[hashIdx(u.ID+u.Name, len(palette))][0], palette[hashIdx(u.ID+u.Name, len(palette))][1]
}

func userStatusText(s identity.UserStatus) string {
	if s == identity.UserStatusBanned {
		return "已封禁"
	}
	return "正常"
}

// UserRowDto 用户列表行（对齐前端 UserRow）。
type UserRowDto struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AvatarBg     string `json:"avatarBg"`
	AvatarColor  string `json:"avatarColor"`
	Phone        string `json:"phone"`
	Area         string `json:"area"`
	OrderCount   int64  `json:"orderCount"`
	Status       string `json:"status"`
	RegisteredAt int64  `json:"registeredAt"`
}

func toUserRow(u *identity.User) UserRowDto {
	bg, color := userAvatarColors(u)
	area := "王家湾镇"
	if u.RegionID != "" {
		area = regionNameFallback(u.RegionID)
	}
	return UserRowDto{
		ID: u.ID, Name: displayName(u), AvatarBg: bg, AvatarColor: color,
		Phone: formatPhone(u.Phone), Area: area, Status: userStatusText(u.Status),
		RegisteredAt: u.RegisteredAt,
	}
}

func displayName(u *identity.User) string {
	if u.Name != "" {
		return u.Name
	}
	if u.Email != "" {
		return u.Email[:strings.Index(u.Email, "@")]
	}
	return u.ID
}

func formatPhone(p string) string {
	if len(p) == 11 {
		return p[:3] + "-" + p[3:7] + "-" + p[7:]
	}
	return p
}

func regionNameFallback(id string) string {
	return "区域 " + id[:minInt(len(id), 6)]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UsersOverviewDto 用户概览（对齐前端 UsersOverview）。
type UsersOverviewDto struct {
	Total       int64   `json:"total"`
	TodayNew    int64   `json:"todayNew"`
	ActiveMonth int64   `json:"activeMonth"`
	ActiveRate  float64 `json:"activeRate"`
	AvgOrders   float64 `json:"avgOrders"`
	Banned      int64   `json:"banned"`
}

// UserDetailDto 用户详情（对齐前端 UserDetail）。
type UserDetailDto struct {
	ID               string               `json:"id"`
	UserNo           string               `json:"userNo"`
	Name             string               `json:"name"`
	AvatarBg         string               `json:"avatarBg"`
	AvatarColor      string               `json:"avatarColor"`
	Level            string               `json:"level"`
	Status           string               `json:"status"`
	Phone            string               `json:"phone"`
	RegisteredAt     int64                `json:"registeredAt"`
	TotalOrders      int64                `json:"totalOrders"`
	TotalSpent       float64              `json:"totalSpent"`
	MonthOrders      int64                `json:"monthOrders"`
	Verified         string               `json:"verified"`
	Nickname         string               `json:"nickname"`
	RealName         string               `json:"realName"`
	Region           string               `json:"region"`
	Address          string               `json:"address"`
	RegisterChannel  string               `json:"registerChannel"`
	BizTypes         string               `json:"bizTypes"`
	Balance          float64              `json:"balance"`
	RedPacketBalance float64              `json:"redPacketBalance"`
	RecentActiveAt   int64                `json:"recentActiveAt"`
	OrderRecordCount int64                `json:"orderRecordCount"`
	ComplaintCount   int64                `json:"complaintCount"`
	RecentOrders     []UserRecentOrderDto `json:"recentOrders"`
}

type UserRecentOrderDto struct {
	ID        string  `json:"id"`
	BizType   string  `json:"bizType"`
	BizBg     string  `json:"bizBg"`
	BizColor  string  `json:"bizColor"`
	Merchant  string  `json:"merchant"`
	Amount    float64 `json:"amount"`
	Rider     string  `json:"rider"`
	Status    string  `json:"status"`
	OrderedAt int64   `json:"orderedAt"`
}

func toUserDetail(u *identity.User, p *userprofile.Profile, totalOrders, monthOrders int64, totalSpent float64, complaintCount int64, recent []*order.Order) UserDetailDto {
	bg, color := userAvatarColors(u)
	level := "Lv.1 新用户"
	verified := "未实名"
	nickname, realName, region, address, channel, bizTypes := "", "", "王家湾镇", "", "微信", "外卖"
	var balance, redPacket float64
	if p != nil {
		if p.Level != "" {
			level = p.Level
		}
		if p.Verified != "" {
			verified = p.Verified
		}
		nickname, realName, region, address = p.Nickname, p.RealName, p.Region, p.Address
		if p.RegisterChannel != "" {
			channel = p.RegisterChannel
		}
		if p.BizTypes != "" {
			bizTypes = p.BizTypes
		}
		balance, redPacket = p.Balance, p.RedPacketBalance
	}
	recentOrders := make([]UserRecentOrderDto, 0, len(recent))
	for _, o := range recent {
		bz := string(o.Type)
		bzbg, bzcolor := orderTypeColors(o.Type)
		riderText := "待分配"
		if o.RiderName != "" {
			riderText = o.RiderName
		}
		merchantText := o.Route
		recentOrders = append(recentOrders, UserRecentOrderDto{
			ID: o.ID, BizType: bz, BizBg: bzbg, BizColor: bzcolor,
			Merchant: merchantText, Amount: o.Amount, Rider: riderText,
			Status: string(o.Status), OrderedAt: o.CreatedAt,
		})
	}
	return UserDetailDto{
		ID: u.ID, UserNo: "U" + u.ID[len(u.ID)-6:], Name: displayName(u),
		AvatarBg: bg, AvatarColor: color, Level: level, Status: userStatusText(u.Status),
		Phone: formatPhone(u.Phone), RegisteredAt: u.RegisteredAt,
		TotalOrders: totalOrders, TotalSpent: totalSpent, MonthOrders: monthOrders,
		Verified: verified, Nickname: nickname, RealName: realName, Region: region,
		Address: address, RegisterChannel: channel, BizTypes: bizTypes,
		Balance: balance, RedPacketBalance: redPacket,
		RecentActiveAt:   u.UpdatedAt,
		OrderRecordCount: totalOrders, ComplaintCount: complaintCount,
		RecentOrders: recentOrders,
	}
}

// ---- 财务 ----

func financeKindColors(kind string) (string, string) {
	switch {
	case strings.Contains(kind, "骑手"):
		return "#E8F0FF", "#2B7FFF"
	case strings.Contains(kind, "商家"):
		return "#FFF3E8", "#FF6B2C"
	}
	return "#FFE8E8", "#FF4D4F"
}

func withdrawStatusText(s finance.WithdrawStatus) string {
	switch s {
	case finance.WithdrawApproved:
		return "已通过"
	case finance.WithdrawRejected:
		return "已驳回"
	}
	return "待审核"
}

// WithdrawRowDto 提现列表行（对齐前端 WithdrawRow）。
type WithdrawRowDto struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	AvatarBg     string  `json:"avatarBg"`
	AvatarColor  string  `json:"avatarColor"`
	Kind         string  `json:"kind"`
	KindBg       string  `json:"kindBg"`
	KindColor    string  `json:"kindColor"`
	BankInfo     string  `json:"bankInfo"`
	RiskTip      string  `json:"riskTip,omitempty"`
	RejectReason string  `json:"rejectReason,omitempty"`
	Amount       float64 `json:"amount"`
	AmountColor  string  `json:"amountColor,omitempty"`
	CreatedAt    int64   `json:"created_at"`
	Status       string  `json:"status"`
}

func toWithdrawRow(w *finance.Withdrawal) WithdrawRowDto {
	kbg, kcol := financeKindColors(w.Kind)
	row := WithdrawRowDto{
		ID: w.ID, Name: w.PartyName, AvatarBg: kbg, AvatarColor: kcol,
		Kind: w.Kind, KindBg: kbg, KindColor: kcol, BankInfo: w.BankInfo,
		Amount: w.Amount, CreatedAt: w.CreatedAt, Status: withdrawStatusText(w.Status),
	}
	if w.RiskWarning != "" {
		row.RiskTip = w.RiskWarning
		row.AmountColor = "#FF4D4F"
	}
	if w.Status == finance.WithdrawRejected {
		row.RejectReason = w.RejectReason
	}
	return row
}

// FinanceStatsDto 财务汇总（对齐前端 FinanceStats）。
type FinanceStatsDto struct {
	TodayIncome             float64 `json:"todayIncome"`
	TodayIncomeDelta        string  `json:"todayIncomeDelta"`
	PendingBalance          float64 `json:"pendingBalance"`
	PendingBalanceBreakdown string  `json:"pendingBalanceBreakdown"`
	WithdrawPendingCount    int64   `json:"withdrawPendingCount"`
	AbnormalBills           int64   `json:"abnormalBills"`
	AbnormalBillsAmount     float64 `json:"abnormalBillsAmount"`
}

// SettlementStatsDto 结算汇总（对齐前端 SettlementStats）。
type SettlementStatsDto struct {
	TodaySettled      float64 `json:"todaySettled"`
	TodaySettledDelta string  `json:"todaySettledDelta"`
	WeekTotal         float64 `json:"weekTotal"`
	WeekBreakdown     string  `json:"weekBreakdown"`
	MonthTotal        float64 `json:"monthTotal"`
	MonthBreakdown    string  `json:"monthBreakdown"`
	PendingAmount     float64 `json:"pendingAmount"`
	PendingCountText  string  `json:"pendingCountText"`
	DateRange         string  `json:"dateRange"`
	ExportedText      string  `json:"exportedText"`
}

// FinanceDto 财务页（对齐前端 FinanceDto）。
type FinanceDto struct {
	Stats           FinanceStatsDto    `json:"stats"`
	SettlementStats SettlementStatsDto `json:"settlementStats"`
	Withdraws       []WithdrawRowDto   `json:"withdraws"`
}

// SettlementRowDto 结算流水行（对齐前端 SettlementRow）。
type SettlementRowDto struct {
	ID        string  `json:"id"`
	Party     string  `json:"party"`
	Type      string  `json:"type"`
	TypeBg    string  `json:"typeBg"`
	TypeColor string  `json:"typeColor"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	Period    string  `json:"period"`
	SettledAt int64   `json:"settled_at"`
	Action    string  `json:"action"`
}

func toSettlementRow(s *finance.Settlement) SettlementRowDto {
	bg, col := financeKindColors(string(s.Type))
	amount := s.Amount
	if s.Type == finance.SettlementRefund {
		amount = -s.Amount
	}
	status := "已到账"
	if s.Status == finance.SettlementProcessing {
		status = "处理中"
	}
	return SettlementRowDto{
		ID: s.ID, Party: s.Party, Type: string(s.Type), TypeBg: bg, TypeColor: col,
		Amount: amount, Status: status, Period: s.Period, SettledAt: s.SettledAt, Action: s.Action,
	}
}

// WithdrawDetailDto 提现详情（对齐前端 WithdrawDetailDto）。
type WithdrawDetailDto struct {
	ID              string          `json:"id"`
	Amount          float64         `json:"amount"`
	Fee             float64         `json:"fee"`
	FeeNote         string          `json:"feeNote"`
	Status          string          `json:"status"`
	Kind            string          `json:"kind"`
	Applicant       string          `json:"applicant"`
	ApplicantNo     string          `json:"applicantNo"`
	CreatedAt       int64           `json:"created_at"`
	WithdrawType    string          `json:"withdrawType"`
	BankInfo        string          `json:"bankInfo"`
	AccountHolder   string          `json:"accountHolder"`
	ArriveExpect    string          `json:"arriveExpect"`
	Purpose         string          `json:"purpose"`
	Balance         float64         `json:"balance"`
	BalanceNote     string          `json:"balanceNote"`
	UsedPercentText string          `json:"usedPercentText"`
	RemainBalance   float64         `json:"remainBalance"`
	RemainNote      string          `json:"remainNote"`
	MonthOrders     int64           `json:"monthOrders"`
	MonthIncome     float64         `json:"monthIncome"`
	LastMonthNote   string          `json:"lastMonthNote"`
	RiskItems       []string        `json:"riskItems"`
	RiskWarning     string          `json:"riskWarning"`
	Timeline        []finance.Event `json:"timeline"`
}

func toWithdrawDetail(w *finance.Withdrawal) WithdrawDetailDto {
	usedPct := "0%"
	if w.Balance > 0 {
		usedPct = pctStr(w.Amount / w.Balance * 100)
	}
	lastMonthNote := "上月同期 " + moneyStr(w.MonthIncome)
	if w.MonthOrders > 0 {
		lastMonthNote = fmt.Sprintf("本月 %d 单 · %s", w.MonthOrders, moneyStr(w.MonthIncome))
	}
	return WithdrawDetailDto{
		ID: w.ID, Amount: w.Amount, Fee: w.Fee, FeeNote: w.FeeNote,
		Status: withdrawStatusText(w.Status), Kind: w.Kind, Applicant: w.Applicant,
		ApplicantNo: w.PartyNo, CreatedAt: w.CreatedAt, WithdrawType: w.WithdrawType,
		BankInfo: w.BankInfo, AccountHolder: w.AccountHolder, ArriveExpect: w.ArriveExpect,
		Purpose: w.Purpose, Balance: w.Balance,
		BalanceNote: "可提现余额", UsedPercentText: usedPct,
		RemainBalance: w.RemainBalance, RemainNote: "提现后余额",
		MonthOrders: w.MonthOrders, MonthIncome: w.MonthIncome,
		LastMonthNote: lastMonthNote,
		RiskItems:     w.RiskItems, RiskWarning: w.RiskWarning, Timeline: w.Timeline,
	}
}

// ---- 投诉 ----

func complaintCategoryColors(category string) (string, string) {
	switch {
	case strings.Contains(category, "配送"):
		return "#FFF3E8", "#FF6B2C"
	case strings.Contains(category, "商家"):
		return "#FFF8E0", "#D48806"
	}
	return "#E8F0FF", "#2B7FFF"
}

// ComplaintRowDto 投诉工单行（对齐前端 ComplaintRow）。
type ComplaintRowDto struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	Complainant string `json:"complainant"`
	Target      string `json:"target"`
	Extra       string `json:"extra"`
	CreatedAt   int64  `json:"created_at"`
	Description string `json:"description,omitempty"`
	DescTone    string `json:"desc_tone,omitempty"`
	Deadline    string `json:"deadline,omitempty"`
	Handler     string `json:"handler,omitempty"`
	Progress    string `json:"progress,omitempty"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
}

func toComplaintRow(c *complaint.Complaint) ComplaintRowDto {
	row := ComplaintRowDto{
		ID: c.ID, Category: c.Category, Title: c.Title, OrderID: c.OrderID,
		Status: string(c.Status), Complainant: c.Complainant, Target: c.Target,
		Extra: c.Extra, CreatedAt: c.CreatedAt,
	}
	if c.Status == complaint.StatusPending {
		row.Description = c.Description
		row.DescTone = c.DescTone
		row.Deadline = c.Deadline
	}
	if c.Status == complaint.StatusProcessing {
		row.Handler = c.Handler
		row.Progress = c.Progress
		row.UpdatedAt = c.UpdatedAt
	}
	return row
}

// ComplaintFlowStepDto 流程节点（对齐前端 ComplaintFlowStep）。
type ComplaintFlowStepDto struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
	State string `json:"state"`
	Step  int    `json:"step,omitempty"`
}

type ComplaintNoteDto struct {
	Avatar      string `json:"avatar"`
	AvatarBg    string `json:"avatar_bg"`
	AvatarColor string `json:"avatar_color"`
	Name        string `json:"name"`
	Time        int64  `json:"time"`
	Content     string `json:"content"`
	Tone        string `json:"tone"`
}

type ComplaintOrderItemDto struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Tone  string `json:"tone,omitempty"`
}

// ComplaintDetailDto 工单详情（对齐前端 ComplaintDetail）。
type ComplaintDetailDto struct {
	ID             string                  `json:"id"`
	Title          string                  `json:"title"`
	Status         string                  `json:"status"`
	Priority       string                  `json:"priority"`
	CreatedAt      int64                   `json:"created_at"`
	Initiator      string                  `json:"initiator"`
	InitiatorExtra string                  `json:"initiator_extra"`
	Target         string                  `json:"target"`
	OrderID        string                  `json:"order_id"`
	Flow           []ComplaintFlowStepDto  `json:"flow"`
	Notes          []ComplaintNoteDto      `json:"notes"`
	Order          []ComplaintOrderItemDto `json:"order"`
	PayAmount      float64                 `json:"pay_amount"`
}

func toComplaintDetail(c *complaint.Complaint) ComplaintDetailDto {
	flow := make([]ComplaintFlowStepDto, 0, len(c.Flow))
	for _, f := range c.Flow {
		flow = append(flow, ComplaintFlowStepDto{Title: f.Title, Desc: f.Desc, State: f.State, Step: f.Step})
	}
	notes := make([]ComplaintNoteDto, 0, len(c.Notes))
	for _, n := range c.Notes {
		notes = append(notes, ComplaintNoteDto{
			Avatar: n.Avatar, AvatarBg: n.AvatarBg, AvatarColor: n.AvatarColor,
			Name: n.Name, Time: n.Time, Content: n.Content, Tone: n.Tone,
		})
	}
	orderItems := make([]ComplaintOrderItemDto, 0, len(c.Order))
	for _, o := range c.Order {
		orderItems = append(orderItems, ComplaintOrderItemDto{Label: o.Label, Value: o.Value, Tone: o.Tone})
	}
	return ComplaintDetailDto{
		ID: c.ID, Title: c.Title, Status: string(c.Status), Priority: c.Priority,
		CreatedAt: c.CreatedAt, Initiator: c.Initiator, InitiatorExtra: c.InitiatorExtra,
		Target: c.Target, OrderID: c.OrderID, Flow: flow, Notes: notes,
		Order: orderItems, PayAmount: c.PayAmount,
	}
}

// ---- 服务区域 / 费率 / 公告 ----

func regionStatusText(s tenant.RegionStatus) string {
	if s == tenant.RegionStatusOpen {
		return "运营中"
	}
	return "筹备中"
}

// ServiceAreaDto 服务区域（对齐前端 ServiceArea）。
type ServiceAreaDto struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	Villages        int     `json:"villages"`
	RadiusKM        float64 `json:"radius_km"`
	OnlineMerchants *int64  `json:"online_merchants"`
	ActiveRiders    *int64  `json:"active_riders"`
	TodayOrders     *int64  `json:"today_orders"`
	StartFee        float64 `json:"start_fee"`
	FreeKM          float64 `json:"free_km"`
	Notice          string  `json:"notice"`
	Action          string  `json:"action"`
	IconBg          string  `json:"icon_bg"`
	IconColor       string  `json:"icon_color"`
}

func toServiceArea(r *tenant.Region, onlineMerchants, activeRiders, todayOrders *int64) ServiceAreaDto {
	notice := r.Notice
	if r.Status == tenant.RegionStatusOpen {
		notice = ""
	}
	action := "停用"
	if r.Status == tenant.RegionStatusPreparing {
		action = "删除"
	}
	iconBg, iconColor := regionIconColors(r.Name)
	return ServiceAreaDto{
		ID: r.ID, Name: r.Name, Status: regionStatusText(r.Status),
		Villages: r.Villages, RadiusKM: r.RadiusKM,
		OnlineMerchants: onlineMerchants, ActiveRiders: activeRiders, TodayOrders: todayOrders,
		StartFee: r.StartFee, FreeKM: r.FreeKM, Notice: notice, Action: action,
		IconBg: iconBg, IconColor: iconColor,
	}
}

func regionIconColors(name string) (string, string) {
	palette := [][2]string{
		{"#FFF3E8", "#FF6B2C"}, {"#E8F0FF", "#2B7FFF"},
		{"#E8F7F0", "#00B578"}, {"#FFF8E0", "#D48806"},
	}
	return palette[hashIdx(name, len(palette))][0], palette[hashIdx(name, len(palette))][1]
}

func feeStatusText(s cms.FeeRuleStatus) string {
	if s == cms.FeeDisabled {
		return "已停用"
	}
	return "已启用"
}

// FeeRuleRowDto 费率规则（对齐前端 FeeRuleRow）。
type FeeRuleRowDto struct {
	ID              string `json:"id"`
	Category        string `json:"category"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	Value           string `json:"value,omitempty"`
	Desc            string `json:"desc,omitempty"`
	Accent          bool   `json:"accent,omitempty"`
	Highlight       bool   `json:"highlight,omitempty"`
	Rule            string `json:"rule,omitempty"`
	Example         string `json:"example,omitempty"`
	Fee             string `json:"fee,omitempty"`
	RiderShare      string `json:"rider_share,omitempty"`
	IconColor       string `json:"icon_color,omitempty"`
	Version         string `json:"version,omitempty"`
	UpdatedAt       int64  `json:"updated_at,omitempty"`
	NextEffectiveAt int64  `json:"next_effective_at,omitempty"`
	Enabled         bool   `json:"enabled,omitempty"`
}

func toFeeRuleRow(r *cms.FeeRule) FeeRuleRowDto {
	return FeeRuleRowDto{
		ID: r.ID, Category: r.Category, Name: r.Name, Status: feeStatusText(r.Status),
		Value: r.Value, Desc: r.Desc, Accent: r.Accent, Highlight: r.Highlight,
		Rule: r.Rule, Example: r.Example, Fee: r.Fee, RiderShare: r.RiderShare,
		IconColor: r.IconColor, Version: r.Version, UpdatedAt: r.UpdatedAt,
		NextEffectiveAt: r.NextEffectiveAt, Enabled: r.Enabled,
	}
}

// AnnouncementRowDto 公告（对齐前端 AnnouncementRow）。
type AnnouncementRowDto struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Type        string       `json:"type"`
	TypeTone    string       `json:"type_tone"`
	Audience    string       `json:"audience"`
	Status      string       `json:"status"`
	PublishedAt *int64       `json:"published_at"`
	PlannedText string       `json:"planned_text"`
	ActionName  string       `json:"action_name"`
	Content     string       `json:"content"`
	Tags        []string     `json:"tags"`
	Stats       []AnnStatDto `json:"stats"`
}

type AnnStatDto struct {
	Label     string `json:"label"`
	Value     string `json:"value"`
	Highlight bool   `json:"highlight,omitempty"`
}

func announcementAction(status cms.AnnouncementStatus) string {
	switch status {
	case cms.AnnPublished:
		return "下线"
	case cms.AnnScheduled:
		return "发布"
	}
	return "撤回"
}

func announcementPlannedText(a *cms.Announcement) string {
	switch a.Status {
	case cms.AnnScheduled:
		if a.PlannedAt > 0 {
			return "计划 " + timeYMD(a.PlannedAt)[5:] + " 发布"
		}
		return "待发布"
	case cms.AnnReviewing:
		return "审核中"
	case cms.AnnPublished:
		return "已发布"
	}
	return "已下线"
}

func toAnnouncementRow(a *cms.Announcement) AnnouncementRowDto {
	var publishedAt *int64
	if a.PublishedAt > 0 {
		pa := a.PublishedAt
		publishedAt = &pa
	}
	stats := make([]AnnStatDto, 0, len(a.Stats))
	for _, s := range a.Stats {
		stats = append(stats, AnnStatDto{Label: s.Label, Value: s.Value, Highlight: s.Highlight})
	}
	return AnnouncementRowDto{
		ID: a.ID, Title: a.Title, Type: a.Type, TypeTone: a.TypeTone,
		Audience: a.Audience, Status: string(a.Status), PublishedAt: publishedAt,
		PlannedText: announcementPlannedText(a), ActionName: announcementAction(a.Status),
		Content: a.Content, Tags: a.Tags, Stats: stats,
	}
}
