// Package payment 支付渠道回调的验签适配。
//
// 把「渠道报文 + 签名」的校验从 HTTP handler 里拿出来：
// handler 只做 HTTP 绑定与错误映射，签名口径（哪些字段、什么顺序、金额格式）
// 和实现（HMAC-SHA256）都收在这里，换真实微信/支付宝时只改这个包。
package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"dawn-errand/internal/application"
)

// DefaultWindow 默认时间戳容忍窗口，与接入文档一致（±5 分钟）。
const DefaultWindow = 5 * time.Minute

// HMACVerifier 统一格式的 HMAC-SHA256 验签器。
type HMACVerifier struct {
	secret []byte
	window time.Duration
}

// NewHMACVerifier 用渠道回调密钥构造验签器。
func NewHMACVerifier(secret string) *HMACVerifier {
	return &HMACVerifier{secret: []byte(secret), window: DefaultWindow}
}

// WithWindow 覆盖时间窗（测试或特殊渠道）。
func (v *HMACVerifier) WithWindow(window time.Duration) *HMACVerifier {
	v.window = window
	return v
}

// canonical 待签名串：channel|order_id|trade_no|amount|status|timestamp。
//
// channel 必须在里面——否则为 A 渠道签的名可以原样打到 B 渠道，把支付单记成另一渠道。
// 金额固定两位小数：避免 1 与 1.00 生成两份合法签名，也避免 float 格式化漂移。
func canonical(n application.PaymentNotify) string {
	return fmt.Sprintf("%s|%s|%s|%.2f|%s|%d",
		n.Channel, n.OrderID, n.TradeNo, n.Amount, n.Status, n.Timestamp)
}

// normalize 去掉字符串字段首尾空白：路径参数与 JSON 都可能带出空白，
// 由验签器统一收敛，避免「同一个回调换个空格就验不过」。
func normalize(n application.PaymentNotify) application.PaymentNotify {
	n.Channel = strings.TrimSpace(n.Channel)
	n.OrderID = strings.TrimSpace(n.OrderID)
	n.TradeNo = strings.TrimSpace(n.TradeNo)
	n.Status = strings.TrimSpace(n.Status)
	return n
}

// Sign 生成签名。渠道模拟器与测试复用同一实现，避免两边各写一份拼串后悄悄分叉。
func (v *HMACVerifier) Sign(n application.PaymentNotify) string {
	mac := hmac.New(sha256.New, v.secret)
	mac.Write([]byte(canonical(normalize(n))))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify 校验签名与时间窗，实现 application.PaymentNotifyVerifier。
func (v *HMACVerifier) Verify(n application.PaymentNotify, sign string, now int64) error {
	if len(v.secret) == 0 {
		return application.ErrPaymentNotConfigured
	}
	n = normalize(n)
	sign = strings.ToLower(strings.TrimSpace(sign))
	if sign == "" || n.Channel == "" || n.OrderID == "" || n.Status == "" || n.Timestamp == 0 {
		return application.ErrPaymentNotifyIncomplete
	}
	// 先判时间窗再比签名：过期报文即使签名正确也不该被处理（防重放）。
	window := int64(v.window / time.Second)
	if delta := now - n.Timestamp; delta > window || delta < -window {
		return application.ErrPaymentTimestamp
	}
	if !hmac.Equal([]byte(sign), []byte(v.Sign(n))) {
		return application.ErrPaymentSignature
	}
	return nil
}
