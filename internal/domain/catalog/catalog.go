// Package catalog 商品目录限界上下文：菜品分类与菜品。
package catalog

import "context"

// Category 菜品分类。
type Category struct {
	ID         string
	MerchantID string
	Name       string
	Sort       int
	CreatedAt  int64
	UpdatedAt  int64
}

// Dish 菜品。DeletedAt > 0 表示已软删除。
type Dish struct {
	ID           string
	MerchantID   string
	CategoryID   string
	Name         string
	Description  string
	ImageURL     string
	Price        float64
	Stock        int
	OnSale       bool
	MonthlySales int64
	Sort         int
	DeletedAt    int64
	CreatedAt    int64
	UpdatedAt    int64
}

// SoldOut 是否售罄（在售但库存为 0）。
func (d *Dish) SoldOut() bool { return d.OnSale && d.Stock <= 0 }

// DishQuery 菜品查询。
type DishQuery struct {
	MerchantID     string
	CategoryID     string
	Keyword        string
	IncludeOffSale bool // 商家端 true；用户端 false
	Page           int
	PageSize       int
}

// Repository 商品目录仓储端口。
type Repository interface {
	// 分类
	SaveCategory(ctx context.Context, c *Category) error
	FindCategory(ctx context.Context, id string) (*Category, error)
	ListCategories(ctx context.Context, merchantID string) ([]*Category, error)
	DeleteCategory(ctx context.Context, id string) error
	CountDishesByCategory(ctx context.Context, categoryID string) (int64, error)
	// 菜品
	SaveDish(ctx context.Context, d *Dish) error
	FindDish(ctx context.Context, id string) (*Dish, error)
	ListDishes(ctx context.Context, q DishQuery) ([]*Dish, int64, error)
	UpdateDishSale(ctx context.Context, id string, onSale bool, stock int, now int64) error
	SoftDeleteDish(ctx context.Context, id string, now int64) error
	CountDishes(ctx context.Context, merchantID string) (int64, error)
	// AdjustStock 原子调整库存：delta<0 扣减（库存不足返回 false），delta>0 释放。
	AdjustStock(ctx context.Context, dishID string, delta int) (bool, error)
}
