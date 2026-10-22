// Package response 统一 API 返回结构：{message, body, request_id, code}。
package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/requestid"
)

// Body 统一响应包。所有接口成功/错误响应 MUST 使用该结构。
type Body struct {
	Message   string `json:"message"`
	Body      any    `json:"body"`
	RequestID string `json:"request_id"`
	Code      int    `json:"code"`
}

// OK 成功响应（HTTP 200，code=0）。
func OK(c *gin.Context, body any) {
	c.AbortWithStatusJSON(http.StatusOK, Body{
		Message:   "ok",
		Body:      body,
		RequestID: requestid.From(c.Request.Context()),
		Code:      errcode.CodeOK,
	})
}

// OKMsg 成功响应并携带自定义 message。
func OKMsg(c *gin.Context, msg string, body any) {
	c.AbortWithStatusJSON(http.StatusOK, Body{
		Message:   msg,
		Body:      body,
		RequestID: requestid.From(c.Request.Context()),
		Code:      errcode.CodeOK,
	})
}

// Err 统一错误响应。根据错误类型映射 HTTP 状态；内部错误不泄露细节。
func Err(c *gin.Context, err error) {
	var e *errcode.Error
	if errors.As(err, &e) {
		c.AbortWithStatusJSON(e.Status, Body{
			Message:   e.Message,
			Body:      nil,
			RequestID: requestid.From(c.Request.Context()),
			Code:      e.Code,
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, Body{
		Message:   "服务器内部错误",
		Body:      nil,
		RequestID: requestid.From(c.Request.Context()),
		Code:      errcode.CodeInternal,
	})
}

// ErrStatus 直接构造错误响应。
func ErrStatus(c *gin.Context, status, code int, msg string) {
	c.AbortWithStatusJSON(status, Body{
		Message:   msg,
		Body:      nil,
		RequestID: requestid.From(c.Request.Context()),
		Code:      code,
	})
}
