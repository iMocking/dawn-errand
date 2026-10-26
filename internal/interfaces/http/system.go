package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/domain/system"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/response"
)

// BootstrapStatus GET /system/bootstrap/status（公开）。
// 仅暴露「是否已初始化」，供安装向导判断入口；init_at / admin_user_id 属于内部信息，
// 系统初始化完成后该端点不再有意义，不对外返回细节。
func (h *Handler) BootstrapStatus(c *gin.Context) {
	st, err := h.System.BootstrapStatus(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{"initialized": st.Initialized})
}

// Bootstrap POST /system/bootstrap。
func (h *Handler) Bootstrap(c *gin.Context) {
	var in struct {
		AdminEmail string `json:"admin_email"`
		InitToken  string `json:"init_token"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	res, err := h.System.Bootstrap(c.Request.Context(), in.AdminEmail, in.InitToken)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"admin_user_id":     res.AdminUserID,
		"admin_email":       res.AdminEmail,
		"admin_name":        res.AdminName,
		"one_time_password": res.OneTimePassword,
	})
}

// GetSMTP GET /system/config/smtp。
func (h *Handler) GetSMTP(c *gin.Context) {
	cfg, err := h.System.GetSMTP(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"enabled":    cfg.Host != "",
		"host":       cfg.Host,
		"port":       cfg.Port,
		"from_email": cfg.FromEmail,
		"encryption": cfg.Encryption,
		"password":   cfg.Password,
	})
}

type smtpBody struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	FromEmail  string `json:"from_email"`
	Password   string `json:"password"`
	Encryption string `json:"encryption"`
	Timeout    int    `json:"timeout"`
}

// UpdateSMTP PUT /system/config/smtp。
func (h *Handler) UpdateSMTP(c *gin.Context) {
	var in smtpBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	err := h.System.UpdateSMTP(c.Request.Context(), system.SMTPConfig{
		Host: in.Host, Port: in.Port, FromEmail: in.FromEmail,
		Password: in.Password, Encryption: in.Encryption, TimeoutSec: in.Timeout,
	}, u.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "SMTP 配置已保存", gin.H{})
}

// TestSMTP POST /system/config/smtp/test。
func (h *Handler) TestSMTP(c *gin.Context) {
	var in smtpBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if err := h.System.TestSMTP(c.Request.Context(), system.SMTPConfig{
		Host: in.Host, Port: in.Port, FromEmail: in.FromEmail,
		Password: in.Password, Encryption: in.Encryption, TimeoutSec: in.Timeout,
	}); err != nil {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeSMTPTestFailed, "SMTP 连接失败："+err.Error())
		return
	}
	response.OKMsg(c, "SMTP 连接成功", gin.H{})
}

// GetSMS GET /system/config/sms。
func (h *Handler) GetSMS(c *gin.Context) {
	cfg, err := h.System.GetSMS(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"enabled":       cfg.AccessKey != "" && cfg.SignName != "",
		"provider":      cfg.Provider,
		"access_key":    cfg.AccessKey,
		"secret_key":    cfg.SecretKey,
		"sign_name":     cfg.SignName,
		"template_code": cfg.TemplateCode,
	})
}

type smsBody struct {
	Provider     string `json:"provider"`
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	SignName     string `json:"sign_name"`
	TemplateCode string `json:"template_code"`
}

// UpdateSMS PUT /system/config/sms。
func (h *Handler) UpdateSMS(c *gin.Context) {
	var in smsBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	err := h.System.UpdateSMS(c.Request.Context(), system.SMSConfig{
		Provider: in.Provider, AccessKey: in.AccessKey, SecretKey: in.SecretKey,
		SignName: in.SignName, TemplateCode: in.TemplateCode,
	}, u.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "短信服务配置已保存", gin.H{})
}

// TestSMS POST /system/config/sms/test。
func (h *Handler) TestSMS(c *gin.Context) {
	var in smsBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if err := h.System.TestSMS(c.Request.Context(), system.SMSConfig{
		Provider: in.Provider, AccessKey: in.AccessKey, SecretKey: in.SecretKey,
		SignName: in.SignName, TemplateCode: in.TemplateCode,
	}); err != nil {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeSMSTestFailed, "短信服务配置校验失败："+err.Error())
		return
	}
	response.OKMsg(c, "短信服务配置校验通过", gin.H{})
}

// GetWithdrawRule GET /system/config/withdraw-rule。
func (h *Handler) GetWithdrawRule(c *gin.Context) {
	cfg, err := h.System.GetWithdrawRule(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"enabled":     cfg.Enabled,
		"fee_rate":    cfg.FeeRate,
		"min_amount":  cfg.MinAmount,
		"max_amount":  cfg.MaxAmount,
		"daily_limit": cfg.DailyLimit,
		"mode":        cfg.Mode,
		"settle_days": cfg.SettleDays,
	})
}

type withdrawRuleBody struct {
	Enabled    bool    `json:"enabled"`
	FeeRate    float64 `json:"fee_rate"`
	MinAmount  float64 `json:"min_amount"`
	MaxAmount  float64 `json:"max_amount"`
	DailyLimit int     `json:"daily_limit"`
	Mode       string  `json:"mode"`
	SettleDays int     `json:"settle_days"`
}

// UpdateWithdrawRule PUT /system/config/withdraw-rule。
func (h *Handler) UpdateWithdrawRule(c *gin.Context) {
	var in withdrawRuleBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	err := h.System.UpdateWithdrawRule(c.Request.Context(), system.WithdrawRuleConfig{
		Enabled: in.Enabled, FeeRate: in.FeeRate, MinAmount: in.MinAmount,
		MaxAmount: in.MaxAmount, DailyLimit: in.DailyLimit, Mode: in.Mode, SettleDays: in.SettleDays,
	}, u.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "提现规则已保存", gin.H{})
}

// GetOSS GET /system/config/oss。
func (h *Handler) GetOSS(c *gin.Context) {
	cfg, err := h.System.GetOSS(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"enabled":    cfg.Endpoint != "" && cfg.Bucket != "",
		"provider":   cfg.Provider,
		"endpoint":   cfg.Endpoint,
		"bucket":     cfg.Bucket,
		"access_key": cfg.AccessKey,
		"secret_key": cfg.SecretKey,
		"region":     cfg.Region,
	})
}

// UpdateOSS PUT /system/config/oss。
func (h *Handler) UpdateOSS(c *gin.Context) {
	var in struct {
		Provider  string `json:"provider"`
		Endpoint  string `json:"endpoint"`
		Bucket    string `json:"bucket"`
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
		Region    string `json:"region"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	err := h.System.UpdateOSS(c.Request.Context(), system.OSSConfig{
		Provider: in.Provider, Endpoint: in.Endpoint, Bucket: in.Bucket,
		AccessKey: in.AccessKey, SecretKey: in.SecretKey, Region: in.Region,
	}, u.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "OSS 配置已保存", gin.H{})
}

// GetLocalStorage GET /system/config/storage/local。
func (h *Handler) GetLocalStorage(c *gin.Context) {
	cfg, err := h.System.GetLocalStorage(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"enabled":         cfg.RootDir != "",
		"root_dir":        cfg.RootDir,
		"public_base_url": cfg.PublicBaseURL,
		"max_size_mb":     cfg.MaxSizeMB,
		"allowed_mime":    cfg.AllowedMIME,
		"path_prefix":     cfg.PathPrefix,
	})
}

// UpdateLocalStorage PUT /system/config/storage/local。
func (h *Handler) UpdateLocalStorage(c *gin.Context) {
	var in struct {
		RootDir       string   `json:"root_dir"`
		PublicBaseURL string   `json:"public_base_url"`
		MaxSizeMB     int64    `json:"max_size_mb"`
		AllowedMIME   []string `json:"allowed_mime"`
		PathPrefix    string   `json:"path_prefix"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	err := h.System.UpdateLocalStorage(c.Request.Context(), system.LocalStorageConfig{
		RootDir: in.RootDir, PublicBaseURL: in.PublicBaseURL, MaxSizeMB: in.MaxSizeMB,
		AllowedMIME: in.AllowedMIME, PathPrefix: in.PathPrefix,
	}, u.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "本地存储配置已保存", gin.H{})
}

// TestLocalStorage POST /system/config/storage/local/test。
func (h *Handler) TestLocalStorage(c *gin.Context) {
	var in struct {
		RootDir    string `json:"root_dir"`
		PathPrefix string `json:"path_prefix"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if err := h.System.TestLocalStorage(c.Request.Context(), system.LocalStorageConfig{
		RootDir: in.RootDir, PathPrefix: in.PathPrefix,
	}); err != nil {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeInternal, "本地存储不可写："+err.Error())
		return
	}
	response.OKMsg(c, "本地存储可用", gin.H{})
}

// GetCasbinConfig GET /system/config/casbin。
func (h *Handler) GetCasbinConfig(c *gin.Context) {
	cfg, err := h.System.GetCasbinConfig(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{"model": cfg.Model, "policy": cfg.Policy, "policy_store": cfg.PolicyStore})
}

// UpdateCasbinConfig PUT /system/config/casbin。
func (h *Handler) UpdateCasbinConfig(c *gin.Context) {
	var in struct {
		Policy [][]string `json:"policy"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	if err := h.System.UpdateCasbinConfig(c.Request.Context(), in.Policy, u.ID); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "权限策略已更新", gin.H{})
}

// DevOutbox GET /system/dev/outbox（仅 dev 环境）。
func (h *Handler) DevOutbox(c *gin.Context) {
	if !h.Cfg.IsDev() {
		response.ErrStatus(c, http.StatusForbidden, errcode.CodeForbidden, "仅开发环境可用")
		return
	}
	emails, err := h.System.DevOutbox(c.Request.Context(), 50)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(emails))
	for _, e := range emails {
		list = append(list, gin.H{
			"id": e.ID, "to_email": e.ToEmail, "subject": e.Subject,
			"body": e.Body, "created_at": e.CreatedAt,
		})
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}
