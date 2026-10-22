package repo

import (
	"context"
	"encoding/json"
	"strings"

	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/infrastructure/ent"
	entsettlement "dawn-errand/internal/infrastructure/ent/settlement"
	entwithdrawal "dawn-errand/internal/infrastructure/ent/withdrawal"
)

func jsonUnmarshal(b []byte, v any) error {
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, v)
}

// WithdrawalRepo 实现 finance.WithdrawalRepository。
type WithdrawalRepo struct{ client *ent.Client }

func (r WithdrawalRepo) Save(w *finance.Withdrawal) error {
	ctx := context.Background()
	if w.ID == "" {
		w.ID = newULID()
	}
	err := r.client.Withdrawal.Create().
		SetID(w.ID).SetRegionID(w.RegionID).
		SetPartyType(w.PartyType).SetPartyName(w.PartyName).SetPartyNo(w.PartyNo).
		SetBankInfo(w.BankInfo).SetAmount(w.Amount).SetFee(w.Fee).SetFeeNote(w.FeeNote).
		SetStatus(string(w.Status)).SetKind(w.Kind).SetApplicant(w.Applicant).
		SetAccountHolder(w.AccountHolder).SetWithdrawType(w.WithdrawType).SetArriveExpect(w.ArriveExpect).
		SetPurpose(w.Purpose).SetBalance(w.Balance).SetRemainBalance(w.RemainBalance).
		SetMonthOrders(w.MonthOrders).SetMonthIncome(w.MonthIncome).
		SetRiskItems(w.RiskItems).SetRiskWarning(w.RiskWarning).SetRejectReason(w.RejectReason).
		SetCreatedAt(w.CreatedAt).SetTimeline(mustJSON(w.Timeline)).
		Exec(ctx)
	return err
}

func toWithdrawal(e *ent.Withdrawal) *finance.Withdrawal {
	if e == nil {
		return nil
	}
	w := &finance.Withdrawal{
		ID: e.ID, RegionID: e.RegionID,
		PartyType: e.PartyType, PartyName: e.PartyName, PartyNo: e.PartyNo,
		BankInfo: e.BankInfo, Amount: e.Amount, Fee: e.Fee, FeeNote: e.FeeNote,
		Status: finance.WithdrawStatus(e.Status), Kind: e.Kind, Applicant: e.Applicant,
		AccountHolder: e.AccountHolder, WithdrawType: e.WithdrawType, ArriveExpect: e.ArriveExpect,
		Purpose: e.Purpose, Balance: e.Balance, RemainBalance: e.RemainBalance,
		MonthOrders: e.MonthOrders, MonthIncome: e.MonthIncome,
		RiskItems: e.RiskItems, RiskWarning: e.RiskWarning, RejectReason: e.RejectReason,
		CreatedAt: e.CreatedAt,
	}
	_ = jsonUnmarshal(e.Timeline, &w.Timeline)
	return w
}

func (r WithdrawalRepo) FindByID(id string) (*finance.Withdrawal, error) {
	e, err := r.client.Withdrawal.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toWithdrawal(e), nil
}

func (r WithdrawalRepo) List(q finance.WithdrawQuery, sc tenant.Scope) ([]*finance.Withdrawal, int64, error) {
	ctx := context.Background()
	bq := r.client.Withdrawal.Query()
	if sc.RegionOnly() {
		bq.Where(entwithdrawal.RegionID(sc.RegionID))
	}
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entwithdrawal.Status(s))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(entwithdrawal.Or(
			entwithdrawal.PartyNameContainsFold(k),
			entwithdrawal.IDContainsFold(k),
			entwithdrawal.ApplicantContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entwithdrawal.FieldCreatedAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*finance.Withdrawal, 0, len(rows))
	for _, e := range rows {
		out = append(out, toWithdrawal(e))
	}
	return out, int64(total), nil
}

func (r WithdrawalRepo) Audit(id string, approved bool, remark, operator string, now int64) error {
	ctx := context.Background()
	status := string(finance.WithdrawApproved)
	if !approved {
		status = string(finance.WithdrawRejected)
	}
	cur, err := r.FindByID(id)
	if err != nil {
		return err
	}
	event := finance.Event{Time: now, Text: "审核通过（打款）", Operator: operator}
	if !approved {
		event.Text = "驳回：" + remark
	}
	cur.Timeline = append(cur.Timeline, event)
	err = r.client.Withdrawal.UpdateOneID(id).
		SetStatus(status).SetRejectReason(remark).SetTimeline(mustJSON(cur.Timeline)).
		Exec(ctx)
	return err
}

func (r WithdrawalRepo) Stats(sc tenant.Scope, now int64) (*finance.Stats, error) {
	ctx := context.Background()
	dayStart := startOfDayUTC(now)
	base := func() *ent.WithdrawalQuery {
		bq := r.client.Withdrawal.Query()
		if sc.RegionOnly() {
			bq.Where(entwithdrawal.RegionID(sc.RegionID))
		}
		return bq
	}
	today, err := sumFloat64(ctx, base().Where(entwithdrawal.CreatedAtGTE(dayStart)), entwithdrawal.FieldAmount)
	if err != nil {
		return nil, err
	}
	yesterday, err := sumFloat64(ctx, base().
		Where(entwithdrawal.CreatedAtGTE(dayStart-86400), entwithdrawal.CreatedAtLT(dayStart)), entwithdrawal.FieldAmount)
	if err != nil {
		return nil, err
	}
	pendingBalance, err := sumFloat64(ctx, base().
		Where(entwithdrawal.Status(string(finance.WithdrawPending))), entwithdrawal.FieldAmount)
	if err != nil {
		return nil, err
	}
	pendingCount, err := base().Where(entwithdrawal.Status(string(finance.WithdrawPending))).Count(ctx)
	if err != nil {
		return nil, err
	}
	// 异常单：金额 0 或账户信息缺失（演示口径）
	abnormal, err := base().Where(entwithdrawal.AmountLTE(0)).Count(ctx)
	if err != nil {
		return nil, err
	}
	abnormalAmt, err := sumFloat64(ctx, base().Where(entwithdrawal.AmountLTE(0)), entwithdrawal.FieldAmount)
	if err != nil {
		return nil, err
	}
	return &finance.Stats{
		TodayIncome: today, YesterdayIncome: yesterday, PendingBalance: pendingBalance,
		WithdrawPending: int64(pendingCount), AbnormalBills: int64(abnormal), AbnormalAmount: abnormalAmt,
	}, nil
}

func (r WithdrawalRepo) CountPending() (int64, error) {
	n, err := r.client.Withdrawal.Query().
		Where(entwithdrawal.Status(string(finance.WithdrawPending))).
		Count(context.Background())
	return int64(n), err
}

// SettlementRepo 实现 finance.SettlementRepository。
type SettlementRepo struct{ client *ent.Client }

func (r SettlementRepo) Save(s *finance.Settlement) error {
	ctx := context.Background()
	if s.ID == "" {
		s.ID = newULID()
	}
	err := r.client.Settlement.Create().
		SetID(s.ID).SetRegionID(s.RegionID).SetParty(s.Party).
		SetType(string(s.Type)).SetAmount(s.Amount).SetStatus(string(s.Status)).
		SetPeriod(s.Period).SetSettledAt(s.SettledAt).SetAction(s.Action).
		Exec(ctx)
	return err
}

func toSettlement(e *ent.Settlement) *finance.Settlement {
	if e == nil {
		return nil
	}
	return &finance.Settlement{
		ID: e.ID, RegionID: e.RegionID, Party: e.Party,
		Type: finance.SettlementType(e.Type), Amount: e.Amount,
		Status: finance.SettlementStatus(e.Status), Period: e.Period,
		SettledAt: e.SettledAt, Action: e.Action,
	}
}

func (r SettlementRepo) List(q finance.SettlementQuery, sc tenant.Scope) ([]*finance.Settlement, int64, error) {
	ctx := context.Background()
	bq := r.client.Settlement.Query()
	if sc.RegionOnly() {
		bq.Where(entsettlement.RegionID(sc.RegionID))
	}
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entsettlement.Status(s))
	}
	if t := strings.TrimSpace(q.Type); t != "" {
		bq.Where(entsettlement.Type(t))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(entsettlement.Or(
			entsettlement.PartyContainsFold(k),
			entsettlement.IDContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entsettlement.FieldSettledAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*finance.Settlement, 0, len(rows))
	for _, e := range rows {
		out = append(out, toSettlement(e))
	}
	return out, int64(total), nil
}

func (r SettlementRepo) Stats(sc tenant.Scope, now int64) (*finance.SettlementStats, error) {
	ctx := context.Background()
	dayStart := startOfDayUTC(now)
	weekStart := daysAgoStart(now, 7)
	monthStart := daysAgoStart(now, 30)
	base := func() *ent.SettlementQuery {
		bq := r.client.Settlement.Query()
		if sc.RegionOnly() {
			bq.Where(entsettlement.RegionID(sc.RegionID))
		}
		return bq
	}
	sumOf := func(q *ent.SettlementQuery) (float64, error) {
		return sumFloat64(ctx, q, entsettlement.FieldAmount)
	}
	today, err := sumOf(base().Where(entsettlement.SettledAtGTE(dayStart)))
	if err != nil {
		return nil, err
	}
	yesterday, err := sumOf(base().
		Where(entsettlement.SettledAtGTE(dayStart-86400), entsettlement.SettledAtLT(dayStart)))
	if err != nil {
		return nil, err
	}
	week, err := sumOf(base().Where(entsettlement.SettledAtGTE(weekStart)))
	if err != nil {
		return nil, err
	}
	month, err := sumOf(base().Where(entsettlement.SettledAtGTE(monthStart)))
	if err != nil {
		return nil, err
	}
	pending, err := sumFloat64(ctx, base().Where(entsettlement.Status(string(finance.SettlementProcessing))), entsettlement.FieldAmount)
	if err != nil {
		return nil, err
	}
	pendingCount, err := base().Where(entsettlement.Status(string(finance.SettlementProcessing))).Count(ctx)
	if err != nil {
		return nil, err
	}
	return &finance.SettlementStats{
		TodaySettled: today, YesterdaySettled: yesterday,
		WeekTotal: week, MonthTotal: month,
		PendingAmount: pending, PendingCount: int64(pendingCount),
	}, nil
}
