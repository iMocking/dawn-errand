package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RiderLocation 骑手最新位置（一人一行，覆盖更新）。
type RiderLocation struct {
	ent.Schema
}

func (RiderLocation) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("rider_id").MaxLen(26).NotEmpty().Unique(),
		field.String("region_id").MaxLen(26).Default(""),
		field.Float("lat").Default(0),
		field.Float("lng").Default(0),
		field.Float("heading").Default(0),
		field.Float("speed").Default(0),
		field.Float("accuracy").Default(0),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (RiderLocation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region_id"),
		index.Fields("updated_at"),
	}
}

// Notification 站内通知（三端共用，按用户维度）。
type Notification struct {
	ent.Schema
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("user_id").MaxLen(26).NotEmpty(),
		field.String("title").MaxLen(128).NotEmpty(),
		field.String("content").MaxLen(512).Default(""),
		field.String("type").MaxLen(16).Default("order"),
		field.String("ref_id").MaxLen(32).Default(""),
		field.Int64("read_at").Default(0),
		field.Int64("created_at").DefaultFunc(now),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
		index.Fields("user_id", "read_at"),
	}
}
