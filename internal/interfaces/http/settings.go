package http

import (
	"time"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/application/admin"
	"dawn-errand/internal/domain/cms"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/response"
	"dawn-errand/internal/pkg/ulidgen"
)

// GetSettings GET /settings。
func (h *Handler) GetSettings(c *gin.Context) {
	res, err := h.Settings.GetSettings(c.Request.Context(), scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	areas := make([]ServiceAreaDto, 0, len(res.Areas))
	now := time.Now().Unix()
	for _, r := range res.Areas {
		var onlineMerchants, activeRiders, todayOrders *int64
		if r.Status == tenant.RegionStatusOpen {
			om, err1 := h.Merchants.Stats(c.Request.Context(), regionScope(c, r.ID))
			ar, err2 := h.Riders.Stats(c.Request.Context(), regionScope(c, r.ID))
			to, err3 := h.Orders.CountByRegion(c.Request.Context(), r.ID, now)
			if err1 == nil {
				v := om.Open
				onlineMerchants = &v
			}
			if err2 == nil {
				v := ar.Online
				activeRiders = &v
			}
			if err3 == nil {
				v := to
				todayOrders = &v
			}
		}
		areas = append(areas, toServiceArea(r, onlineMerchants, activeRiders, todayOrders))
	}
	response.OK(c, gin.H{"total_towns": res.TotalTowns, "areas": areas})
}

// regionScope 构建区域站长视角的 scope（仅用于统计）。
func regionScope(c *gin.Context, regionID string) tenant.Scope {
	return tenant.Scope{Role: "region_station", RegionID: regionID}
}

type saveSettingsBody struct {
	Action string         `json:"action"`
	Area   *tenant.Region `json:"area"`
}

// SaveSettings POST /settings（区域新增/编辑/停用/删除）。
func (h *Handler) SaveSettings(c *gin.Context) {
	var in saveSettingsBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.Action == "" {
		response.ErrStatus(c, 400, 1301, "缺少 action（create/update/disable/delete）")
		return
	}
	if (in.Action == "create" || in.Action == "update") && in.Area != nil && in.Area.ID == "" {
		in.Area.ID = ulidgen.New()
	}
	u := middleware.User(c)
	if err := h.Settings.SaveArea(c.Request.Context(), admin.SaveAreaAction{Action: in.Action, Area: in.Area}, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "操作成功", gin.H{"action": in.Action, "id": areaID(in)})
}

func areaID(in saveSettingsBody) string {
	if in.Area != nil {
		return in.Area.ID
	}
	return ""
}

// ListFeeRules GET /settings/fee-rules。
func (h *Handler) ListFeeRules(c *gin.Context) {
	rules, err := h.Settings.GetFeeRules(c.Request.Context(), scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]FeeRuleRowDto, 0, len(rules))
	for _, r := range rules {
		list = append(list, toFeeRuleRow(r))
	}
	response.OK(c, gin.H{"list": list, "total": len(list), "page": 1, "pageSize": len(list)})
}

// SaveFeeRule POST /settings/fee-rules。
func (h *Handler) SaveFeeRule(c *gin.Context) {
	var in cms.FeeRule
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.ID == "" {
		in.ID = ulidgen.New()
	}
	u := middleware.User(c)
	if err := h.Settings.SaveFeeRule(c.Request.Context(), in, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "规则已保存", gin.H{"id": in.ID})
}

// ListAnnouncements GET /settings/announcements。
func (h *Handler) ListAnnouncements(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Settings.ListAnnouncements(c.Request.Context(), cms.AnnouncementQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: c.Query("status"),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]AnnouncementRowDto, 0, len(rows))
	for _, a := range rows {
		list = append(list, toAnnouncementRow(a))
	}
	pageResult(c, list, total, page, size)
}

type saveAnnouncementBody struct {
	Action       string           `json:"action"`
	Announcement cms.Announcement `json:"announcement"`
}

// SaveAnnouncement POST /settings/announcements。
func (h *Handler) SaveAnnouncement(c *gin.Context) {
	var in saveAnnouncementBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.Action == "" {
		in.Action = "save"
	}
	switch in.Action {
	case "发布":
		in.Action = "publish"
	case "下线":
		in.Action = "offline"
	case "撤回":
		in.Action = "withdraw"
	}
	if in.Announcement.ID == "" {
		if in.Action != "save" {
			response.ErrStatus(c, 400, errcode.CodeBadRequest, "公告 id 必填")
			return
		}
		in.Announcement.ID = ulidgen.New()
	}
	u := middleware.User(c)
	if err := h.Settings.SaveAnnouncement(c.Request.Context(), in.Announcement, in.Action, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "操作成功", gin.H{"id": in.Announcement.ID, "action": in.Action})
}
