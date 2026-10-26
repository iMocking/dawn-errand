package http

import (
	"strconv"

	"dawn-errand/internal/domain/rider"
)

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
	// IDCardMasked 身份证号脱敏（明文直存字段，脱敏后展示）。
	IDCardMasked string `json:"idCardMasked"`
	NoticeType   string `json:"noticeType"`
	Notice       string `json:"notice"`
	DocCount     int    `json:"docCount"`
	Docs         string `json:"docs"`
}

func maskIDCard(s string) string {
	runes := []rune(s)
	if len(runes) <= 8 {
		return "******"
	}
	return string(runes[:6]) + "********" + string(runes[len(runes)-4:])
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
		IDCheck: r.IDCheck, IDCardMasked: maskIDCard(r.IDCard), NoticeType: noticeType, Notice: r.Notice,
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
			Name: r.Name, Phone: r.Phone, IDCard: maskIDCard(r.IDCard), Area: r.Area,
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
