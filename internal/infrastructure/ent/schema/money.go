package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// WalletAccount 钱包账户（用户/商家/骑手）。
type WalletAccount struct {
	ent.Schema
}

func (WalletAccount) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("party_type").MaxLen(16).NotEmpty(), // user / merchant / rider
		field.String("party_id").MaxLen(26).NotEmpty(),
		field.Float("balance").Default(0),
		field.Float("frozen").Default(0),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (WalletAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("party_type", "party_id").Unique(),
	}
}

// WalletTransaction 钱包流水（只增不改）。
type WalletTransaction struct {
	ent.Schema
}

func (WalletTransaction) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("account_id").MaxLen(26).NotEmpty(),
		field.String("party_type").MaxLen(16).Default(""),
		field.String("party_id").MaxLen(26).Default(""),
		field.String("order_id").MaxLen(32).Default(""),
		field.String("type").MaxLen(32).NotEmpty(),
		field.Float("amount").Default(0),
		field.Float("balance_after").Default(0),
		field.String("note").MaxLen(255).Default(""),
		field.Int64("created_at").DefaultFunc(now),
	}
}

func (WalletTransaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id", "created_at"),
		index.Fields("party_type", "party_id", "created_at"),
		index.Fields("order_id"),
	}
}

// Review 订单评价。
type Review struct {
	ent.Schema
}

func (Review) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("order_id").MaxLen(32).NotEmpty().Unique(),
		field.String("user_id").MaxLen(26).NotEmpty(),
		field.String("merchant_id").MaxLen(26).Default(""),
		field.String("rider_id").MaxLen(26).Default(""),
		field.Int("rating").Default(5),
		field.Strings("tags").Optional(),
		field.String("content").MaxLen(1024).Default(""),
		field.Strings("images").Optional(),
		field.String("reply").MaxLen(1024).Default(""),
		field.Int64("replied_at").Default(0),
		field.Int64("created_at").DefaultFunc(now),
	}
}

func (Review) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("merchant_id", "created_at"),
		index.Fields("rider_id", "created_at"),
	}
}

// Refund 退款/售后单。
type Refund struct {
	ent.Schema
}

func (Refund) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("order_id").MaxLen(32).NotEmpty(),
		field.String("user_id").MaxLen(26).Default(""),
		field.String("merchant_id").MaxLen(26).Default(""),
		field.Float("amount").Default(0),
		field.String("reason").MaxLen(512).Default(""),
		field.String("status").MaxLen(16).Default("requested"),
		field.String("prev_status").MaxLen(16).Default(""),
		field.String("operator").MaxLen(26).Default(""),
		field.String("reject_reason").MaxLen(255).Default(""),
		field.Int64("created_at").DefaultFunc(now),
		field.Int64("handled_at").Default(0),
	}
}

func (Refund) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("status", "created_at"),
	}
}
