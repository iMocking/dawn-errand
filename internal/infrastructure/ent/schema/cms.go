package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Announcement 公告。
type Announcement struct {
	ent.Schema
}

func (Announcement) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("title").MaxLen(128).NotEmpty(),
		field.String("type").MaxLen(16).Default("提醒"),
		field.String("type_tone").MaxLen(16).Default("blue"),
		field.String("audience").MaxLen(64).Default("全部用户"),
		field.String("status").MaxLen(16).Default("待发布"),
		field.Int64("published_at").Default(0),
		field.Int64("planned_at").Default(0),
		field.String("action_name").MaxLen(16).Default("发布"),
		field.String("content").MaxLen(2048).Default(""),
		field.Strings("tags").Optional(),
		field.Bytes("stats").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("created_by").MaxLen(26).Default(""),
		field.Int64("created_at").DefaultFunc(now),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (Announcement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "planned_at"),
		index.Fields("type"),
	}
}

// FeeRule 费率规则。
type FeeRule struct {
	ent.Schema
}

func (FeeRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("category").MaxLen(16).NotEmpty(),
		field.String("name").MaxLen(64).NotEmpty(),
		field.String("status").MaxLen(16).Default("已启用"),
		field.String("value").MaxLen(64).Default(""),
		field.String("desc").MaxLen(128).Default(""),
		field.Bool("accent").Default(false),
		field.Bool("highlight").Default(false),
		field.String("rule").MaxLen(128).Default(""),
		field.String("example").MaxLen(128).Default(""),
		field.String("fee").MaxLen(64).Default(""),
		field.String("rider_share").MaxLen(64).Default(""),
		field.String("icon_color").MaxLen(32).Default(""),
		field.String("version").MaxLen(16).Default(""),
		field.Int64("updated_at").DefaultFunc(now),
		field.Int64("next_effective_at").Default(0),
		field.Bool("enabled").Default(true),
	}
}

func (FeeRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category"),
	}
}

// MailOutbox dev 邮件发件箱（仅 dev/test 环境启用）。
type MailOutbox struct {
	ent.Schema
}

func (MailOutbox) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("to_email").MaxLen(255).NotEmpty(),
		field.String("subject").MaxLen(255).Default(""),
		field.String("body").MaxLen(4096).Default(""),
		field.Int64("created_at").DefaultFunc(now),
	}
}

func (MailOutbox) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
	}
}

// CasbinRule Casbin 策略持久化。
type CasbinRule struct {
	ent.Schema
}

func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("ptype").MaxLen(16).NotEmpty(),
		field.String("v0").MaxLen(64).Default(""),
		field.String("v1").MaxLen(64).Default(""),
		field.String("v2").MaxLen(64).Default(""),
		field.String("v3").MaxLen(64).Default(""),
		field.String("v4").MaxLen(64).Default(""),
		field.String("v5").MaxLen(64).Default(""),
	}
}

func (CasbinRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ptype", "v0"),
	}
}

// AuditLog 审计日志（认证/密码/权限/配置/初始化）。
type AuditLog struct {
	ent.Schema
}

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("user_id").MaxLen(26).Default(""),
		field.String("action").MaxLen(64).NotEmpty(),
		field.String("object").MaxLen(255).Default(""),
		field.String("result").MaxLen(16).Default("success"),
		field.String("request_id").MaxLen(32).Default(""),
		field.String("operator").MaxLen(64).Default(""),
		field.String("ip").MaxLen(64).Default(""),
		field.Int64("created_at").DefaultFunc(now),
	}
}

func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
		index.Fields("action"),
	}
}
