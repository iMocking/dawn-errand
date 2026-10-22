package http

import (
	"math"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/pkg/response"
)

// StatCardDto 顶部指标卡（对齐前端 StatCardDto）。
type StatCardDto struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Value     string `json:"value"`
	SubValue  string `json:"subValue,omitempty"`
	Delta     string `json:"delta"`
	DeltaType string `json:"deltaType"`
	Foot      string `json:"foot"`
	Icon      string `json:"icon"`
	IconBg    string `json:"iconBg"`
	IconColor string `json:"iconColor"`
}

type TrendPointDto struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type TrendTipDto struct {
	X    int    `json:"x"`
	Y    int    `json:"y"`
	W    int    `json:"w"`
	Text string `json:"text"`
}

type TrendDto struct {
	XLabels   []string        `json:"xLabels"`
	OrderLine []TrendPointDto `json:"orderLine"`
	GmvLine   []TrendPointDto `json:"gmvLine"`
	OrderTip  TrendTipDto     `json:"orderTip"`
	GmvTip    TrendTipDto     `json:"gmvTip"`
}

type DistItemDto struct {
	Label   string `json:"label"`
	Percent int    `json:"percent"`
	Color   string `json:"color"`
}

type RecentOrderDto struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	TypeBg      string `json:"typeBg"`
	TypeColor   string `json:"typeColor"`
	Route       string `json:"route"`
	Status      string `json:"status"`
	StatusColor string `json:"statusColor"`
	Time        string `json:"time"`
}

type TodoDto struct {
	Icon  string `json:"icon"`
	Bg    string `json:"bg"`
	Color string `json:"color"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// DashboardDto 数据总览（对齐前端 DashboardDto）。
type DashboardDto struct {
	PendingCount int64            `json:"pendingCount"`
	Stats        []StatCardDto    `json:"stats"`
	Trend        TrendDto         `json:"trend"`
	Distribution []DistItemDto    `json:"distribution"`
	OnTimeRate   string           `json:"onTimeRate"`
	OnTimeTarget string           `json:"onTimeTarget"`
	RecentOrders []RecentOrderDto `json:"recentOrders"`
	Todos        []TodoDto        `json:"todos"`
}

// GetDashboard GET /dashboard。
func (h *Handler) GetDashboard(c *gin.Context) {
	data, err := h.Dashboard.Get(c.Request.Context(), scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	now := time.Now().Unix()
	orderDelta := 0.0
	if data.YesterdayOrders > 0 {
		orderDelta = (float64(data.TodayOrders) - float64(data.YesterdayOrders)) / float64(data.YesterdayOrders) * 100
	}
	gmvDelta := 0.0
	if data.YesterdayGMV > 0 {
		gmvDelta = (data.TodayGMV - data.YesterdayGMV) / data.YesterdayGMV * 100
	}
	onlineRate := 0.0
	if data.TotalRiders > 0 {
		onlineRate = float64(data.OnlineRiders) / float64(data.TotalRiders) * 100
	}

	stats := []StatCardDto{
		{Key: "todayOrders", Label: "今日订单", Value: thousandSep(data.TodayOrders),
			Delta: deltaStr(orderDelta), DeltaType: deltaType(orderDelta), Foot: "较昨日",
			Icon: "order", IconBg: "#FFF3E8", IconColor: "#FF6B2C"},
		{Key: "todayGmv", Label: "今日交易额", Value: moneyStr(data.TodayGMV),
			Delta: deltaStr(gmvDelta), DeltaType: deltaType(gmvDelta), Foot: "较昨日",
			Icon: "money", IconBg: "#E8F7F0", IconColor: "#00B578"},
		{Key: "onlineRiders", Label: "在线骑手", Value: thousandSep(data.OnlineRiders),
			SubValue: "/" + thousandSep(data.TotalRiders), Delta: pctStr(onlineRate),
			DeltaType: deltaType(onlineRate - 80), Foot: "在线率",
			Icon: "grid", IconBg: "#E8F0FF", IconColor: "#2B7FFF"},
		{Key: "merchants", Label: "入驻商家", Value: thousandSep(data.Merchants),
			Delta: "▲ " + thousandSep(data.MerchantsWeekNew) + "家", DeltaType: "green", Foot: "本周新增",
			Icon: "merchant", IconBg: "#FFF8E0", IconColor: "#D48806"},
	}

	trend := buildTrend(data.Trend, now)
	dist := buildDistribution(data.Distribution)

	recent := make([]RecentOrderDto, 0, len(data.RecentOrders))
	for _, o := range data.RecentOrders {
		_, tcol := orderTypeColors(o.Type)
		recent = append(recent, RecentOrderDto{
			ID: o.ID, Type: string(o.Type), TypeBg: orderTypeBg(o.Type), TypeColor: tcol,
			Route: o.Route, Status: string(o.Status), StatusColor: orderStatusColor[o.Status],
			Time: timeHM(o.CreatedAt),
		})
	}

	todos := []TodoDto{
		{Icon: "merchant", Bg: "#FFF3E8", Color: "#FF6B2C", Label: "商家入驻待审核", Count: data.TodoMerchantAudit},
		{Icon: "grid", Bg: "#E8F0FF", Color: "#2B7FFF", Label: "骑手入驻待审核", Count: data.TodoRiderAudit},
		{Icon: "money", Bg: "#FFF8E0", Color: "#D48806", Label: "骑手提现待处理", Count: data.TodoWithdrawAudit},
		{Icon: "complaint", Bg: "#FFE8E8", Color: "#FF4D4F", Label: "投诉工单待处理", Count: data.TodoComplaints},
	}
	pending := data.TodoMerchantAudit + data.TodoRiderAudit + data.TodoWithdrawAudit + data.TodoComplaints

	response.OK(c, DashboardDto{
		PendingCount: pending, Stats: stats, Trend: trend, Distribution: dist,
		OnTimeRate: pctStr(data.OnTimeRate), OnTimeTarget: "95%",
		RecentOrders: recent, Todos: todos,
	})
}

func deltaType(v float64) string {
	if v >= 0 {
		return "green"
	}
	return "red"
}

func orderTypeBg(t order.OrderType) string {
	bg, _ := orderTypeColors(t)
	return bg
}

func buildTrend(days []order.DayTrend, now int64) TrendDto {
	n := 7
	if len(days) > 0 {
		n = len(days)
	}
	xLabels := make([]string, 0, n)
	orders := make([]int64, 0, n)
	gmvs := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		var day order.DayTrend
		if i < len(days) {
			day = days[i]
		} else {
			day = order.DayTrend{Day: timeMD(now - int64(n-i-1)*86400)}
		}
		if day.Day == "" {
			day.Day = timeMD(now - int64(n-i-1)*86400)
		}
		if i == n-1 {
			xLabels = append(xLabels, "今天")
		} else {
			xLabels = append(xLabels, day.Day)
		}
		orders = append(orders, day.Orders)
		gmvs = append(gmvs, day.GMV)
	}
	minO, maxO := minMax64(orders)
	minG, maxG := minMaxF(gmvs)
	orderLine := make([]TrendPointDto, 0, n)
	gmvLine := make([]TrendPointDto, 0, n)
	for i := 0; i < n; i++ {
		x := 40 + i*97
		orderLine = append(orderLine, TrendPointDto{X: x, Y: normY(float64(orders[i]), float64(minO), float64(maxO))})
		gmvLine = append(gmvLine, TrendPointDto{X: x, Y: normY(gmvs[i], minG, maxG)})
	}
	return TrendDto{
		XLabels: xLabels, OrderLine: orderLine, GmvLine: gmvLine,
		OrderTip: TrendTipDto{X: 40 + (n-1)*97, Y: normY(float64(orders[n-1]), float64(minO), float64(maxO)) - 76, W: 56, Text: thousandSep(orders[n-1])},
		GmvTip:   TrendTipDto{X: 35 + (n-1)*97, Y: normY(gmvs[n-1], minG, maxG) - 80, W: 90, Text: moneyStr(gmvs[n-1])},
	}
}

func minMax64(v []int64) (int64, int64) {
	if len(v) == 0 {
		return 0, 1
	}
	mn, mx := v[0], v[0]
	for _, x := range v[1:] {
		if x < mn {
			mn = x
		}
		if x > mx {
			mx = x
		}
	}
	return mn, mx
}

func minMaxF(v []float64) (float64, float64) {
	if len(v) == 0 {
		return 0, 1
	}
	mn, mx := v[0], v[0]
	for _, x := range v[1:] {
		if x < mn {
			mn = x
		}
		if x > mx {
			mx = x
		}
	}
	return mn, mx
}

// normY 归一化到 40..200 绘图坐标。
func normY(v, mn, mx float64) int {
	if mx <= mn {
		return 150
	}
	ratio := (v - mn) / (mx - mn)
	y := 200 - ratio*140
	return int(math.Round(y))
}

var distColor = map[order.OrderType]string{
	order.OrderTypeFood:   "linear-gradient(90deg,#FF8A3D,#FF6B2C)",
	order.OrderTypeBuy:    "#2B7FFF",
	order.OrderTypeSend:   "#00B578",
	order.OrderTypePick:   "#D48806",
	order.OrderTypeErrand: "#722ED1",
}

var distLabel = map[order.OrderType]string{
	order.OrderTypeFood:   "外卖配送",
	order.OrderTypeBuy:    "帮买商品",
	order.OrderTypeSend:   "帮送物品",
	order.OrderTypePick:   "帮取快递",
	order.OrderTypeErrand: "帮办业务",
}

func buildDistribution(counts map[order.OrderType]int64) []DistItemDto {
	total := int64(0)
	for _, n := range counts {
		total += n
	}
	if total == 0 {
		return []DistItemDto{}
	}
	out := make([]DistItemDto, 0, len(counts))
	for t, n := range counts {
		pct := int(math.Round(float64(n) / float64(total) * 100))
		out = append(out, DistItemDto{Label: distLabel[t], Percent: pct, Color: distColor[t]})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Percent > out[j].Percent })
	return out
}
