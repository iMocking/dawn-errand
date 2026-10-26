package http

import (
	"net/http"
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

// areaBody 服务区域写入字段白名单（后台面 camelCase，与 apps/admin 现发送口径一致）。
// 不再直绑 tenant.Region：实体一旦直绑，created_at/updated_at 就成了可写字段，
// 而且新增实体字段会自动变成新的注入面。
type areaBody struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	Villages int     `json:"villages"`
	RadiusKM float64 `json:"radiusKM"`
	StartFee float64 `json:"startFee"`
	FreeKM   float64 `json:"freeKM"`
	Notice   string  `json:"notice"`
}

func (b areaBody) toInput() admin.SaveAreaInput {
	return admin.SaveAreaInput{
		ID: b.ID, Name: b.Name, Status: b.Status, Villages: b.Villages,
		RadiusKM: b.RadiusKM, StartFee: b.StartFee, FreeKM: b.FreeKM, Notice: b.Notice,
	}
}

type saveSettingsBody struct {
	Action string    `json:"action"`
	Area   *areaBody `json:"area"`
}

// SaveSettings POST /settings（区域新增/编辑/停用/删除）。
func (h *Handler) SaveSettings(c *gin.Context) {
	var in saveSettingsBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.Action == "" {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "缺少 action（create/update/disable/delete）")
		return
	}
	var area *admin.SaveAreaInput
	if in.Area != nil {
		if (in.Action == "create" || in.Action == "update") && in.Area.ID == "" {
			in.Area.ID = ulidgen.New() // 预生成，响应体需要回传新建 id
		}
		input := in.Area.toInput()
		area = &input
	}
	u := middleware.User(c)
	if err := h.Settings.SaveArea(c.Request.Context(), admin.SaveAreaAction{Action: in.Action, Area: area}, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	id := ""
	if in.Area != nil {
		id = in.Area.ID
	}
	response.OKMsg(c, "操作成功", gin.H{"action": in.Action, "id": id})
}

// feeRuleBody 费率规则写入字段白名单。
type feeRuleBody struct {
	ID              string `json:"id"`
	Category        string `json:"category"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	Value           string `json:"value"`
	Desc            string `json:"desc"`
	Accent          bool   `json:"accent"`
	Highlight       bool   `json:"highlight"`
	Rule            string `json:"rule"`
	Example         string `json:"example"`
	Fee             string `json:"fee"`
	RiderShare      string `json:"riderShare"`
	IconColor       string `json:"iconColor"`
	Version         string `json:"version"`
	NextEffectiveAt int64  `json:"nextEffectiveAt"`
}

func (b feeRuleBody) toInput() admin.SaveFeeRuleInput {
	return admin.SaveFeeRuleInput{
		ID: b.ID, Category: b.Category, Name: b.Name, Status: b.Status, Value: b.Value,
		Desc: b.Desc, Accent: b.Accent, Highlight: b.Highlight, Rule: b.Rule, Example: b.Example,
		Fee: b.Fee, RiderShare: b.RiderShare, IconColor: b.IconColor, Version: b.Version,
		NextEffectiveAt: b.NextEffectiveAt,
	}
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
	var in feeRuleBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.ID == "" {
		in.ID = ulidgen.New()
	}
	u := middleware.User(c)
	if err := h.Settings.SaveFeeRule(c.Request.Context(), in.toInput(), u.ID, scope(c)); err != nil {
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

// announcementBody 公告写入字段白名单。
// created_by / created_at / action_name / stats 不再由请求体决定（原直绑 cms.Announcement 时可被任意覆盖）。
type announcementBody struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	TypeTone    string   `json:"typeTone"`
	Audience    string   `json:"audience"`
	Status      string   `json:"status"`
	Content     string   `json:"content"`
	Tags        []string `json:"tags"`
	PlannedAt   int64    `json:"plannedAt"`
	PublishedAt int64    `json:"publishedAt"`
}

func (b announcementBody) toInput() admin.SaveAnnouncementInput {
	return admin.SaveAnnouncementInput{
		ID: b.ID, Title: b.Title, Type: b.Type, TypeTone: b.TypeTone, Audience: b.Audience,
		Status: b.Status, Content: b.Content, Tags: b.Tags,
		PlannedAt: b.PlannedAt, PublishedAt: b.PublishedAt,
	}
}

type saveAnnouncementBody struct {
	Action       string           `json:"action"`
	Announcement announcementBody `json:"announcement"`
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
	if in.Announcement.ID == "" && in.Action == "save" {
		in.Announcement.ID = ulidgen.New()
	}
	u := middleware.User(c)
	if err := h.Settings.SaveAnnouncement(c.Request.Context(), in.Announcement.toInput(), in.Action, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "操作成功", gin.H{"id": in.Announcement.ID, "action": in.Action})
}
