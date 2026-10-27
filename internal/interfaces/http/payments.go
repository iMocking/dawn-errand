package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/application"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/response"
)

// PaymentCallback POST /api/payments/callback/:channel（支付渠道回调）。
//
// 这里只做 HTTP 绑定与错误映射：验签已下沉到 infrastructure/payment 的适配层，
// 待签名串包含 channel，因此为 A 渠道签的名打不到 B 渠道上。
// 真实渠道适配层负责把微信/支付宝报文转换为该统一格式后再调用本接口。
func (h *Handler) PaymentCallback(c *gin.Context) {
	var in struct {
		OrderID   string  `json:"order_id"`
		TradeNo   string  `json:"trade_no"`
		Amount    float64 `json:"amount"`
		Status    string  `json:"status"`
		Timestamp int64   `json:"timestamp"`
		Sign      string  `json:"sign"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if h.Payments == nil {
		err := application.ErrPaymentNotConfigured
		_ = c.Error(err)
		response.ErrStatus(c, http.StatusInternalServerError, errcode.CodeInternal, "支付回调验签未配置")
		return
	}

	notify := application.PaymentNotify{
		Channel:   strings.TrimSpace(c.Param("channel")),
		OrderID:   strings.TrimSpace(in.OrderID),
		TradeNo:   strings.TrimSpace(in.TradeNo),
		Amount:    in.Amount,
		Status:    strings.TrimSpace(in.Status),
		Timestamp: in.Timestamp,
	}
	switch err := h.Payments.Verify(notify, in.Sign, time.Now().Unix()); {
	case err == nil:
	case errors.Is(err, application.ErrPaymentNotifyIncomplete):
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "回调参数不完整")
		return
	case errors.Is(err, application.ErrPaymentTimestamp):
		response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeConflict, "回调时间戳超出窗口")
		return
	case errors.Is(err, application.ErrPaymentNotConfigured):
		_ = c.Error(err)
		response.ErrStatus(c, http.StatusInternalServerError, errcode.CodeInternal, "支付回调验签未配置")
		return
	default:
		response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeForbidden, "回调签名校验失败")
		return
	}

	if notify.Status != "paid" {
		response.OKMsg(c, "已忽略非支付成功回调", gin.H{"status": notify.Status})
		return
	}
	o, err := h.OrderApp.PayByCallback(c.Request.Context(), notify.OrderID, notify.Channel, notify.TradeNo, notify.Amount)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "回调处理成功", gin.H{
		"order_id": o.ID, "status": o.PaymentStatus,
		"pay_amount": o.PayAmount,
	})
}
