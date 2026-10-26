package admin

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/coupon"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// CouponService 管理端优惠券发放/查询。
type CouponService struct {
	Repos *application.RepoUnit
}

// IssueCouponInput 发放参数。
type IssueCouponInput struct {
	UserID     string
	Phone      string
	Title      string
	Amount     float64
	MinSpend   float64
	MerchantID string
	ExpiresAt  int64
}

// Issue 向指定用户发放优惠券。
func (s *CouponService) Issue(ctx context.Context, in IssueCouponInput) (*coupon.Coupon, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" || utf8.RuneCountInString(in.Title) > 64 {
		return nil, errcode.New(errcode.CodeBadRequest, "券标题需为 1-64 字")
	}
	if in.Amount <= 0 || in.MinSpend < 0 {
		return nil, errcode.New(errcode.CodeBadRequest, "券金额必须大于 0，门槛不能为负")
	}
	if in.ExpiresAt == 0 {
		in.ExpiresAt = time.Now().AddDate(0, 0, 30).Unix()
	}
	if in.ExpiresAt <= time.Now().Unix() {
		return nil, errcode.New(errcode.CodeBadRequest, "过期时间必须晚于当前时间")
	}
	userID := strings.TrimSpace(in.UserID)
	if userID == "" && in.Phone != "" {
		u, err := s.Repos.Users.FindByPhone(ctx, strings.TrimSpace(in.Phone))
		if err != nil {
			return nil, errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
		}
		userID = u.ID
	}
	if userID == "" {
		return nil, errcode.New(errcode.CodeBadRequest, "需指定 userId 或 phone")
	}
	if _, err := s.Repos.Users.FindByID(ctx, userID); err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
	}
	merchantID := strings.TrimSpace(in.MerchantID)
	if merchantID != "" {
		if _, err := s.Repos.Merchants.FindByID(ctx, merchantID); err != nil {
			return nil, errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
		}
	}
	c := &coupon.Coupon{
		ID: ulidgen.New(), UserID: userID, MerchantID: merchantID, Title: in.Title,
		Amount: in.Amount, MinSpend: in.MinSpend, Status: coupon.StatusUnused,
		ExpiresAt: in.ExpiresAt,
	}
	if err := s.Repos.Coupons.Save(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// List 管理端券列表。
func (s *CouponService) List(ctx context.Context, page, pageSize int, status, userID string) ([]*coupon.Coupon, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.Repos.Coupons.ListAll(ctx, page, pageSize, status, userID)
}
