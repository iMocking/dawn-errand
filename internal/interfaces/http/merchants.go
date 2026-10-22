package http

import (
	"strings"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/response"
)

// ListMerchants GET /merchants。
func (h *Handler) ListMerchants(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Merchants.List(c.Request.Context(), merchant.MerchantQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: mapMerchantStatus(c.Query("status")),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]MerchantRowDto, 0, len(rows))
	for _, m := range rows {
		list = append(list, toMerchantRow(m))
	}
	pageResult(c, list, total, page, size)
}

type createMerchantBody struct {
	Name           string  `json:"name"`
	Category       string  `json:"category"`
	Contact        string  `json:"contact"`
	Phone          string  `json:"phone"`
	Address        string  `json:"address"`
	AddressSub     string  `json:"addressSub"`
	OpenHours      string  `json:"openHours"`
	DeliveryRange  string  `json:"deliveryRange"`
	Commission     float64 `json:"commission"`
	DeliveryMethod string  `json:"deliveryMethod"`
	RegionID       string  `json:"regionId"`
}

// CreateMerchant POST /merchants。
func (h *Handler) CreateMerchant(c *gin.Context) {
	var in createMerchantBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Phone) == "" {
		response.ErrStatus(c, 400, 1301, "商家名称与联系电话为必填项")
		return
	}
	m := &merchant.Merchant{
		RegionID: in.RegionID, Name: strings.TrimSpace(in.Name), Category: in.Category,
		Contact: in.Contact, Phone: in.Phone, Address: in.Address, AddressSub: in.AddressSub,
		OpenHours: in.OpenHours, DeliveryRange: in.DeliveryRange, Commission: in.Commission,
		DeliveryMethod: in.DeliveryMethod,
		Status: merchant.MerchantStatusRest, Operator: in.Contact,
	}
	if err := h.Merchants.Create(c.Request.Context(), m, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已提交入驻申请", gin.H{"id": m.ID, "applyNo": m.ApplyNo, "auditStatus": "pending"})
}

// GetMerchantStats GET /merchants/stats。
func (h *Handler) GetMerchantStats(c *gin.Context) {	st, err := h.Merchants.Stats(c.Request.Context(), scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, MerchantStatsDto{
		Total: st.Total, WeekNew: st.WeekNew, Open: st.Open,
		OpenRate: pctStr(st.OpenRate), Pending: st.Pending, Closed: st.Closed,
		ClosedDetail: "休息中 " + thousandSep(st.Closed) + " 家",
	})
}

// ListMerchantAudit GET /merchants/audit。
func (h *Handler) ListMerchantAudit(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Merchants.AuditList(c.Request.Context(), merchant.AuditQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: mapAuditStatus(c.Query("status")),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]MerchantAuditRowDto, 0, len(rows))
	for _, m := range rows {
		list = append(list, toMerchantAuditRow(m))
	}
	pageResult(c, list, total, page, size)
}

// GetMerchant GET /merchants/:id。
func (h *Handler) GetMerchant(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	m, err := h.Merchants.Detail(c.Request.Context(), id, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toMerchantDetail(m))
}

type auditMerchantBody struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// AuditMerchant POST /merchants/audit/:id。
func (h *Handler) AuditMerchant(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	var in auditMerchantBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.Action != "approve" && in.Action != "reject" {
		response.ErrStatus(c, 400, 1301, "action 仅支持 approve/reject")
		return
	}
	u := middleware.User(c)
	if err := h.Merchants.AuditAction(c.Request.Context(), id, in.Action, in.Reason, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "操作成功", gin.H{"id": id, "action": in.Action})
}
