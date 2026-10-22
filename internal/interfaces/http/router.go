package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/response"
)

// NewRouter 组装 Gin 路由（/api 前缀，统一返回结构，对齐前端契约）。
func NewRouter(h *Handler, mw *middleware.Middleware) *gin.Engine {
	if h.Cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(mw.RequestID(), mw.Metrics(), mw.AccessLog(), mw.Recovery())

	// 健康检查与指标
	r.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 404/405 统一返回
	r.NoRoute(func(c *gin.Context) {
		response.ErrStatus(c, http.StatusNotFound, 1401, "接口不存在")
	})

	api := r.Group("/api")
	{
		// ---- 公开：认证 ----
		authPub := api.Group("/auth")
		authPub.POST("/login", mw.RateLimit("auth:login", 30, time.Minute), h.Login)
		authPub.POST("/register", mw.RateLimit("auth:register", 10, time.Minute), h.Register)
		authPub.POST("/refresh", mw.RateLimit("auth:refresh", 30, time.Minute), h.Refresh)
		authPub.GET("/email/verify", h.VerifyEmail)
		authPub.POST("/email/resend", mw.RateLimit("auth:resend", 10, time.Minute), h.ResendVerification)
		authPub.POST("/password/forgot", mw.RateLimit("auth:forgot", 10, time.Minute), h.ForgotPassword)
		authPub.POST("/password/reset", mw.RateLimit("auth:reset", 10, time.Minute), h.ResetPassword)

		// ---- 公开：初始化 ----
		sysPub := api.Group("/system")
		sysPub.GET("/bootstrap/status", h.BootstrapStatus)
		sysPub.POST("/bootstrap", mw.RateLimit("system:bootstrap", 5, time.Minute), h.Bootstrap)

		// ---- 需登录 ----
		authd := api.Group("")
		authd.Use(mw.Auth())
		{
			auth := authd.Group("/auth")
			auth.POST("/logout", h.Logout)
			auth.GET("/me", h.Me)
			auth.POST("/password/change", h.ChangePassword)

			// 系统配置（仅 super_admin）
			sys := authd.Group("/system", mw.RequirePerm("system", "config"))
			sys.GET("/config/smtp", h.GetSMTP)
			sys.PUT("/config/smtp", h.UpdateSMTP)
			sys.POST("/config/smtp/test", h.TestSMTP)
			sys.GET("/config/sms", h.GetSMS)
			sys.PUT("/config/sms", h.UpdateSMS)
			sys.POST("/config/sms/test", h.TestSMS)
			sys.GET("/config/withdraw-rule", h.GetWithdrawRule)
			sys.PUT("/config/withdraw-rule", h.UpdateWithdrawRule)
			sys.GET("/config/oss", h.GetOSS)
			sys.PUT("/config/oss", h.UpdateOSS)
			sys.GET("/config/storage/local", h.GetLocalStorage)
			sys.PUT("/config/storage/local", h.UpdateLocalStorage)
			sys.POST("/config/storage/local/test", h.TestLocalStorage)
			sys.GET("/config/casbin", h.GetCasbinConfig)
			sys.PUT("/config/casbin", h.UpdateCasbinConfig)
			if h.Cfg.IsDev() {
				sys.GET("/dev/outbox", h.DevOutbox)
			}

			// 数据总览
			authd.GET("/dashboard", mw.RequirePerm("dashboard", "view"), h.GetDashboard)

			// 订单
			orders := authd.Group("/orders")
			orders.GET("", mw.RequirePerm("order", "list"), h.ListOrders)
			orders.GET("/:id", mw.RequirePerm("order", "detail"), h.GetOrder)

			// 商家
			merchants := authd.Group("/merchants")
			merchants.GET("", mw.RequirePerm("merchant", "list"), h.ListMerchants)
			merchants.POST("", mw.RequirePerm("merchant", "create"), h.CreateMerchant)
			merchants.GET("/stats", mw.RequirePerm("merchant", "list"), h.GetMerchantStats)
			merchants.GET("/audit", mw.RequirePerm("merchant", "audit"), h.ListMerchantAudit)
			merchants.POST("/audit/:id", mw.RequirePerm("merchant", "audit"), h.AuditMerchant)
			merchants.GET("/:id", mw.RequirePerm("merchant", "detail"), h.GetMerchant)

			// 骑手
			riders := authd.Group("/riders")
			riders.GET("", mw.RequirePerm("rider", "list"), h.ListRiders)
			riders.POST("", mw.RequirePerm("rider", "create"), h.CreateRider)
			riders.GET("/stats", mw.RequirePerm("rider", "list"), h.GetRiderStats)
			riders.GET("/audit", mw.RequirePerm("rider", "audit"), h.ListRiderAudit)
			riders.POST("/audit", mw.RequirePerm("rider", "audit"), h.AuditRider)
			riders.POST("/freeze", mw.RequirePerm("rider", "freeze"), h.FreezeRider)
			riders.GET("/:id", mw.RequirePerm("rider", "detail"), h.GetRider)

			// 用户
			users := authd.Group("/users")
			users.GET("", mw.RequirePerm("user", "list"), h.ListUsers)
			users.GET("/overview", mw.RequirePerm("user", "list"), h.GetUsersOverview)
			users.POST("/:id/ban", mw.RequirePerm("user", "ban"), h.BanUser)
			users.POST("/:id/unban", mw.RequirePerm("user", "unban"), h.UnbanUser)
			users.POST("/:id/role", mw.RequirePerm("user", "role"), h.UpdateUserRole)
			users.GET("/:id/orders", mw.RequirePerm("user", "detail"), h.GetUserOrders)
			users.GET("/:id/complaints", mw.RequirePerm("user", "detail"), h.GetUserComplaints)
			users.GET("/:id", mw.RequirePerm("user", "detail"), h.GetUser)

			// 财务
			finance := authd.Group("/finance")
			finance.GET("", mw.RequirePerm("finance", "view"), h.GetFinance)
			finance.GET("/settlement", mw.RequirePerm("finance", "view"), h.ListSettlements)
			finance.GET("/withdraws", mw.RequirePerm("finance", "view"), h.ListWithdraws)
			finance.GET("/withdraw/:id", mw.RequirePerm("finance", "view"), h.GetWithdraw)
			finance.POST("/withdraw/:id/audit", mw.RequirePerm("finance", "audit"), h.AuditWithdraw)

			// 投诉
			complaints := authd.Group("/complaints")
			complaints.GET("", mw.RequirePerm("complaint", "list"), h.ListComplaints)
			complaints.GET("/:id", mw.RequirePerm("complaint", "detail"), h.GetComplaint)

			// 系统设置（服务区域/费率/公告）
			settings := authd.Group("/settings")
			settings.GET("", mw.RequirePerm("settings", "list"), h.GetSettings)
			settings.POST("", mw.RequirePerm("settings", "area"), h.SaveSettings)
			settings.GET("/fee-rules", mw.RequirePerm("settings", "list"), h.ListFeeRules)
			settings.POST("/fee-rules", mw.RequirePerm("settings", "save"), h.SaveFeeRule)
			settings.GET("/announcements", mw.RequirePerm("settings", "list"), h.ListAnnouncements)
			settings.POST("/announcements", mw.RequirePerm("settings", "save"), h.SaveAnnouncement)
		}
	}
	return r
}
