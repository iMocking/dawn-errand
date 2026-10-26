package repo

import (
	"context"
	"strings"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/infrastructure/ent"
	entrider "dawn-errand/internal/infrastructure/ent/rider"
)

// RiderRepo 实现 rider.RiderRepository。Cipher 非空时身份证号透明加解密（v1: 前缀）。
type RiderRepo struct {
	client *ent.Client
	Cipher application.ConfigCipher
}

// sealSecret 写入前加密（每次写入生成新密文，读路径始终返回解密值）。
func (r RiderRepo) sealSecret(v string) string {
	if v == "" || r.Cipher == nil {
		return v
	}
	sealed, err := r.Cipher.Encrypt(v)
	if err != nil {
		return v
	}
	return sealed
}

// unsealSecret 读取后解密；非 v1 前缀视为存量明文原样返回（由上层脱敏兜底）。
func (r RiderRepo) unsealSecret(v string) string {
	if v == "" || r.Cipher == nil || !strings.HasPrefix(v, "v1:") {
		return v
	}
	plain, err := r.Cipher.Decrypt(v)
	if err != nil {
		return ""
	}
	return plain
}

func (r RiderRepo) Save(m *rider.Rider) error {
	ctx := context.Background()
	if m.ID == "" {
		m.ID = newULID()
	}
	exists, err := r.client.Rider.Query().Where(entrider.ID(m.ID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err := r.client.Rider.Update().
			Where(entrider.ID(m.ID)).
			SetRegionID(m.RegionID).SetUserID(m.UserID).
			SetName(m.Name).SetPhone(m.Phone).SetArea(m.Area).SetTool(m.Tool).SetLevel(m.Level).
			SetIDCard(r.sealSecret(m.IDCard)).SetVehicle(m.Vehicle).SetAccount(m.Account).
			SetStatus(string(m.Status)).SetAuditStatus(string(m.AuditStatus)).
			SetApplyNo(m.ApplyNo).SetSubmittedAt(m.SubmittedAt).SetOrderTypes(m.OrderTypes).
			SetIDCheck(m.IDCheck).SetNoticeType(m.NoticeType).SetNotice(m.Notice).
			SetDocCount(m.DocCount).SetDocs(m.Docs).
			SetServiceType(m.ServiceType).SetJoinedAt(m.JoinedAt).SetServiceMonths(m.ServiceMonths).
			SetRating(m.Rating).SetTodayOrders(m.TodayOrders).SetTodayOrdersDelta(m.TodayOrdersDelta).
			SetOnlineHours(m.OnlineHours).SetMonthIncome(m.MonthIncome).SetOnTimeRate(m.OnTimeRate).
			SetComplaintsToday(m.ComplaintsToday).SetAvgMinutes(m.AvgMinutes).
			SetLastAction(m.LastAction).SetLastActionAt(m.LastActionAt).SetUpdatedAt(nowSec()).
			Save(ctx)
		return err
	}
	return r.client.Rider.Create().
		SetID(m.ID).SetRegionID(m.RegionID).SetUserID(m.UserID).
		SetName(m.Name).SetPhone(m.Phone).SetArea(m.Area).SetTool(m.Tool).SetLevel(m.Level).
		SetIDCard(r.sealSecret(m.IDCard)).SetVehicle(m.Vehicle).SetAccount(m.Account).
		SetStatus(string(m.Status)).SetAuditStatus(string(m.AuditStatus)).
		SetApplyNo(m.ApplyNo).SetSubmittedAt(m.SubmittedAt).SetOrderTypes(m.OrderTypes).
		SetIDCheck(m.IDCheck).SetNoticeType(m.NoticeType).SetNotice(m.Notice).
		SetDocCount(m.DocCount).SetDocs(m.Docs).
		SetServiceType(m.ServiceType).SetJoinedAt(m.JoinedAt).SetServiceMonths(m.ServiceMonths).
		SetRating(m.Rating).SetTodayOrders(m.TodayOrders).SetTodayOrdersDelta(m.TodayOrdersDelta).
		SetOnlineHours(m.OnlineHours).SetMonthIncome(m.MonthIncome).SetOnTimeRate(m.OnTimeRate).
		SetComplaintsToday(m.ComplaintsToday).SetAvgMinutes(m.AvgMinutes).
		SetLastAction(m.LastAction).SetLastActionAt(m.LastActionAt).SetUpdatedAt(nowSec()).
		Exec(ctx)
}

func (r RiderRepo) toRider(e *ent.Rider) *rider.Rider {
	if e == nil {
		return nil
	}
	return &rider.Rider{
		ID: e.ID, RegionID: e.RegionID, UserID: e.UserID,
		Name: e.Name, Phone: e.Phone, Area: e.Area, Tool: e.Tool, Level: e.Level,
		IDCard: r.unsealSecret(e.IDCard), Vehicle: e.Vehicle, Account: e.Account,
		Status: rider.RiderStatus(e.Status), AuditStatus: rider.AuditStatus(e.AuditStatus),
		ApplyNo: e.ApplyNo, SubmittedAt: e.SubmittedAt, OrderTypes: e.OrderTypes,
		IDCheck: e.IDCheck, NoticeType: e.NoticeType, Notice: e.Notice,
		DocCount: e.DocCount, Docs: e.Docs,
		ServiceType: e.ServiceType, JoinedAt: e.JoinedAt, ServiceMonths: e.ServiceMonths,
		Rating: e.Rating, TodayOrders: e.TodayOrders, TodayOrdersDelta: e.TodayOrdersDelta,
		OnlineHours: e.OnlineHours, MonthIncome: e.MonthIncome, OnTimeRate: e.OnTimeRate,
		ComplaintsToday: e.ComplaintsToday, AvgMinutes: e.AvgMinutes,
		LastAction: e.LastAction, LastActionAt: e.LastActionAt,
	}
}

func (r RiderRepo) FindByID(id string) (*rider.Rider, error) {
	e, err := r.client.Rider.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.toRider(e), nil
}

func (r RiderRepo) List(q rider.RiderQuery, sc tenant.Scope) ([]*rider.Rider, int64, error) {
	ctx := context.Background()
	bq := r.client.Rider.Query()
	scopeRider(bq, sc)
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entrider.Status(s))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(entrider.Or(
			entrider.NameContainsFold(k),
			entrider.PhoneContainsFold(k),
			entrider.AreaContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entrider.FieldUpdatedAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*rider.Rider, 0, len(rows))
	for _, e := range rows {
		out = append(out, r.toRider(e))
	}
	return out, int64(total), nil
}

func (r RiderRepo) AuditList(q rider.AuditQuery, sc tenant.Scope) ([]*rider.Rider, int64, error) {
	ctx := context.Background()
	bq := r.client.Rider.Query()
	scopeRider(bq, sc)
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entrider.AuditStatus(s))
	} else {
		bq.Where(entrider.AuditStatusIn(
			string(rider.AuditPending), string(rider.AuditApproved), string(rider.AuditRejected),
		))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(entrider.Or(
			entrider.NameContainsFold(k),
			entrider.PhoneContainsFold(k),
			entrider.ApplyNoContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entrider.FieldSubmittedAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*rider.Rider, 0, len(rows))
	for _, e := range rows {
		out = append(out, r.toRider(e))
	}
	return out, int64(total), nil
}

func (r RiderRepo) UpdateAudit(id string, approved bool, reason, operator string, now int64) error {
	ctx := context.Background()
	status := string(rider.AuditApproved)
	if !approved {
		status = string(rider.AuditRejected)
	}
	// 身份核验随审核联动：通过 → 已通过；驳回 → 已驳回（重新提交时恢复待补充）。
	idCheck := "已驳回"
	if approved {
		idCheck = "已通过"
	}
	upd := r.client.Rider.UpdateOneID(id).
		SetAuditStatus(status).SetIDCheck(idCheck).SetUpdatedAt(now)
	if approved {
		upd.SetStatus(string(rider.RiderOffline)).SetJoinedAt(now)
	}
	err := upd.SetLastAction(operator + " 审核" + status).SetLastActionAt(now).Exec(ctx)
	return err
}

func (r RiderRepo) UpdateStatus(id string, status rider.RiderStatus, now int64) error {
	err := r.client.Rider.UpdateOneID(id).
		SetStatus(string(status)).SetUpdatedAt(now).
		SetLastActionAt(now).Exec(context.Background())
	return err
}

func (r RiderRepo) Stats(sc tenant.Scope) (*rider.Stats, error) {
	ctx := context.Background()
	base := func() *ent.RiderQuery {
		bq := r.client.Rider.Query()
		scopeRider(bq, sc)
		return bq
	}
	total, err := base().Count(ctx)
	if err != nil {
		return nil, err
	}
	online, err := base().Where(entrider.StatusIn(
		string(rider.RiderOnline), string(rider.RiderBusy),
	)).Count(ctx)
	if err != nil {
		return nil, err
	}
	pending, err := base().Where(entrider.AuditStatus(string(rider.AuditPending))).Count(ctx)
	if err != nil {
		return nil, err
	}
	frozen, err := base().Where(entrider.Status(string(rider.RiderFrozen))).Count(ctx)
	if err != nil {
		return nil, err
	}
	rate := 0.0
	if total > 0 {
		rate = float64(online) / float64(total) * 100
	}

	// 审核汇总
	today, err := base().Where(entrider.SubmittedAtGTE(startOfDayUTC(nowSec()))).Count(ctx)
	if err != nil {
		return nil, err
	}
	approved, err := base().Where(entrider.AuditStatus(string(rider.AuditApproved))).Count(ctx)
	if err != nil {
		return nil, err
	}
	rejected, err := base().Where(entrider.AuditStatus(string(rider.AuditRejected))).Count(ctx)
	if err != nil {
		return nil, err
	}
	monthApproved, err := base().Where(
		entrider.AuditStatus(string(rider.AuditApproved)),
		entrider.JoinedAtGTE(daysAgoStart(nowSec(), 30)),
	).Count(ctx)
	if err != nil {
		return nil, err
	}

	return &rider.Stats{
		Total:        int64(total),
		Online:       int64(online),
		OnlineRate:   rate,
		PendingAudit: int64(pending),
		Frozen:       int64(frozen),
		WeekDelta:    "较上周",
		Audit: rider.AuditStats{
			TodayNew: int64(today), Approved: int64(approved), Rejected: int64(rejected),
			MonthApproved: int64(monthApproved), MomRate: "环比",
		},
	}, nil
}

func (r RiderRepo) CountByRegion(regionID string) (int64, error) {
	n, err := r.client.Rider.Query().Where(entrider.RegionID(regionID)).Count(context.Background())
	return int64(n), err
}

func (r RiderRepo) ActiveCountByRegion(regionID string) (int64, error) {
	n, err := r.client.Rider.Query().
		Where(entrider.RegionID(regionID), entrider.StatusIn(
			string(rider.RiderOnline), string(rider.RiderBusy), string(rider.RiderOffline),
		)).
		Count(context.Background())
	return int64(n), err
}

func (r RiderRepo) OnlineCount() (int64, error) {
	n, err := r.client.Rider.Query().
		Where(entrider.StatusIn(string(rider.RiderOnline), string(rider.RiderBusy))).
		Count(context.Background())
	return int64(n), err
}

func (r RiderRepo) TotalCount() (int64, error) {
	n, err := r.client.Rider.Query().Count(context.Background())
	return int64(n), err
}
