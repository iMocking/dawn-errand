package http

import (
	"github.com/gin-gonic/gin"

	"dawn-errand/internal/pkg/response"
)

// GetMessages GET /messages：管理端消息中心 + 左侧菜单待办角标。
func (h *Handler) GetMessages(c *gin.Context) {
	list, badges, err := h.Workspace.Messages(c.Request.Context(), 20, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"list":   list,
		"badges": badges,
		"total":  len(list),
	})
}
