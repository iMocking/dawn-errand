package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/response"
)

// PaymentCallback POST /api/payments/callback/:channel（支付渠道回调，HMAC 验签）。
// 真实渠道适配层需将微信/支付宝报文转换为该统一格式后再调用本接口。
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
	if in.OrderID == "" || in.Sign == "" || in.Timestamp == 0 {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "回调参数不完整")
		return
	}
	if delta := time.Now().Unix() - in.Timestamp; delta > 300 || delta < -300 {
		response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeConflict, "回调时间戳超出窗口")
		return
	}
	payload := fmt.Sprintf("%s|%s|%.2f|%s|%d",
		in.OrderID, in.TradeNo, in.Amount, in.Status, in.Timestamp)
	mac := hmac.New(sha256.New, []byte(h.Cfg.PaymentCallbackSecret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(in.Sign)), []byte(expected)) {
		response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeForbidden, "回调签名校验失败")
		return
	}
	if in.Status != "paid" {
		response.OKMsg(c, "已忽略非支付成功回调", gin.H{"status": in.Status})
		return
	}
	channel := strings.TrimSpace(c.Param("channel"))
	o, err := h.OrderApp.PayByCallback(c.Request.Context(), in.OrderID, channel, in.TradeNo, in.Amount)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "回调处理成功", gin.H{
		"order_id": o.ID, "status": o.PaymentStatus,
		"pay_amount": o.PayAmount,
	})
}
