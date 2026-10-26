package http

import (
	"strings"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/response"
)

// ListRiders GET /riders。
func (h *Handler) ListRiders(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Riders.List(c.Request.Context(), rider.RiderQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: mapRiderStatus(c.Query("status")),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]RiderRowDto, 0, len(rows))
	for _, r := range rows {
		list = append(list, toRiderRow(r))
	}
	pageResult(c, list, total, page, size)
}

type createRiderBody struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Area        string `json:"area"`
	Tool        string `json:"tool"`
	OrderTypes  string `json:"orderTypes"`
	IDCard      string `json:"idCard"`
	ServiceType string `json:"serviceType"`
	Vehicle     string `json:"vehicle"`
	Account     string `json:"account"`
	RegionID    string `json:"regionId"`
}

// CreateRider POST /riders。
func (h *Handler) CreateRider(c *gin.Context) {
	var in createRiderBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Phone) == "" {
		response.ErrStatus(c, 400, 1301, "骑手姓名与联系电话为必填项")
		return
	}
	r := &rider.Rider{
		RegionID: in.RegionID, Name: strings.TrimSpace(in.Name), Phone: in.Phone,
		Area: in.Area, Tool: in.Tool, OrderTypes: in.OrderTypes, IDCard: in.IDCard,
		ServiceType: in.ServiceType, Vehicle: in.Vehicle, Account: in.Account,
		Status: rider.RiderOffline, IDCheck: "待补充", NoticeType: "none",
	}
	if err := h.Riders.Create(c.Request.Context(), r, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已提交入驻申请", gin.H{"id": r.ID, "applyNo": r.ApplyNo, "auditStatus": "pending"})
}

// GetRiderStats GET /riders/stats。
func (h *Handler) GetRiderStats(c *gin.Context) {
	st, err := h.Riders.Stats(c.Request.Context(), scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	weekDelta := st.WeekDelta
	if weekDelta == "" {
		weekDelta = "▲ 8"
	}
	response.OK(c, RiderStatsDto{
		Total: st.Total, Online: st.Online,
		OnlineRate: pctStr(st.OnlineRate), PendingAudit: st.PendingAudit,
		Frozen: st.Frozen, FrozenNote: "冻结" + thousandSep(st.Frozen) + " 名", WeekDelta: weekDelta,
		Audit: RiderAuditStatsDto{
			TodayNew: st.Audit.TodayNew, Approved: st.Audit.Approved,
			Rejected: st.Audit.Rejected, MonthApproved: st.Audit.MonthApproved,
			MomRate: st.Audit.MomRate,
		},
	})
}

// ListRiderAudit GET /riders/audit。
func (h *Handler) ListRiderAudit(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Riders.AuditList(c.Request.Context(), rider.AuditQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: mapAuditStatus(c.Query("status")),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]RiderAuditRowDto, 0, len(rows))
	for _, r := range rows {
		list = append(list, toRiderAuditRow(r))
	}
	pageResult(c, list, total, page, size)
}

// GetRider GET /riders/:id。
func (h *Handler) GetRider(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	r, err := h.Riders.Detail(c.Request.Context(), id, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toRiderDetail(r))
}

// GetRiderWorkspace GET /riders/:id/workspace：详情页各 Tab 数据。
func (h *Handler) GetRiderWorkspace(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	ws, err := h.Riders.Workspace(c.Request.Context(), id, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	orders := make([]gin.H, 0, len(ws.Orders))
	for _, o := range ws.Orders {
		orders = append(orders, orderBrief(o))
	}
	abnormal := make([]gin.H, 0, len(ws.Abnormal))
	for _, o := range ws.Abnormal {
		abnormal = append(abnormal, orderBrief(o))
	}
	reviews := make([]gin.H, 0, len(ws.Reviews))
	for _, rv := range ws.Reviews {
		reviews = append(reviews, gin.H{
			"id": rv.ID, "orderId": rv.OrderID, "userId": rv.UserID,
			"userName": ws.UserNames[rv.UserID], "rating": rv.Rating,
			"tags": rv.Tags, "content": rv.Content, "reply": rv.Reply,
			"createdAt": rv.CreatedAt,
		})
	}
	r := ws.Rider
	response.OK(c, gin.H{
		"records":  ws.Records,
		"orders":   orders,
		"abnormal": abnormal,
		"trend":    ws.Trend,
		"reviews":  gin.H{"list": reviews, "total": len(reviews), "rating": r.Rating},
		"profile": gin.H{
			"idCard": maskIDCard(r.IDCard), "idCheck": r.IDCheck, "tool": r.Tool,
			"vehicle": r.Vehicle, "account": r.Account, "docs": r.Docs,
			"docCount": r.DocCount, "auditStatus": riderAuditText(r.AuditStatus),
			"submittedAt": r.SubmittedAt, "joinedAt": r.JoinedAt,
			"serviceType": r.ServiceType, "orderTypes": r.OrderTypes,
		},
		"performance": gin.H{
			"rating": r.Rating, "onTimeRate": pctStr(r.OnTimeRate),
			"avgMinutes": r.AvgMinutes, "complaintsToday": r.ComplaintsToday,
			"todayOrders": r.TodayOrders, "monthIncome": r.MonthIncome,
			"onlineHours": r.OnlineHours,
		},
	})
}

type riderAuditBody struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// AuditRider POST /riders/audit。
func (h *Handler) AuditRider(c *gin.Context) {
	var in riderAuditBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.Action != "approve" && in.Action != "reject" {
		response.ErrStatus(c, 400, 1301, "action 仅支持 approve/reject")
		return
	}
	u := middleware.User(c)
	account, err := h.Riders.AuditAction(c.Request.Context(), in.ID, in.Action, in.Reason, u.ID, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	body := gin.H{"id": in.ID, "action": in.Action}
	if account != nil {
		body["account"] = gin.H{
			"userId": account.UserID, "email": account.Email, "phone": account.Phone,
			"initialPassword": account.InitialPassword,
		}
	}
	response.OKMsg(c, "操作成功", body)
}

// ResetRiderAccount POST /riders/:id/account/reset。
func (h *Handler) ResetRiderAccount(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	res, err := h.Accounts.ResetRiderPassword(c.Request.Context(), id, u.ID, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "密码已重置", gin.H{
		"account": gin.H{
			"userId": res.UserID, "email": res.Email, "phone": res.Phone,
			"initialPassword": res.InitialPassword,
		},
	})
}

// FreezeRider POST /riders/freeze。
func (h *Handler) FreezeRider(c *gin.Context) {
	var in riderAuditBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.Action != "freeze" && in.Action != "unfreeze" {
		response.ErrStatus(c, 400, 1301, "action 仅支持 freeze/unfreeze")
		return
	}
	u := middleware.User(c)
	if err := h.Riders.Freeze(c.Request.Context(), in.ID, in.Action, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "操作成功", gin.H{"id": in.ID, "action": in.Action})
}
