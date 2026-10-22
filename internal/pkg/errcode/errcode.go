// Package errcode 定义统一业务错误码与错误类型。
package errcode

import (
	"errors"
	"fmt"
	"net/http"
)

// 统一返回结构：code=0 成功，非 0 业务错误。错误响应不得泄露堆栈/SQL/密钥。
const (
	CodeOK             = 0
	CodeBadRequest     = 1301 // 参数错误
	CodeInvalidID      = 1302 // 非法 ID
	CodeUnauthorized   = 1001 // 未认证 / 凭证无效
	CodeForbidden      = 1201 // 无权限
	CodeNotFound       = 1401 // 资源不存在
	CodeConflict       = 1402 // 状态冲突
	CodeTooManyRequest = 1007 // 触发限流
	CodeInternal       = 1500 // 内部错误

	// 认证域 100x
	CodeInvalidCredentials = 1002 // 账号或密码错误（防枚举：不区分账号不存在/密码错误）
	CodeEmailNotVerified   = 1003 // 邮箱未验证
	CodeEmailAlreadyExists = 1004 // 邮箱已注册
	CodeUserBanned         = 1005 // 账号被封禁
	CodeTokenInvalid       = 1006 // token 无效或过期
	CodeCaptchaInvalid     = 1008 // 验证码错误
	CodePasswordPolicy     = 1009 // 密码不符合策略
	CodeOldPasswordWrong   = 1010 // 原密码错误
	CodeEmailSendFailed    = 1011 // 邮件发送失败

	// 系统域 11xx
	CodeNotInitialized   = 1101 // 系统未初始化
	CodeAlreadyInit      = 1102 // 系统已初始化
	CodeBootstrapInvalid = 1103 // 初始化令牌无效
	CodeSettingNotFound  = 1104 // 配置不存在
	CodeSettingInvalid   = 1105 // 配置不合法
	CodeSMTPTestFailed   = 1106 // SMTP 测试连接失败
	CodeOSSTestFailed    = 1107 // OSS 测试连接失败
	CodeSMSTestFailed    = 1108 // 短信服务配置校验失败

	// 业务域 15xx
	CodeMerchantNotFound  = 1501
	CodeMerchantStatus    = 1502
	CodeRiderNotFound     = 1503
	CodeRiderStatus       = 1504
	CodeUserNotFound      = 1505
	CodeUserStatus        = 1506
	CodeOrderNotFound     = 1507
	CodeOrderStatus       = 1508
	CodeComplaintNotFound = 1509
	CodeWithdrawNotFound  = 1510
	CodeWithdrawStatus    = 1511
	CodeAnnouncementError = 1512
	CodeAreaConflict      = 1513
	CodeFeeRuleError      = 1514
	CodeRegionNotFound    = 1515
	CodeTenantScopeDenied = 1516 // 租户数据越权（不应出现的兜底）
)

// Error 业务错误。Code 与 HTTP 状态分离，HTTP 状态 MUST 正确。
type Error struct {
	Code    int
	Message string
	Status  int // HTTP 状态码
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

// New 构造业务错误（默认 400）。
func New(code int, msg string) *Error {
	return &Error{Code: code, Message: msg, Status: http.StatusBadRequest}
}

// NewStatus 构造带 HTTP 状态的业务错误。
func NewStatus(status, code int, msg string) *Error {
	return &Error{Code: code, Message: msg, Status: status}
}

// Wrap 包装底层错误并保留业务错误码。
func Wrap(err error, code int, msg string) *Error {
	return &Error{Code: code, Message: msg, Status: http.StatusBadRequest, Cause: err}
}

// As 提取 *Error。
func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
