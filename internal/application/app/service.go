// Package app App 面应用服务：三端上下文、区域归属与个人资料。
package app

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/rider"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/domain/userprofile"
	"dawn-errand/internal/pkg/errcode"
)

// Service App 面用例。
type Service struct {
	Repos *application.RepoUnit
	Tx    application.TxRunner
}

// ClientMe 用户端上下文。
type ClientMe struct {
	User    *identity.User
	Profile *userprofile.Profile
	Region  *tenant.Region
}

// StoreMe 商家端上下文。
type StoreMe struct {
	Merchant *merchant.Merchant
	Today    *order.Stats
}

// DriverMe 骑手端上下文。
type DriverMe struct {
	Rider *rider.Rider
	Today *order.Stats
}

// ClientMe 读取当前用户、资料与区域。
func (s *Service) ClientMe(ctx context.Context, userID string) (*ClientMe, error) {
	u, err := s.Repos.Users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
		}
		return nil, err
	}
	out := &ClientMe{User: u}
	if p, err := s.Repos.Profiles.FindByUserID(ctx, userID); err == nil {
		out.Profile = p
	} else if !errors.Is(err, application.ErrNotFound) {
		return nil, err
	}
	if u.RegionID != "" {
		if r, err := s.Repos.Regions.FindByID(ctx, u.RegionID); err == nil {
			out.Region = r
		} else if !errors.Is(err, application.ErrNotFound) {
			return nil, err
		}
	}
	return out, nil
}

// UpdateClientProfile 修改昵称/头像（仅本人）。
func (s *Service) UpdateClientProfile(ctx context.Context, userID, name, avatar string) (*ClientMe, error) {
	name = strings.TrimSpace(name)
	avatar = strings.TrimSpace(avatar)
	if name == "" && avatar == "" {
		return nil, errcode.New(errcode.CodeBadRequest, "昵称与头像至少提供一项")
	}
	if name != "" {
		if n := len([]rune(name)); n < 1 || n > 32 {
			return nil, errcode.New(errcode.CodeBadRequest, "昵称长度需为 1-32 字")
		}
	}
	if len(avatar) > 512 {
		return nil, errcode.New(errcode.CodeBadRequest, "头像地址过长")
	}

	var out *ClientMe
	err := s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		u, err := tx.Users.FindByID(ctx, userID)
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				return errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
			}
			return err
		}
		now := time.Now().Unix()
		if name != "" {
			u.Name = name
		}
		if avatar != "" {
			u.Avatar = avatar
		}
		u.UpdatedAt = now
		if err := tx.Users.Save(ctx, u); err != nil {
			return err
		}
		p, err := tx.Profiles.FindByUserID(ctx, userID)
		if err != nil {
			if !errors.Is(err, application.ErrNotFound) {
				return err
			}
			p = &userprofile.Profile{UserID: userID, Level: "Lv.1 新用户", Verified: "未实名"}
		}
		if name != "" {
			p.Nickname = name
		}
		p.UpdatedAt = now
		if err := tx.Profiles.Save(ctx, p); err != nil {
			return err
		}
		out = &ClientMe{User: u, Profile: p}
		if u.RegionID != "" {
			if r, err := tx.Regions.FindByID(ctx, u.RegionID); err == nil {
				out.Region = r
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListRegions 可用服务区域（仅运营中）。
func (s *Service) ListRegions(ctx context.Context) ([]*tenant.Region, error) {
	rows, err := s.Repos.Regions.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*tenant.Region, 0, len(rows))
	for _, r := range rows {
		if r.Status == tenant.RegionStatusOpen {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// SetClientRegion 设置当前用户服务区域（区域必须运营中）。
func (s *Service) SetClientRegion(ctx context.Context, userID, regionID string) (*ClientMe, error) {
	regionID = strings.TrimSpace(regionID)
	if regionID == "" {
		return nil, errcode.New(errcode.CodeBadRequest, "region_id 必填")
	}
	region, err := s.Repos.Regions.FindByID(ctx, regionID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeRegionNotFound, "服务区域不存在")
		}
		return nil, err
	}
	if region.Status != tenant.RegionStatusOpen {
		return nil, errcode.New(errcode.CodeBadRequest, "该区域暂未开通服务")
	}

	var out *ClientMe
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		u, err := tx.Users.FindByID(ctx, userID)
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				return errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
			}
			return err
		}
		now := time.Now().Unix()
		u.RegionID = region.ID
		u.UpdatedAt = now
		if err := tx.Users.Save(ctx, u); err != nil {
			return err
		}
		p, err := tx.Profiles.FindByUserID(ctx, userID)
		if err != nil {
			if !errors.Is(err, application.ErrNotFound) {
				return err
			}
			p = &userprofile.Profile{UserID: userID, Level: "Lv.1 新用户", Verified: "未实名"}
		}
		p.Region = region.Name
		p.UpdatedAt = now
		if err := tx.Profiles.Save(ctx, p); err != nil {
			return err
		}
		out = &ClientMe{User: u, Profile: p, Region: region}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StoreMe 商家端上下文（店铺 + 今日经营数据）。
func (s *Service) StoreMe(ctx context.Context, merchantID string) (*StoreMe, error) {
	m, err := s.Repos.Merchants.FindByID(ctx, merchantID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
		}
		return nil, err
	}
	stats, err := s.Repos.Orders.Stats(ctx, tenant.Scope{Role: identity.RoleMerchant, MerchantID: merchantID}, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	return &StoreMe{Merchant: m, Today: stats}, nil
}

// DriverMe 骑手端上下文（档案 + 今日履约数据）。
func (s *Service) DriverMe(ctx context.Context, riderID string) (*DriverMe, error) {
	r, err := s.Repos.Riders.FindByID(ctx, riderID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeRiderNotFound, "骑手不存在")
		}
		return nil, err
	}
	stats, err := s.Repos.Orders.Stats(ctx, tenant.Scope{Role: identity.RoleDriver, RiderID: riderID}, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	return &DriverMe{Rider: r, Today: stats}, nil
}
