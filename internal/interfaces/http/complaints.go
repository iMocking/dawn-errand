package http

import (
	"github.com/gin-gonic/gin"

	"dawn-errand/internal/domain/complaint"
	"dawn-errand/internal/pkg/response"
)

// ListComplaints GET /complaints。
func (h *Handler) ListComplaints(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Complaints.List(c.Request.Context(), complaint.Query{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: c.Query("status"),
	}, scope(c))
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

// GetComplaint GET /complaints/:id。
func (h *Handler) GetComplaint(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	cc, err := h.Complaints.Detail(c.Request.Context(), id, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toComplaintDetail(cc))
}
