package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OrderEvent 订单状态/操作事件。
type OrderEvent struct {
	ent.Schema
}

func (OrderEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("order_id").MaxLen(32).NotEmpty(),
		field.String("from_status").MaxLen(16).Default(""),
		field.String("to_status").MaxLen(16).Default(""),
		field.String("actor_type").MaxLen(16).Default("system"),
		field.String("actor_id").MaxLen(26).Default(""),
		field.String("note").MaxLen(255).Default(""),
		field.Int64("created_at").DefaultFunc(now),
	}
}

func (OrderEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id", "created_at"),
	}
}

// Payment 支付单。
type Payment struct {
	ent.Schema
}

func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("order_id").MaxLen(32).NotEmpty(),
		field.String("user_id").MaxLen(26).Default(""),
		field.String("channel").MaxLen(16).Default(""),
		field.Float("amount").Default(0),
		field.String("status").MaxLen(16).Default("created"),
		field.String("trade_no").MaxLen(64).Default(""),
		field.Int64("paid_at").Default(0),
		field.Int64("created_at").DefaultFunc(now),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("trade_no"),
	}
}
