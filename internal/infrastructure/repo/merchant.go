package repo

import (
	"context"
	"strings"

	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/infrastructure/ent"
	entmerchant "dawn-errand/internal/infrastructure/ent/merchant"
)

// MerchantRepo 实现 merchant.MerchantRepository。
type MerchantRepo struct{ client *ent.Client }

func (r MerchantRepo) Save(m *merchant.Merchant) error {
	ctx := context.Background()
	if m.ID == "" {
		m.ID = newULID()
	}
	exists, err := r.client.Merchant.Query().Where(entmerchant.ID(m.ID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err := r.client.Merchant.Update().
			Where(entmerchant.ID(m.ID)).
			SetRegionID(m.RegionID).SetOwnerUserID(m.OwnerUserID).
			SetName(m.Name).SetCategory(m.Category).SetAddress(m.Address).SetAddressSub(m.AddressSub).
			SetContact(m.Contact).SetPhone(m.Phone).SetOpenHours(m.OpenHours).SetDeliveryRange(m.DeliveryRange).
			SetCommission(m.Commission).SetBankAccount(m.BankAccount).SetDeliveryMethod(m.DeliveryMethod).
			SetLat(m.Lat).SetLng(m.Lng).SetAutoAccept(m.AutoAccept).
			SetMinPrice(m.MinPrice).SetDeliveryFee(m.DeliveryFee).SetPackingFee(m.PackingFee).
			SetAnnouncement(m.Announcement).SetServicePhone(m.ServicePhone).
			SetStatus(string(m.Status)).SetAuditStatus(string(m.AuditStatus)).
			SetApplyNo(m.ApplyNo).SetOperator(m.Operator).SetAppliedAt(m.AppliedAt).
			SetCertLabels(m.CertLabels).SetRejectReason(m.RejectReason).
			SetRating(m.Rating).SetMonthlyOrders(m.MonthlyOrders).SetMonthlyRevenue(m.MonthlyRevenue).
			SetAvgOrder(m.AvgOrder).SetGoodRate(m.GoodRate).SetJoinedAt(m.JoinedAt).
			SetLastUpdatedAt(m.LastUpdatedAt).SetCompleteness(m.Completeness).SetReviewCount(m.ReviewCount).
			SetLastOperation(m.LastOperation).
			SetTodayOrders(m.TodayOrders).SetTodayRevenue(m.TodayRevenue).SetOnTimeRate(m.OnTimeRate).
			SetPendingAbnormal(m.PendingAbnormal).SetUpdatedAt(nowSec()).
			Save(ctx)
		return err
	}
	return r.client.Merchant.Create().
		SetID(m.ID).SetRegionID(m.RegionID).SetOwnerUserID(m.OwnerUserID).
		SetName(m.Name).SetCategory(m.Category).SetAddress(m.Address).SetAddressSub(m.AddressSub).
		SetContact(m.Contact).SetPhone(m.Phone).SetOpenHours(m.OpenHours).SetDeliveryRange(m.DeliveryRange).
		SetCommission(m.Commission).SetBankAccount(m.BankAccount).SetDeliveryMethod(m.DeliveryMethod).
		SetLat(m.Lat).SetLng(m.Lng).SetAutoAccept(m.AutoAccept).
		SetMinPrice(m.MinPrice).SetDeliveryFee(m.DeliveryFee).SetPackingFee(m.PackingFee).
		SetAnnouncement(m.Announcement).SetServicePhone(m.ServicePhone).
		SetStatus(string(m.Status)).SetAuditStatus(string(m.AuditStatus)).
		SetApplyNo(m.ApplyNo).SetOperator(m.Operator).SetAppliedAt(m.AppliedAt).
		SetCertLabels(m.CertLabels).SetRejectReason(m.RejectReason).
		SetRating(m.Rating).SetMonthlyOrders(m.MonthlyOrders).SetMonthlyRevenue(m.MonthlyRevenue).
		SetAvgOrder(m.AvgOrder).SetGoodRate(m.GoodRate).SetJoinedAt(m.JoinedAt).
		SetLastUpdatedAt(m.LastUpdatedAt).SetCompleteness(m.Completeness).SetReviewCount(m.ReviewCount).
		SetLastOperation(m.LastOperation).
		SetTodayOrders(m.TodayOrders).SetTodayRevenue(m.TodayRevenue).SetOnTimeRate(m.OnTimeRate).
		SetPendingAbnormal(m.PendingAbnormal).SetUpdatedAt(nowSec()).
		Exec(ctx)
}

func toMerchant(e *ent.Merchant) *merchant.Merchant {
	if e == nil {
		return nil
	}
	return &merchant.Merchant{
		ID: e.ID, RegionID: e.RegionID, OwnerUserID: e.OwnerUserID,
		Name: e.Name, Category: e.Category, Address: e.Address, AddressSub: e.AddressSub,
		Contact: e.Contact, Phone: e.Phone, OpenHours: e.OpenHours, DeliveryRange: e.DeliveryRange,
		Commission: e.Commission, BankAccount: e.BankAccount, DeliveryMethod: e.DeliveryMethod,
		Lat: e.Lat, Lng: e.Lng, AutoAccept: e.AutoAccept,
		MinPrice: e.MinPrice, DeliveryFee: e.DeliveryFee, PackingFee: e.PackingFee,
		Announcement: e.Announcement, ServicePhone: e.ServicePhone,
		Status: merchant.MerchantStatus(e.Status), AuditStatus: merchant.AuditStatus(e.AuditStatus),
		ApplyNo: e.ApplyNo, Operator: e.Operator, AppliedAt: e.AppliedAt,
		CertLabels: e.CertLabels, RejectReason: e.RejectReason,
		Rating: e.Rating, MonthlyOrders: e.MonthlyOrders, MonthlyRevenue: e.MonthlyRevenue,
		AvgOrder: e.AvgOrder, GoodRate: e.GoodRate, JoinedAt: e.JoinedAt,
		LastUpdatedAt: e.LastUpdatedAt, Completeness: e.Completeness, ReviewCount: e.ReviewCount,
		LastOperation: e.LastOperation,
		TodayOrders:   e.TodayOrders, TodayRevenue: e.TodayRevenue,
		OnTimeRate: e.OnTimeRate, PendingAbnormal: e.PendingAbnormal,
	}
}

func (r MerchantRepo) FindByID(id string) (*merchant.Merchant, error) {
	e, err := r.client.Merchant.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toMerchant(e), nil
}

func (r MerchantRepo) List(q merchant.MerchantQuery, sc tenant.Scope) ([]*merchant.Merchant, int64, error) {
	ctx := context.Background()
	bq := r.client.Merchant.Query()
	scopeMerchant(bq, sc)
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entmerchant.Status(s))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(entmerchant.Or(
			entmerchant.NameContainsFold(k),
			entmerchant.ContactContainsFold(k),
			entmerchant.PhoneContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entmerchant.FieldUpdatedAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*merchant.Merchant, 0, len(rows))
	for _, e := range rows {
		out = append(out, toMerchant(e))
	}
	return out, int64(total), nil
}

// ListForClient C 端商家列表：区域内、审核通过、未封禁；营业中优先。
func (r MerchantRepo) ListForClient(regionID, keyword, category string, page, pageSize int) ([]*merchant.Merchant, int64, error) {
	ctx := context.Background()
	bq := r.client.Merchant.Query().
		Where(entmerchant.RegionID(regionID),
			entmerchant.AuditStatus(string(merchant.AuditApproved)),
			entmerchant.StatusNEQ(string(merchant.MerchantStatusBanned)))
	if c := strings.TrimSpace(category); c != "" {
		bq.Where(entmerchant.Category(c))
	}
	if k := strings.TrimSpace(keyword); k != "" {
		bq.Where(entmerchant.Or(
			entmerchant.NameContainsFold(k),
			entmerchant.CategoryContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(page, pageSize)
	rows, err := bq.Order(ent.Asc(entmerchant.FieldStatus), ent.Desc(entmerchant.FieldRating)).
		Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*merchant.Merchant, 0, len(rows))
	for _, e := range rows {
		out = append(out, toMerchant(e))
	}
	return out, int64(total), nil
}

func (r MerchantRepo) AuditList(q merchant.AuditQuery, sc tenant.Scope) ([]*merchant.Merchant, int64, error) {
	ctx := context.Background()
	bq := r.client.Merchant.Query()
	scopeMerchant(bq, sc)
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entmerchant.AuditStatus(s))
	} else {
		bq.Where(entmerchant.AuditStatusIn(
			string(merchant.AuditPending), string(merchant.AuditApproved), string(merchant.AuditRejected),
		))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(entmerchant.Or(
			entmerchant.NameContainsFold(k),
			entmerchant.OperatorContainsFold(k),
			entmerchant.ApplyNoContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entmerchant.FieldAppliedAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*merchant.Merchant, 0, len(rows))
	for _, e := range rows {
		out = append(out, toMerchant(e))
	}
	return out, int64(total), nil
}

func (r MerchantRepo) UpdateAudit(id string, approved bool, reason, operator string, now int64) error {
	ctx := context.Background()
	status := string(merchant.AuditApproved)
	if !approved {
		status = string(merchant.AuditRejected)
	}
	upd := r.client.Merchant.UpdateOneID(id).
		SetAuditStatus(status).SetRejectReason(reason).SetUpdatedAt(now)
	if approved {
		upd.SetStatus(string(merchant.MerchantStatusOpen)).SetJoinedAt(now)
	}
	err := upd.SetLastOperation(operator + " 审核" + status).Exec(ctx)
	return err
}

func (r MerchantRepo) UpdateStatus(id string, status merchant.MerchantStatus, now int64) error {
	err := r.client.Merchant.UpdateOneID(id).
		SetStatus(string(status)).SetUpdatedAt(now).Exec(context.Background())
	return err
}

func (r MerchantRepo) Stats(sc tenant.Scope) (*merchant.Stats, error) {
	ctx := context.Background()
	base := func() *ent.MerchantQuery {
		bq := r.client.Merchant.Query()
		scopeMerchant(bq, sc)
		return bq
	}
	total, err := base().Count(ctx)
	if err != nil {
		return nil, err
	}
	open, err := base().Where(entmerchant.Status(string(merchant.MerchantStatusOpen))).Count(ctx)
	if err != nil {
		return nil, err
	}
	pending, err := base().Where(entmerchant.AuditStatus(string(merchant.AuditPending))).Count(ctx)
	if err != nil {
		return nil, err
	}
	rest, err := base().Where(entmerchant.StatusIn(
		string(merchant.MerchantStatusRest), string(merchant.MerchantStatusBanned),
	)).Count(ctx)
	if err != nil {
		return nil, err
	}
	weekNew, err := base().Where(entmerchant.JoinedAtGTE(daysAgoStart(nowSec(), 7))).Count(ctx)
	if err != nil {
		return nil, err
	}
	rate := 0.0
	if total > 0 {
		rate = float64(open) / float64(total) * 100
	}
	return &merchant.Stats{
		Total: int64(total), WeekNew: int64(weekNew), Open: int64(open),
		OpenRate: rate, Pending: int64(pending), Closed: int64(rest),
	}, nil
}

func (r MerchantRepo) CountByRegion(regionID string) (int64, error) {
	n, err := r.client.Merchant.Query().Where(entmerchant.RegionID(regionID)).Count(context.Background())
	return int64(n), err
}

func (r MerchantRepo) OnlineCountByRegion(regionID string) (int64, error) {
	n, err := r.client.Merchant.Query().
		Where(entmerchant.RegionID(regionID), entmerchant.Status(string(merchant.MerchantStatusOpen))).
		Count(context.Background())
	return int64(n), err
}

func (r MerchantRepo) CountAll() (int64, error) {
	n, err := r.client.Merchant.Query().Count(context.Background())
	return int64(n), err
}

func (r MerchantRepo) CountWeekNew(now int64) (int64, error) {
	n, err := r.client.Merchant.Query().
		Where(entmerchant.JoinedAtGTE(daysAgoStart(now, 7))).
		Count(context.Background())
	return int64(n), err
}
