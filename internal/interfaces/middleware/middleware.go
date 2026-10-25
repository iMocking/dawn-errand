// Package middleware Gin 中间件：请求 ID、访问日志、恢复、指标、限流、认证、权限。
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/metrics"
	"dawn-errand/internal/pkg/requestid"
	"dawn-errand/internal/pkg/response"
)

// Middleware 中间件依赖集合。
type Middleware struct {
	Logger   *zap.Logger
	Tokens   application.TokenManager
	Users    identity.UserRepository
	Creds    identity.CredentialRepository
	Limiter  application.RateLimiter
	Enforcer application.PolicyManager
}

// gin.Context 键（Get/Set 需要 string 键）。
const (
	ctxPrincipalKey = "ctx_principal"
	ctxUserKey      = "ctx_user"
)

// Principal 从上下文取数据隔离主体。
func Principal(c *gin.Context) *tenant.Principal {
	v, _ := c.Get(ctxPrincipalKey)
	return v.(*tenant.Principal)
}

// User 从上下文取当前用户。
func User(c *gin.Context) *identity.User {
	v, _ := c.Get(ctxUserKey)
	return v.(*identity.User)
}

// RequestID 生成/透传请求 ID（X-Request-ID），并注入审计元数据。
func (m *Middleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = requestid.New()
		}
		ctx := requestid.WithID(c.Request.Context(), id)
		ctx = application.WithRequestMeta(ctx, id, c.ClientIP())
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// AccessLog 访问日志（method/path/status/latency/request_id/ip）。
func (m *Middleware) AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		m.Logger.Info("http",
			zap.String("request_id", requestid.From(c.Request.Context())),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
		)
	}
}

// Recovery 捕获 panic 并返回统一 500。
func (m *Middleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				m.Logger.Error("panic", zap.Any("panic", r), zap.String("request_id", requestid.From(c.Request.Context())))
				response.ErrStatus(c, http.StatusInternalServerError, errcode.CodeInternal, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}

// Metrics Prometheus 指标。
func (m *Middleware) Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		c.Next()
		metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, itoa(c.Writer.Status())).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}

// SecurityHeaders 基础安全响应头。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

// RateLimit 按 IP+前缀限流。
func (m *Middleware) RateLimit(prefix string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rl:" + prefix + ":" + c.ClientIP()
		ok, err := m.Limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil || !ok {
			response.ErrStatus(c, http.StatusTooManyRequests, errcode.CodeTooManyRequest, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

// Auth JWT 认证：校验令牌、会话版本、用户状态，注入 Principal。
func (m *Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := ""
		if strings.HasPrefix(header, "Bearer ") {
			token = strings.TrimPrefix(header, "Bearer ")
		}
		if token == "" {
			metrics.AuthFailuresTotal.Inc()
			response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeUnauthorized, "未登录或凭证缺失")
			c.Abort()
			return
		}
		claims, err := m.Tokens.ParseAccessToken(token)
		if err != nil {
			metrics.AuthFailuresTotal.Inc()
			response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeTokenInvalid, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		u, err := m.Users.FindByID(claims.UserID)
		if err != nil {
			metrics.AuthFailuresTotal.Inc()
			response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeTokenInvalid, "用户不存在")
			c.Abort()
			return
		}
		if !u.IsActive() {
			metrics.AuthFailuresTotal.Inc()
			response.ErrStatus(c, http.StatusForbidden, errcode.CodeUserBanned, "账号已被封禁")
			c.Abort()
			return
		}
		cred, err := m.Creds.FindByUserID(u.ID)
		if err != nil || cred.SessionVersion != claims.SessionVersion {
			metrics.AuthFailuresTotal.Inc()
			response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeTokenInvalid, "登录状态已失效，请重新登录")
			c.Abort()
			return
		}
		principal := &tenant.Principal{
			UserID: u.ID, Role: u.Role,
			RegionID: u.RegionID, MerchantID: u.MerchantID, RiderID: u.RiderID,
		}
		c.Set(ctxPrincipalKey, principal)
		c.Set(ctxUserKey, u)
		c.Next()
	}
}

// RequirePerm 权限校验：obj:act 必须被当前角色策略放行。
func (m *Middleware) RequirePerm(obj, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := Principal(c)
		ok, err := m.Enforcer.Enforce(p.Role, obj+":"+act, act)
		if err != nil || !ok {
			response.ErrStatus(c, http.StatusForbidden, errcode.CodeForbidden, "无权限执行该操作")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAppRole App 面角色守卫：校验角色并强制主体绑定。
// merchant 必须绑定 MerchantID、driver 必须绑定 RiderID，否则返回 1202；
// user 允许尚未选择区域（区域接口自身放行，业务接口由用例校验）。
func (m *Middleware) RequireAppRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		p := Principal(c)
		if !allowed[p.Role] {
			response.ErrStatus(c, http.StatusForbidden, errcode.CodeRoleDenied, "当前账号无权访问该端")
			c.Abort()
			return
		}
		switch p.Role {
		case identity.RoleMerchant:
			if p.MerchantID == "" {
				response.ErrStatus(c, http.StatusForbidden, errcode.CodeAccountUnbound, "账号未绑定商家，请联系平台开通")
				c.Abort()
				return
			}
		case identity.RoleDriver:
			if p.RiderID == "" {
				response.ErrStatus(c, http.StatusForbidden, errcode.CodeAccountUnbound, "账号未绑定骑手，请联系平台开通")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
