package repo

import (
	"context"
	"strings"

	"dawn-errand/internal/domain/catalog"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/dish"
	"dawn-errand/internal/infrastructure/ent/dishcategory"
)

// CatalogRepo implements catalog.Repository.
type CatalogRepo struct{ client *ent.Client }

func (r CatalogRepo) SaveCategory(ctx context.Context, c *catalog.Category) error {
	if c.ID == "" {
		c.ID = newULID()
	}
	exists, err := r.client.DishCategory.Query().Where(dishcategory.ID(c.ID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		err = r.client.DishCategory.UpdateOneID(c.ID).
			SetMerchantID(c.MerchantID).SetName(c.Name).SetSort(c.Sort).
			SetUpdatedAt(nowSec()).Exec(ctx)
		return err
	}
	return r.client.DishCategory.Create().
		SetID(c.ID).SetMerchantID(c.MerchantID).SetName(c.Name).SetSort(c.Sort).
		SetCreatedAt(nowSec()).SetUpdatedAt(nowSec()).Exec(ctx)
}

func (r CatalogRepo) FindCategory(ctx context.Context, id string) (*catalog.Category, error) {
	e, err := r.client.DishCategory.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &catalog.Category{
		ID: e.ID, MerchantID: e.MerchantID, Name: e.Name, Sort: e.Sort,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}, nil
}

func (r CatalogRepo) ListCategories(ctx context.Context, merchantID string) ([]*catalog.Category, error) {
	rows, err := r.client.DishCategory.Query().
		Where(dishcategory.MerchantID(merchantID)).
		Order(ent.Asc(dishcategory.FieldSort), ent.Asc(dishcategory.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*catalog.Category, 0, len(rows))
	for _, e := range rows {
		out = append(out, &catalog.Category{
			ID: e.ID, MerchantID: e.MerchantID, Name: e.Name, Sort: e.Sort,
			CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
		})
	}
	return out, nil
}

func (r CatalogRepo) DeleteCategory(ctx context.Context, id string) error {
	return r.client.DishCategory.DeleteOneID(id).Exec(ctx)
}

func (r CatalogRepo) CountDishesByCategory(ctx context.Context, categoryID string) (int64, error) {
	n, err := r.client.Dish.Query().
		Where(dish.CategoryID(categoryID), dish.DeletedAt(0)).
		Count(ctx)
	return int64(n), err
}

func (r CatalogRepo) SaveDish(ctx context.Context, d *catalog.Dish) error {
	if d.ID == "" {
		d.ID = newULID()
	}
	exists, err := r.client.Dish.Query().Where(dish.ID(d.ID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		err = r.client.Dish.UpdateOneID(d.ID).
			SetMerchantID(d.MerchantID).SetCategoryID(d.CategoryID).
			SetName(d.Name).SetDescription(d.Description).SetImageURL(d.ImageURL).
			SetPrice(d.Price).SetStock(d.Stock).SetOnSale(d.OnSale).SetSort(d.Sort).
			SetUpdatedAt(nowSec()).Exec(ctx)
		return err
	}
	return r.client.Dish.Create().
		SetID(d.ID).SetMerchantID(d.MerchantID).SetCategoryID(d.CategoryID).
		SetName(d.Name).SetDescription(d.Description).SetImageURL(d.ImageURL).
		SetPrice(d.Price).SetStock(d.Stock).SetOnSale(d.OnSale).SetSort(d.Sort).
		SetMonthlySales(d.MonthlySales).SetCreatedAt(nowSec()).SetUpdatedAt(nowSec()).
		Exec(ctx)
}

func (r CatalogRepo) FindDish(ctx context.Context, id string) (*catalog.Dish, error) {
	e, err := r.client.Dish.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toDish(e), nil
}

func toDish(e *ent.Dish) *catalog.Dish {
	if e == nil {
		return nil
	}
	return &catalog.Dish{
		ID: e.ID, MerchantID: e.MerchantID, CategoryID: e.CategoryID,
		Name: e.Name, Description: e.Description, ImageURL: e.ImageURL,
		Price: e.Price, Stock: e.Stock, OnSale: e.OnSale, MonthlySales: e.MonthlySales,
		Sort: e.Sort, DeletedAt: e.DeletedAt, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

func (r CatalogRepo) ListDishes(ctx context.Context, q catalog.DishQuery) ([]*catalog.Dish, int64, error) {
	bq := r.client.Dish.Query().Where(dish.DeletedAt(0))
	if q.MerchantID != "" {
		bq.Where(dish.MerchantID(q.MerchantID))
	}
	if q.CategoryID != "" {
		bq.Where(dish.CategoryID(q.CategoryID))
	}
	if !q.IncludeOffSale {
		bq.Where(dish.OnSale(true))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(dish.NameContainsFold(k))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Asc(dish.FieldSort), ent.Desc(dish.FieldCreatedAt)).
		Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*catalog.Dish, 0, len(rows))
	for _, e := range rows {
		out = append(out, toDish(e))
	}
	return out, int64(total), nil
}

func (r CatalogRepo) UpdateDishSale(ctx context.Context, id string, onSale bool, stock int, now int64) error {
	upd := r.client.Dish.UpdateOneID(id).SetOnSale(onSale).SetUpdatedAt(now)
	if stock >= 0 {
		upd.SetStock(stock)
	}
	return upd.Exec(ctx)
}

func (r CatalogRepo) SoftDeleteDish(ctx context.Context, id string, now int64) error {
	return r.client.Dish.UpdateOneID(id).
		SetDeletedAt(now).SetOnSale(false).SetUpdatedAt(now).
		Exec(ctx)
}

func (r CatalogRepo) CountDishes(ctx context.Context, merchantID string) (int64, error) {
	n, err := r.client.Dish.Query().
		Where(dish.MerchantID(merchantID), dish.DeletedAt(0)).
		Count(ctx)
	return int64(n), err
}

// AdjustStock 原子调整库存；delta<0 时库存不足返回 false。
func (r CatalogRepo) AdjustStock(ctx context.Context, dishID string, delta int) (bool, error) {
	q := r.client.Dish.Update().Where(dish.ID(dishID))
	if delta < 0 {
		q = q.Where(dish.StockGTE(-delta))
	}
	affected, err := q.AddStock(delta).Save(ctx)
	return affected > 0, err
}
