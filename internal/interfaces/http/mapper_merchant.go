package http

import (
	"strconv"
	"strings"

	"dawn-errand/internal/domain/merchant"
)

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
	ID           string               `json:"id"`
	ApplyNo      string               `json:"applyNo"`
	Name         string               `json:"name"`
	Operator     string               `json:"operator"`
	Type         string               `json:"type"`
	Address      string               `json:"address"`
	Contact      string               `json:"contact"`
	Phone        string               `json:"phone"`
	AppliedAt    int64                `json:"appliedAt"`
	SubmittedAt  int64                `json:"submittedAt"`
	LogoBg       string               `json:"logoBg"`
	LogoColor    string               `json:"logoColor"`
	LogoIcon     string               `json:"logoIcon"`
	Certs        []CertItem           `json:"certs"`
	Application  merchant.Application `json:"application"`
	Status       string               `json:"status"`
	RejectReason string               `json:"rejectReason"`
	VerifiedAt   int64                `json:"verifiedAt"`
}

type CertItem struct {
	Label string `json:"label"`
	Icon  string `json:"icon"`
	URL   string `json:"url,omitempty"`
}

func merchantCerts(m *merchant.Merchant) []CertItem {
	out := []CertItem{{Label: "营业执照", Icon: "license", URL: m.Application.LicenseImage}}
	switch merchantLogoIcon(m.Category) {
	case "pharmacy":
		out = append(out, CertItem{Label: "药品经营许可证", Icon: "pharmacylicense"})
	case "food":
		out = append(out, CertItem{Label: "食品经营许可证", Icon: "foodlicense", URL: m.Application.FoodLicenseImage})
	}
	if m.Application.IDCardFront != "" || m.Application.IDCardBack != "" {
		out = append(out, CertItem{Label: "法人身份证", Icon: "idcard", URL: m.Application.IDCardFront})
	}
	if len(m.Application.StorefrontImages) > 0 {
		out = append(out, CertItem{Label: "门店照片", Icon: "storefront", URL: m.Application.StorefrontImages[0]})
	} else {
		out = append(out, CertItem{Label: "门店照片", Icon: "storefront"})
	}
	return out
}

func toMerchantAuditRow(m *merchant.Merchant) MerchantAuditRowDto {
	bg, color, icon := merchantLogoColors(m)
	return MerchantAuditRowDto{
		ID: m.ID, ApplyNo: m.ApplyNo, Name: m.Name, Operator: m.Operator,
		Type: m.Category, Address: m.Address, Contact: m.Contact, Phone: m.Phone,
		AppliedAt: m.AppliedAt, SubmittedAt: m.SubmittedAt,
		LogoBg: bg, LogoColor: color, LogoIcon: icon,
		Certs: merchantCerts(m), Application: m.Application,
		Status: merchantAuditText(m.AuditStatus), RejectReason: m.RejectReason,
		VerifiedAt: m.VerifiedAt,
	}
}

// MerchantDetailDto 商家详情（对齐前端 MerchantDetail）。
type MerchantDetailDto struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Status         string               `json:"status"`
	Category       string               `json:"category"`
	CategoryBg     string               `json:"categoryBg"`
	CategoryColor  string               `json:"categoryColor"`
	JoinedAt       int64                `json:"joinedAt"`
	Rating         float64              `json:"rating"`
	MonthlyOrders  int64                `json:"monthlyOrders"`
	MonthlyRevenue float64              `json:"monthlyRevenue"`
	AvgOrder       float64              `json:"avgOrder"`
	GoodRate       string               `json:"goodRate"`
	LogoBg         string               `json:"logoBg"`
	LogoColor      string               `json:"logoColor"`
	Basic          MerchantBasicDto     `json:"basic"`
	LastUpdatedAt  int64                `json:"lastUpdatedAt"`
	Completeness   int                  `json:"completeness"`
	Today          MerchantTodayDto     `json:"today"`
	ReviewCount    int                  `json:"reviewCount"`
	LastOperation  string               `json:"lastOperation"`
	AuditStatus    string               `json:"auditStatus"`
	RejectReason   string               `json:"rejectReason"`
	SubmittedAt    int64                `json:"submittedAt"`
	VerifiedAt     int64                `json:"verifiedAt"`
	OwnerUserID    string               `json:"ownerUserId"`
	Application    merchant.Application `json:"application"`
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
		AuditStatus: merchantAuditText(m.AuditStatus), RejectReason: m.RejectReason,
		SubmittedAt: m.SubmittedAt, VerifiedAt: m.VerifiedAt, OwnerUserID: m.OwnerUserID,
		Application: m.Application,
	}
}
