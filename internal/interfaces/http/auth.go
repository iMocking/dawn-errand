package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/application/auth"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/response"
)

// AuthUserDto 认证用户（对齐前端 AuthUser）。
type AuthUserDto struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Role   string `json:"role"`
}

// LoginResultDto 登录结果（对齐前端 LoginResult；含 refresh_token 扩展字段）。
type LoginResultDto struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         AuthUserDto `json:"user"`
	Permissions  []string    `json:"permissions"`
}

type loginBody struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Captcha  string `json:"captcha"`
}

func (h *Handler) toLoginResult(lr *auth.LoginResult) LoginResultDto {
	return LoginResultDto{
		Token: lr.AccessToken, RefreshToken: lr.RefreshToken,
		User:        AuthUserDto{ID: lr.User.ID, Name: lr.User.Name, Avatar: lr.User.Avatar, Role: lr.User.Role},
		Permissions: lr.Permissions,
	}
}

// Login POST /auth/login。
func (h *Handler) Login(c *gin.Context) {
	var in loginBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.Account == "" || in.Password == "" {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "账号与密码必填")
		return
	}
	lr, err := h.Auth.Login(c.Request.Context(), in.Account, in.Password)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, h.toLoginResult(lr))
}

// SendSMSCode POST /auth/sms/send：发送手机验证码（dev 环境回显 dev_code）。
func (h *Handler) SendSMSCode(c *gin.Context) {
	var in struct {
		Phone string `json:"phone"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if strings.TrimSpace(in.Phone) == "" {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "手机号必填")
		return
	}
	res, err := h.Auth.SendSMSCode(c.Request.Context(), in.Phone)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "验证码已发送", res)
}

// LoginBySMS POST /auth/sms/login：手机号 + 验证码登录（未注册自动创建）。
func (h *Handler) LoginBySMS(c *gin.Context) {
	var in struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if strings.TrimSpace(in.Phone) == "" || strings.TrimSpace(in.Code) == "" {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "手机号与验证码必填")
		return
	}
	lr, err := h.Auth.LoginBySMS(c.Request.Context(), in.Phone, in.Code)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, h.toLoginResult(lr))
}

// Refresh POST /auth/refresh（公开入口：凭 refresh_token 换新令牌，无 access token）。
// all_devices=true 时同时吊销该用户其他设备的会话；缺省保持只轮换当前令牌的老契约。
func (h *Handler) Refresh(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
		AllDevices   bool   `json:"all_devices"`
	}
	if err := bind(c, &in); err != nil || in.RefreshToken == "" {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "refresh_token 必填")
		return
	}
	uid, err := h.Auth.RefreshUserID(c.Request.Context(), in.RefreshToken)
	if err != nil {
		response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeTokenInvalid, "刷新令牌无效或已过期")
		return
	}
	var lr *auth.LoginResult
	if in.AllDevices {
		lr, err = h.Auth.RefreshAll(c.Request.Context(), uid, in.RefreshToken)
	} else {
		lr, err = h.Auth.Refresh(c.Request.Context(), uid, in.RefreshToken)
	}
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, h.toLoginResult(lr))
}

// Logout POST /auth/logout。
// scope=current 仅登出当前设备；缺省或其他值一律按全端登出（最严格语义）。
func (h *Handler) Logout(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	_ = c.ShouldBindJSON(&in)
	u := middleware.User(c)
	allDevices := !strings.EqualFold(strings.TrimSpace(in.Scope), "current")
	if err := h.Auth.Logout(c.Request.Context(), u.ID, in.RefreshToken, allDevices); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已退出登录", gin.H{})
}

// Me GET /auth/me。
func (h *Handler) Me(c *gin.Context) {
	u := middleware.User(c)
	lr, err := h.Auth.Me(c.Request.Context(), u.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"user":        AuthUserDto{ID: lr.User.ID, Name: lr.User.Name, Avatar: lr.User.Avatar, Role: lr.User.Role},
		"permissions": lr.Permissions,
	})
}

// Register POST /auth/register。
func (h *Handler) Register(c *gin.Context) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Phone    string `json:"phone"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	res, err := h.Auth.Register(c.Request.Context(), auth.RegisterInput{
		Email: in.Email, Password: in.Password, Name: in.Name, Phone: in.Phone,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "注册成功，请查收验证邮件", gin.H{"id": res.User.ID})
}

// VerifyEmail GET /auth/email/verify。
func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "缺少 token")
		return
	}
	if err := h.Auth.VerifyEmail(c.Request.Context(), token); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "邮箱验证成功", gin.H{})
}

// ResendVerification POST /auth/email/resend。
func (h *Handler) ResendVerification(c *gin.Context) {
	var in struct {
		Email string `json:"email"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if err := h.Auth.ResendVerification(c.Request.Context(), in.Email); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "验证邮件已发送", gin.H{})
}

// ForgotPassword POST /auth/password/forgot。
func (h *Handler) ForgotPassword(c *gin.Context) {
	var in struct {
		Email string `json:"email"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if err := h.Auth.ForgotPassword(c.Request.Context(), in.Email); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "重置邮件已发送", gin.H{})
}

// ResetPassword POST /auth/password/reset。
func (h *Handler) ResetPassword(c *gin.Context) {
	var in struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if err := h.Auth.ResetPassword(c.Request.Context(), in.Token, in.NewPassword); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "密码已重置，请重新登录", gin.H{})
}

// ChangePassword POST /auth/password/change。
func (h *Handler) ChangePassword(c *gin.Context) {
	var in struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	if err := h.Auth.ChangePassword(c.Request.Context(), u.ID, in.OldPassword, in.NewPassword); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "密码已修改，请重新登录", gin.H{})
}

// MyRealName GET /auth/realname-verification：查询当前用户实名认证状态。
func (h *Handler) MyRealName(c *gin.Context) {
	u := middleware.User(c)
	v, err := h.Auth.MyRealName(c.Request.Context(), u.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toRealNameVerification(v))
}

type submitRealNameBody struct {
	RealName string `json:"realName"`
	IDCard   string `json:"idCard"`
}

// SubmitRealName POST /auth/realname-verification：提交实名认证申请。
func (h *Handler) SubmitRealName(c *gin.Context) {
	var in submitRealNameBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	v, err := h.Auth.SubmitRealName(c.Request.Context(), u.ID, in.RealName, in.IDCard)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "实名认证申请已提交，等待审核", toRealNameVerification(v))
}
