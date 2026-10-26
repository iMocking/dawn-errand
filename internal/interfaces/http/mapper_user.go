package http

import (
	"strings"

	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/realname"
	"dawn-errand/internal/domain/userprofile"
)

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
	Role         string `json:"role"`
	RoleName     string `json:"roleName"`
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
		RegisteredAt: u.RegisteredAt, Role: u.Role, RoleName: roleDisplayName(u.Role),
	}
}

func roleDisplayName(role string) string {
	if n, ok := identity.RoleDisplayName[role]; ok {
		return n
	}
	if role == "" {
		return "普通用户"
	}
	return role
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
	Role             string               `json:"role"`
	RoleName         string               `json:"roleName"`
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
	RealNameVerify   *RealNameVerifyDto   `json:"realNameVerify"`
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

// toRecentOrder 订单 → 用户详情订单行 DTO（基本信息卡/订单记录 tab 复用）。
func toRecentOrder(o *order.Order) UserRecentOrderDto {
	if o == nil {
		return UserRecentOrderDto{}
	}
	bz := string(o.Type)
	bzbg, bzcolor := orderTypeColors(o.Type)
	riderText := "待分配"
	if o.RiderName != "" {
		riderText = o.RiderName
	}
	merchantText := o.Route
	return UserRecentOrderDto{
		ID: o.ID, BizType: bz, BizBg: bzbg, BizColor: bzcolor,
		Merchant: merchantText, Amount: o.Amount, Rider: riderText,
		Status: string(o.Status), OrderedAt: o.CreatedAt,
	}
}

func toUserDetail(u *identity.User, p *userprofile.Profile, totalOrders, monthOrders int64, totalSpent float64, complaintCount int64, recent []*order.Order, rn *realname.View) UserDetailDto {
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
		recentOrders = append(recentOrders, toRecentOrder(o))
	}
	return UserDetailDto{
		ID: u.ID, UserNo: "U" + u.ID[len(u.ID)-6:], Name: displayName(u),
		AvatarBg: bg, AvatarColor: color, Level: level, Status: userStatusText(u.Status),
		Role: u.Role, RoleName: roleDisplayName(u.Role),
		Phone: formatPhone(u.Phone), RegisteredAt: u.RegisteredAt,
		TotalOrders: totalOrders, TotalSpent: totalSpent, MonthOrders: monthOrders,
		Verified: verified, Nickname: nickname, RealName: realName, Region: region,
		Address: address, RegisterChannel: channel, BizTypes: bizTypes,
		Balance: balance, RedPacketBalance: redPacket,
		RecentActiveAt:   u.UpdatedAt,
		OrderRecordCount: totalOrders, ComplaintCount: complaintCount,
		RecentOrders:   recentOrders,
		RealNameVerify: toRealNameVerification(rn),
	}
}

// RealNameVerifyDto 实名认证申请（身份证号脱敏：保留前 6 后 4）。
type RealNameVerifyDto struct {
	ID           string `json:"id"`
	UserID       string `json:"userId"`
	RealName     string `json:"realName"`
	IDCardMasked string `json:"idCardMasked"`
	Status       string `json:"status"`
	RejectReason string `json:"rejectReason"`
	AppliedAt    int64  `json:"appliedAt"`
	ReviewedAt   int64  `json:"reviewedAt"`
}

func toRealNameVerification(v *realname.View) *RealNameVerifyDto {
	if v == nil {
		return nil
	}
	return &RealNameVerifyDto{
		ID: v.ID, UserID: v.UserID, RealName: v.RealName,
		IDCardMasked: v.IDCardMasked, Status: v.Status,
		RejectReason: v.RejectReason, AppliedAt: v.AppliedAt, ReviewedAt: v.ReviewedAt,
	}
}
