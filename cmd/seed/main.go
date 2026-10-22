// cmd/seed 演示数据种子：初始化 schema → bootstrap 超管 → 写入对齐 apps/admin/mock 形态的演示数据。
// 用法：go run ./cmd/seed（可传 --email 覆盖超管邮箱，默认 admin@shuguang）
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"dawn-errand/internal/application"
	"dawn-errand/internal/assembly"
	"dawn-errand/internal/config"
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

func main() {
	email := "admin@shuguang"
	if len(os.Args) > 1 {
		for i, a := range os.Args {
			if a == "--email" && i+1 < len(os.Args) {
				email = os.Args[i+1]
			}
		}
	}

	ctx := context.Background()
	cfg := config.Load()
	app, err := assembly.Build(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "装配失败:", err)
		os.Exit(1)
	}
	defer app.Client.Close()
	defer app.Rdb.Close()

	if err := app.Client.Schema.Create(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "schema 迁移失败:", err)
		os.Exit(1)
	}

	// 1. bootstrap 超管（幂等）。
	st, _ := app.System.BootstrapStatus(ctx)
	if st.Initialized {
		fmt.Println("系统已初始化，跳过 bootstrap")
	} else {
		res, err := app.System.Bootstrap(ctx, email, cfg.BootstrapToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "bootstrap 失败:", err)
			os.Exit(1)
		}
		fmt.Printf("超管已创建: %s (%s)，一次性密码: %s\n", res.AdminName, res.AdminEmail, res.OneTimePassword)
	}

	// 2. 演示数据（幂等：已存在则跳过）。
	now := time.Now().Unix()
	repos := app.Store.Repos()

	seedRegion(ctx, repos, now)
	seedUser(ctx, repos, now)
	seedMerchant(ctx, repos, now)
	seedRider(ctx, repos, now)
	seedOrder(ctx, repos, now)
	seedWithdraw(ctx, repos, now)
	seedComplaint(ctx, repos, now)
	seedFeeRules(ctx, repos, now)
	seedAnnouncements(ctx, repos, now)

	fmt.Println("种子完成 ✔")
}

func exists(ctx context.Context, repos *application.RepoUnit, find func() error) bool {
	err := find()
	return err == nil
}

func seedRegion(ctx context.Context, repos *application.RepoUnit, now int64) {
	regions, err := repos.Regions.List()
	if err == nil {
		for _, r := range regions {
			if r.Name == "王家湾镇" {
				return
			}
		}
	}
	if err := repos.Regions.Save(&tenant.Region{
		ID: "area-001", Name: "王家湾镇", Status: tenant.RegionStatusOpen,
		Villages: 12, RadiusKM: 5, StartFee: 5, FreeKM: 2,
		Notice: "", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		fmt.Println("  ! 区域写入失败:", err)
		return
	}
	fmt.Println("  + 区域 area-001 王家湾镇")
}

func seedUser(ctx context.Context, repos *application.RepoUnit, now int64) {
	if _, err := repos.Users.FindByID("U86540"); err == nil {
		return
	}
	if err := repos.Users.Save(&identity.User{
		ID: "U86540", Email: "wangxiulan@shuguang.local", Phone: "13800008640",
		Name: "王秀兰", Status: identity.UserStatusActive, EmailVerified: true,
		Role: identity.RoleUser, RegionID: "area-001",
		RegisteredAt: now - 300*86400, UpdatedAt: now,
	}); err != nil {
		fmt.Println("  ! 用户写入失败:", err)
		return
	}
	_ = repos.Profiles.Save(&userprofile.Profile{
		UserID: "U86540", Level: "Lv.5 乡镇常客", Nickname: "王秀兰", RealName: "王秀兰",
		Verified: "已实名", Region: "王家湾镇", Address: "王家湾镇中心村 12 号",
		RegisterChannel: "微信", BizTypes: "外卖,帮买", Balance: 88.5, RedPacketBalance: 12.0,
		RecentActiveAt: now, UpdatedAt: now,
	})
	fmt.Println("  + 用户 U86540 王秀兰")
}

func seedMerchant(ctx context.Context, repos *application.RepoUnit, now int64) {
	if _, err := repos.Merchants.FindByID("M00268"); err == nil {
		return
	}
	if err := repos.Merchants.Save(&merchant.Merchant{
		ID: "M00268", RegionID: "area-001", Name: "村口老陈家常菜", Category: "餐饮",
		Address: "王家湾镇正街 28 号", AddressSub: "距镇政府 300m", Contact: "老陈", Phone: "13900002680",
		OpenHours: "09:00-21:00", DeliveryRange: "全镇 5km", Commission: 12,
		BankAccount: "6222 0210 **** 8690", DeliveryMethod: "平台骑手",
		Status: merchant.MerchantStatusOpen, AuditStatus: merchant.AuditApproved,
		ApplyNo: "MA20261001001", Operator: "陈建国", AppliedAt: now - 60*86400,
		CertLabels: []string{"license", "foodlicense", "storefront"},
		Rating:     4.8, MonthlyOrders: 268, MonthlyRevenue: 13400, AvgOrder: 50, GoodRate: 97.5,
		JoinedAt: now - 60*86400, LastUpdatedAt: now, Completeness: 100, ReviewCount: 86,
		LastOperation: "陈建国 更新了菜品价格",
		TodayOrders:   12, TodayRevenue: 620, OnTimeRate: 96, PendingAbnormal: 0,
	}); err != nil {
		fmt.Println("  ! 商家写入失败:", err)
		return
	}
	fmt.Println("  + 商家 M00268 村口老陈家常菜")
}

func seedRider(ctx context.Context, repos *application.RepoUnit, now int64) {
	if _, err := repos.Riders.FindByID("R00086"); err == nil {
		return
	}
	if err := repos.Riders.Save(&rider.Rider{
		ID: "R00086", RegionID: "area-001", UserID: "U86541",
		Name: "赵铁柱", Phone: "13700008686", Area: "王家湾镇", Tool: "电动两轮车",
		Level: "Lv.5", IDCard: "4201**********0817", Vehicle: "台铃小绿", Account: "6222 0202 **** 8612",
		Status: rider.RiderOnline, AuditStatus: rider.AuditApproved,
		ApplyNo: "RA20261001086", SubmittedAt: now - 45*86400,
		OrderTypes: "外卖,帮送,帮取", IDCheck: "已通过", NoticeType: "none",
		DocCount: 3, Docs: "身份证,驾驶证,车辆证明",
		ServiceType: "外卖/帮送", JoinedAt: now - 45*86400, ServiceMonths: 14,
		Rating: 4.9, TodayOrders: 23, TodayOrdersDelta: "▲ 5",
		OnlineHours: 8.5, MonthIncome: 3860, OnTimeRate: 97,
		ComplaintsToday: 0, AvgMinutes: 28,
		LastAction: "赵铁柱 完成订单 SG202610220086", LastActionAt: now,
	}); err != nil {
		fmt.Println("  ! 骑手写入失败:", err)
		return
	}
	fmt.Println("  + 骑手 R00086 赵铁柱")
}

func seedOrder(ctx context.Context, repos *application.RepoUnit, now int64) {
	if _, err := repos.Orders.FindByID("SG202610220086"); err == nil {
		return
	}
	if err := repos.Orders.Save(&order.Order{
		ID: "SG202610220086", OrderNo: "202610220086", RegionID: "area-001",
		MerchantID: "M00268", UserID: "U86540", RiderID: "R00086",
		Type: order.OrderTypeFood, Status: order.OrderDelivering,
		Route: "村口老陈家常菜 → 王家湾镇中心村 12 号", UserName: "王秀兰", RiderName: "赵铁柱",
		Amount: 48.0, CreatedAt: now - 1800, Channel: "微信小程序",
		EstimatedArrival: now + 900, Remark: "少放辣，多送一份米饭",
		Goods: []order.OrderGoods{
			{Name: "尖椒炒肉盖饭", Spec: "中份", Price: 22},
			{Name: "西红柿蛋汤", Spec: "大份", Price: 12},
			{Name: "米饭", Spec: "加饭", Price: 3},
		},
		Fees: []order.OrderFee{
			{Label: "商品金额", Value: 37},
			{Label: "打包费", Value: 2},
			{Label: "配送费", Value: 5},
			{Label: "餐具费", Value: 1},
			{Label: "新客立减", Value: -2, Negative: true},
		},
		Steps: []order.OrderStep{
			{Title: "已下单", Time: now - 1800, State: "done", Sub: "微信小程序"},
			{Title: "商家接单", Time: now - 1650, State: "done", Sub: "已确认"},
			{Title: "骑手取餐", Time: now - 900, State: "active", Sub: "配送中"},
			{Title: "送达完成", Time: 0, State: "pending", Sub: ""},
		},
		Merchant:  order.MerchantSnapshot{Name: "村口老陈家常菜", Address: "王家湾镇正街 28 号", Phone: "13900002680", Status: "营业中"},
		User:      order.UserSnapshot{Name: "王秀兰", Address: "王家湾镇中心村 12 号", Phone: "13800008640"},
		Rider:     order.RiderSnapshot{Name: "赵铁柱", Vehicle: "台铃小绿", Phone: "13700008686", Service: "外卖", Score: "4.9"},
		PayAmount: 43, PayMethod: "微信支付", FundFlow: "余额支付 ¥43.00（订单已结算至商家）",
	}); err != nil {
		fmt.Println("  ! 订单写入失败:", err)
		return
	}
	fmt.Println("  + 订单 SG202610220086")
}

func seedWithdraw(ctx context.Context, repos *application.RepoUnit, now int64) {
	if _, err := repos.Withdrawals.FindByID("TX20261022016"); err == nil {
		return
	}
	if err := repos.Withdrawals.Save(&finance.Withdrawal{
		ID: "TX20261022016", RegionID: "area-001",
		PartyType: "rider", PartyName: "赵铁柱", PartyNo: "R00086",
		BankInfo: "尾号 8612 · 工商银行", Amount: 860, Fee: 0,
		FeeNote: "本单免手续费", Status: finance.WithdrawPending, Kind: "骑手提现",
		Applicant: "赵铁柱", AccountHolder: "赵铁柱", WithdrawType: "银行卡提现",
		ArriveExpect: "预计 2 小时内到账", Purpose: "跑腿收入提现",
		Balance: 1286.5, RemainBalance: 426.5,
		MonthOrders: 286, MonthIncome: 3860,
		RiskItems:   []string{"近 7 日新增绑卡", "单笔金额超过月均 3 倍"},
		RiskWarning: "新卡首次提现，建议人工复核",
		CreatedAt:   now - 3600,
		Timeline: []finance.Event{
			{Time: now - 3600, Text: "赵铁柱 提交提现申请 ¥860.00", Operator: "赵铁柱"},
			{Time: now - 1800, Text: "系统风控标记 2 项风险", Operator: "风控引擎"},
		},
	}); err != nil {
		fmt.Println("  ! 提现写入失败:", err)
		return
	}
	fmt.Println("  + 提现 TX20261022016")
}

func seedComplaint(ctx context.Context, repos *application.RepoUnit, now int64) {
	if _, err := repos.Complaints.FindByID("GD20261022006"); err == nil {
		return
	}
	if err := repos.Complaints.Save(&complaint.Complaint{
		ID: "GD20261022006", OrderID: "SG202610220086", RegionID: "area-001", UserID: "U86540",
		Category: "配送异常", Title: "配送超时 25 分钟，餐品已凉", Status: complaint.StatusPending,
		Priority: "普通", Complainant: "王秀兰", Target: "赵铁柱", Extra: "骑手",
		Description: "预计 12:10 送达，实际 12:35 才送到，汤已经凉了。", DescTone: "yellow",
		Deadline: "今日 18:00", Initiator: "王秀兰", InitiatorExtra: "13800008640",
		PayAmount: 43, CreatedAt: now - 3600, UpdatedAt: now - 3600,
		Flow: []complaint.FlowStep{
			{Title: "用户提交", Desc: "已提交工单", State: "done", Step: 1},
			{Title: "客服受理", Desc: "待分配客服", State: "current", Step: 2},
			{Title: "处理完成", Desc: "", State: "pending", Step: 3},
		},
		Notes: []complaint.Note{
			{Avatar: "王秀兰", AvatarBg: "#FFF3E8", AvatarColor: "#FF6B2C", Name: "王秀兰", Time: now - 3600, Content: "汤凉了，申请部分退款", Tone: "yellow"},
		},
		Order: []complaint.OrderItem{
			{Label: "订单号", Value: "202610220086", Tone: "blue"},
			{Label: "骑手", Value: "赵铁柱 (R00086)", Tone: "primary"},
			{Label: "配送时长", Value: "47 分钟", Tone: "red"},
		},
	}); err != nil {
		fmt.Println("  ! 投诉写入失败:", err)
		return
	}
	fmt.Println("  + 投诉 GD20261022006")
}

func seedFeeRules(ctx context.Context, repos *application.RepoUnit, now int64) {
	seed := func(id, category, name, value, desc string, accent bool, rule, example, fee, share, color string) {
		if _, err := repos.FeeRules.FindByID(id); err == nil {
			return
		}
		if err := repos.FeeRules.Save(&cms.FeeRule{
			ID: id, Category: category, Name: name, Status: cms.FeeEnabled,
			Value: value, Desc: desc, Accent: accent, Highlight: accent,
			Rule: rule, Example: example, Fee: fee, RiderShare: share,
			IconColor: color, Version: "v1.2", UpdatedAt: now,
			Enabled: true,
		}); err != nil {
			fmt.Println("  ! 费率规则写入失败:", id, err)
		}
	}
	seed("fee-summary", "summary", "综合费率", "12%", "平台综合费率 12%", true,
		"订单实付金额 × 12%", "示例：订单 50 元 → 平台费 6 元", "6 元", "3.6 元 (60%)", "#FF6B2C")
	seed("fee-basic-1", "basic", "起步价", "5 元", "起步价含 2 公里", false,
		"配送距离 ≤ 2km 时按起步价", "示例：1.5km → 5 元", "5 元", "2.5 元 (50%)", "#2B7FFF")
	seed("fee-basic-2", "basic", "里程加价", "1.5 元/km", "超出 2km 后每公里加价", false,
		"超出部分 × 1.5 元/km", "示例：5km → 5+3×1.5=9.5 元", "9.5 元", "4.75 元 (50%)", "#2B7FFF")
	seed("fee-ladder-1", "ladder", "晚高峰加成", "+20%", "17:00-19:00 时段加成", true,
		"订单配送费 × 120%", "示例：配送费 5 元 → 6 元", "6 元", "3.6 元 (60%)", "#FF8A3D")
	seed("fee-special-1", "special", "恶劣天气附加费", "+30%", "暴雨/大雪天附加费", false,
		"订单配送费 × 130%", "示例：配送费 5 元 → 6.5 元", "6.5 元", "3.9 元 (60%)", "#722ED1")
	seed("fee-scope-1", "scope", "偏远村组附加费", "+5 元", "覆盖偏远村组配送", false,
		"订单配送费 +5 元", "示例：配送费 5 元 → 10 元", "10 元", "5 元 (50%)", "#00B578")
	fmt.Println("  + 费率规则 6 条")
}

func seedAnnouncements(ctx context.Context, repos *application.RepoUnit, now int64) {
	seed := func(id, title, typ, tone, audience string, status cms.AnnouncementStatus, publishedAt, plannedAt int64, content string, tags []string) {
		if _, err := repos.Announcements.FindByID(id); err == nil {
			return
		}
		if err := repos.Announcements.Save(&cms.Announcement{
			ID: id, Title: title, Type: typ, TypeTone: tone, Audience: audience,
			Status: status, PublishedAt: publishedAt, PlannedAt: plannedAt,
			Content: content, Tags: tags, CreatedBy: "seed",
			CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			fmt.Println("  ! 公告写入失败:", id, err)
		}
	}
	seed("ann-001", "关于配送费调整的通知", "费率", "orange", "全部用户", cms.AnnPublished,
		now-3*86400, 0, "自下月起，起步价调整至 5 元（含 2km）。", []string{"费率", "重要"})
	seed("ann-002", "十一假期运力保障安排", "活动", "blue", "全部骑手", cms.AnnScheduled,
		0, now+2*86400, "假期高峰时段平台将启用晚高峰加成。", []string{"运力"})
	seed("ann-003", "商家食品安全自查周", "提醒", "yellow", "全部商家", cms.AnnReviewing,
		0, 0, "请各商家配合完成食品安全自查。", []string{"商家"})
	fmt.Println("  + 公告 3 条")
}
