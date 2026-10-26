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

func (r WithdrawalRepo) Save(ctx context.Context, w *finance.Withdrawal) error {
	if w.ID == "" {
		w.ID = newULID()
	}
	err := r.client.Withdrawal.Create().
		SetID(w.ID).SetRegionID(w.RegionID).
		SetPartyType(w.PartyType).SetPartyID(w.PartyID).SetPartyName(w.PartyName).SetPartyNo(w.PartyNo).
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
		PartyType: e.PartyType, PartyID: e.PartyID, PartyName: e.PartyName, PartyNo: e.PartyNo,
		BankInfo: e.BankInfo, Amount: e.Amount, Fee: e.Fee, FeeNote: e.FeeNote,
		Status: finance.WithdrawStatus(e.Status), Kind: e.Kind, Applicant: e.Applicant,
		AccountHolder: e.AccountHolder, WithdrawType: e.WithdrawType, ArriveExpect: e.ArriveExpect,
		Purpose: e.Purpose, Balance: e.Balance, RemainBalance: e.RemainBalance,
		MonthOrders: e.MonthOrders, MonthIncome: e.MonthIncome,
		RiskItems: e.RiskItems, RiskWarning: e.RiskWarning, RejectReason: e.RejectReason,
		CreatedAt: e.CreatedAt,
	}
	_ = jsonUnmarshal(e.Timeline, &w.Timeline)
	for _, ev := range w.Timeline {
		if strings.HasPrefix(ev.Text, "打款成功") {
			w.PayoutRef = ev.Operator
			w.PayoutAt = ev.Time
		}
	}
	return w
}

func (r WithdrawalRepo) FindByID(ctx context.Context, id string) (*finance.Withdrawal, error) {
	e, err := r.client.Withdrawal.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toWithdrawal(e), nil
}

func (r WithdrawalRepo) List(ctx context.Context, q finance.WithdrawQuery, sc tenant.Scope) ([]*finance.Withdrawal, int64, error) {
	bq := r.client.Withdrawal.Query()
	if sc.RegionOnly() {
		bq.Where(entwithdrawal.RegionID(sc.RegionID))
	}
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entwithdrawal.Status(s))
	}
	if q.PartyType != "" {
		bq.Where(entwithdrawal.PartyType(q.PartyType))
	}
	if q.PartyID != "" {
		bq.Where(entwithdrawal.PartyID(q.PartyID))
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

// ClaimAudit 条件认领审核：UPDATE ... WHERE status=待审核，命中 0 行说明已被并发审核。
func (r WithdrawalRepo) ClaimAudit(ctx context.Context, id string, approved bool, remark, operator string, now int64) (bool, error) {
	cur, err := r.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	status := string(finance.WithdrawApproved)
	event := finance.Event{Time: now, Text: "审核通过（打款）", Operator: operator}
	if !approved {
		status = string(finance.WithdrawRejected)
		event.Text = "驳回：" + remark
	}
	cur.Timeline = append(cur.Timeline, event)
	affected, err := r.client.Withdrawal.Update().
		Where(entwithdrawal.ID(id), entwithdrawal.Status(string(finance.WithdrawPending))).
		SetStatus(status).SetRejectReason(remark).SetTimeline(mustJSON(cur.Timeline)).
		Save(ctx)
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

// RevertAuditToPending 回退状态供重试（仅当前状态为 fromStatus 时生效，回执留痕）。
func (r WithdrawalRepo) RevertAuditToPending(ctx context.Context, id, fromStatus, operator string, now int64) error {
	cur, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	cur.Timeline = append(cur.Timeline, finance.Event{Time: now, Text: "回退待审核（重试）", Operator: operator})
	return r.client.Withdrawal.Update().
		Where(entwithdrawal.ID(id), entwithdrawal.Status(fromStatus)).
		SetStatus(string(finance.WithdrawPending)).SetTimeline(mustJSON(cur.Timeline)).
		Exec(ctx)
}

func (r WithdrawalRepo) MarkPayout(ctx context.Context, id, ref string, now int64) error {
	cur, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	cur.Timeline = append(cur.Timeline, finance.Event{Time: now, Text: "打款成功", Operator: ref})
	return r.client.Withdrawal.UpdateOneID(id).
		SetTimeline(mustJSON(cur.Timeline)).Exec(ctx)
}

func (r WithdrawalRepo) Stats(ctx context.Context, sc tenant.Scope, now int64) (*finance.Stats, error) {
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

func (r WithdrawalRepo) CountPending(ctx context.Context) (int64, error) {
	n, err := r.client.Withdrawal.Query().
		Where(entwithdrawal.Status(string(finance.WithdrawPending))).
		Count(ctx)
	return int64(n), err
}

func (r WithdrawalRepo) CountByPartySince(ctx context.Context, partyType, partyID string, since int64) (int64, error) {
	n, err := r.client.Withdrawal.Query().
		Where(entwithdrawal.PartyType(partyType), entwithdrawal.PartyID(partyID),
			entwithdrawal.CreatedAtGTE(since)).
		Count(ctx)
	return int64(n), err
}

func (r WithdrawalRepo) SumPendingByParty(ctx context.Context, partyType, partyID string) (float64, error) {
	return sumFloat64(ctx, r.client.Withdrawal.Query().
		Where(entwithdrawal.PartyType(partyType), entwithdrawal.PartyID(partyID),
			entwithdrawal.Status(string(finance.WithdrawPending))),
		entwithdrawal.FieldAmount)
}

// SettlementRepo 实现 finance.SettlementRepository。
type SettlementRepo struct{ client *ent.Client }

func (r SettlementRepo) Save(ctx context.Context, s *finance.Settlement) error {
	if s.ID == "" {
		s.ID = newULID()
	}
	err := r.client.Settlement.Create().
		SetID(s.ID).SetRegionID(s.RegionID).SetParty(s.Party).
		SetPartyType(s.PartyType).SetPartyID(s.PartyID).
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
		PartyType: e.PartyType, PartyID: e.PartyID,
		Type: finance.SettlementType(e.Type), Amount: e.Amount,
		Status: finance.SettlementStatus(e.Status), Period: e.Period,
		SettledAt: e.SettledAt, Action: e.Action,
	}
}

func (r SettlementRepo) List(ctx context.Context, q finance.SettlementQuery, sc tenant.Scope) ([]*finance.Settlement, int64, error) {
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
	if q.PartyType != "" {
		bq.Where(entsettlement.PartyType(q.PartyType))
	}
	if q.PartyID != "" {
		bq.Where(entsettlement.PartyID(q.PartyID))
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

func (r SettlementRepo) SumPendingByParty(ctx context.Context, partyType, partyID string) (float64, error) {
	return sumFloat64(ctx, r.client.Settlement.Query().
		Where(entsettlement.PartyType(partyType), entsettlement.PartyID(partyID),
			entsettlement.Status(string(finance.SettlementProcessing))),
		entsettlement.FieldAmount)
}

func (r SettlementRepo) Stats(ctx context.Context, sc tenant.Scope, now int64) (*finance.SettlementStats, error) {
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
