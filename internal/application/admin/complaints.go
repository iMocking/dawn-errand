package admin

import (
	"context"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/complaint"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
)

// ComplaintsService 投诉工单。
type ComplaintsService struct {
	Repos *application.RepoUnit
}

// List 工单列表。
func (s *ComplaintsService) List(ctx context.Context, q complaint.Query, sc tenant.Scope) ([]*complaint.Complaint, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Complaints.List(q, sc)
}

// Detail 工单详情（对象级隔离）。
func (s *ComplaintsService) Detail(ctx context.Context, id string, sc tenant.Scope) (*complaint.Complaint, error) {
	c, err := s.Repos.Complaints.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, errcode.NewStatus(404, errcode.CodeComplaintNotFound, "工单不存在")
		}
		return nil, err
	}
	if sc.RegionOnly() && c.RegionID != sc.RegionID {
		return nil, tenantDenied()
	}
	return c, nil
}
