package app

import (
	"context"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/catalog"
	"dawn-errand/internal/domain/coupon"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/notification"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/refund"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/system"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/userprofile"
	"dawn-errand/internal/domain/wallet"
)

// 单元测试替身：只实现被测用例真正走到的方法，其余端口以「嵌入接口 + 空实现」占位，
// 被测代码一旦越界调用未实现方法就会 panic，反而能暴露隐式依赖。
// 条件更新的语义（余额/库存/状态不足则 affected=0）在这里逐条复刻仓储实现，
// 目的是验证应用层是否依赖条件更新而不是「先查再写」，与真实 SQL 行为对齐。

// --- 事务：直接执行，不做回滚（应用层的事务边界由 calls 计数验证）---

type fakeTx struct {
	unit     *application.RepoUnit
	runCount int
}

func (t *fakeTx) RunInTx(ctx context.Context, fn func(*application.RepoUnit) error) error {
	t.runCount++
	return fn(t.unit)
}

// --- 钱包 ---

type fakeWallet struct {
	wallet.Repository
	accounts map[string]*wallet.Account
	txns     []*wallet.Transaction
	adjustOK *bool // 非 nil 时强制覆盖 Adjust 的成功判定，用于模拟并发失败
	err      error
}

func newFakeWallet() *fakeWallet {
	return &fakeWallet{accounts: map[string]*wallet.Account{}}
}

func walletKey(partyType, partyID string) string { return partyType + "|" + partyID }

func (f *fakeWallet) Ensure(ctx context.Context, partyType, partyID string, now int64) (*wallet.Account, error) {
	if f.err != nil {
		return nil, f.err
	}
	k := walletKey(partyType, partyID)
	acc, ok := f.accounts[k]
	if !ok {
		acc = &wallet.Account{ID: "WA" + k, PartyType: partyType, PartyID: partyID, Balance: 0, UpdatedAt: now}
		f.accounts[k] = acc
	}
	return acc, nil
}

func (f *fakeWallet) Find(ctx context.Context, partyType, partyID string) (*wallet.Account, error) {
	acc, ok := f.accounts[walletKey(partyType, partyID)]
	if !ok {
		return nil, application.ErrNotFound
	}
	return acc, nil
}

// Adjust 复刻 WalletRepo.Adjust：负向且不允许透支时附加 balance >= -delta 条件，
// 条件不命中即 affected=0 返回 (false, nil, nil)。
func (f *fakeWallet) Adjust(ctx context.Context, partyType, partyID string, delta float64, allowNegative bool, now int64) (bool, *wallet.Account, error) {
	if f.err != nil {
		return false, nil, f.err
	}
	acc, err := f.Ensure(ctx, partyType, partyID, now)
	if err != nil {
		return false, nil, err
	}
	if delta < 0 && !allowNegative && acc.Balance+delta < -0.005 {
		return false, nil, nil
	}
	acc.Balance = round2(acc.Balance + delta)
	acc.UpdatedAt = now
	if f.adjustOK != nil && !*f.adjustOK {
		return false, nil, nil
	}
	return true, acc, nil
}

func (f *fakeWallet) SaveTransaction(ctx context.Context, t *wallet.Transaction) error {
	if f.err != nil {
		return f.err
	}
	if t.ID == "" {
		t.ID = "WT" + string(rune('a'+len(f.txns)))
	}
	f.txns = append(f.txns, t)
	return nil
}

// --- 用户资料余额 ---

type fakeProfiles struct {
	userprofile.ProfileRepository
	balances map[string]float64
	missing  map[string]bool // FindByUserID 返回 ErrNotFound 的用户
	err      error
	adds     []string // 记录 "userID:delta:min"，用于断言未发生 Find-then-Save
}

func newFakeProfiles() *fakeProfiles {
	return &fakeProfiles{balances: map[string]float64{}, missing: map[string]bool{}}
}

func (f *fakeProfiles) FindByUserID(ctx context.Context, userID string) (*userprofile.Profile, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.missing[userID] {
		return nil, application.ErrNotFound
	}
	return &userprofile.Profile{UserID: userID, Balance: f.balances[userID]}, nil
}

// AddBalance 复刻 ProfileRepo.AddBalance：minBalance>=0 时附加 balance>=minBalance 条件。
func (f *fakeProfiles) AddBalance(ctx context.Context, userID string, delta, minBalance float64, now int64) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.missing[userID] {
		return 0, nil
	}
	cur := f.balances[userID]
	if minBalance >= 0 && cur < minBalance-0.005 {
		return 0, nil
	}
	f.balances[userID] = round2(cur + delta)
	f.adds = append(f.adds, userID)
	return 1, nil
}

// --- 优惠券 ---

type fakeCoupons struct {
	coupon.Repository
	byID     map[string]*coupon.Coupon
	markOK   bool
	released []string
}

func newFakeCoupons() *fakeCoupons {
	return &fakeCoupons{byID: map[string]*coupon.Coupon{}, markOK: true}
}

func (f *fakeCoupons) FindByID(ctx context.Context, id string) (*coupon.Coupon, error) {
	c, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return c, nil
}

func (f *fakeCoupons) MarkUsed(ctx context.Context, id, userID, orderID string, now int64) (bool, error) {
	c, ok := f.byID[id]
	if !ok || !f.markOK || c.Status != coupon.StatusUnused {
		return false, nil
	}
	c.Status = coupon.StatusUsed
	c.UsedOrderID = orderID
	c.UserID = userID
	return true, nil
}

func (f *fakeCoupons) ReleaseByOrder(ctx context.Context, orderID string, now int64) error {
	f.released = append(f.released, orderID)
	return nil
}

// --- 订单 ---

type fakeOrders struct {
	order.OrderRepository
	byID     map[string]*order.Order
	updateOK bool
	updates  []string // "id:from->to"
	assignOK bool
	saved    []*order.Order

	// List/ListPool 记录最后一次入参：应用层必须把租户 scope 与 Tab 状态
	// 下推给仓储，而不是取全量回内存筛（R3-5 的默认拒绝就靠这层传参）。
	lastQuery order.Query
	lastScope tenant.Scope
	lastPool  string // regionID
}

func newFakeOrders() *fakeOrders {
	return &fakeOrders{byID: map[string]*order.Order{}, updateOK: true, assignOK: true}
}

func (f *fakeOrders) Save(ctx context.Context, o *order.Order) error {
	f.saved = append(f.saved, o)
	f.byID[o.ID] = o
	return nil
}

func (f *fakeOrders) FindByID(ctx context.Context, id string) (*order.Order, error) {
	o, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return o, nil
}

// UpdateStatus 复刻条件流转：from 不匹配或显式关闭时返回 false（并发抢占），
// 命中则同时按 patch 写入非零字段，与仓储层保持一致。
func (f *fakeOrders) UpdateStatus(ctx context.Context, id string, from, to order.OrderStatus, patch order.StatusPatch, now int64) (bool, error) {
	o, ok := f.byID[id]
	if !ok {
		return false, nil
	}
	if !f.updateOK || o.Status != from {
		return false, nil
	}
	o.Status = to
	applyPatch(o, patch, now)
	f.updates = append(f.updates, string(from)+"->"+string(to))
	return true, nil
}

func applyPatch(o *order.Order, p order.StatusPatch, now int64) {
	if p.PaymentStatus != "" {
		o.PaymentStatus = p.PaymentStatus
	}
	if p.PayMethod != "" {
		o.PayMethod = p.PayMethod
	}
	if p.FundFlow != "" {
		o.FundFlow = p.FundFlow
	}
	if p.RiderID != "" {
		o.RiderID = p.RiderID
	}
	if p.RiderName != "" {
		o.RiderName = p.RiderName
	}
	if p.PickupCode != "" {
		o.PickupCode = p.PickupCode
	}
	if p.DeliveryCode != "" {
		o.DeliveryCode = p.DeliveryCode
	}
	if p.CancelReason != "" {
		o.CancelReason = p.CancelReason
	}
	if p.AcceptedAt != 0 {
		o.AcceptedAt = p.AcceptedAt
	}
	if p.ReadyAt != 0 {
		o.ReadyAt = p.ReadyAt
	}
	if p.PickedAt != 0 {
		o.PickedAt = p.PickedAt
	}
	if p.DeliveredAt != 0 {
		o.DeliveredAt = p.DeliveredAt
	}
	if p.ConfirmedAt != 0 {
		o.ConfirmedAt = p.ConfirmedAt
	}
	if p.CanceledAt != 0 {
		o.CanceledAt = p.CanceledAt
	}
	if p.RiderEarning != nil {
		o.RiderEarning = *p.RiderEarning
	}
	_ = now
}

// AssignRider 复刻抢单互斥：仅「待取货且未分配骑手」可占单，并发第二人返回 false。
func (f *fakeOrders) AssignRider(ctx context.Context, id, riderID, riderName string, now int64) (bool, error) {
	o, ok := f.byID[id]
	if !ok || o.Status != order.StatusReady || o.RiderID != "" {
		return false, nil
	}
	o.RiderID, o.RiderName = riderID, riderName
	return true, nil
}

// List 复刻列表查询：按主体归属与状态过滤，并保留入参供断言。
func (f *fakeOrders) List(ctx context.Context, q order.Query, scope tenant.Scope) ([]*order.Order, int64, error) {
	f.lastQuery, f.lastScope = q, scope
	out := make([]*order.Order, 0, len(f.byID))
	for _, o := range f.byID {
		if q.UserID != "" && o.UserID != q.UserID {
			continue
		}
		if scope.MerchantID != "" && o.MerchantID != scope.MerchantID {
			continue
		}
		if scope.UserID != "" && o.UserID != scope.UserID {
			continue
		}
		if scope.RiderID != "" && o.RiderID != scope.RiderID {
			continue
		}
		if len(q.Statuses) > 0 && !containsStr(q.Statuses, string(o.Status)) {
			continue
		}
		out = append(out, o)
	}
	return out, int64(len(out)), nil
}

// ListPool 复刻抢单池：本区域待取货且未分配骑手。
func (f *fakeOrders) ListPool(ctx context.Context, regionID string, page, pageSize int) ([]*order.Order, int64, error) {
	f.lastPool = regionID
	out := make([]*order.Order, 0, len(f.byID))
	for _, o := range f.byID {
		if o.RegionID != regionID || o.Status != order.StatusReady || o.RiderID != "" {
			continue
		}
		out = append(out, o)
	}
	return out, int64(len(out)), nil
}

func containsStr(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// FindByClientOrderNo 复刻幂等回读：同用户同客户端单号命中已有订单。
func (f *fakeOrders) FindByClientOrderNo(ctx context.Context, userID, clientOrderNo string) (*order.Order, error) {
	for _, o := range f.byID {
		if o.UserID == userID && o.ClientOrderNo == clientOrderNo {
			return o, nil
		}
	}
	return nil, application.ErrNotFound
}

// --- 用户 ---

type fakeUsers struct {
	identity.UserRepository
	byID map[string]*identity.User
}

func (f *fakeUsers) FindByID(ctx context.Context, id string) (*identity.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return u, nil
}

func (f *fakeUsers) FindByPhone(ctx context.Context, phone string) (*identity.User, error) {
	for _, u := range f.byID {
		if u.Phone == phone {
			return u, nil
		}
	}
	return nil, application.ErrNotFound
}

// --- 区域 ---

type fakeRegions struct {
	tenant.RegionRepository
	byID map[string]*tenant.Region
}

func (f *fakeRegions) FindByID(ctx context.Context, id string) (*tenant.Region, error) {
	r, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return r, nil
}

// --- 订单事件 ---

type fakeEvents struct {
	order.EventRepository
	saved []*order.Event
}

func (f *fakeEvents) Save(ctx context.Context, e *order.Event) error {
	f.saved = append(f.saved, e)
	return nil
}

func (f *fakeEvents) ListByOrder(ctx context.Context, orderID string) ([]*order.Event, error) {
	out := make([]*order.Event, 0, len(f.saved))
	for _, e := range f.saved {
		if e.OrderID == orderID {
			out = append(out, e)
		}
	}
	return out, nil
}

// --- 支付单 ---

type fakePayments struct {
	order.PaymentRepository
	byOrder  map[string]*order.Payment
	saved    []*order.Payment
	statuses []string
}

func newFakePayments() *fakePayments {
	return &fakePayments{byOrder: map[string]*order.Payment{}}
}

func (f *fakePayments) Save(ctx context.Context, p *order.Payment) error {
	if p.ID == "" {
		p.ID = "P" + p.OrderID
	}
	f.byOrder[p.OrderID] = p
	f.saved = append(f.saved, p)
	return nil
}

func (f *fakePayments) FindByOrder(ctx context.Context, orderID string) (*order.Payment, error) {
	p, ok := f.byOrder[orderID]
	if !ok {
		return nil, application.ErrNotFound
	}
	return p, nil
}

func (f *fakePayments) UpdateStatus(ctx context.Context, id string, status order.PaymentStatus, tradeNo string, paidAt int64) error {
	f.statuses = append(f.statuses, string(status))
	for _, p := range f.byOrder {
		if p.ID == id {
			p.Status = status
		}
	}
	return nil
}

// --- 退款 ---

type fakeRefunds struct {
	refund.Repository
	byOrder map[string]*refund.Refund
	updates []refund.Status
}

func newFakeRefunds() *fakeRefunds {
	return &fakeRefunds{byOrder: map[string]*refund.Refund{}}
}

func (f *fakeRefunds) Save(ctx context.Context, r *refund.Refund) error {
	f.byOrder[r.OrderID] = r
	return nil
}

func (f *fakeRefunds) FindByOrder(ctx context.Context, orderID string) (*refund.Refund, error) {
	r, ok := f.byOrder[orderID]
	if !ok {
		return nil, application.ErrNotFound
	}
	return r, nil
}

func (f *fakeRefunds) Update(ctx context.Context, id string, status refund.Status, operator, rejectReason string, handledAt int64) error {
	f.updates = append(f.updates, status)
	for _, r := range f.byOrder {
		if r.ID == id {
			r.Status = status
			r.RejectReason = rejectReason
		}
	}
	return nil
}

// --- 商家 / 骑手 ---

type fakeMerchants struct {
	merchant.MerchantRepository
	byID map[string]*merchant.Merchant
}

func (f *fakeMerchants) FindByID(ctx context.Context, id string) (*merchant.Merchant, error) {
	m, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return m, nil
}

type fakeRiders struct {
	rider.RiderRepository
	byID      map[string]*rider.Rider
	updateErr error
	statuses  []rider.RiderStatus
}

func (f *fakeRiders) FindByID(ctx context.Context, id string) (*rider.Rider, error) {
	r, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return r, nil
}

func (f *fakeRiders) UpdateStatus(ctx context.Context, id string, status rider.RiderStatus, now int64) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	r, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	r.Status = status
	f.statuses = append(f.statuses, status)
	return nil
}

// --- 库存 ---

type fakeCatalog struct {
	catalog.Repository
	stocks map[string]int
	calls  []int
}

func (f *fakeCatalog) AdjustStock(ctx context.Context, dishID string, delta int) (bool, error) {
	f.calls = append(f.calls, delta)
	cur := f.stocks[dishID]
	if delta < 0 && cur+delta < 0 {
		return false, nil
	}
	f.stocks[dishID] = cur + delta
	return true, nil
}

// --- 通知 ---

type fakeNotifications struct {
	notification.Repository
	saved []*notification.Notification
}

func (f *fakeNotifications) Save(ctx context.Context, n *notification.Notification) error {
	f.saved = append(f.saved, n)
	return nil
}

// --- 系统配置 ---

type fakeSettings struct {
	system.SettingRepository
	valueJSON map[string]string
	err       error
}

func (f *fakeSettings) Get(ctx context.Context, key system.SettingKey) (*system.SystemSetting, error) {
	if f.err != nil {
		return nil, f.err
	}
	v, ok := f.valueJSON[string(key)]
	if !ok {
		return nil, application.ErrNotFound
	}
	return &system.SystemSetting{Key: key, ValueJSON: v}, nil
}

// --- 提现单 ---

type fakeWithdrawals struct {
	finance.WithdrawalRepository
	saved       []*finance.Withdrawal
	byID        map[string]*finance.Withdrawal
	countSince  int64
	pendingSum  float64
	claimOK     bool
	claimStatus []string
}

func newFakeWithdrawals() *fakeWithdrawals {
	return &fakeWithdrawals{byID: map[string]*finance.Withdrawal{}, claimOK: true}
}

func (f *fakeWithdrawals) Save(ctx context.Context, w *finance.Withdrawal) error {
	f.saved = append(f.saved, w)
	f.byID[w.ID] = w
	return nil
}

func (f *fakeWithdrawals) FindByID(ctx context.Context, id string) (*finance.Withdrawal, error) {
	w, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return w, nil
}

func (f *fakeWithdrawals) CountByPartySince(ctx context.Context, partyType, partyID string, since int64) (int64, error) {
	return f.countSince, nil
}

func (f *fakeWithdrawals) SumPendingByParty(ctx context.Context, partyType, partyID string) (float64, error) {
	return f.pendingSum, nil
}

// ClaimAudit 复刻条件认领：仅「待审核」可流转，重复审核返回 false。
func (f *fakeWithdrawals) ClaimAudit(ctx context.Context, id string, approved bool, remark, operator string, now int64) (bool, error) {
	w, ok := f.byID[id]
	if !ok || !f.claimOK || w.Status != finance.WithdrawPending {
		return false, nil
	}
	if approved {
		w.Status = finance.WithdrawApproved
	} else {
		w.Status = finance.WithdrawRejected
	}
	f.claimStatus = append(f.claimStatus, string(w.Status))
	return true, nil
}

// --- 结算 ---

type fakeSettlements struct {
	finance.SettlementRepository
	pending float64
	saved   []*finance.Settlement
}

func (f *fakeSettlements) SumPendingByParty(ctx context.Context, partyType, partyID string) (float64, error) {
	return f.pending, nil
}

func (f *fakeSettlements) Save(ctx context.Context, s *finance.Settlement) error {
	f.saved = append(f.saved, s)
	return nil
}

// --- 审计 ---

type fakeAudit struct {
	entries []application.AuditEntry
}

func (f *fakeAudit) Record(ctx context.Context, e application.AuditEntry) {
	f.entries = append(f.entries, e)
}

// --- 装配 ---

type unit struct {
	repos   *application.RepoUnit
	tx      *fakeTx
	users   *fakeUsers
	region  *fakeRegions
	wallet  *fakeWallet
	prof    *fakeProfiles
	coupon  *fakeCoupons
	orders  *fakeOrders
	refund  *fakeRefunds
	withd   *fakeWithdrawals
	settle  *fakeSettlements
	audit   *fakeAudit
	set     *fakeSettings
	merch   *fakeMerchants
	riders  *fakeRiders
	catalog *fakeCatalog
	pay     *fakePayments
	noti    *fakeNotifications
	evts    *fakeEvents
}

func newUnit() *unit {
	u := &unit{
		wallet:  newFakeWallet(),
		prof:    newFakeProfiles(),
		coupon:  newFakeCoupons(),
		orders:  newFakeOrders(),
		refund:  newFakeRefunds(),
		withd:   newFakeWithdrawals(),
		settle:  &fakeSettlements{},
		audit:   &fakeAudit{},
		set:     &fakeSettings{valueJSON: map[string]string{}},
		merch:   &fakeMerchants{byID: map[string]*merchant.Merchant{}},
		catalog: &fakeCatalog{stocks: map[string]int{}},
		pay:     newFakePayments(),
		noti:    &fakeNotifications{},
		evts:    &fakeEvents{},
		riders:  &fakeRiders{byID: map[string]*rider.Rider{}},
		users:   &fakeUsers{byID: map[string]*identity.User{}},
		region:  &fakeRegions{byID: map[string]*tenant.Region{}},
	}
	u.repos = &application.RepoUnit{
		Users: u.users, Regions: u.region,
		Wallets: u.wallet, Profiles: u.prof, Coupons: u.coupon, Orders: u.orders,
		Refunds: u.refund, Withdrawals: u.withd, Settings: u.set, Merchants: u.merch,
		Catalog: u.catalog, Payments: u.pay, Notifications: u.noti, OrderEvents: u.evts,
		Audit: u.audit, Riders: u.riders, Settlements: u.settle,
	}
	u.tx = &fakeTx{unit: u.repos}
	return u
}
