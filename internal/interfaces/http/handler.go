// Package http Gin 接口层：路由、Handler、DTO（对齐前端 apps/admin 契约）。
package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/application"
	"dawn-errand/internal/application/admin"
	"dawn-errand/internal/application/auth"
	"dawn-errand/internal/application/system"
	"dawn-errand/internal/config"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/response"
)

// Handler 所有 HTTP Handler 的依赖集合。
type Handler struct {
	Cfg    *config.Config
	MW     *middleware.Middleware
	Auth   *auth.Service
	System *system.Service

	Dashboard  *admin.DashboardService
	Orders     *admin.OrdersService
	Merchants  *admin.MerchantsService
	Riders     *admin.RidersService
	Users      *admin.UsersService
	Finance    *admin.FinanceService
	Complaints *admin.ComplaintsService
	Settings   *admin.SettingsService
}

// principal 取当前主体。
func principal(c *gin.Context) *tenant.Principal {
	return middleware.Principal(c)
}

// scope 取数据隔离作用域。
func scope(c *gin.Context) tenant.Scope {
	p := principal(c)
	return p.Scope()
}

// pathID 解析路径参数 id。
func pathID(c *gin.Context) (string, error) {
	id := c.Param("id")
	if id == "" {
		return "", errcode.NewStatus(http.StatusBadRequest, errcode.CodeInvalidID, "缺少资源 ID")
	}
	return id, nil
}

// bind 绑定 JSON 请求体。
func bind(c *gin.Context, v any) error {
	if err := c.ShouldBindJSON(v); err != nil {
		return errcode.NewStatus(http.StatusBadRequest, errcode.CodeBadRequest, "请求参数错误："+err.Error())
	}
	return nil
}

// pageQuery 解析分页参数。
type pageQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
}

func (q *pageQuery) page() int { return maxInt(q.Page, 1) }
func (q *pageQuery) pageSize() int {
	if q.PageSize < 1 || q.PageSize > 200 {
		return 10
	}
	return q.PageSize
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// pageResult 统一分页返回。
func pageResult(c *gin.Context, list any, total int64, page, pageSize int) {
	response.OK(c, gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// parsePage 从 query 解析分页（兼容 form binding）。
func parsePage(c *gin.Context) (page, size int) {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	s, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if p < 1 {
		p = 1
	}
	if s < 1 || s > 200 {
		s = 10
	}
	return p, s
}

// errNotFound 通用资源不存在。
func notFound(msg string) error {
	return errcode.NewStatus(http.StatusNotFound, errcode.CodeNotFound, msg)
}

// isNotFound 判断是否未找到错误。
func isNotFound(err error) bool { return errors.Is(err, application.ErrNotFound) }

// ---- 查询状态映射：前端列表按中文状态筛选，repo 层比对域内英文枚举 ----

// mapMerchantStatus 商家营业状态：中文 → 英文枚举。
func mapMerchantStatus(s string) string {
	switch s {
	case "营业中":
		return "open"
	case "休息中":
		return "rest"
	case "已封禁":
		return "banned"
	}
	return s
}

// mapRiderStatus 骑手状态：中文 → 英文枚举。
func mapRiderStatus(s string) string {
	switch s {
	case "在线":
		return "online"
	case "忙碌":
		return "busy"
	case "离线":
		return "offline"
	case "已冻结":
		return "frozen"
	}
	return s
}

// mapUserStatus 用户状态：中文 → 英文枚举。
func mapUserStatus(s string) string {
	switch s {
	case "正常":
		return "active"
	case "已封禁":
		return "banned"
	}
	return s
}

// mapAuditStatus 入驻/审核状态：中文 → 英文枚举。
func mapAuditStatus(s string) string {
	switch s {
	case "已通过":
		return "approved"
	case "已驳回":
		return "rejected"
	case "待审核":
		return "pending"
	}
	return s
}
