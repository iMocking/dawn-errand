package app

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/coupon"
)

// CouponService 用户优惠券查询。
type CouponService struct {
	Repos *application.RepoUnit
}

// List 用户券列表。filter: available / used / expired / all。
func (s *CouponService) List(ctx context.Context, userID, filter string) ([]*coupon.Coupon, error) {
	now := time.Now().Unix()
	_ = s.Repos.Coupons.Expire(userID, now)
	rows, err := s.Repos.Coupons.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	out := make([]*coupon.Coupon, 0, len(rows))
	for _, c := range rows {
		switch filter {
		case "available":
			if c.Status != coupon.StatusUnused {
				continue
			}
		case "used":
			if c.Status != coupon.StatusUsed {
				continue
			}
		case "expired":
			if c.Status != coupon.StatusExpired {
				continue
			}
		}
		out = append(out, c)
	}
	return out, nil
}
