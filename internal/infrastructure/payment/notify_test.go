package payment

import (
	"errors"
	"strings"
	"testing"

	"dawn-errand/internal/application"
)

// R2-5：验签是渠道资金入账的唯一闸门。签名串里必须带 channel，
// 时间窗/字段完整性/密钥缺失要各自可判读——排障时要能分清「渠道配错」与「有人重放」。

func sampleNotify() application.PaymentNotify {
	return application.PaymentNotify{
		Channel: "wechat", OrderID: "SG202610260001", TradeNo: "WX001",
		Amount: 56.7, Status: "paid", Timestamp: 1_700_000_000,
	}
}

func TestVerifyRoundTrip(t *testing.T) {
	v := NewHMACVerifier("secret")
	n := sampleNotify()
	sign := v.Sign(n)

	if err := v.Verify(n, sign, n.Timestamp); err != nil {
		t.Fatalf("合法签名被拒: %v", err)
	}
	// 渠道侧大小写/空白差异不应导致验签失败（hex 统一按小写比较）。
	if err := v.Verify(n, "  "+strings.ToUpper(sign)+"  ", n.Timestamp+1); err != nil {
		t.Errorf("大小写/空白容错失效: %v", err)
	}
	// 路径参数/JSON 带出的首尾空白由验签器收敛，不应变成「签名不符」。
	padded := n
	padded.Channel = " " + n.Channel + " "
	padded.OrderID = n.OrderID + "  "
	if err := v.Verify(padded, sign, n.Timestamp); err != nil {
		t.Errorf("字段空白应被收敛, got=%v", err)
	}
}

// 跨渠道重放：为 wechat 签的名打到 alipay 路径上，必须拒。
func TestVerifyRejectsChannelReplay(t *testing.T) {
	v := NewHMACVerifier("secret")
	n := sampleNotify()
	sign := v.Sign(n)

	n.Channel = "alipay"
	if err := v.Verify(n, sign, n.Timestamp); !errors.Is(err, application.ErrPaymentSignature) {
		t.Fatalf("跨渠道重放应判签名不符, got=%v", err)
	}
}

// 任一进签名串的字段被改都必须验不过（金额尤甚）。
func TestVerifyRejectsTamperedFields(t *testing.T) {
	v := NewHMACVerifier("secret")
	base := sampleNotify()
	sign := v.Sign(base)

	mutations := map[string]func(*application.PaymentNotify){
		"金额":  func(n *application.PaymentNotify) { n.Amount = 0.01 },
		"订单":  func(n *application.PaymentNotify) { n.OrderID = "SG202610260002" },
		"流水号": func(n *application.PaymentNotify) { n.TradeNo = "WX002" },
		"状态":  func(n *application.PaymentNotify) { n.Status = "refunded" },
		"时间戳": func(n *application.PaymentNotify) { n.Timestamp++ },
	}
	for name, mutate := range mutations {
		n := base
		mutate(&n)
		if err := v.Verify(n, sign, n.Timestamp); !errors.Is(err, application.ErrPaymentSignature) {
			t.Errorf("%s被改动后仍通过验签: %v", name, err)
		}
	}
}

func TestVerifyTimestampWindow(t *testing.T) {
	v := NewHMACVerifier("secret")
	n := sampleNotify()
	sign := v.Sign(n)

	for _, now := range []int64{n.Timestamp + 301, n.Timestamp - 301} {
		if err := v.Verify(n, sign, now); !errors.Is(err, application.ErrPaymentTimestamp) {
			t.Errorf("超出 ±5 分钟应判过期, now=%d got=%v", now, err)
		}
	}
	// 边界（正好 300 秒）放行：渠道时钟抖动不应过严。
	for _, now := range []int64{n.Timestamp + 300, n.Timestamp - 300} {
		if err := v.Verify(n, sign, now); err != nil {
			t.Errorf("边界应放行, now=%d got=%v", now, err)
		}
	}
}

func TestVerifyIncompleteAndUnconfigured(t *testing.T) {
	v := NewHMACVerifier("secret")
	n := sampleNotify()

	if err := v.Verify(n, "", n.Timestamp); !errors.Is(err, application.ErrPaymentNotifyIncomplete) {
		t.Errorf("缺签名应判参数不完整, got=%v", err)
	}
	blank := map[string]func(*application.PaymentNotify){
		"渠道":  func(x *application.PaymentNotify) { x.Channel = "  " },
		"订单":  func(x *application.PaymentNotify) { x.OrderID = "" },
		"状态":  func(x *application.PaymentNotify) { x.Status = "" },
		"时间戳": func(x *application.PaymentNotify) { x.Timestamp = 0 },
	}
	for name, mutate := range blank {
		x := n
		mutate(&x)
		if err := v.Verify(x, v.Sign(x), n.Timestamp); !errors.Is(err, application.ErrPaymentNotifyIncomplete) {
			t.Errorf("缺%s应判参数不完整, got=%v", name, err)
		}
	}

	// 未配置密钥时 fail closed，且要能和「签名不符」区分开（否则运维会去查渠道签名）。
	if err := NewHMACVerifier("").Verify(n, "whatever", n.Timestamp); !errors.Is(err, application.ErrPaymentNotConfigured) {
		t.Errorf("空密钥应判未配置, got=%v", err)
	}
}

// 渠道侧是照着文档拼串的，格式一变线上全验不过。这里用黄金串锁住字段顺序与金额两位小数。
func TestCanonicalFormatIsFrozen(t *testing.T) {
	n := application.PaymentNotify{
		Channel: "wechat", OrderID: "O1", TradeNo: "T1",
		Amount: 56.7, Status: "paid", Timestamp: 1700000000,
	}
	if got, want := canonical(n), "wechat|O1|T1|56.70|paid|1700000000"; got != want {
		t.Errorf("待签名串已变（渠道按旧格式签名会全失败）: got=%q want=%q", got, want)
	}
	n.Amount = 7
	if got, want := canonical(n), "wechat|O1|T1|7.00|paid|1700000000"; got != want {
		t.Errorf("整数金额也应固定两位小数: got=%q want=%q", got, want)
	}
}
