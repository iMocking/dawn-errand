package application

import "errors"

// PaymentNotify 统一后的支付渠道回调报文。
//
// 微信/支付宝的原始报文由渠道适配层翻译成该结构，再交给 PaymentNotifyVerifier 验签。
// 验签刻意不放在 HTTP handler：字段拼接口径、签名算法、时间窗都属于「渠道协议」，
// 和 HTTP 绑定混在一起会让「换渠道」变成「改 handler」，也让协议本身无法单测。
type PaymentNotify struct {
	// Channel 渠道标识（来自回调路径 :channel，例如 wechat / alipay）。
	Channel string
	OrderID string
	TradeNo string
	Amount  float64
	Status  string
	// Timestamp 渠道侧签名时间（秒），用于时间窗校验。
	Timestamp int64
}

// PaymentNotifyVerifier 渠道回调验签端口（实现见 infrastructure/payment）。
type PaymentNotifyVerifier interface {
	// Verify 校验签名与时间窗。now 由调用方注入（秒），便于测试固定时间。
	Verify(n PaymentNotify, sign string, now int64) error
}

var (
	// ErrPaymentNotifyIncomplete 回调缺少签名或必填字段。
	ErrPaymentNotifyIncomplete = errors.New("payment notify incomplete")
	// ErrPaymentSignature 签名不匹配（包含「为 A 渠道签的名打到 B 渠道」的跨渠道重放）。
	ErrPaymentSignature = errors.New("payment signature mismatch")
	// ErrPaymentTimestamp 时间戳超出容忍窗口。
	ErrPaymentTimestamp = errors.New("payment timestamp out of window")
	// ErrPaymentNotConfigured 未配置回调密钥（生产环境启动校验会拦，这里是兜底）。
	ErrPaymentNotConfigured = errors.New("payment callback not configured")
)
