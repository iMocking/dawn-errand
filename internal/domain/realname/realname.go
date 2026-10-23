// Package realname 用户实名认证限界上下文：认证申请实体、状态机与仓储端口。
package realname

import (
	"context"
	"errors"
	"regexp"
)

// 认证状态（与 UserProfile.verified 文案对齐：未实名 / 已实名）。
const (
	StatusPending  = "待审核"
	StatusApproved = "已通过"
	StatusRejected = "已驳回"
)

// VerifiedDone 通过后写入 UserProfile.verified 的值。
const VerifiedDone = "已实名"

// Verification 实名认证申请（一人仅保留一条当前申请，重提覆盖）。
type Verification struct {
	ID           string
	UserID       string
	RealName     string
	IDCardCipher string // 身份证号 AES-GCM 密文
	Status       string
	RejectReason string
	ReviewedBy   string
	AppliedAt    int64
	ReviewedAt   int64
	UpdatedAt    int64
}

// View 对外脱敏视图（身份证号仅掩码，永不外泄明文/密文）。
type View struct {
	ID           string
	UserID       string
	RealName     string
	IDCardMasked string
	Status       string
	RejectReason string
	AppliedAt    int64
	ReviewedAt   int64
}

// ToView 由仓储实体构造脱敏视图。
func ToView(v *Verification, idCardMasked string) *View {
	if v == nil {
		return nil
	}
	return &View{
		ID: v.ID, UserID: v.UserID, RealName: v.RealName,
		IDCardMasked: idCardMasked, Status: v.Status,
		RejectReason: v.RejectReason, AppliedAt: v.AppliedAt, ReviewedAt: v.ReviewedAt,
	}
}

// Query 列表查询。
type Query struct {
	Page     int
	PageSize int
	Status   string // 空 = 全部
}

// Repository 实名认证仓储端口。
type Repository interface {
	// Save 按 user_id 唯一 upsert（新申请覆盖旧申请）。
	Save(ctx context.Context, v *Verification) error
	// FindByID 按申请 ID 查询。
	FindByID(ctx context.Context, id string) (*Verification, error)
	// FindByUserID 按用户查询当前申请；无记录返回 application.ErrNotFound。
	FindByUserID(ctx context.Context, userID string) (*Verification, error)
	// List 分页查询（status 可为空），返回行与总数。
	List(ctx context.Context, q Query) ([]*Verification, int64, error)
	// UpdateReview 更新审核结果（状态/驳回原因/审核人/审核时间）。
	UpdateReview(ctx context.Context, id, status, reason, reviewer string, reviewedAt int64) error
}

var idCardRe = regexp.MustCompile(`^[0-9]{15,18}([0-9Xx])?$`)

// ValidateIDCard 基础身份证号格式校验（15/18 位数字，末位可为 X）。
func ValidateIDCard(s string) bool { return idCardRe.MatchString(s) }

// ErrBusy 状态冲突（重复提交/已处理）统一错误标记。
var ErrBusy = errors.New("verification busy")
