package admin

import (
	"context"
	"time"

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

// ---- 服务区域 ----

// SettingsResult GET /settings 返回。
type SettingsResult struct {
	TotalTowns int64
	Areas      []*tenant.Region
}

// GetSettings 服务区域列表（区域站长仅本区域）。
func (s *SettingsService) GetSettings(ctx context.Context, sc tenant.Scope) (*SettingsResult, error) {
	all, err := s.Repos.Regions.List()
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

// SaveAreaAction 区域操作。
type SaveAreaAction struct {
	Action string         // create / update / disable / delete
	Area   *tenant.Region `json:"area,omitempty"`
}

// SaveArea 新增/编辑/停用/删除服务区域。
func (s *SettingsService) SaveArea(ctx context.Context, in SaveAreaAction, operator string, sc tenant.Scope) error {
	if sc.RegionOnly() {
		return tenantDenied()
	}
	now := time.Now().Unix()
	switch in.Action {
	case "create", "update":
		if in.Area == nil || in.Area.Name == "" {
			return errcode.New(errcode.CodeBadRequest, "区域名称必填")
		}
		if in.Area.ID == "" {
			in.Area.ID = ulidgen.New()
		}
		in.Area.UpdatedAt = now
		if in.Area.CreatedAt == 0 {
			in.Area.CreatedAt = now
		}
		if in.Area.Status == "" {
			in.Area.Status = tenant.RegionStatusPreparing
		}
		if err := s.Repos.Regions.Save(in.Area); err != nil {
			return err
		}
	case "disable":
		r, err := s.Repos.Regions.FindByID(in.Area.ID)
		if err != nil {
			return err
		}
		r.Status = tenant.RegionStatusPreparing
		r.UpdatedAt = now
		if err := s.Repos.Regions.Save(r); err != nil {
			return err
		}
	case "delete":
		if err := s.Repos.Regions.Delete(in.Area.ID); err != nil {
			return err
		}
	default:
		return errcode.New(errcode.CodeBadRequest, "未知操作："+in.Action)
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "settings.area." + in.Action, Object: in.Area.ID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// ---- 费率规则 ----

// GetFeeRules 费率规则列表。
func (s *SettingsService) GetFeeRules(ctx context.Context, sc tenant.Scope) ([]*cms.FeeRule, error) {
	rules, err := s.Repos.FeeRules.List(sc)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return []*cms.FeeRule{}, nil
	}
	return rules, nil
}

// SaveFeeRule 新增/编辑费率规则。
func (s *SettingsService) SaveFeeRule(ctx context.Context, in cms.FeeRule, operator string, sc tenant.Scope) error {
	if sc.RegionOnly() {
		return tenantDenied()
	}
	if in.Name == "" {
		return errcode.New(errcode.CodeBadRequest, "规则名称必填")
	}
	now := time.Now().Unix()
	if in.ID == "" {
		in.ID = ulidgen.New()
	}
	if in.Status == "" {
		in.Status = cms.FeeEnabled
	}
	in.Enabled = in.Status == cms.FeeEnabled
	in.UpdatedAt = now
	if err := s.Repos.FeeRules.Save(&in); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "settings.fee-rule.save", Object: in.ID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// ---- 公告 ----

// AnnouncementAction 公告操作。
type AnnouncementAction struct {
	Action string // save / publish / offline / review / withdraw
}

// ListAnnouncements 公告列表。
func (s *SettingsService) ListAnnouncements(ctx context.Context, q cms.AnnouncementQuery, sc tenant.Scope) ([]*cms.Announcement, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Announcements.List(q, sc)
}

// SaveAnnouncement 新增/编辑/发布/下线/撤回公告。
func (s *SettingsService) SaveAnnouncement(ctx context.Context, in cms.Announcement, action, operator string, sc tenant.Scope) error {
	if sc.RegionOnly() {
		return tenantDenied()
	}
	now := time.Now().Unix()
	switch action {
	case "save":
		if in.Title == "" {
			return errcode.New(errcode.CodeBadRequest, "公告标题必填")
		}
		if in.ID == "" {
			in.ID = ulidgen.New()
		}
		if in.Status == "" {
			in.Status = cms.AnnReviewing
		}
		in.CreatedBy = operator
		in.UpdatedAt = now
		if in.CreatedAt == 0 {
			in.CreatedAt = now
		}
		if err := s.Repos.Announcements.Save(&in); err != nil {
			return err
		}
	case "publish":
		if err := s.Repos.Announcements.UpdateStatus(in.ID, cms.AnnPublished, now); err != nil {
			return err
		}
	case "offline":
		if err := s.Repos.Announcements.UpdateStatus(in.ID, cms.AnnOffline, now); err != nil {
			return err
		}
	case "review":
		if err := s.Repos.Announcements.UpdateStatus(in.ID, cms.AnnReviewing, now); err != nil {
			return err
		}
	case "withdraw":
		if err := s.Repos.Announcements.UpdateStatus(in.ID, cms.AnnScheduled, now); err != nil {
			return err
		}
	default:
		return errcode.New(errcode.CodeBadRequest, "未知操作："+action)
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "settings.announcement." + action, Object: in.ID, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}
