package app

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/catalog"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// StoreService 商家端用例：营业设置、状态、菜单管理。
type StoreService struct {
	Repos *application.RepoUnit
	Tx    application.TxRunner
}

// BusinessSettings 门店经营设置（全量保存）。
type BusinessSettings struct {
	OpenHours     string
	DeliveryRange string
	AutoAccept    bool
	MinPrice      float64
	DeliveryFee   float64
	PackingFee    float64
	Announcement  string
	ServicePhone  string
}

// Menu 商家菜单（含下架/售罄）。
type Menu struct {
	Categories []*MenuCategory
}

// MenuCategory 分类 + 菜品。
type MenuCategory struct {
	Category *catalog.Category
	Dishes   []*catalog.Dish
}

// CategoryInput 分类创建/修改入参。
type CategoryInput struct {
	Name string
	Sort int
}

// DishInput 菜品创建/修改入参。OnSale 为 nil 时：创建默认上架、修改保持原状态。
type DishInput struct {
	CategoryID  string
	Name        string
	Description string
	ImageURL    string
	Price       float64
	Stock       int
	OnSale      *bool
	Sort        int
}

// UpdateStatus 营业/休息切换（封禁商家不可自行操作）。
func (s *StoreService) UpdateStatus(ctx context.Context, merchantID string, open bool, operator string) (*merchant.Merchant, error) {
	m, err := s.Repos.Merchants.FindByID(ctx, merchantID)
	if err != nil {
		return nil, merchantErr(err)
	}
	if m.Status == merchant.MerchantStatusBanned {
		return nil, errcode.New(errcode.CodeMerchantStatus, "商家已被封禁，请联系平台")
	}
	status := merchant.MerchantStatusRest
	if open {
		status = merchant.MerchantStatusOpen
	}
	now := time.Now().Unix()
	if err := s.Repos.Merchants.UpdateStatus(ctx, merchantID, status, now); err != nil {
		return nil, err
	}
	m.Status = status
	m.LastUpdatedAt = now
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "store.status", Object: merchantID,
		Result: "success", RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
		Operator: operator,
	})
	return m, nil
}

// GetBusiness 读取经营设置。
func (s *StoreService) GetBusiness(ctx context.Context, merchantID string) (*merchant.Merchant, error) {
	m, err := s.Repos.Merchants.FindByID(ctx, merchantID)
	if err != nil {
		return nil, merchantErr(err)
	}
	return m, nil
}

// UpdateBusiness 保存经营设置（仅本店）。
func (s *StoreService) UpdateBusiness(ctx context.Context, merchantID string, in BusinessSettings, operator string) (*merchant.Merchant, error) {
	in.OpenHours = strings.TrimSpace(in.OpenHours)
	in.DeliveryRange = strings.TrimSpace(in.DeliveryRange)
	in.Announcement = strings.TrimSpace(in.Announcement)
	in.ServicePhone = strings.TrimSpace(in.ServicePhone)
	if n := utf8.RuneCountInString(in.OpenHours); n < 1 || n > 64 {
		return nil, errcode.New(errcode.CodeBadRequest, "营业时间需为 1-64 字")
	}
	if n := utf8.RuneCountInString(in.DeliveryRange); n < 1 || n > 64 {
		return nil, errcode.New(errcode.CodeBadRequest, "配送范围需为 1-64 字")
	}
	if n := utf8.RuneCountInString(in.Announcement); n > 512 {
		return nil, errcode.New(errcode.CodeBadRequest, "公告不能超过 512 字")
	}
	if len(in.ServicePhone) > 32 {
		return nil, errcode.New(errcode.CodeBadRequest, "客服电话过长")
	}
	for name, v := range map[string]float64{
		"起送价": in.MinPrice, "配送费": in.DeliveryFee, "打包费": in.PackingFee,
	} {
		if v < 0 || v > 9999 {
			return nil, errcode.New(errcode.CodeBadRequest, name+"需在 0-9999 元之间")
		}
	}
	in.MinPrice = round2(in.MinPrice)
	in.DeliveryFee = round2(in.DeliveryFee)
	in.PackingFee = round2(in.PackingFee)

	m, err := s.Repos.Merchants.FindByID(ctx, merchantID)
	if err != nil {
		return nil, merchantErr(err)
	}
	m.OpenHours = in.OpenHours
	m.DeliveryRange = in.DeliveryRange
	m.AutoAccept = in.AutoAccept
	m.MinPrice = in.MinPrice
	m.DeliveryFee = in.DeliveryFee
	m.PackingFee = in.PackingFee
	m.Announcement = in.Announcement
	m.ServicePhone = in.ServicePhone
	m.LastUpdatedAt = time.Now().Unix()
	if err := s.Repos.Merchants.Save(ctx, m); err != nil {
		return nil, err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "store.business.update", Object: merchantID,
		Result: "success", RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
		Operator: operator,
	})
	return m, nil
}

// Menu 读取本店菜单（含下架与售罄）。
func (s *StoreService) Menu(ctx context.Context, merchantID string) (*Menu, error) {
	if _, err := s.Repos.Merchants.FindByID(ctx, merchantID); err != nil {
		return nil, merchantErr(err)
	}
	cats, err := s.Repos.Catalog.ListCategories(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	dishes, _, err := s.Repos.Catalog.ListDishes(ctx, catalog.DishQuery{
		MerchantID: merchantID, IncludeOffSale: true, Page: 1, PageSize: 500,
	})
	if err != nil {
		return nil, err
	}
	byCat := map[string][]*catalog.Dish{}
	for _, d := range dishes {
		byCat[d.CategoryID] = append(byCat[d.CategoryID], d)
	}
	out := &Menu{Categories: make([]*MenuCategory, 0, len(cats))}
	for _, c := range cats {
		out.Categories = append(out.Categories, &MenuCategory{Category: c, Dishes: byCat[c.ID]})
	}
	return out, nil
}

// CreateCategory 新增分类。
func (s *StoreService) CreateCategory(ctx context.Context, merchantID string, in CategoryInput, operator string) (*catalog.Category, error) {
	name, err := validateCategoryName(in.Name)
	if err != nil {
		return nil, err
	}
	if _, err := s.Repos.Merchants.FindByID(ctx, merchantID); err != nil {
		return nil, merchantErr(err)
	}
	cats, err := s.Repos.Catalog.ListCategories(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	if len(cats) >= 50 {
		return nil, errcode.New(errcode.CodeBadRequest, "分类数量已达上限 50")
	}
	c := &catalog.Category{ID: ulidgen.New(), MerchantID: merchantID, Name: name, Sort: in.Sort}
	if err := s.Repos.Catalog.SaveCategory(ctx, c); err != nil {
		return nil, err
	}
	s.audit(ctx, operator, "store.category.create", c.ID)
	return c, nil
}

// UpdateCategory 修改分类。
func (s *StoreService) UpdateCategory(ctx context.Context, merchantID, categoryID string, in CategoryInput, operator string) (*catalog.Category, error) {
	name, err := validateCategoryName(in.Name)
	if err != nil {
		return nil, err
	}
	c, err := s.ownCategory(ctx, merchantID, categoryID)
	if err != nil {
		return nil, err
	}
	c.Name = name
	c.Sort = in.Sort
	if err := s.Repos.Catalog.SaveCategory(ctx, c); err != nil {
		return nil, err
	}
	s.audit(ctx, operator, "store.category.update", c.ID)
	return c, nil
}

// DeleteCategory 删除分类（分类下仍有菜品时拒绝）。
func (s *StoreService) DeleteCategory(ctx context.Context, merchantID, categoryID, operator string) error {
	c, err := s.ownCategory(ctx, merchantID, categoryID)
	if err != nil {
		return err
	}
	n, err := s.Repos.Catalog.CountDishesByCategory(ctx, c.ID)
	if err != nil {
		return err
	}
	if n > 0 {
		return errcode.New(errcode.CodeCategoryInUse, "分类下仍有菜品，请先移动或删除菜品")
	}
	if err := s.Repos.Catalog.DeleteCategory(ctx, c.ID); err != nil {
		return err
	}
	s.audit(ctx, operator, "store.category.delete", c.ID)
	return nil
}

// CreateDish 新增菜品。
func (s *StoreService) CreateDish(ctx context.Context, merchantID string, in DishInput, operator string) (*catalog.Dish, error) {
	in, err := validateDishInput(in)
	if err != nil {
		return nil, err
	}
	if _, err := s.ownCategory(ctx, merchantID, in.CategoryID); err != nil {
		return nil, err
	}
	d := &catalog.Dish{
		ID: ulidgen.New(), MerchantID: merchantID, CategoryID: in.CategoryID,
		Name: in.Name, Description: in.Description, ImageURL: in.ImageURL,
		Price: in.Price, Stock: in.Stock, OnSale: true, Sort: in.Sort,
	}
	if in.OnSale != nil {
		d.OnSale = *in.OnSale
	}
	if err := s.Repos.Catalog.SaveDish(ctx, d); err != nil {
		return nil, err
	}
	s.audit(ctx, operator, "store.dish.create", d.ID)
	return d, nil
}

// UpdateDish 修改菜品（保留历史销量）。
func (s *StoreService) UpdateDish(ctx context.Context, merchantID, dishID string, in DishInput, operator string) (*catalog.Dish, error) {
	in, err := validateDishInput(in)
	if err != nil {
		return nil, err
	}
	d, err := s.ownDish(ctx, merchantID, dishID)
	if err != nil {
		return nil, err
	}
	if _, err := s.ownCategory(ctx, merchantID, in.CategoryID); err != nil {
		return nil, err
	}
	d.CategoryID = in.CategoryID
	d.Name = in.Name
	d.Description = in.Description
	d.ImageURL = in.ImageURL
	d.Price = in.Price
	d.Stock = in.Stock
	if in.OnSale != nil {
		d.OnSale = *in.OnSale
	}
	d.Sort = in.Sort
	if err := s.Repos.Catalog.SaveDish(ctx, d); err != nil {
		return nil, err
	}
	s.audit(ctx, operator, "store.dish.update", d.ID)
	return d, nil
}

// SetDishSale 上架/下架/改库存。stock < 0 表示不改库存。
func (s *StoreService) SetDishSale(ctx context.Context, merchantID, dishID string, onSale bool, stock int, operator string) (*catalog.Dish, error) {
	d, err := s.ownDish(ctx, merchantID, dishID)
	if err != nil {
		return nil, err
	}
	if stock >= 0 && stock > 99999 {
		return nil, errcode.New(errcode.CodeBadRequest, "库存超出范围")
	}
	if err := s.Repos.Catalog.UpdateDishSale(ctx, d.ID, onSale, stock, time.Now().Unix()); err != nil {
		return nil, err
	}
	d.OnSale = onSale
	if stock >= 0 {
		d.Stock = stock
	}
	s.audit(ctx, operator, "store.dish.sale", d.ID)
	return d, nil
}

// DeleteDish 软删除菜品（历史订单快照不受影响）。
func (s *StoreService) DeleteDish(ctx context.Context, merchantID, dishID, operator string) error {
	d, err := s.ownDish(ctx, merchantID, dishID)
	if err != nil {
		return err
	}
	if err := s.Repos.Catalog.SoftDeleteDish(ctx, d.ID, time.Now().Unix()); err != nil {
		return err
	}
	s.audit(ctx, operator, "store.dish.delete", d.ID)
	return nil
}

func (s *StoreService) ownCategory(ctx context.Context, merchantID, categoryID string) (*catalog.Category, error) {
	c, err := s.Repos.Catalog.FindCategory(ctx, categoryID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeCategoryNotFound, "分类不存在")
		}
		return nil, err
	}
	if c.MerchantID != merchantID {
		return nil, errcode.NewStatus(404, errcode.CodeCategoryNotFound, "分类不存在")
	}
	return c, nil
}

func (s *StoreService) ownDish(ctx context.Context, merchantID, dishID string) (*catalog.Dish, error) {
	d, err := s.Repos.Catalog.FindDish(ctx, dishID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeDishNotFound, "菜品不存在")
		}
		return nil, err
	}
	if d.MerchantID != merchantID || d.DeletedAt > 0 {
		return nil, errcode.NewStatus(404, errcode.CodeDishNotFound, "菜品不存在")
	}
	return d, nil
}

func (s *StoreService) audit(ctx context.Context, operator, action, object string) {
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: action, Object: object, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx), Operator: operator,
	})
}

func validateCategoryName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if n := utf8.RuneCountInString(name); n < 1 || n > 32 {
		return "", errcode.New(errcode.CodeBadRequest, "分类名称需为 1-32 字")
	}
	return name, nil
}

func validateDishInput(in DishInput) (DishInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	if n := utf8.RuneCountInString(in.Name); n < 1 || n > 64 {
		return in, errcode.New(errcode.CodeBadRequest, "菜品名称需为 1-64 字")
	}
	if n := utf8.RuneCountInString(in.Description); n > 512 {
		return in, errcode.New(errcode.CodeBadRequest, "菜品描述不能超过 512 字")
	}
	if len(in.ImageURL) > 512 {
		return in, errcode.New(errcode.CodeBadRequest, "图片地址过长")
	}
	if in.Price < 0.01 || in.Price > 9999 {
		return in, errcode.New(errcode.CodeBadRequest, "价格需在 0.01-9999 元之间")
	}
	if in.Stock < 0 || in.Stock > 99999 {
		return in, errcode.New(errcode.CodeBadRequest, "库存需在 0-99999 之间")
	}
	in.Price = round2(in.Price)
	return in, nil
}

func merchantErr(err error) error {
	if errors.Is(err, application.ErrNotFound) {
		return errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
	}
	return err
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
