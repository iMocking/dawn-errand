package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Merchant 商家。
type Merchant struct {
	ent.Schema
}

func (Merchant) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("region_id").MaxLen(26).NotEmpty(),
		field.String("owner_user_id").MaxLen(26).Default(""),
		field.String("name").MaxLen(128).NotEmpty(),
		field.String("category").MaxLen(64).Default(""),
		field.String("address").MaxLen(255).Default(""),
		field.String("address_sub").MaxLen(128).Default(""),
		field.String("contact").MaxLen(64).Default(""),
		field.String("phone").MaxLen(32).Default(""),
		field.String("open_hours").MaxLen(64).Default(""),
		field.String("delivery_range").MaxLen(64).Default(""),
		field.Float("commission").Default(0),
		field.String("bank_account").MaxLen(128).Default(""),
		field.String("delivery_method").MaxLen(64).Default(""),
		// 门店位置与经营设置（App 商家端）。
		field.Float("lat").Default(0),
		field.Float("lng").Default(0),
		field.Bool("auto_accept").Default(false),
		field.Float("min_price").Default(0),
		field.Float("delivery_fee").Default(0),
		field.Float("packing_fee").Default(0),
		field.String("announcement").MaxLen(512).Default(""),
		field.String("service_phone").MaxLen(32).Default(""),
		field.String("status").MaxLen(16).Default("open"),
		field.String("audit_status").MaxLen(16).Default("approved"),
		field.String("apply_no").MaxLen(32).Default(""),
		field.String("operator").MaxLen(64).Default(""),
		field.Int64("applied_at").Default(0),
		field.Strings("cert_labels").Optional(),
		field.String("reject_reason").MaxLen(255).Default(""),
		field.Float("rating").Default(0),
		field.Int64("monthly_orders").Default(0),
		field.Float("monthly_revenue").Default(0),
		field.Float("avg_order").Default(0),
		field.Float("good_rate").Default(0),
		field.Int64("joined_at").Default(0),
		field.Int64("last_updated_at").DefaultFunc(now),
		field.Int("completeness").Default(0),
		field.Int("review_count").Default(0),
		field.String("last_operation").MaxLen(128).Default(""),
		field.Int64("today_orders").Default(0),
		field.Float("today_revenue").Default(0),
		field.Float("on_time_rate").Default(0),
		field.Int("pending_abnormal").Default(0),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (Merchant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region_id", "status"),
		index.Fields("audit_status"),
		index.Fields("name"),
	}
}

// Rider 骑手（跑手）。
type Rider struct {
	ent.Schema
}

func (Rider) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("region_id").MaxLen(26).NotEmpty(),
		field.String("user_id").MaxLen(26).Default(""),
		field.String("name").MaxLen(64).NotEmpty(),
		field.String("phone").MaxLen(32).Default(""),
		field.String("area").MaxLen(128).Default(""),
		field.String("tool").MaxLen(32).Default(""),
		field.String("level").MaxLen(16).Default("Lv.1"),
		field.String("id_card").MaxLen(64).Default(""),
		field.String("vehicle").MaxLen(64).Default(""),
		field.String("account").MaxLen(128).Default(""),
		field.String("status").MaxLen(16).Default("offline"),
		field.String("audit_status").MaxLen(16).Default("pending"),
		field.String("apply_no").MaxLen(32).Default(""),
		field.Int64("submitted_at").Default(0),
		field.String("order_types").MaxLen(128).Default(""),
		field.String("id_check").MaxLen(16).Default("待补充"),
		field.String("notice_type").MaxLen(16).Default("none"),
		field.String("notice").MaxLen(255).Default(""),
		field.Int("doc_count").Default(0),
		field.String("docs").MaxLen(255).Default(""),
		field.String("service_type").MaxLen(64).Default(""),
		field.Int64("joined_at").Default(0),
		field.Int("service_months").Default(0),
		field.Float("rating").Default(0),
		field.Int64("today_orders").Default(0),
		field.String("today_orders_delta").MaxLen(32).Default(""),
		field.Float("online_hours").Default(0),
		field.Float("month_income").Default(0),
		field.Float("on_time_rate").Default(0),
		field.Int("complaints_today").Default(0),
		field.Float("avg_minutes").Default(0),
		field.String("last_action").MaxLen(64).Default(""),
		field.Int64("last_action_at").Default(0),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (Rider) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region_id", "status"),
		index.Fields("audit_status"),
		index.Fields("phone"),
	}
}

// Order 订单。
type Order struct {
	ent.Schema
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("order_no").MaxLen(32).NotEmpty().Unique(),
		field.String("region_id").MaxLen(26).NotEmpty(),
		field.String("merchant_id").MaxLen(26).Default(""),
		field.String("user_id").MaxLen(26).NotEmpty(),
		field.String("rider_id").MaxLen(26).Default(""),
		field.String("type").MaxLen(16).NotEmpty(),
		field.String("status").MaxLen(16).NotEmpty(),
		field.String("route").MaxLen(255).Default(""),
		field.String("user_name").MaxLen(64).Default(""),
		field.String("rider_name").MaxLen(64).Default(""),
		field.Float("amount").Default(0),
		field.Int64("created_at").DefaultFunc(now),
		field.String("channel").MaxLen(64).Default(""),
		field.Int64("estimated_arrival").Default(0),
		field.String("remark").MaxLen(512).Default(""),
		field.Bytes("goods").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Bytes("fees").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Bytes("steps").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Bytes("merchant").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Bytes("user").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Bytes("rider").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Float("pay_amount").Default(0),
		field.String("pay_method").MaxLen(64).Default(""),
		field.String("fund_flow").MaxLen(255).Default(""),
		// 交易闭环字段（M2）。
		field.String("client_order_no").MaxLen(64).Default(""),
		field.String("address_id").MaxLen(26).Default(""),
		field.Bytes("address_snapshot").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Float("pickup_lat").Default(0),
		field.Float("pickup_lng").Default(0),
		field.Float("delivery_lat").Default(0),
		field.Float("delivery_lng").Default(0),
		field.Float("delivery_distance_km").Default(0),
		field.Float("delivery_fee").Default(0),
		field.Float("packing_fee").Default(0),
		field.Float("discount_amount").Default(0),
		field.Float("coupon_amount").Default(0),
		field.String("coupon_id").MaxLen(26).Default(""),
		field.String("payment_status").MaxLen(16).Default("unpaid"),
		field.Int64("accepted_at").Default(0),
		field.Int64("ready_at").Default(0),
		field.Int64("picked_at").Default(0),
		field.Int64("delivered_at").Default(0),
		field.Int64("canceled_at").Default(0),
		field.Int64("confirmed_at").Default(0),
		field.Int64("reviewed_at").Default(0),
		field.String("cancel_reason").MaxLen(255).Default(""),
		field.Float("rider_earning").Default(0),
		field.String("pickup_code").MaxLen(8).Default(""),
		field.String("delivery_code").MaxLen(8).Default(""),
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region_id", "status", "created_at"),
		index.Fields("merchant_id", "created_at"),
		index.Fields("user_id", "created_at"),
		index.Fields("rider_id"),
		index.Fields("user_id", "client_order_no"),
		index.Fields("region_id", "status", "rider_id"),
	}
}

// Complaint 投诉工单。
type Complaint struct {
	ent.Schema
}

func (Complaint) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("order_id").MaxLen(32).Default(""),
		field.String("region_id").MaxLen(26).Default(""),
		field.String("user_id").MaxLen(26).Default(""),
		field.String("category").MaxLen(32).NotEmpty(),
		field.String("title").MaxLen(128).NotEmpty(),
		field.String("status").MaxLen(16).Default("待处理"),
		field.String("priority").MaxLen(16).Default("普通"),
		field.String("complainant").MaxLen(128).Default(""),
		field.String("target").MaxLen(128).Default(""),
		field.String("extra").MaxLen(128).Default(""),
		field.String("description").MaxLen(1024).Default(""),
		field.String("desc_tone").MaxLen(16).Default("yellow"),
		field.String("deadline").MaxLen(64).Default(""),
		field.String("handler").MaxLen(64).Default(""),
		field.String("progress").MaxLen(255).Default(""),
		field.String("initiator").MaxLen(64).Default(""),
		field.String("initiator_extra").MaxLen(128).Default(""),
		field.Float("pay_amount").Default(0),
		field.Int64("created_at").DefaultFunc(now),
		field.Int64("updated_at").DefaultFunc(now),
		field.Bytes("flow").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Bytes("notes").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Bytes("order_items").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (Complaint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "created_at"),
		index.Fields("order_id"),
	}
}

// Withdrawal 提现单。
type Withdrawal struct {
	ent.Schema
}

func (Withdrawal) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("region_id").MaxLen(26).Default(""),
		field.String("party_type").MaxLen(16).Default("rider"),
		field.String("party_id").MaxLen(26).Default(""),
		field.String("party_name").MaxLen(64).Default(""),
		field.String("party_no").MaxLen(32).Default(""),
		field.String("bank_info").MaxLen(128).Default(""),
		field.Float("amount").Default(0),
		field.Float("fee").Default(0),
		field.String("fee_note").MaxLen(64).Default(""),
		field.String("status").MaxLen(16).Default("待审核"),
		field.String("kind").MaxLen(32).Default("骑手提现"),
		field.String("applicant").MaxLen(64).Default(""),
		field.String("account_holder").MaxLen(64).Default(""),
		field.String("withdraw_type").MaxLen(64).Default(""),
		field.String("arrive_expect").MaxLen(128).Default(""),
		field.String("purpose").MaxLen(255).Default(""),
		field.Float("balance").Default(0),
		field.Float("remain_balance").Default(0),
		field.Int64("month_orders").Default(0),
		field.Float("month_income").Default(0),
		field.Strings("risk_items").Optional(),
		field.String("risk_warning").MaxLen(255).Default(""),
		field.String("reject_reason").MaxLen(255).Default(""),
		field.Int64("created_at").DefaultFunc(now),
		field.Bytes("timeline").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (Withdrawal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "created_at"),
		index.Fields("region_id"),
	}
}

// Settlement 结算流水。
type Settlement struct {
	ent.Schema
}

func (Settlement) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("region_id").MaxLen(26).Default(""),
		field.String("party").MaxLen(64).Default(""),
		field.String("party_type").MaxLen(16).Default(""),
		field.String("party_id").MaxLen(26).Default(""),
		field.String("type").MaxLen(32).Default(""),
		field.Float("amount").Default(0),
		field.String("status").MaxLen(16).Default("已到账"),
		field.String("period").MaxLen(64).Default(""),
		field.Int64("settled_at").DefaultFunc(now),
		field.String("action").MaxLen(64).Default(""),
	}
}

func (Settlement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("settled_at"),
		index.Fields("type"),
	}
}
