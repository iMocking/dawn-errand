package http

import (
	"crypto/subtle"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/response"
)

// defaultBodyLimitBytes 非上传接口的请求体上限（1MB）。
const defaultBodyLimitBytes int64 = 1 << 20

// NewRouter 组装 Gin 路由（/api 前缀，统一返回结构，对齐前端契约）。
func NewRouter(h *Handler, mw *middleware.Middleware) *gin.Engine {
	if h.Cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// 请求体上限：JSON 1MB，上传路由按 STORAGE_MAX_SIZE_MB + 1MB 余量放宽。
	uploadMaxMB := h.Cfg.StorageMaxSizeMB
	if uploadMaxMB <= 0 {
		uploadMaxMB = 10
	}
	r.Use(mw.RequestID(), mw.Metrics(), mw.AccessLog(), mw.Recovery(), middleware.SecurityHeaders(),
		middleware.BodyLimit(defaultBodyLimitBytes, (uploadMaxMB+1)*1024*1024))

	// 健康检查与指标
	r.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if h.Ready == nil {
			response.OK(c, gin.H{"status": "ready"})
			return
		}
		if err := h.Ready(c.Request.Context()); err != nil {
			response.ErrStatus(c, http.StatusServiceUnavailable, 1500, "依赖未就绪")
			return
		}
		response.OK(c, gin.H{"status": "ready"})
	})
	r.GET("/metrics", h.metricsHandler())

	// 本地上传文件静态访问：仅挂载公开桶子目录 <root>/<mount>；
	// 私有桶（证件等 PII）位于 <root>/private，不在静态路由内，只能经 /api/uploads/private 鉴权读取。
	mount := h.StorageMount
	if mount == "" {
		mount = "uploads"
	}
	r.Static("/"+strings.Trim(mount, "/"), filepath.Join(h.Cfg.StorageRootDir, strings.Trim(mount, "/")))

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
		authPub.POST("/sms/send", mw.RateLimit("auth:sms:send", 30, time.Minute), h.SendSMSCode)
		authPub.POST("/sms/login", mw.RateLimit("auth:sms:login", 30, time.Minute), h.LoginBySMS)

		// ---- 公开：支付回调（HMAC 验签）----
		api.POST("/payments/callback/:channel", mw.RateLimit("payment:callback", 60, time.Minute), h.PaymentCallback)

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
			auth.GET("/realname-verification", h.MyRealName)
			auth.POST("/realname-verification", h.SubmitRealName)

			// 私有桶受控读取（需登录，归属校验在 handler 内）。
			authd.GET("/uploads/private/*rel", h.PrivateUploadDownload)

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
			authd.GET("/messages", mw.RequirePerm("dashboard", "view"), h.GetMessages)
			// 管理端上传：仅限平台/区域管理角色（三端 App 走 /api/app/common/uploads），
			// 避免任意登录用户借该路由向存储桶写文件。
			authd.POST("/uploads",
				mw.RequireAppRole(identity.RoleSuperAdmin, identity.RoleCSManager, identity.RoleCSAgent, identity.RoleRegionStation),
				mw.RateLimit("upload", 60, time.Minute), h.AppUpload)

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
			merchants.POST("/:id/account", mw.RequirePerm("merchant", "create"), h.ProvisionMerchantAccount)
			merchants.POST("/:id/account/reset", mw.RequirePerm("merchant", "create"), h.ResetMerchantAccount)
			merchants.GET("/:id/workspace", mw.RequirePerm("merchant", "detail"), h.GetMerchantWorkspace)
			merchants.PUT("/:id/application", mw.RequirePerm("merchant", "create"), h.UpdateMerchantApplication)
			merchants.GET("/:id", mw.RequirePerm("merchant", "detail"), h.GetMerchant)

			// 骑手
			riders := authd.Group("/riders")
			riders.GET("", mw.RequirePerm("rider", "list"), h.ListRiders)
			riders.POST("", mw.RequirePerm("rider", "create"), h.CreateRider)
			riders.GET("/stats", mw.RequirePerm("rider", "list"), h.GetRiderStats)
			riders.GET("/audit", mw.RequirePerm("rider", "audit"), h.ListRiderAudit)
			riders.POST("/audit", mw.RequirePerm("rider", "audit"), h.AuditRider)
			riders.POST("/freeze", mw.RequirePerm("rider", "freeze"), h.FreezeRider)
			riders.POST("/:id/account", mw.RequirePerm("rider", "create"), h.ProvisionRiderAccount)
			riders.POST("/:id/account/reset", mw.RequirePerm("rider", "create"), h.ResetRiderAccount)
			riders.GET("/:id/workspace", mw.RequirePerm("rider", "detail"), h.GetRiderWorkspace)
			riders.GET("/:id", mw.RequirePerm("rider", "detail"), h.GetRider)

			// 用户
			users := authd.Group("/users")
			users.GET("", mw.RequirePerm("user", "list"), h.ListUsers)
			users.GET("/overview", mw.RequirePerm("user", "list"), h.GetUsersOverview)
			users.POST("/:id/ban", mw.RequirePerm("user", "ban"), h.BanUser)
			users.POST("/:id/unban", mw.RequirePerm("user", "unban"), h.UnbanUser)
			users.POST("/:id/role", mw.RequirePerm("user", "role"), h.UpdateUserRole)
			users.GET("/realname-verifications", mw.RequirePerm("user", "audit"), h.ListRealNames)
			users.POST("/realname-verifications/:id/audit", mw.RequirePerm("user", "audit"), h.AuditRealName)
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
			finance.GET("/reconciliation", mw.RequirePerm("finance", "view"), h.GetReconciliation)

			// 投诉
			complaints := authd.Group("/complaints")
			complaints.GET("", mw.RequirePerm("complaint", "list"), h.ListComplaints)
			complaints.GET("/:id", mw.RequirePerm("complaint", "detail"), h.GetComplaint)

			// 优惠券运营
			coupons := authd.Group("/coupons")
			{
				coupons.GET("", mw.RequirePerm("coupon", "view"), h.ListCoupons)
				coupons.POST("", mw.RequirePerm("coupon", "manage"), h.IssueCoupon)
			}

			// 系统设置（服务区域/费率/公告）
			settings := authd.Group("/settings")
			settings.GET("", mw.RequirePerm("settings", "list"), h.GetSettings)
			settings.POST("", mw.RequirePerm("settings", "area"), h.SaveSettings)
			settings.GET("/fee-rules", mw.RequirePerm("settings", "list"), h.ListFeeRules)
			settings.POST("/fee-rules", mw.RequirePerm("settings", "save"), h.SaveFeeRule)
			settings.GET("/announcements", mw.RequirePerm("settings", "list"), h.ListAnnouncements)
			settings.POST("/announcements", mw.RequirePerm("settings", "save"), h.SaveAnnouncement)
		}

		// ---- App 面（三端专用，独立命名空间，snake_case 契约）----
		appAPI := api.Group("/app")
		{
			client := appAPI.Group("/client", mw.Auth(), mw.RequireAppRole(identity.RoleUser))
			client.GET("/me", h.AppClientMe)
			client.PUT("/me", h.AppUpdateClientProfile)
			client.GET("/regions", h.AppListRegions)
			client.PUT("/region", h.AppSetRegion)
			client.GET("/merchants", h.AppListMerchants)
			client.GET("/merchants/:id", h.AppGetMerchant)
			client.GET("/addresses", h.AppListAddresses)
			client.POST("/addresses", h.AppCreateAddress)
			client.PUT("/addresses/:id", h.AppUpdateAddress)
			client.DELETE("/addresses/:id", h.AppDeleteAddress)
			client.POST("/addresses/:id/default", h.AppSetDefaultAddress)
			client.POST("/orders", mw.RateLimit("app:order:create", 30, time.Minute), h.AppCreateOrder)
			client.GET("/orders", h.AppListClientOrders)
			client.GET("/orders/:id", h.AppGetClientOrder)
			client.POST("/orders/:id/cancel", h.AppCancelClientOrder)
			client.POST("/orders/:id/pay", h.AppPayOrder)
			client.POST("/orders/:id/confirm", h.AppConfirmOrder)
			client.GET("/orders/:id/track", h.AppTrackOrder)
			client.POST("/orders/:id/review", h.AppSubmitReview)
			client.POST("/orders/:id/after-sale", h.AppApplyAfterSale)
			client.GET("/coupons", h.AppListCoupons)

			store := appAPI.Group("/store", mw.Auth(), mw.RequireAppRole(identity.RoleMerchant))
			store.GET("/me", h.AppStoreMe)
			store.PUT("/status", h.AppUpdateStoreStatus)
			store.GET("/business", h.AppGetStoreBusiness)
			store.PUT("/business", h.AppUpdateStoreBusiness)
			store.GET("/menu", h.AppStoreMenu)
			store.POST("/categories", h.AppCreateCategory)
			store.PUT("/categories/:id", h.AppUpdateCategory)
			store.DELETE("/categories/:id", h.AppDeleteCategory)
			store.POST("/dishes", h.AppCreateDish)
			store.PUT("/dishes/:id", h.AppUpdateDish)
			store.DELETE("/dishes/:id", h.AppDeleteDish)
			store.POST("/dishes/:id/on-sale", h.AppSetDishSale)
			store.GET("/orders", h.AppListStoreOrders)
			store.GET("/orders/:id", h.AppGetStoreOrder)
			store.POST("/orders/:id/accept", h.AppAcceptOrder)
			store.POST("/orders/:id/reject", h.AppRejectOrder)
			store.POST("/orders/:id/ready", h.AppReadyOrder)
			store.POST("/orders/:id/refund/decision", h.AppDecideRefund)
			store.GET("/reviews", h.AppStoreReviews)
			store.POST("/reviews/:id/reply", h.AppReplyReview)
			store.GET("/finance/summary", h.AppStoreFinanceSummary)
			store.GET("/finance/records", h.AppStoreFinanceRecords)
			store.GET("/settlements", h.AppStoreSettlements)
			store.GET("/withdrawals", h.AppStoreWithdrawals)
			store.POST("/withdrawals", mw.RateLimit("app:withdraw", 10, time.Hour), h.AppStoreCreateWithdrawal)

			driver := appAPI.Group("/driver", mw.Auth(), mw.RequireAppRole(identity.RoleDriver))
			driver.GET("/me", h.AppDriverMe)
			driver.PUT("/status", h.AppSetDriverStatus)
			driver.GET("/order-pool", h.AppDriverPool)
			driver.GET("/orders", h.AppListDriverOrders)
			driver.GET("/orders/:id", h.AppGetDriverOrder)
			driver.POST("/orders/:id/grab", h.AppGrabOrder)
			driver.POST("/orders/:id/pickup", h.AppPickupOrder)
			driver.POST("/orders/:id/deliver", h.AppDeliverOrder)
			driver.POST("/location", mw.RateLimit("driver:loc", 120, time.Minute), h.AppReportLocation)
			driver.GET("/finance/summary", h.AppDriverFinanceSummary)
			driver.GET("/finance/records", h.AppDriverFinanceRecords)
			driver.GET("/settlements", h.AppDriverSettlements)
			driver.GET("/withdrawals", h.AppDriverWithdrawals)
			driver.POST("/withdrawals", mw.RateLimit("app:withdraw", 10, time.Hour), h.AppDriverCreateWithdrawal)

			common := appAPI.Group("/common", mw.Auth(),
				mw.RequireAppRole(identity.RoleUser, identity.RoleMerchant, identity.RoleDriver))
			common.POST("/uploads", mw.RateLimit("app:upload", 60, time.Minute), h.AppUpload)
			common.GET("/notifications", h.AppListNotifications)
			common.POST("/notifications/read-all", h.AppReadAllNotifications)
			common.POST("/notifications/:id/read", h.AppReadNotification)
		}
	}
	return r
}

// metricsHandler /metrics 访问控制。
// Prometheus 指标暴露业务量级与 Go 运行时信息（含内部路径、连接数），不应公网裸奔：
// 配置 METRICS_TOKEN 后必须带 X-Metrics-Token；未配置时生产环境直接关闭，非生产环境放行便于本机调试。
func (h *Handler) metricsHandler() gin.HandlerFunc {
	prom := gin.WrapH(promhttp.Handler())
	return func(c *gin.Context) {
		token := h.Cfg.MetricsToken
		if token == "" {
			if h.Cfg.AppEnv == "prod" {
				response.ErrStatus(c, http.StatusNotFound, 1401, "接口不存在")
				return
			}
			prom(c)
			return
		}
		if subtle.ConstantTimeCompare([]byte(c.GetHeader("X-Metrics-Token")), []byte(token)) != 1 {
			response.ErrStatus(c, http.StatusUnauthorized, 1001, "指标访问令牌无效")
			return
		}
		prom(c)
	}
}
