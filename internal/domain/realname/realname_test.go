package realname

import "testing"

// R5-5：身份证号格式是加密入库前的唯一挡板，非数字/带空白/长度离谱必须拦住。
// 注意：当前正则是 `^[0-9]{15,18}([0-9Xx])?$`，接受 15-19 位连续数字（末尾可为 X），
// 比真实身份证（恰好 15 或 18 位）更宽——16/17/19 位也会放行。本用例把这一现状冻结为
// 回归基线，收紧为「恰好 15/18 位 + 校验位算法」属功能变更，需前端与存量数据确认后再做。
func TestValidateIDCard(t *testing.T) {
	ok := []string{
		"11010119900307721X", "11010119900307721x", "110101199003077219",
		"110101900307721", // 15 位老式
		"123456789012345678",
		"1101011990030772",    // 16 位：当前正则接受（已知偏宽）
		"110101199003077211X", // 19 位：当前正则接受（已知偏宽）
	}
	bad := []string{
		"", " ", "11010119900307", "12345678901234", // 14 位及以下
		"1234567890123456789012", // 20 位以上
		"11010119900307721A", "110101199003077 19", " 1010119900307721",
		"11010119900307721X ", "身份证号1101011990030772", "11010119900307721-",
		"1101011990-307721",
	}
	for _, s := range ok {
		if !ValidateIDCard(s) {
			t.Errorf("合法证件号被拒: %q", s)
		}
	}
	for _, s := range bad {
		if ValidateIDCard(s) {
			t.Errorf("非法证件号被放行: %q", s)
		}
	}
}

// R5-5：View 是实名信息对外的唯一出口——密文字段绝不出现在视图里，
// 掩码由调用方（仓储/服务）传入，视图只负责搬运。
func TestToView(t *testing.T) {
	if got := ToView(nil, "110101********7219"); got != nil {
		t.Fatalf("nil 入参应返回 nil, got=%+v", got)
	}
	v := &Verification{
		ID: "RN1", UserID: "U1", RealName: "张三",
		IDCardCipher: "AES-GCM-BLOB", Status: StatusApproved,
		AppliedAt: 100, ReviewedAt: 200,
	}
	view := ToView(v, "110101********7219")
	if view.ID != v.ID || view.UserID != v.UserID || view.RealName != v.RealName {
		t.Fatalf("视图字段搬运错误: %+v", view)
	}
	if view.IDCardMasked != "110101********7219" {
		t.Errorf("掩码未按入参透传: %q", view.IDCardMasked)
	}
	if view.Status != StatusApproved || view.RejectReason != v.RejectReason {
		t.Errorf("审核状态/驳回原因错误: %+v", view)
	}
	if view.AppliedAt != 100 || view.ReviewedAt != 200 {
		t.Errorf("时间戳错误: %+v", view)
	}
}

// R5-5：状态常量与 UserProfile.verified 文案是两套系统对齐的锚点，改名会静默断链。
func TestStatusConstants(t *testing.T) {
	if StatusPending != "待审核" || StatusApproved != "已通过" || StatusRejected != "已驳回" {
		t.Errorf("审核状态常量被改动: %q %q %q", StatusPending, StatusApproved, StatusRejected)
	}
	if VerifiedDone != "已实名" {
		t.Errorf("VerifiedDone 被改动: %q", VerifiedDone)
	}
}
