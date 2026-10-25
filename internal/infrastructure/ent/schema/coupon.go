package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserCoupon 用户优惠券。
type UserCoupon struct {
	ent.Schema
}

func (UserCoupon) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("user_id").MaxLen(26).NotEmpty(),
		field.String("merchant_id").MaxLen(26).Default(""),
		field.String("title").MaxLen(64).NotEmpty(),
		field.Float("amount").Default(0),
		field.Float("min_spend").Default(0),
		field.String("status").MaxLen(16).Default("unused"),
		field.Int64("expires_at").Default(0),
		field.String("used_order_id").MaxLen(32).Default(""),
		field.Int64("created_at").DefaultFunc(now),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (UserCoupon) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "status", "expires_at"),
		index.Fields("used_order_id"),
	}
}
