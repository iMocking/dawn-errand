package admin

import (
	"context"
	"fmt"
	"sort"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/cms"
	"dawn-errand/internal/domain/complaint"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
)

// WorkspaceService 管理端消息中心与待办角标。
type WorkspaceService struct {
	Repos *application.RepoUnit
}

// BadgeCounts 左侧菜单角标（真实待办数）。
type BadgeCounts struct {
	Complaints     int64 `json:"complaints"`
	MerchantAudits int64 `json:"merchantAudits"`
	RiderAudits    int64 `json:"riderAudits"`
	Withdrawals    int64 `json:"withdrawals"`
}

// Message 消息中心条目（由真实待办/公告聚合）。
type Message struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Link      string `json:"link"`
	Level     string `json:"level"`
	CreatedAt int64  `json:"createdAt"`
}

// Messages 聚合待处理投诉、入驻审核、提现审核与已发布公告。
func (s *WorkspaceService) Messages(ctx context.Context, limit int, sc tenant.Scope) ([]Message, BadgeCounts, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var badges BadgeCounts
	var list []Message

	complaints, total, err := s.Repos.Complaints.List(ctx, complaint.Query{
		Page: 1, PageSize: limit, Status: string(complaint.StatusPending),
	}, sc)
	if err != nil {
		return nil, badges, err
	}
	badges.Complaints = total
	for _, c := range complaints {
		list = append(list, Message{
			ID: "complaint:" + c.ID, Type: "complaint", Title: "客诉待处理 · " + c.Title,
			Content: fmt.Sprintf("%s · 责任方 %s", c.Complainant, c.Target),
			Link:    "/complaints/" + c.ID, Level: "warning", CreatedAt: c.CreatedAt,
		})
	}

	merchants, total, err := s.Repos.Merchants.AuditList(ctx, merchant.AuditQuery{
		Page: 1, PageSize: limit, Status: string(merchant.AuditPending),
	}, sc)
	if err != nil {
		return nil, badges, err
	}
	badges.MerchantAudits = total
	for _, m := range merchants {
		list = append(list, Message{
			ID: "merchant:" + m.ID, Type: "merchant", Title: "商家入驻待审核 · " + m.Name,
			Content: "申请单 " + m.ApplyNo, Link: "/merchants/audit", Level: "info", CreatedAt: m.JoinedAt,
		})
	}

	riders, total, err := s.Repos.Riders.AuditList(ctx, rider.AuditQuery{
		Page: 1, PageSize: limit, Status: string(rider.AuditPending),
	}, sc)
	if err != nil {
		return nil, badges, err
	}
	badges.RiderAudits = total
	for _, r := range riders {
		list = append(list, Message{
			ID: "rider:" + r.ID, Type: "rider", Title: "骑手入驻待审核 · " + r.Name,
			Content: "申请单 " + r.ApplyNo, Link: "/riders/audit", Level: "info", CreatedAt: r.JoinedAt,
		})
	}

	withdrawals, total, err := s.Repos.Withdrawals.List(ctx, finance.WithdrawQuery{
		Page: 1, PageSize: limit, Status: string(finance.WithdrawPending),
	}, sc)
	if err != nil {
		return nil, badges, err
	}
	badges.Withdrawals = total
	for _, w := range withdrawals {
		list = append(list, Message{
			ID: "withdraw:" + w.ID, Type: "withdraw", Title: "提现待审核 · " + w.PartyName,
			Content: fmt.Sprintf("¥%.2f · %s", w.Amount, w.Kind),
			Link:    "/finance/withdraw/" + w.ID, Level: "warning", CreatedAt: w.CreatedAt,
		})
	}

	anns, _, err := s.Repos.Announcements.List(ctx, cms.AnnouncementQuery{
		Page: 1, PageSize: 5, Status: string(cms.AnnPublished),
	}, sc)
	if err != nil {
		return nil, badges, err
	}
	for _, a := range anns {
		list = append(list, Message{
			ID: "announcement:" + a.ID, Type: "system", Title: "平台公告 · " + a.Title,
			Content: "已发布", Link: "/settings/announcements", Level: "info", CreatedAt: a.PublishedAt,
		})
	}

	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt > list[j].CreatedAt })
	if len(list) > limit {
		list = list[:limit]
	}
	if list == nil {
		list = []Message{}
	}
	return list, badges, nil
}
