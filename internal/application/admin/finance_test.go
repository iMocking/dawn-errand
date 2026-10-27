package admin

import (
	"context"
	"errors"
	"testing"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/wallet"
	"dawn-errand/internal/pkg/errcode"
)

// R5-5 管理端资金单测（提现审核 + 对账）。R2-3 的修复把「并发双审核」压到仓储的
// 条件认领上，但「认领失败后不能打款」「打款失败要回退」「驳回失败要回退」这三条
// 编排语义在应用层，只有单测能廉价地把每个失败分支都走一遍。

// --- 提现 ---

type fakeWithdrawals struct {
	finance.WithdrawalRepository
	byID        map[string]*finance.Withdrawal
	claimOK     bool
	claimErr    error
	claimStatus []string
	reverts     [][2]string
	payoutRefs  []string
	listCalls   int
	stats       *finance.Stats
}

func newFakeWithdrawals() *fakeWithdrawals {
	return &fakeWithdrawals{byID: map[string]*finance.Withdrawal{}, claimOK: true, stats: &finance.Stats{}}
}

func (f *fakeWithdrawals) FindByID(ctx context.Context, id string) (*finance.Withdrawal, error) {
	w, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return w, nil
}

// ClaimAudit 复刻条件认领：只有「待审核」能被认领，重复审核返回 false。
func (f *fakeWithdrawals) ClaimAudit(ctx context.Context, id string, approved bool, remark, operator string, now int64) (bool, error) {
	if f.claimErr != nil {
		return false, f.claimErr
	}
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

func (f *fakeWithdrawals) RevertAuditToPending(ctx context.Context, id, fromStatus, operator string, now int64) error {
	f.reverts = append(f.reverts, [2]string{id, fromStatus})
	if w, ok := f.byID[id]; ok {
		w.Status = finance.WithdrawPending
	}
	return nil
}

func (f *fakeWithdrawals) MarkPayout(ctx context.Context, id, ref string, now int64) error {
	f.payoutRefs = append(f.payoutRefs, ref)
	return nil
}

func (f *fakeWithdrawals) List(ctx context.Context, q finance.WithdrawQuery, sc tenant.Scope) ([]*finance.Withdrawal, int64, error) {
	f.listCalls++
	out := make([]*finance.Withdrawal, 0, len(f.byID))
	for _, w := range f.byID {
		if q.Status != "" && string(w.Status) != q.Status {
			continue
		}
		if sc.RegionOnly() && w.RegionID != sc.RegionID {
			continue
		}
		out = append(out, w)
	}
	return out, int64(len(out)), nil
}

func (f *fakeWithdrawals) Stats(ctx context.Context, sc tenant.Scope, now int64) (*finance.Stats, error) {
	return f.stats, nil
}

// --- 结算 ---

type fakeSettlements struct {
	finance.SettlementRepository
	stats     *finance.SettlementStats
	listCalls int
}

func (f *fakeSettlements) Stats(ctx context.Context, sc tenant.Scope, now int64) (*finance.SettlementStats, error) {
	if f.stats == nil {
		return &finance.SettlementStats{}, nil
	}
	return f.stats, nil
}

func (f *fakeSettlements) List(ctx context.Context, q finance.SettlementQuery, sc tenant.Scope) ([]*finance.Settlement, int64, error) {
	f.listCalls++
	return nil, 0, nil
}

// --- 钱包 ---

type fakeWallets struct {
	wallet.Repository
	accounts  map[string]*wallet.Account
	adjustOK  bool
	adjustErr error
	txns      []*wallet.Transaction
}

func newFakeWallets() *fakeWallets {
	return &fakeWallets{accounts: map[string]*wallet.Account{}, adjustOK: true}
}

func walletKey(partyType, partyID string) string { return partyType + "|" + partyID }

func (f *fakeWallets) Ensure(ctx context.Context, partyType, partyID string, now int64) (*wallet.Account, error) {
	k := walletKey(partyType, partyID)
	acc, ok := f.accounts[k]
	if !ok {
		acc = &wallet.Account{ID: "WA" + k, PartyType: partyType, PartyID: partyID}
		f.accounts[k] = acc
	}
	return acc, nil
}

func (f *fakeWallets) Adjust(ctx context.Context, partyType, partyID string, delta float64, allowNegative bool, now int64) (bool, *wallet.Account, error) {
	if f.adjustErr != nil {
		return false, nil, f.adjustErr
	}
	acc, err := f.Ensure(ctx, partyType, partyID, now)
	if err != nil {
		return false, nil, err
	}
	if !f.adjustOK {
		return false, nil, nil
	}
	acc.Balance += delta
	return true, acc, nil
}

func (f *fakeWallets) SaveTransaction(ctx context.Context, t *wallet.Transaction) error {
	f.txns = append(f.txns, t)
	return nil
}

// --- 支付对账 ---

type fakePayments struct {
	order.PaymentRepository
	sum     *order.ReconSummary
	lastReg string
	start   int64
	end     int64
}

func (f *fakePayments) SummarizeRecon(ctx context.Context, start, end int64, regionID string) (*order.ReconSummary, error) {
	f.start, f.end, f.lastReg = start, end, regionID
	if f.sum == nil {
		return &order.ReconSummary{}, nil
	}
	return f.sum, nil
}

// --- 审计 ---

type fakeAudit struct {
	entries []application.AuditEntry
}

func (f *fakeAudit) Record(ctx context.Context, e application.AuditEntry) {
	f.entries = append(f.entries, e)
}

// --- 打款通道 ---

type fakePayout struct {
	ref string
	err error
	got []string
}

func (p *fakePayout) Payout(ctx context.Context, withdrawalID, partyType, partyID string, amount float64) (string, error) {
	p.got = append(p.got, withdrawalID)
	return p.ref, p.err
}

// --- 事务 ---

type fakeTx struct {
	unit     *application.RepoUnit
	runCount int
}

func (t *fakeTx) RunInTx(ctx context.Context, fn func(*application.RepoUnit) error) error {
	t.runCount++
	return fn(t.unit)
}

// --- 装配 ---

type moneyUnit struct {
	repos   *application.RepoUnit
	tx      *fakeTx
	withd   *fakeWithdrawals
	settle  *fakeSettlements
	wallets *fakeWallets
	pay     *fakePayments
	audit   *fakeAudit
}

func newMoneyUnit() *moneyUnit {
	u := &moneyUnit{
		withd:   newFakeWithdrawals(),
		settle:  &fakeSettlements{},
		wallets: newFakeWallets(),
		pay:     &fakePayments{},
		audit:   &fakeAudit{},
	}
	u.repos = &application.RepoUnit{
		Withdrawals: u.withd, Settlements: u.settle, Wallets: u.wallets,
		Payments: u.pay, Audit: u.audit,
	}
	u.tx = &fakeTx{unit: u.repos}
	return u
}

func (u *moneyUnit) addWithdrawal(id, regionID string, amount float64) *finance.Withdrawal {
	w := &finance.Withdrawal{
		ID: id, RegionID: regionID, PartyType: "rider", PartyID: "R1",
		PartyName: "小王", Amount: amount, Fee: 0.6, Status: finance.WithdrawPending,
	}
	u.withd.byID[id] = w
	return w
}

func adminCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var e *errcode.Error
	if errors.As(err, &e) {
		return e.Code
	}
	t.Fatalf("非业务错误，无法断言错误码: %v", err)
	return -1
}

func TestAuditWithdrawPassRecordsPayout(t *testing.T) {
	ctx := context.Background()
	u := newMoneyUnit()
	u.addWithdrawal("W1", "RG1", 100)
	payout := &fakePayout{ref: "PO-1"}
	s := &FinanceService{Repos: u.repos, Tx: u.tx, Payout: payout}

	if err := s.AuditWithdraw(ctx, "W1", "pass", "资料齐全", "admin1", tenant.Scope{Role: "super_admin"}); err != nil {
		t.Fatalf("审核通过失败: %v", err)
	}
	if len(payout.got) != 1 || payout.got[0] != "W1" {
		t.Errorf("未发起打款: %v", payout.got)
	}
	if len(u.withd.payoutRefs) != 1 || u.withd.payoutRefs[0] != "PO-1" {
		t.Errorf("未记录打款回执: %v", u.withd.payoutRefs)
	}
	if u.withd.byID["W1"].Status != finance.WithdrawApproved {
		t.Errorf("状态未置为已通过: %q", u.withd.byID["W1"].Status)
	}
	if len(u.audit.entries) != 1 || u.audit.entries[0].Action != "withdraw.audit" {
		t.Errorf("审核未留审计: %+v", u.audit.entries)
	}
	if len(u.wallets.txns) != 0 {
		t.Error("审核通过不应再动钱包账本（申请时已扣减）")
	}

	// 重复审核：状态已变，第二人必须被拦下，且不会二次打款。
	before := len(payout.got)
	if code := adminCodeOf(t, s.AuditWithdraw(ctx, "W1", "pass", "", "admin2", tenant.Scope{Role: "super_admin"})); code != errcode.CodeWithdrawStatus {
		t.Errorf("重复审核应拒, got=%d", code)
	}
	if len(payout.got) != before {
		t.Error("重复审核触发了二次打款")
	}
}

func TestAuditWithdrawGuards(t *testing.T) {
	ctx := context.Background()
	u := newMoneyUnit()
	u.addWithdrawal("W1", "RG1", 100)
	s := &FinanceService{Repos: u.repos, Tx: u.tx, Payout: &fakePayout{ref: "PO"}}

	if code := adminCodeOf(t, s.AuditWithdraw(ctx, "W404", "pass", "", "op", tenant.Scope{Role: "super_admin"})); code != errcode.CodeWithdrawNotFound {
		t.Errorf("提现单不存在应 404, got=%d", code)
	}
	// 区域站长只能审本区域。
	if code := adminCodeOf(t, s.AuditWithdraw(ctx, "W1", "pass", "", "op", tenant.Scope{Role: "region_station", RegionID: "RG2"})); code != errcode.CodeTenantScopeDenied {
		t.Errorf("跨区域审核应 403, got=%d", code)
	}
	// 打款通道未配置：必须在认领前拒绝，否则钱扣了没人打款。
	noPayout := &FinanceService{Repos: u.repos, Tx: u.tx}
	if code := adminCodeOf(t, noPayout.AuditWithdraw(ctx, "W1", "pass", "", "op", tenant.Scope{Role: "super_admin"})); code != errcode.CodePaymentFailed {
		t.Errorf("无打款通道应拒绝通过, got=%d", code)
	}
	if u.withd.byID["W1"].Status != finance.WithdrawPending {
		t.Error("无通道拒绝后状态应保持待审核")
	}
	// 已被并发认领：条件更新 false，不得继续打款。
	u.withd.claimOK = false
	defer func() { u.withd.claimOK = true }()
	if code := adminCodeOf(t, s.AuditWithdraw(ctx, "W1", "pass", "", "op", tenant.Scope{Role: "super_admin"})); code != errcode.CodeWithdrawStatus {
		t.Errorf("认领失败应报已审核, got=%d", code)
	}
	if len(u.withd.payoutRefs) != 0 {
		t.Error("认领失败却记录了打款")
	}
}

func TestAuditWithdrawPayoutFailureReverts(t *testing.T) {
	ctx := context.Background()
	u := newMoneyUnit()
	u.addWithdrawal("W1", "RG1", 100)
	s := &FinanceService{Repos: u.repos, Tx: u.tx, Payout: &fakePayout{err: errors.New("channel timeout")}}

	err := s.AuditWithdraw(ctx, "W1", "pass", "", "op", tenant.Scope{Role: "super_admin"})
	if code := adminCodeOf(t, err); code != errcode.CodePaymentFailed {
		t.Errorf("打款失败应报 payment failed, got=%d", code)
	}
	if len(u.withd.reverts) != 1 || u.withd.reverts[0][1] != string(finance.WithdrawApproved) {
		t.Errorf("打款失败应回退待审核: %v", u.withd.reverts)
	}
	if u.withd.byID["W1"].Status != finance.WithdrawPending {
		t.Errorf("回退后状态应待审核, got=%q", u.withd.byID["W1"].Status)
	}
	if len(u.audit.entries) != 0 {
		t.Error("打款失败不应记成功审计")
	}
}

func TestAuditWithdrawRejectRefundsWallet(t *testing.T) {
	ctx := context.Background()
	u := newMoneyUnit()
	u.addWithdrawal("W1", "RG1", 100)
	s := &FinanceService{Repos: u.repos, Tx: u.tx}

	if err := s.AuditWithdraw(ctx, "W1", "reject", "账号信息有误", "op", tenant.Scope{Role: "super_admin"}); err != nil {
		t.Fatalf("驳回失败: %v", err)
	}
	acc := u.wallets.accounts[walletKey("rider", "R1")]
	if acc == nil || acc.Balance != 100 {
		t.Errorf("驳回未退回钱包余额: %+v", acc)
	}
	if len(u.wallets.txns) != 1 || u.wallets.txns[0].Type != "withdraw_refund" || u.wallets.txns[0].Amount != 100 {
		t.Errorf("驳回未写退款流水: %+v", u.wallets.txns)
	}
	if u.withd.byID["W1"].Status != finance.WithdrawRejected {
		t.Errorf("状态未置为已驳回: %q", u.withd.byID["W1"].Status)
	}

	// 入账失败（Adjust 条件不命中）：必须回退到待审核，让人能重试驳回。
	u2 := newMoneyUnit()
	u2.addWithdrawal("W2", "RG1", 50)
	u2.wallets.adjustOK = false
	s2 := &FinanceService{Repos: u2.repos, Tx: u2.tx}
	err := s2.AuditWithdraw(ctx, "W2", "reject", "x", "op", tenant.Scope{Role: "super_admin"})
	if code := adminCodeOf(t, err); code != errcode.CodeWithdrawStatus {
		t.Errorf("退款失败应报提现状态错误, got=%d", code)
	}
	if len(u2.withd.reverts) != 1 || u2.withd.reverts[0][1] != string(finance.WithdrawRejected) {
		t.Errorf("退款失败应回退待审核: %v", u2.withd.reverts)
	}
	if len(u2.wallets.txns) != 0 {
		t.Error("退款失败仍写了流水")
	}
}

func TestReconciliationScopesAndAggregation(t *testing.T) {
	ctx := context.Background()
	u := newMoneyUnit()
	u.pay.sum = &order.ReconSummary{PaidCount: 3, PaidAmount: 100, PaidOrderCount: 2, OrderAmount: 100.6, MissingOrders: 1, RefundCount: 1, RefundAmount: 9.999}
	s := &FinanceService{Repos: u.repos, Tx: u.tx}

	// 非平台角色（商家/司机/会员）不得看对账。
	for _, role := range []string{"merchant", "driver", "user", ""} {
		if code := adminCodeOf(t, mustReconErr(s.Reconciliation(ctx, "2026-10-26", tenant.Scope{Role: role}))); code != errcode.CodeTenantScopeDenied {
			t.Errorf("%q 角色看对账应 403, got=%d", role, code)
		}
	}
	if code := adminCodeOf(t, mustReconErr(s.Reconciliation(ctx, "2026/10/26", tenant.Scope{Role: "super_admin"}))); code != errcode.CodeBadRequest {
		t.Errorf("非法日期应 400, got=%d", code)
	}

	res, err := s.Reconciliation(ctx, "2026-10-26", tenant.Scope{Role: "super_admin"})
	if err != nil {
		t.Fatalf("对账失败: %v", err)
	}
	if res.PaidOrders != 2 || res.PaymentCount != 3 || res.MissingOrders != 1 {
		t.Errorf("对账计数错误: %+v", res)
	}
	// 金额按分取整后做差，避免浮点噪声把 0.60 算成 0.5999999。
	if res.OrderAmount != 100.6 || res.PaymentAmount != 100 || res.RefundAmount != 10 {
		t.Errorf("对账金额未取整: %+v", res)
	}
	if res.Diff != 0.6 {
		t.Errorf("差额应为 0.6, got=%v", res.Diff)
	}
	if u.pay.lastReg != "" {
		t.Errorf("全局对账不应带区域: %q", u.pay.lastReg)
	}
	// 日期窗口应是 [当日 00:00, 次日 00:00)。
	if u.pay.end-u.pay.start != 24*3600 {
		t.Errorf("对账窗口应为一整天: %d", u.pay.end-u.pay.start)
	}

	// 站长只看本区域：区域条件下推给仓储，而不是取全量再筛。
	if _, err := s.Reconciliation(ctx, "2026-10-26", tenant.Scope{Role: "region_station", RegionID: "RG1"}); err != nil {
		t.Fatalf("站长对账失败: %v", err)
	}
	if u.pay.lastReg != "RG1" {
		t.Errorf("站长对账未下推区域: %q", u.pay.lastReg)
	}
}

func TestFinanceOverviewAndDetailScopes(t *testing.T) {
	ctx := context.Background()
	u := newMoneyUnit()
	u.withd.stats = &finance.Stats{TodayIncome: 100, WithdrawPending: 1}
	u.settle.stats = &finance.SettlementStats{TodaySettled: 50}
	u.addWithdrawal("W1", "RG1", 100)
	u.addWithdrawal("W2", "RG2", 200)
	s := &FinanceService{Repos: u.repos, Tx: u.tx}

	ov, err := s.Get(ctx, tenant.Scope{Role: "region_station", RegionID: "RG1"})
	if err != nil {
		t.Fatalf("财务总览失败: %v", err)
	}
	if ov.Stats.TodayIncome != 100 || ov.SettlementStats.TodaySettled != 50 {
		t.Errorf("总览统计未透传: %+v", ov)
	}
	if ov.WithdrawTotal != 1 || len(ov.Withdraws) != 1 || ov.Withdraws[0].ID != "W1" {
		t.Errorf("总览列表未按区域隔离: %+v", ov.Withdraws)
	}

	if code := adminCodeOf(t, mustWithdrawErr(s.WithdrawDetail(ctx, "W404", tenant.Scope{Role: "super_admin"}))); code != errcode.CodeWithdrawNotFound {
		t.Errorf("不存在应 404, got=%d", code)
	}
	if code := adminCodeOf(t, mustWithdrawErr(s.WithdrawDetail(ctx, "W2", tenant.Scope{Role: "region_station", RegionID: "RG1"}))); code != errcode.CodeTenantScopeDenied {
		t.Errorf("跨区域详情应 403, got=%d", code)
	}
	w, err := s.WithdrawDetail(ctx, "W1", tenant.Scope{Role: "region_station", RegionID: "RG1"})
	if err != nil {
		t.Fatalf("本区域详情失败: %v", err)
	}
	if w.ID != "W1" {
		t.Errorf("详情错误: %+v", w)
	}

	// 列表分页参数归一（0 值不交给仓储当 LIMIT 0）。
	if _, _, err := s.WithdrawList(ctx, finance.WithdrawQuery{}, tenant.Scope{Role: "super_admin"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.SettlementList(ctx, finance.SettlementQuery{Page: -1, PageSize: 999}, tenant.Scope{Role: "super_admin"}); err != nil {
		t.Fatal(err)
	}
}

func mustReconErr(_ *Reconciliation, err error) error        { return err }
func mustWithdrawErr(_ *finance.Withdrawal, err error) error { return err }
