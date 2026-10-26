package admin

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/cms"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// SettingsService 运营配置：服务区域、费率规则、公告。
type SettingsService struct {
	Repos *application.RepoUnit
}

// overLen 按字符数判长度：库表用 varchar(N)，Postgres 按字符计（中文算 1 字），
// 上限与 migrations/001 的列宽对齐，超限在应用层拦成 400，而不是让驱动报错冒成 1500。
func overLen(s string, max int) bool { return utf8.RuneCountInString(s) > max }

// lenField 一个待判长度的字段。
type lenField struct {
	label string
	value string
	max   int
}

// checkLens 逐字段校验长度，返回第一个超限字段的提示。
func checkLens(fields ...lenField) error {
	for _, f := range fields {
		if overLen(f.value, f.max) {
			return errcode.New(errcode.CodeBadRequest, fmt.Sprintf("%s不能超过 %d 字", f.label, f.max))
		}
	}
	return nil
}

// ---- 服务区域 ----

// SettingsResult GET /settings 返回。
type SettingsResult struct {
	TotalTowns int64
	Areas      []*tenant.Region
}

// GetSettings 服务区域列表（区域站长仅本区域）。
func (s *SettingsService) GetSettings(ctx context.Context, sc tenant.Scope) (*SettingsResult, error) {
	all, err := s.Repos.Regions.List(ctx)
	if err != nil {
		return nil, err
	}
	res := &SettingsResult{TotalTowns: int64(len(all))}
	for _, r := range all {
		if sc.RegionOnly() && r.ID != sc.RegionID {
			continue
		}
		res.Areas = append(res.Areas, r)
	}
	return res, nil
}

// SaveAreaInput 服务区域写入白名单（R4-2）。
// 直绑 tenant.Region 等于把 CreatedAt/UpdatedAt 交给请求体，且没有任何长度与枚举约束；
// 显式入参后「可写字段集合」就是边界，时间戳一律由服务端生成。
type SaveAreaInput struct {
	ID       string
	Name     string
	Status   string
	Villages int
	RadiusKM float64
	StartFee float64
	FreeKM   float64
	Notice   string
}

// SaveAreaAction 区域操作。
type SaveAreaAction struct {
	Action string // create / update / disable / delete
	Area   *SaveAreaInput
}

// toRegion 校验并转换为领域实体。
func (in *SaveAreaInput) toRegion(now int64) (*tenant.Region, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || overLen(name, 64) {
		return nil, errcode.New(errcode.CodeBadRequest, "区域名称需为 1-64 字")
	}
	if err := checkLens(lenField{"区域 id", in.ID, 26}, lenField{"开通提示", in.Notice, 255}); err != nil {
		return nil, err
	}
	status := tenant.RegionStatus(strings.TrimSpace(in.Status))
	switch status {
	case "":
		status = tenant.RegionStatusPreparing
	case tenant.RegionStatusOpen, tenant.RegionStatusPreparing:
	default:
		return nil, errcode.New(errcode.CodeBadRequest, "区域状态仅支持 open / preparing")
	}
	if in.Villages < 0 || in.RadiusKM < 0 || in.StartFee < 0 || in.FreeKM < 0 {
		return nil, errcode.New(errcode.CodeBadRequest, "村组数、服务半径与价格不能为负")
	}
	return &tenant.Region{
		ID: strings.TrimSpace(in.ID), Name: name, Status: status, Villages: in.Villages,
		RadiusKM: in.RadiusKM, StartFee: in.StartFee, FreeKM: in.FreeKM,
		Notice: strings.TrimSpace(in.Notice), CreatedAt: now, UpdatedAt: now,
	}, nil
}

// SaveArea 新增/编辑/停用/删除服务区域。
func (s *SettingsService) SaveArea(ctx context.Context, in SaveAreaAction, operator string, sc tenant.Scope) error {
	if sc.RegionOnly() {
		return tenantDenied()
	}
	now := time.Now().Unix()
	objectID := ""
	switch in.Action {
	case "create", "update":
		// 此前无判空：body 省略 area 时 in.Area.ID 直接空指针 panic。
		if in.Area == nil {
			return errcode.New(errcode.CodeBadRequest, "缺少区域信息")
		}
		area, err := in.Area.toRegion(now)
		if err != nil {
			return err
		}
		if area.ID == "" {
			area.ID = ulidgen.New()
		}
		objectID = area.ID
		if err := s.Repos.Regions.Save(ctx, area); err != nil {
			return err
		}
	case "disable":
		if in.Area == nil || in.Area.ID == "" {
			return errcode.New(errcode.CodeBadRequest, "缺少区域 id")
		}
		objectID = in.Area.ID
		r, err := s.Repos.Regions.FindByID(ctx, objectID)
		if err != nil {
			return err
		}
		r.Status = tenant.RegionStatusPreparing
		r.UpdatedAt = now
		if err := s.Repos.Regions.Save(ctx, r); err != nil {
			return err
		}
	case "delete":
		if in.Area == nil || in.Area.ID == "" {
			return errcode.New(errcode.CodeBadRequest, "缺少区域 id")
		}
		objectID = in.Area.ID
		if err := s.Repos.Regions.Delete(ctx, objectID); err != nil {
			return err
		}
	default:
		// 不回显 in.Action：入参原样出现在响应里等于给探测者一份回执。
		return errcode.New(errcode.CodeBadRequest, "不支持的区域操作")
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "settings.area." + in.Action, Object: objectID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// ---- 费率规则 ----

// GetFeeRules 费率规则列表。
func (s *SettingsService) GetFeeRules(ctx context.Context, sc tenant.Scope) ([]*cms.FeeRule, error) {
	rules, err := s.Repos.FeeRules.List(ctx, sc)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return []*cms.FeeRule{}, nil
	}
	return rules, nil
}

// SaveFeeRuleInput 费率规则写入白名单。
// Enabled 不在其中：它由 Status 派生，接收客户端布尔只会造出「enabled=true + status=已停用」的矛盾行。
type SaveFeeRuleInput struct {
	ID              string
	Category        string
	Name            string
	Status          string
	Value           string
	Desc            string
	Accent          bool
	Highlight       bool
	Rule            string
	Example         string
	Fee             string
	RiderShare      string
	IconColor       string
	Version         string
	NextEffectiveAt int64
}

// SaveFeeRule 新增/编辑费率规则。
func (s *SettingsService) SaveFeeRule(ctx context.Context, in SaveFeeRuleInput, operator string, sc tenant.Scope) error {
	if sc.RegionOnly() {
		return tenantDenied()
	}
	name := strings.TrimSpace(in.Name)
	if name == "" || overLen(name, 64) {
		return errcode.New(errcode.CodeBadRequest, "规则名称需为 1-64 字")
	}
	category := strings.TrimSpace(in.Category)
	if category == "" || overLen(category, 16) {
		return errcode.New(errcode.CodeBadRequest, "规则分类需为 1-16 字")
	}
	if err := checkLens(
		lenField{"规则 id", in.ID, 26},
		lenField{"展示值", in.Value, 64},
		lenField{"说明", in.Desc, 128},
		lenField{"规则文案", in.Rule, 128},
		lenField{"示例", in.Example, 128},
		lenField{"费用文案", in.Fee, 64},
		lenField{"骑手分成", in.RiderShare, 64},
		lenField{"图标颜色", in.IconColor, 32},
		lenField{"版本号", in.Version, 16},
	); err != nil {
		return err
	}
	status := cms.FeeRuleStatus(strings.TrimSpace(in.Status))
	switch status {
	case "":
		status = cms.FeeEnabled
	case cms.FeeEnabled, cms.FeeDisabled:
	default:
		return errcode.New(errcode.CodeBadRequest, "规则状态仅支持 已启用 / 已停用")
	}
	now := time.Now().Unix()
	rule := &cms.FeeRule{
		ID: strings.TrimSpace(in.ID), Category: category, Name: name, Status: status,
		Value: in.Value, Desc: in.Desc, Accent: in.Accent, Highlight: in.Highlight,
		Rule: in.Rule, Example: in.Example, Fee: in.Fee, RiderShare: in.RiderShare,
		IconColor: in.IconColor, Version: in.Version,
		NextEffectiveAt: in.NextEffectiveAt, UpdatedAt: now, Enabled: status == cms.FeeEnabled,
	}
	if rule.ID == "" {
		rule.ID = ulidgen.New()
	}
	if err := s.Repos.FeeRules.Save(ctx, rule); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "settings.fee-rule.save", Object: rule.ID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// ---- 公告 ----

// ListAnnouncements 公告列表。
func (s *SettingsService) ListAnnouncements(ctx context.Context, q cms.AnnouncementQuery, sc tenant.Scope) ([]*cms.Announcement, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Announcements.List(ctx, q, sc)
}

// SaveAnnouncementInput 公告写入白名单。
// CreatedBy/CreatedAt/ActionName/Stats 不进白名单：归属与推送统计由服务端维护，
// 但仓储 Save 是无条件 SetXxx 整行覆盖，所以编辑时需要从原行继承（见 SaveAnnouncement）。
type SaveAnnouncementInput struct {
	ID          string
	Title       string
	Type        string
	TypeTone    string
	Audience    string
	Status      string
	Content     string
	Tags        []string
	PlannedAt   int64
	PublishedAt int64
}

// toEntity 校验并转换为领域实体。
func (in *SaveAnnouncementInput) toEntity(now int64, operator string) (*cms.Announcement, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || overLen(title, 128) {
		return nil, errcode.New(errcode.CodeBadRequest, "公告标题需为 1-128 字")
	}
	if err := checkLens(
		lenField{"公告 id", in.ID, 26},
		lenField{"公告类型", in.Type, 16},
		lenField{"标题色调", in.TypeTone, 16},
		lenField{"受众", in.Audience, 64},
		lenField{"公告状态", in.Status, 16},
		lenField{"正文", in.Content, 2048},
	); err != nil {
		return nil, err
	}
	status := cms.AnnouncementStatus(strings.TrimSpace(in.Status))
	switch status {
	case "":
		status = cms.AnnReviewing
	case cms.AnnPublished, cms.AnnScheduled, cms.AnnReviewing, cms.AnnOffline:
	default:
		return nil, errcode.New(errcode.CodeBadRequest, "公告状态取值不合法")
	}
	tags := make([]string, 0, len(in.Tags))
	for _, t := range in.Tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if overLen(t, 16) {
			return nil, errcode.New(errcode.CodeBadRequest, "标签不能超过 16 字")
		}
		tags = append(tags, t)
	}
	if len(tags) > 10 {
		return nil, errcode.New(errcode.CodeBadRequest, "标签最多 10 个")
	}
	or := func(v, def string) string {
		if v == "" {
			return def
		}
		return v
	}
	aud := &cms.Announcement{
		ID: strings.TrimSpace(in.ID), Title: title,
		Type:     or(strings.TrimSpace(in.Type), "提醒"),
		TypeTone: or(strings.TrimSpace(in.TypeTone), "blue"),
		Audience: or(strings.TrimSpace(in.Audience), "全部用户"),
		Status:   status, Content: in.Content, Tags: tags,
		PlannedAt: in.PlannedAt, PublishedAt: in.PublishedAt,
		CreatedBy: operator, UpdatedAt: now,
	}
	if aud.ID == "" {
		aud.ID = ulidgen.New()
	}
	return aud, nil
}

// SaveAnnouncement 新增/编辑/发布/下线/撤回公告。
func (s *SettingsService) SaveAnnouncement(ctx context.Context, in SaveAnnouncementInput, action, operator string, sc tenant.Scope) error {
	if sc.RegionOnly() {
		return tenantDenied()
	}
	now := time.Now().Unix()
	objectID := strings.TrimSpace(in.ID)
	switch action {
	case "save":
		ann, err := in.toEntity(now, operator)
		if err != nil {
			return err
		}
		objectID = ann.ID
		// 编辑保留原行的 created_at / action_name / stats：白名单不含这些字段，
		// 若按零值整行覆盖，公告的创建时间、按钮文案与推送统计会被清空。
		if prev, ferr := s.Repos.Announcements.FindByID(ctx, ann.ID); ferr == nil && prev != nil {
			ann.CreatedAt = prev.CreatedAt
			ann.ActionName = prev.ActionName
			ann.Stats = prev.Stats
		} else {
			ann.CreatedAt = now
			ann.ActionName = "发布"
		}
		if err := s.Repos.Announcements.Save(ctx, ann); err != nil {
			return err
		}
	case "publish", "offline", "review", "withdraw":
		if objectID == "" || overLen(objectID, 26) {
			return errcode.New(errcode.CodeBadRequest, "公告 id 必填")
		}
		target := map[string]cms.AnnouncementStatus{
			"publish":  cms.AnnPublished,
			"offline":  cms.AnnOffline,
			"review":   cms.AnnReviewing,
			"withdraw": cms.AnnScheduled,
		}[action]
		if err := s.Repos.Announcements.UpdateStatus(ctx, objectID, target, now); err != nil {
			return err
		}
	default:
		return errcode.New(errcode.CodeBadRequest, "不支持的公告操作")
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "settings.announcement." + action, Object: objectID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}
