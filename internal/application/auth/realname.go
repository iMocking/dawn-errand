package auth

import (
	"context"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/realname"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// MyRealName 查询当前用户的实名认证状态；无申请返回 (nil, nil)。
// 返回脱敏视图（身份证号掩码，不泄露明文与密文）。
func (s *Service) MyRealName(ctx context.Context, userID string) (*realname.View, error) {
	v, err := s.Repos.RealNames.FindByUserID(ctx, userID)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return s.toView(v)
}

// toView 解密身份证号并构造脱敏视图；解密失败返回空掩码（不阻断查询）。
func (s *Service) toView(v *realname.Verification) (*realname.View, error) {
	masked := "******"
	if plain, err := s.Cipher.Decrypt(v.IDCardCipher); err == nil {
		masked = maskIDCard(plain)
	}
	return realname.ToView(v, masked), nil
}

func maskIDCard(s string) string {
	runes := []rune(s)
	if len(runes) <= 8 {
		return "******"
	}
	return string(runes[:6]) + "********" + string(runes[len(runes)-4:])
}

// SubmitRealName 提交实名认证申请（一人一条，重提覆盖旧申请）。
// 已实名（已通过）拒绝重复提交；待审核/已驳回可重新提交。
// 身份证号 AES-GCM 加密存储，对外仅脱敏展示。
func (s *Service) SubmitRealName(ctx context.Context, userID, realName, idCard string) (*realname.View, error) {
	realName = strings.TrimSpace(realName)
	idCard = strings.TrimSpace(strings.ToUpper(idCard))
	if realName == "" || len([]rune(realName)) > 32 {
		return nil, errcode.New(errcode.CodeBadRequest, "请填写真实姓名（2-32 字）")
	}
	if !realname.ValidateIDCard(idCard) {
		return nil, errcode.New(errcode.CodeBadRequest, "身份证号格式不正确（15 或 18 位，末位可为 X）")
	}
	cur, err := s.Repos.RealNames.FindByUserID(ctx, userID)
	if err != nil && err != application.ErrNotFound {
		return nil, err
	}
	if cur != nil && cur.Status == realname.StatusApproved {
		return nil, errcode.NewStatus(409, errcode.CodeConflict, "已实名认证，无需重复提交")
	}

	cipher, err := s.Cipher.Encrypt(idCard)
	if err != nil {
		return nil, errcode.Wrap(err, errcode.CodeInternal, "加密失败")
	}
	now := time.Now().Unix()
	v := &realname.Verification{
		ID: ulidgen.New(), UserID: userID, RealName: realName,
		IDCardCipher: cipher, Status: realname.StatusPending,
		AppliedAt: now, UpdatedAt: now,
	}
	if err := s.Repos.RealNames.Save(ctx, v); err != nil {
		return nil, err
	}
	return s.toView(v)
}
