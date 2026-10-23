package http

import (
	"github.com/gin-gonic/gin"

	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/realname"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/response"
)

// ListUsers GET /users。
func (h *Handler) ListUsers(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Users.List(c.Request.Context(), identity.UserQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: mapUserStatus(c.Query("status")),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]UserRowDto, 0, len(rows))
	for _, u := range rows {
		list = append(list, toUserRow(u))
	}
	pageResult(c, list, total, page, size)
}

// GetUsersOverview GET /users/overview。
func (h *Handler) GetUsersOverview(c *gin.Context) {
	ov, err := h.Users.Overview(c.Request.Context(), scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	activeRate := 0.0
	if ov.Total > 0 {
		activeRate = float64(ov.ActiveMonth) / float64(ov.Total) * 100
	}
	response.OK(c, UsersOverviewDto{
		Total: ov.Total, TodayNew: ov.TodayNew, ActiveMonth: ov.ActiveMonth,
		ActiveRate: round1(activeRate), AvgOrders: round1(ov.AvgOrders), Banned: ov.Banned,
	})
}

func round1(v float64) float64 {
	return float64(int64(v*10+0.5)) / 10
}

// GetUser GET /users/:id。
func (h *Handler) GetUser(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	d, err := h.Users.Detail(c.Request.Context(), id, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toUserDetail(d.User, d.Profile, d.TotalOrders, d.MonthOrders, d.TotalSpent, d.ComplaintCount, d.RecentOrders, d.RealName))
}

// GetUserOrders GET /users/:id/orders（用户详情订单记录，分页）。
func (h *Handler) GetUserOrders(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	page, size := parsePage(c)
	rows, total, err := h.Users.OrdersByUser(c.Request.Context(), id, page, size, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]UserRecentOrderDto, 0, len(rows))
	for _, o := range rows {
		list = append(list, toRecentOrder(o))
	}
	pageResult(c, list, total, page, size)
}

// GetUserComplaints GET /users/:id/complaints（用户详情投诉售后，分页）。
func (h *Handler) GetUserComplaints(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	page, size := parsePage(c)
	rows, total, err := h.Users.ComplaintsByUser(c.Request.Context(), id, page, size, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]ComplaintRowDto, 0, len(rows))
	for _, cc := range rows {
		list = append(list, toComplaintRow(cc))
	}
	pageResult(c, list, total, page, size)
}

// BanUser POST /users/:id/ban。
func (h *Handler) BanUser(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	if err := h.Users.Ban(c.Request.Context(), id, true, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已封禁", gin.H{"id": id, "status": "已封禁"})
}

// UnbanUser POST /users/:id/unban。
func (h *Handler) UnbanUser(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	if err := h.Users.Ban(c.Request.Context(), id, false, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已解封", gin.H{"id": id, "status": "正常"})
}

// UpdateUserRole POST /users/:id/role（账号权限集成到用户管理）。
func (h *Handler) UpdateUserRole(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	var in struct {
		Role string `json:"role"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	if err := h.Users.SetRole(c.Request.Context(), id, in.Role, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "角色已更新", gin.H{"id": id, "role": in.Role})
}

// ListRealNames GET /users/realname-verifications（实名认证审核列表）。
func (h *Handler) ListRealNames(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Users.RealNames(c.Request.Context(), realname.Query{
		Page: page, PageSize: size, Status: c.Query("status"),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]RealNameVerifyDto, 0, len(rows))
	for _, v := range rows {
		list = append(list, *toRealNameVerification(v))
	}
	pageResult(c, list, total, page, size)
}

type auditRealNameBody struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// AuditRealName POST /users/realname-verifications/:id/audit（通过/驳回实名认证）。
func (h *Handler) AuditRealName(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	var in auditRealNameBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	if err := h.Users.AuditRealName(c.Request.Context(), id, in.Action, in.Reason, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	msg := "已通过实名认证"
	if in.Action == "reject" {
		msg = "已驳回实名认证申请"
	}
	response.OKMsg(c, msg, gin.H{"id": id, "action": in.Action})
}
