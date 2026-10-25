package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DishCategory 菜品分类（商家维度）。
type DishCategory struct {
	ent.Schema
}

func (DishCategory) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("merchant_id").MaxLen(26).NotEmpty(),
		field.String("name").MaxLen(64).NotEmpty(),
		field.Int("sort").Default(0),
		field.Int64("created_at").DefaultFunc(now),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (DishCategory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("merchant_id", "sort"),
	}
}

// Dish 菜品（软删除：deleted_at > 0 表示已删除）。
type Dish struct {
	ent.Schema
}

func (Dish) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("merchant_id").MaxLen(26).NotEmpty(),
		field.String("category_id").MaxLen(26).NotEmpty(),
		field.String("name").MaxLen(128).NotEmpty(),
		field.String("description").MaxLen(512).Default(""),
		field.String("image_url").MaxLen(512).Default(""),
		field.Float("price").Default(0),
		field.Int("stock").Default(0),
		field.Bool("on_sale").Default(true),
		field.Int64("monthly_sales").Default(0),
		field.Int("sort").Default(0),
		field.Int64("deleted_at").Default(0),
		field.Int64("created_at").DefaultFunc(now),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (Dish) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("merchant_id", "on_sale"),
		index.Fields("category_id"),
		index.Fields("name"),
	}
}

// UserAddress 用户收货地址。
type UserAddress struct {
	ent.Schema
}

func (UserAddress) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("user_id").MaxLen(26).NotEmpty(),
		field.String("contact_name").MaxLen(64).Default(""),
		field.String("phone").MaxLen(32).Default(""),
		field.String("region_id").MaxLen(26).Default(""),
		field.String("detail").MaxLen(255).NotEmpty(),
		field.String("tag").MaxLen(16).Default(""),
		field.Float("lat").Default(0),
		field.Float("lng").Default(0),
		field.Bool("is_default").Default(false),
		field.Int64("created_at").DefaultFunc(now),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (UserAddress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
	}
}
