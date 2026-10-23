package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

func now() int64 { return time.Now().Unix() }

func ulid() string { return newULID() }

// User 用户（正式会员/平台角色）。
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("email").MaxLen(255).NotEmpty(),
		field.String("phone").MaxLen(32).Default(""),
		field.String("name").MaxLen(64).NotEmpty(),
		field.String("avatar").MaxLen(512).Default(""),
		field.String("status").MaxLen(16).Default("active"),
		field.Bool("email_verified").Default(false),
		field.String("role").MaxLen(32).Default("user"),
		field.String("region_id").MaxLen(26).Default(""),
		field.String("merchant_id").MaxLen(26).Default(""),
		field.String("rider_id").MaxLen(26).Default(""),
		field.Int64("registered_at").DefaultFunc(now),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
		index.Fields("phone"),
		index.Fields("status"),
		index.Fields("region_id"),
		index.Fields("role"),
	}
}

// Credential 登录凭证。
type Credential struct {
	ent.Schema
}

func (Credential) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("user_id").MaxLen(26).NotEmpty().Unique(),
		field.String("password_hash").MaxLen(255).NotEmpty(),
		field.Int64("session_version").Default(0),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

// EmailVerificationToken 邮箱验证令牌。
type EmailVerificationToken struct {
	ent.Schema
}

func (EmailVerificationToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("user_id").MaxLen(26).NotEmpty(),
		field.String("token_hash").MaxLen(64).NotEmpty().Unique(),
		field.Int64("expires_at").Default(0),
		field.Int64("used_at").Default(0),
	}
}

func (EmailVerificationToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}

// PasswordResetToken 密码重置令牌。
type PasswordResetToken struct {
	ent.Schema
}

func (PasswordResetToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("user_id").MaxLen(26).NotEmpty(),
		field.String("token_hash").MaxLen(64).NotEmpty().Unique(),
		field.Int64("expires_at").Default(0),
		field.Int64("used_at").Default(0),
	}
}

func (PasswordResetToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}

// Region 服务区域（租户单元）。
type Region struct {
	ent.Schema
}

func (Region) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("name").MaxLen(64).NotEmpty().Unique(),
		field.String("status").MaxLen(16).Default("open"),
		field.Int("villages").Default(0),
		field.Float("radius_km").Default(0),
		field.Float("start_fee").Default(0),
		field.Float("free_km").Default(0),
		field.String("notice").MaxLen(255).Default(""),
		field.Int64("created_at").DefaultFunc(now),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

// SystemSetting 系统配置。
type SystemSetting struct {
	ent.Schema
}

func (SystemSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("key").MaxLen(64).NotEmpty().Unique(),
		field.String("status").MaxLen(16).Default("disabled"),
		field.Bytes("value_json").Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("updated_by").MaxLen(26).Default(""),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

// UserProfile 用户资料。
type UserProfile struct {
	ent.Schema
}

func (UserProfile) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("user_id").MaxLen(26).NotEmpty().Unique(),
		field.String("level").MaxLen(64).Default("Lv.1 新用户"),
		field.String("nickname").MaxLen(64).Default(""),
		field.String("real_name").MaxLen(64).Default(""),
		field.String("verified").MaxLen(16).Default("未实名"),
		field.String("region").MaxLen(128).Default(""),
		field.String("address").MaxLen(255).Default(""),
		field.String("register_channel").MaxLen(128).Default(""),
		field.String("biz_types").MaxLen(128).Default(""),
		field.Float("balance").Default(0),
		field.Float("red_packet_balance").Default(0),
		field.Int64("recent_active_at").Default(0),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

// RealNameVerification 用户实名认证申请（一人仅一条当前申请）。
type RealNameVerification struct {
	ent.Schema
}

func (RealNameVerification) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(26).DefaultFunc(ulid),
		field.String("user_id").MaxLen(26).NotEmpty().Unique(),
		field.String("real_name").MaxLen(64).NotEmpty(),
		// id_card 存储加密密文；对外仅返回脱敏（前 6 后 4）。
		field.String("id_card_cipher").MaxLen(512).NotEmpty(),
		field.String("status").MaxLen(16).Default("待审核"),
		field.String("reject_reason").MaxLen(255).Default(""),
		field.String("reviewed_by").MaxLen(26).Default(""),
		field.Int64("applied_at").DefaultFunc(now),
		field.Int64("reviewed_at").Default(0),
		field.Int64("updated_at").DefaultFunc(now),
	}
}

func (RealNameVerification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
		index.Fields("status"),
	}
}
