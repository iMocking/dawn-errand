package http

import (
	"dawn-errand/internal/domain/cms"
	"dawn-errand/internal/domain/tenant"
)

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
