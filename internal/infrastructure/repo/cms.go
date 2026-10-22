package repo

import (
	"context"
	"strings"

	"dawn-errand/internal/domain/cms"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/infrastructure/ent"
	entannouncement "dawn-errand/internal/infrastructure/ent/announcement"
	entfeerule "dawn-errand/internal/infrastructure/ent/feerule"
)

// AnnouncementRepo 实现 cms.AnnouncementRepository。
type AnnouncementRepo struct{ client *ent.Client }

func (r AnnouncementRepo) Save(a *cms.Announcement) error {
	ctx := context.Background()
	if a.ID == "" {
		a.ID = newULID()
	}
	exists, err := r.client.Announcement.Query().Where(entannouncement.ID(a.ID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err := r.client.Announcement.Update().
			Where(entannouncement.ID(a.ID)).
			SetTitle(a.Title).SetType(a.Type).SetTypeTone(a.TypeTone).
			SetAudience(a.Audience).SetStatus(string(a.Status)).
			SetPublishedAt(a.PublishedAt).SetPlannedAt(a.PlannedAt).
			SetActionName(a.ActionName).SetContent(a.Content).SetTags(a.Tags).
			SetStats(mustJSON(a.Stats)).SetCreatedBy(a.CreatedBy).
			SetCreatedAt(a.CreatedAt).SetUpdatedAt(a.UpdatedAt).
			Save(ctx)
		return err
	}
	return r.client.Announcement.Create().
		SetID(a.ID).SetTitle(a.Title).SetType(a.Type).SetTypeTone(a.TypeTone).
		SetAudience(a.Audience).SetStatus(string(a.Status)).
		SetPublishedAt(a.PublishedAt).SetPlannedAt(a.PlannedAt).
		SetActionName(a.ActionName).SetContent(a.Content).SetTags(a.Tags).
		SetStats(mustJSON(a.Stats)).SetCreatedBy(a.CreatedBy).
		SetCreatedAt(a.CreatedAt).SetUpdatedAt(a.UpdatedAt).
		Exec(ctx)
}

func toAnnouncement(e *ent.Announcement) *cms.Announcement {
	if e == nil {
		return nil
	}
	a := &cms.Announcement{
		ID: e.ID, Title: e.Title, Type: e.Type, TypeTone: e.TypeTone,
		Audience: e.Audience, Status: cms.AnnouncementStatus(e.Status),
		PublishedAt: e.PublishedAt, PlannedAt: e.PlannedAt,
		ActionName: e.ActionName, Content: e.Content, Tags: e.Tags,
		CreatedBy: e.CreatedBy, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
	_ = jsonUnmarshal(e.Stats, &a.Stats)
	return a
}

func (r AnnouncementRepo) FindByID(id string) (*cms.Announcement, error) {
	e, err := r.client.Announcement.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toAnnouncement(e), nil
}

func (r AnnouncementRepo) List(q cms.AnnouncementQuery, sc tenant.Scope) ([]*cms.Announcement, int64, error) {
	ctx := context.Background()
	bq := r.client.Announcement.Query()
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entannouncement.Status(s))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(entannouncement.Or(
			entannouncement.TitleContainsFold(k),
			entannouncement.ContentContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entannouncement.FieldCreatedAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*cms.Announcement, 0, len(rows))
	for _, e := range rows {
		out = append(out, toAnnouncement(e))
	}
	return out, int64(total), nil
}

func (r AnnouncementRepo) ListDueScheduled(now int64) ([]*cms.Announcement, error) {
	rows, err := r.client.Announcement.Query().
		Where(
			entannouncement.Status(string(cms.AnnScheduled)),
			entannouncement.PlannedAtGT(0),
			entannouncement.PlannedAtLTE(now),
		).
		All(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]*cms.Announcement, 0, len(rows))
	for _, e := range rows {
		out = append(out, toAnnouncement(e))
	}
	return out, nil
}

func (r AnnouncementRepo) MarkPublished(id string, now int64) error {
	err := r.client.Announcement.UpdateOneID(id).
		SetStatus(string(cms.AnnPublished)).SetPublishedAt(now).SetActionName("下线").
		SetUpdatedAt(now).
		Exec(context.Background())
	return err
}

func (r AnnouncementRepo) UpdateStatus(id string, status cms.AnnouncementStatus, now int64) error {
	err := r.client.Announcement.UpdateOneID(id).
		SetStatus(string(status)).SetUpdatedAt(now).
		Exec(context.Background())
	return err
}

// FeeRuleRepo 实现 cms.FeeRuleRepository。
type FeeRuleRepo struct{ client *ent.Client }

func (r FeeRuleRepo) Save(f *cms.FeeRule) error {
	ctx := context.Background()
	if f.ID == "" {
		f.ID = newULID()
	}
	exists, err := r.client.FeeRule.Query().Where(entfeerule.ID(f.ID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err := r.client.FeeRule.Update().
			Where(entfeerule.ID(f.ID)).
			SetCategory(f.Category).SetName(f.Name).SetStatus(string(f.Status)).
			SetValue(f.Value).SetDesc(f.Desc).SetAccent(f.Accent).SetHighlight(f.Highlight).
			SetRule(f.Rule).SetExample(f.Example).SetFee(f.Fee).SetRiderShare(f.RiderShare).
			SetIconColor(f.IconColor).SetVersion(f.Version).SetUpdatedAt(f.UpdatedAt).
			SetNextEffectiveAt(f.NextEffectiveAt).SetEnabled(f.Enabled).
			Save(ctx)
		return err
	}
	return r.client.FeeRule.Create().
		SetID(f.ID).SetCategory(f.Category).SetName(f.Name).SetStatus(string(f.Status)).
		SetValue(f.Value).SetDesc(f.Desc).SetAccent(f.Accent).SetHighlight(f.Highlight).
		SetRule(f.Rule).SetExample(f.Example).SetFee(f.Fee).SetRiderShare(f.RiderShare).
		SetIconColor(f.IconColor).SetVersion(f.Version).SetUpdatedAt(f.UpdatedAt).
		SetNextEffectiveAt(f.NextEffectiveAt).SetEnabled(f.Enabled).
		Exec(ctx)
}

func toFeeRule(e *ent.FeeRule) *cms.FeeRule {
	if e == nil {
		return nil
	}
	return &cms.FeeRule{
		ID: e.ID, Category: e.Category, Name: e.Name, Status: cms.FeeRuleStatus(e.Status),
		Value: e.Value, Desc: e.Desc, Accent: e.Accent, Highlight: e.Highlight,
		Rule: e.Rule, Example: e.Example, Fee: e.Fee, RiderShare: e.RiderShare,
		IconColor: e.IconColor, Version: e.Version, UpdatedAt: e.UpdatedAt,
		NextEffectiveAt: e.NextEffectiveAt, Enabled: e.Enabled,
	}
}

func (r FeeRuleRepo) List(sc tenant.Scope) ([]*cms.FeeRule, error) {
	rows, err := r.client.FeeRule.Query().
		Order(ent.Asc(entfeerule.FieldCategory), ent.Asc(entfeerule.FieldName)).
		All(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]*cms.FeeRule, 0, len(rows))
	for _, e := range rows {
		out = append(out, toFeeRule(e))
	}
	return out, nil
}

func (r FeeRuleRepo) FindByID(id string) (*cms.FeeRule, error) {
	e, err := r.client.FeeRule.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toFeeRule(e), nil
}

func (r FeeRuleRepo) Delete(id string) error {
	return r.client.FeeRule.DeleteOneID(id).Exec(context.Background())
}
