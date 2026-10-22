package http

import (
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
	if err := h.Riders.AuditAction(c.Request.Context(), in.ID, in.Action, in.Reason, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "操作成功", gin.H{"id": in.ID, "action": in.Action})
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
