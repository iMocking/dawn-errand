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
	"dawn-errand/internal/domain/address"
	"dawn-errand/internal/domain/catalog"
	"dawn-errand/internal/domain/cms"
	"dawn-errand/internal/domain/complaint"
	"dawn-errand/internal/domain/coupon"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/review"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/userprofile"
	"dawn-errand/internal/pkg/ulidgen"
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

	// 2. 内置系统级身份（幂等；随初始化自动创建）。
	if err := app.System.EnsureSystemUsers(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "系统身份创建失败:", err)
		os.Exit(1)
	}
	fmt.Printf("系统身份已就绪: %d 个\n", len(identity.SystemIdentities))

	// 3. 演示数据（幂等：已存在则跳过）。
	now := time.Now().Unix()
	repos := app.Store.Repos()

	seedRegion(ctx, repos, now)
	seedUser(ctx, repos, now)
	seedMerchant(ctx, repos, now)
	seedRider(ctx, repos, now)
	seedAccounts(ctx, repos, app.Auth.Hasher, now)
	seedCatalog(ctx, repos, now)
	seedAddresses(ctx, repos, now)
	seedCoupons(ctx, repos, now)
	seedOrder(ctx, repos, now)
	seedMerchantApplication(ctx, repos)
	seedReview(ctx, repos, now)
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

// seedAccounts 三端联调账号（幂等）：客户端 / 商家端 / 骑手端。
// 商家与骑手账号同时写入 users.merchant_id / users.rider_id 与主体表反向绑定。
func seedAccounts(ctx context.Context, repos *application.RepoUnit, hasher application.PasswordHasher, now int64) {
	type seedAccount struct {
		Email      string
		Password   string
		Name       string
		Role       string
		RegionID   string
		MerchantID string
		RiderID    string
	}
	accounts := []seedAccount{
		{Email: "client.test@dawnerrand.local", Password: "Client123456", Name: "联调用户", Role: identity.RoleUser, RegionID: "area-001"},
		{Email: "store.owner@dawnerrand.local", Password: "Store123456", Name: "老陈", Role: identity.RoleMerchant, RegionID: "area-001", MerchantID: "M00268"},
		{Email: "rider.owner@dawnerrand.local", Password: "Rider123456", Name: "赵铁柱", Role: identity.RoleDriver, RegionID: "area-001", RiderID: "R00086"},
	}
	for _, a := range accounts {
		u, err := repos.Users.FindByEmail(a.Email)
		if err != nil {
			hash, herr := hasher.Hash(a.Password)
			if herr != nil {
				fmt.Println("  ! 账号密码哈希失败:", a.Email, herr)
				continue
			}
			u = &identity.User{
				ID: ulidgen.New(), Email: a.Email, Name: a.Name,
				Status: identity.UserStatusActive, EmailVerified: true,
				Role: a.Role, RegionID: a.RegionID,
				MerchantID: a.MerchantID, RiderID: a.RiderID,
				RegisteredAt: now, UpdatedAt: now,
			}
			if err := repos.Users.Save(u); err != nil {
				fmt.Println("  ! 联调账号写入失败:", a.Email, err)
				continue
			}
			if err := repos.Credentials.Save(&identity.Credential{
				UserID: u.ID, PasswordHash: hash, SessionVersion: 0, UpdatedAt: now,
			}); err != nil {
				fmt.Println("  ! 联调账号凭证写入失败:", a.Email, err)
				continue
			}
			_ = repos.Profiles.Save(&userprofile.Profile{
				UserID: u.ID, Level: "Lv.3 联调账号", Nickname: a.Name,
				Verified: "未实名", Region: "王家湾镇", RegisterChannel: "seed",
				RecentActiveAt: now, UpdatedAt: now,
			})
			fmt.Println("  + 联调账号", a.Email, "("+identity.RoleDisplayName[a.Role]+")")
		} else {
			// 已存在：只补齐角色与绑定，不重置密码，避免覆盖联调中修改的凭据。
			u.Role = a.Role
			u.RegionID = a.RegionID
			if a.MerchantID != "" {
				u.MerchantID = a.MerchantID
			}
			if a.RiderID != "" {
				u.RiderID = a.RiderID
			}
			u.EmailVerified = true
			u.UpdatedAt = now
			if err := repos.Users.Save(u); err != nil {
				fmt.Println("  ! 联调账号绑定更新失败:", a.Email, err)
			}
			if _, err := repos.Credentials.FindByUserID(u.ID); err != nil {
				hash, herr := hasher.Hash(a.Password)
				if herr == nil {
					_ = repos.Credentials.Save(&identity.Credential{
						UserID: u.ID, PasswordHash: hash, UpdatedAt: now,
					})
				}
			}
			fmt.Println("  = 联调账号已存在", a.Email)
		}
	}

	// 主体表反向绑定（幂等）。
	if m, err := repos.Merchants.FindByID("M00268"); err == nil {
		storeUser, uerr := repos.Users.FindByEmail("store.owner@dawnerrand.local")
		if uerr == nil && m.OwnerUserID != storeUser.ID {
			m.OwnerUserID = storeUser.ID
			m.LastUpdatedAt = now
			if err := repos.Merchants.Save(m); err != nil {
				fmt.Println("  ! 商家反向绑定失败:", err)
			}
		}
	}
	if r, err := repos.Riders.FindByID("R00086"); err == nil {
		riderUser, uerr := repos.Users.FindByEmail("rider.owner@dawnerrand.local")
		if uerr == nil && r.UserID != riderUser.ID {
			r.UserID = riderUser.ID
			if err := repos.Riders.Save(r); err != nil {
				fmt.Println("  ! 骑手反向绑定失败:", err)
			}
		}
	}
}

// seedCatalog 商家经营设置 + 菜品分类/菜品（幂等）。
func seedCatalog(ctx context.Context, repos *application.RepoUnit, now int64) {
	if m, err := repos.Merchants.FindByID("M00268"); err == nil {
		// 仅首次初始化经营设置，避免覆盖联调/运营修改。
		if m.Lat == 0 && m.Lng == 0 && m.MinPrice == 0 && m.DeliveryFee == 0 && m.PackingFee == 0 && m.Announcement == "" {
			m.Lat, m.Lng = 30.4201, 115.3290
			m.MinPrice = 15
			m.DeliveryFee = 5
			m.PackingFee = 2
			m.Announcement = "今日有新鲜土鸡，欢迎下单"
			m.ServicePhone = "13900002680"
			m.LastUpdatedAt = now
			if err := repos.Merchants.Save(m); err != nil {
				fmt.Println("  ! 商家经营设置写入失败:", err)
			}
		}
	}

	categories := []catalog.Category{
		{ID: "C001", MerchantID: "M00268", Name: "招牌热菜", Sort: 1},
		{ID: "C002", MerchantID: "M00268", Name: "主食汤品", Sort: 2},
	}
	for i := range categories {
		if _, err := repos.Catalog.FindCategory(categories[i].ID); err == nil {
			continue
		}
		if err := repos.Catalog.SaveCategory(&categories[i]); err != nil {
			fmt.Println("  ! 菜品分类写入失败:", categories[i].ID, err)
		}
	}

	dishes := []catalog.Dish{
		{ID: "D001", MerchantID: "M00268", CategoryID: "C001", Name: "尖椒炒肉盖饭", Description: "本地辣椒，现炒", Price: 22, Stock: 50, OnSale: true, MonthlySales: 320, Sort: 1},
		{ID: "D002", MerchantID: "M00268", CategoryID: "C001", Name: "红烧土鸡块", Description: "土鸡现烧，微辣", Price: 58, Stock: 20, OnSale: true, MonthlySales: 128, Sort: 2},
		{ID: "D003", MerchantID: "M00268", CategoryID: "C001", Name: "手撕包菜", Description: "大火快炒", Price: 12, Stock: 0, OnSale: true, MonthlySales: 52, Sort: 3},
		{ID: "D004", MerchantID: "M00268", CategoryID: "C002", Name: "西红柿蛋汤", Description: "农家土鸡蛋", Price: 12, Stock: 30, OnSale: true, MonthlySales: 96, Sort: 1},
		{ID: "D005", MerchantID: "M00268", CategoryID: "C002", Name: "米饭", Description: "东北大米", Price: 3, Stock: 100, OnSale: true, MonthlySales: 560, Sort: 2},
		{ID: "D006", MerchantID: "M00268", CategoryID: "C001", Name: "农家小炒肉", Description: "季节限定，暂未上架", Price: 28, Stock: 0, OnSale: false, MonthlySales: 0, Sort: 4},
	}
	for i := range dishes {
		if _, err := repos.Catalog.FindDish(dishes[i].ID); err == nil {
			continue
		}
		if err := repos.Catalog.SaveDish(&dishes[i]); err != nil {
			fmt.Println("  ! 菜品写入失败:", dishes[i].ID, err)
		}
	}
	fmt.Println("  + 菜品分类 2 个、菜品 6 个")
}

// seedAddresses 联调用户地址（幂等，仅首次写入）。
func seedAddresses(ctx context.Context, repos *application.RepoUnit, now int64) {
	u, err := repos.Users.FindByEmail("client.test@dawnerrand.local")
	if err != nil {
		fmt.Println("  ! 联调用户不存在，跳过地址种子")
		return
	}
	if n, _ := repos.Addresses.CountByUser(u.ID); n > 0 {
		return
	}
	rows := []address.Address{
		{ID: "ADDR001", UserID: u.ID, ContactName: "联调用户", Phone: "13800008640", RegionID: "area-001", Detail: "王家湾镇中心村 12 号", Tag: "家", Lat: 30.4310, Lng: 115.3350, IsDefault: true},
		{ID: "ADDR002", UserID: u.ID, ContactName: "联调用户", Phone: "13800008640", RegionID: "area-001", Detail: "王家湾镇正街 6 号供销社", Tag: "公司", Lat: 30.4205, Lng: 115.3295},
	}
	for i := range rows {
		if err := repos.Addresses.Save(&rows[i]); err != nil {
			fmt.Println("  ! 地址写入失败:", rows[i].ID, err)
		}
	}
	fmt.Println("  + 用户地址 2 条")
}

// seedCoupons 联调用户优惠券（幂等）。
func seedCoupons(ctx context.Context, repos *application.RepoUnit, now int64) {
	u, err := repos.Users.FindByEmail("client.test@dawnerrand.local")
	if err != nil {
		return
	}
	rows := []coupon.Coupon{
		{ID: "CPN001", UserID: u.ID, MerchantID: "M00268", Title: "老陈家常菜满减券", Amount: 3, MinSpend: 15, Status: coupon.StatusUnused, ExpiresAt: now + 30*86400},
		{ID: "CPN002", UserID: u.ID, Title: "平台通用立减券", Amount: 5, MinSpend: 30, Status: coupon.StatusUnused, ExpiresAt: now + 30*86400},
	}
	count := 0
	for i := range rows {
		if _, err := repos.Coupons.FindByID(rows[i].ID); err == nil {
			continue
		}
		if err := repos.Coupons.Save(&rows[i]); err != nil {
			fmt.Println("  ! 优惠券写入失败:", rows[i].ID, err)
			continue
		}
		count++
	}
	if count > 0 {
		fmt.Println("  + 优惠券", count, "张")
	}
}

// seedMerchantApplication 为联调商家补齐入驻资料（仅当历史数据为空）。
func seedMerchantApplication(ctx context.Context, repos *application.RepoUnit) {
	m, err := repos.Merchants.FindByID("M00268")
	if err != nil || m.Application.LicenseNo != "" {
		return
	}
	m.Application = merchant.Application{
		LicenseNo: "91330000DEMO002680", LicenseImage: "",
		LegalName: "陈建国", LegalIDCard: "330100197801010018",
		FoodLicenseNo: "JY330100DEMO00268",
		BankName:      "工商银行", BankAccountName: "陈建国", BankAccountNo: "62220210000000268",
		AccountPhone: m.Phone,
	}
	m.Completeness = 100
	m.SubmittedAt = m.AppliedAt
	m.VerifiedAt = m.JoinedAt
	if err := repos.Merchants.Save(m); err != nil {
		fmt.Println("  ! 商家入驻资料写入失败:", err)
		return
	}
	fmt.Println("  + 商家入驻资料 M00268")
}

// seedReview 联调商家评价（幂等，按订单去重）。
func seedReview(ctx context.Context, repos *application.RepoUnit, now int64) {
	if _, err := repos.Reviews.FindByOrder("SG202610220086"); err == nil {
		return
	}
	if err := repos.Reviews.Save(&review.Review{
		ID: "RV20261022001", OrderID: "SG202610220086", UserID: "U86540",
		MerchantID: "M00268", RiderID: "R00086", Rating: 5,
		Tags:    []string{"味道好", "分量足", "送得快"},
		Content: "菜量很足，汤还是热的，骑手也很快。", CreatedAt: now - 1200,
		Reply: "感谢支持，欢迎常来！", RepliedAt: now - 600,
	}); err != nil {
		fmt.Println("  ! 评价写入失败:", err)
		return
	}
	fmt.Println("  + 商家评价 RV20261022001")
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
