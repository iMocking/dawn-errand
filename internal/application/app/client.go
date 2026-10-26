package app

import (
	"context"
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/address"
	"dawn-errand/internal/domain/catalog"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// ClientService 用户端用例：商家目录、地址簿。
type ClientService struct {
	Repos *application.RepoUnit
	Tx    application.TxRunner
}

// MerchantCard 商家列表卡片。
type MerchantCard struct {
	Merchant        *merchant.Merchant
	DistanceKM      float64
	DeliveryMinutes int
}

// MerchantDetail 商家详情（分类 + 在售菜品）。
type MerchantDetail struct {
	Merchant        *merchant.Merchant
	Categories      []*catalog.Category
	Dishes          []*catalog.Dish
	DistanceKM      float64
	DeliveryMinutes int
}

// AddressInput 地址创建/修改入参。
type AddressInput struct {
	ContactName string
	Phone       string
	RegionID    string
	Detail      string
	Tag         string
	Lat         float64
	Lng         float64
	IsDefault   bool
}

// ListMerchants 区域内商家列表（审核通过、未封禁）。
func (s *ClientService) ListMerchants(ctx context.Context, userID, keyword, category string, page, pageSize int, lat, lng float64) ([]*MerchantCard, int64, error) {
	u, err := s.Repos.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, 0, userErr(err)
	}
	if u.RegionID == "" {
		return nil, 0, errcode.NewStatus(400, errcode.CodeRegionRequired, "请先选择服务区域")
	}
	page, pageSize = normalizePageParams(page, pageSize)
	rows, total, err := s.Repos.Merchants.ListForClient(ctx, u.RegionID, keyword, category, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*MerchantCard, 0, len(rows))
	for _, m := range rows {
		card := &MerchantCard{Merchant: m}
		if validCoord(lat, lng) && validCoord(m.Lat, m.Lng) {
			card.DistanceKM = round2(haversine(lat, lng, m.Lat, m.Lng))
			card.DeliveryMinutes = estimateMinutes(card.DistanceKM)
		}
		out = append(out, card)
	}
	return out, total, nil
}

// GetMerchant 商家详情（仅本区域、审核通过）。
func (s *ClientService) GetMerchant(ctx context.Context, userID, merchantID string, lat, lng float64) (*MerchantDetail, error) {
	u, err := s.Repos.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, userErr(err)
	}
	if u.RegionID == "" {
		return nil, errcode.NewStatus(400, errcode.CodeRegionRequired, "请先选择服务区域")
	}
	m, err := s.Repos.Merchants.FindByID(ctx, merchantID)
	if err != nil {
		return nil, merchantErr(err)
	}
	if m.RegionID != u.RegionID || m.AuditStatus != merchant.AuditApproved || m.Status == merchant.MerchantStatusBanned {
		return nil, errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
	}
	cats, err := s.Repos.Catalog.ListCategories(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	dishes, _, err := s.Repos.Catalog.ListDishes(ctx, catalog.DishQuery{
		MerchantID: m.ID, IncludeOffSale: false, Page: 1, PageSize: 500,
	})
	if err != nil {
		return nil, err
	}
	out := &MerchantDetail{Merchant: m, Categories: cats, Dishes: dishes}
	if validCoord(lat, lng) && validCoord(m.Lat, m.Lng) {
		out.DistanceKM = round2(haversine(lat, lng, m.Lat, m.Lng))
		out.DeliveryMinutes = estimateMinutes(out.DistanceKM)
	}
	return out, nil
}

// ListAddresses 本人地址列表。
func (s *ClientService) ListAddresses(ctx context.Context, userID string) ([]*address.Address, error) {
	return s.Repos.Addresses.ListByUser(ctx, userID)
}

// CreateAddress 新增地址（首条自动默认；上限 20 条）。
func (s *ClientService) CreateAddress(ctx context.Context, userID string, in AddressInput) (*address.Address, error) {
	in, err := s.validateAddressInput(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	n, err := s.Repos.Addresses.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if n >= 20 {
		return nil, errcode.New(errcode.CodeBadRequest, "地址数量已达上限 20 条")
	}
	if n == 0 {
		in.IsDefault = true
	}
	a := &address.Address{
		ID: ulidgen.New(), UserID: userID, ContactName: in.ContactName, Phone: in.Phone,
		RegionID: in.RegionID, Detail: in.Detail, Tag: in.Tag,
		Lat: in.Lat, Lng: in.Lng, IsDefault: in.IsDefault,
	}
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if a.IsDefault {
			if err := tx.Addresses.ClearDefault(ctx, userID); err != nil {
				return err
			}
		}
		return tx.Addresses.Save(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// UpdateAddress 修改地址（仅本人）。
func (s *ClientService) UpdateAddress(ctx context.Context, userID, addressID string, in AddressInput) (*address.Address, error) {
	a, err := s.ownAddress(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}
	in, err = s.validateAddressInput(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	a.ContactName = in.ContactName
	a.Phone = in.Phone
	a.RegionID = in.RegionID
	a.Detail = in.Detail
	a.Tag = in.Tag
	a.Lat = in.Lat
	a.Lng = in.Lng
	// 默认标记不允许通过编辑取消（只能切换到另一条地址）。
	if a.IsDefault {
		in.IsDefault = true
	}
	a.IsDefault = in.IsDefault
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if a.IsDefault {
			if err := tx.Addresses.ClearDefault(ctx, userID); err != nil {
				return err
			}
		}
		return tx.Addresses.Save(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// DeleteAddress 删除地址；删除默认地址后自动选取最近一条作为默认。
func (s *ClientService) DeleteAddress(ctx context.Context, userID, addressID string) error {
	a, err := s.ownAddress(ctx, userID, addressID)
	if err != nil {
		return err
	}
	return s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if err := tx.Addresses.Delete(ctx, a.ID); err != nil {
			return err
		}
		if !a.IsDefault {
			return nil
		}
		list, err := tx.Addresses.ListByUser(ctx, userID)
		if err != nil || len(list) == 0 {
			return err
		}
		next := list[0]
		next.IsDefault = true
		return tx.Addresses.Save(ctx, next)
	})
}

// SetDefaultAddress 设为默认地址。
func (s *ClientService) SetDefaultAddress(ctx context.Context, userID, addressID string) (*address.Address, error) {
	a, err := s.ownAddress(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if err := tx.Addresses.ClearDefault(ctx, userID); err != nil {
			return err
		}
		a.IsDefault = true
		return tx.Addresses.Save(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *ClientService) ownAddress(ctx context.Context, userID, addressID string) (*address.Address, error) {
	a, err := s.Repos.Addresses.FindByID(ctx, addressID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeAddressNotFound, "地址不存在")
		}
		return nil, err
	}
	if a.UserID != userID {
		return nil, errcode.NewStatus(404, errcode.CodeAddressNotFound, "地址不存在")
	}
	return a, nil
}

func (s *ClientService) validateAddressInput(ctx context.Context, userID string, in AddressInput) (AddressInput, error) {
	in.ContactName = strings.TrimSpace(in.ContactName)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Detail = strings.TrimSpace(in.Detail)
	in.Tag = strings.TrimSpace(in.Tag)
	in.RegionID = strings.TrimSpace(in.RegionID)
	if n := utf8.RuneCountInString(in.ContactName); n < 1 || n > 32 {
		return in, errcode.New(errcode.CodeBadRequest, "联系人需为 1-32 字")
	}
	if !identity.ValidatePhone(in.Phone) {
		return in, errcode.New(errcode.CodeBadRequest, "手机号格式不正确")
	}
	if n := utf8.RuneCountInString(in.Detail); n < 2 || n > 120 {
		return in, errcode.New(errcode.CodeBadRequest, "详细地址需为 2-120 字")
	}
	if n := utf8.RuneCountInString(in.Tag); n > 8 {
		return in, errcode.New(errcode.CodeBadRequest, "地址标签不能超过 8 字")
	}
	if in.Lat != 0 || in.Lng != 0 {
		if !validCoord(in.Lat, in.Lng) {
			return in, errcode.New(errcode.CodeBadRequest, "经纬度不合法")
		}
	}
	if in.RegionID == "" {
		u, err := s.Repos.Users.FindByID(ctx, userID)
		if err != nil {
			return in, userErr(err)
		}
		in.RegionID = u.RegionID
	}
	if in.RegionID == "" {
		return in, errcode.NewStatus(400, errcode.CodeRegionRequired, "请先选择服务区域")
	}
	region, err := s.Repos.Regions.FindByID(ctx, in.RegionID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return in, errcode.NewStatus(404, errcode.CodeRegionNotFound, "服务区域不存在")
		}
		return in, err
	}
	if region.Status != tenant.RegionStatusOpen {
		return in, errcode.NewStatus(400, errcode.CodeRegionRequired, "该区域暂未开通服务")
	}
	return in, nil
}

func userErr(err error) error {
	if errors.Is(err, application.ErrNotFound) {
		return errcode.NewStatus(404, errcode.CodeUserNotFound, "用户不存在")
	}
	return err
}

func normalizePageParams(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}

func validCoord(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180 && !(lat == 0 && lng == 0)
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKM = 6371.0
	toRad := func(v float64) float64 { return v * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadiusKM * math.Asin(math.Sqrt(a))
}

func estimateMinutes(distanceKM float64) int {
	// 15 分钟备餐/取货 + 20km/h 骑行时间，向上取整。
	return int(math.Ceil(15 + distanceKM/20*60))
}
