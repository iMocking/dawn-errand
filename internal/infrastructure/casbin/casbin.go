// Package casbin 基于 Casbin 的 RBAC 权限实现。
package casbin

import (
	"context"
	_ "embed"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"

	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/casbinrule"
)

//go:embed model.conf
var modelConf string

// Adapter 基于 ent 的 Casbin 策略适配器（持久化到 casbin_rules 表）。
type Adapter struct {
	client *ent.Client
}

// NewAdapter 构建适配器。
func NewAdapter(client *ent.Client) *Adapter { return &Adapter{client: client} }

func (a *Adapter) LoadPolicy(m model.Model) error {
	rows, err := a.client.CasbinRule.Query().Order(ent.Asc(casbinrule.FieldID)).All(context.Background())
	if err != nil {
		return err
	}
	for _, r := range rows {
		line := policyLine(r.Ptype, r.V0, r.V1, r.V2, r.V3, r.V4, r.V5)
		if err := persist.LoadPolicyLine(line, m); err != nil {
			return err
		}
	}
	return nil
}

func policyLine(ptype string, v ...string) string {
	last := len(v)
	for last > 0 && v[last-1] == "" {
		last--
	}
	parts := make([]string, 0, last+1)
	parts = append(parts, ptype)
	parts = append(parts, v[:last]...)
	return strings.Join(parts, ", ")
}

func (a *Adapter) SavePolicy(m model.Model) error {
	var rules []*ent.CasbinRule
	for ptype, ast := range m["p"] {
		for _, rule := range ast.Policy {
			rules = append(rules, a.buildRule(ptype, rule))
		}
	}
	for ptype, ast := range m["g"] {
		for _, rule := range ast.Policy {
			rules = append(rules, a.buildRule(ptype, rule))
		}
	}
	if err := a.replaceAll(rules); err != nil {
		return err
	}
	return nil
}

func (a *Adapter) buildRule(ptype string, rule []string) *ent.CasbinRule {
	fill := make([]string, 6)
	copy(fill, rule)
	return a.client.CasbinRule.Create().
		SetPtype(ptype).SetV0(fill[0]).SetV1(fill[1]).SetV2(fill[2]).
		SetV3(fill[3]).SetV4(fill[4]).SetV5(fill[5]).
		SaveX(context.Background())
}

func (a *Adapter) replaceAll(rules []*ent.CasbinRule) error {
	ctx := context.Background()
	tx, err := a.client.Tx(ctx)
	if err != nil {
		return err
	}
	if _, err := tx.CasbinRule.Delete().Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, r := range rules {
		if _, err := tx.CasbinRule.Create().
			SetPtype(r.Ptype).SetV0(r.V0).SetV1(r.V1).SetV2(r.V2).
			SetV3(r.V3).SetV4(r.V4).SetV5(r.V5).
			Save(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (a *Adapter) AddPolicy(sec, ptype string, rule []string) error {
	return a.saveSingle(ptype, rule)
}

func (a *Adapter) RemovePolicy(sec, ptype string, rule []string) error {
	q := a.client.CasbinRule.Delete().Where(casbinrule.Ptype(ptype))
	return a.removeByRule(q, rule)
}

func (a *Adapter) removeByRule(q *ent.CasbinRuleDelete, rule []string) error {
	if len(rule) > 0 && rule[0] != "" {
		q.Where(casbinrule.V0(rule[0]))
	}
	if len(rule) > 1 && rule[1] != "" {
		q.Where(casbinrule.V1(rule[1]))
	}
	if len(rule) > 2 && rule[2] != "" {
		q.Where(casbinrule.V2(rule[2]))
	}
	_, err := q.Exec(context.Background())
	return err
}

func (a *Adapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	q := a.client.CasbinRule.Delete().Where(casbinrule.Ptype(ptype))
	for i, v := range fieldValues {
		if v == "" {
			continue
		}
		switch fieldIndex + i {
		case 0:
			q.Where(casbinrule.V0(v))
		case 1:
			q.Where(casbinrule.V1(v))
		case 2:
			q.Where(casbinrule.V2(v))
		case 3:
			q.Where(casbinrule.V3(v))
		case 4:
			q.Where(casbinrule.V4(v))
		case 5:
			q.Where(casbinrule.V5(v))
		}
	}
	_, err := q.Exec(context.Background())
	return err
}

func (a *Adapter) saveSingle(ptype string, rule []string) error {
	fill := make([]string, 6)
	copy(fill, rule)
	_, err := a.client.CasbinRule.Create().
		SetPtype(ptype).SetV0(fill[0]).SetV1(fill[1]).SetV2(fill[2]).
		SetV3(fill[3]).SetV4(fill[4]).SetV5(fill[5]).
		Save(context.Background())
	return err
}

// Enforcer 封装 Casbin 执行器。
type Enforcer struct {
	e *casbin.Enforcer
}

// NewEnforcer 构建执行器（model 内嵌，policy 从 DB 加载）。
func NewEnforcer(client *ent.Client) (*Enforcer, error) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(m, NewAdapter(client))
	if err != nil {
		return nil, err
	}
	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}
	return &Enforcer{e: e}, nil
}

// Enforce 权限判定：sub 角色对 obj 资源执行 act。
func (en *Enforcer) Enforce(role, obj, act string) (bool, error) {
	return en.e.Enforce(role, obj, act)
}

// PermissionsForRole 返回角色直接拥有的权限列表（obj:act；super_admin 返回 *）。
func (en *Enforcer) PermissionsForRole(role string) []string {
	if role == identity.RoleSuperAdmin {
		return []string{"*"}
	}
	var out []string
	seen := map[string]bool{}
	for _, rule := range en.e.GetModel()["p"]["p"].Policy {
		if len(rule) >= 3 && rule[0] == role {
			perm := rule[1] + ":" + rule[2]
			if !seen[perm] {
				seen[perm] = true
				out = append(out, perm)
			}
		}
	}
	return out
}

// SavePolicies 全量替换策略并重载（用于初始化与配置管理，策略变更需审计）。
func (en *Enforcer) SavePolicies(policies [][]string) error {
	if err := en.e.GetAdapter().SavePolicy(modelFromPolicies(policies)); err != nil {
		return err
	}
	return en.e.LoadPolicy()
}

func modelFromPolicies(policies [][]string) model.Model {
	m, _ := model.NewModelFromString(modelConf)
	p := m.AddDef("p", "p", "sub, obj, act")
	g := m.AddDef("g", "g", "_, _")
	_ = p
	for _, rule := range policies {
		if len(rule) == 0 {
			continue
		}
		ptype := rule[0]
		params := rule[1:]
		switch ptype {
		case "p":
			if len(params) == 3 {
				_ = m.AddPolicy("p", "p", params)
			}
		case "g":
			if len(params) == 2 {
				_ = m.AddPolicy("g", "g", params)
			}
		}
	}
	_ = g
	return m
}

// Policies 返回当前策略（含 ptype 前缀）。
func (en *Enforcer) Policies() [][]string {
	var out [][]string
	for ptype, section := range en.e.GetModel() {
		if ptype != "p" && ptype != "g" {
			continue
		}
		ast, ok := section[ptype]
		if !ok {
			continue
		}
		for _, rule := range ast.Policy {
			row := append([]string{ptype}, rule...)
			out = append(out, row)
		}
	}
	return out
}

// HasRole 角色是否具有某权限（通配符展开）。
func (en *Enforcer) HasRole(role string) bool { return role != "" }

// DefaultPolicies 默认策略（bootstrap/种子时写入 DB）。
func DefaultPolicies() [][]string {
	return [][]string{
		{"p", identity.RoleSuperAdmin, "*", "*"},

		{"p", identity.RoleCSManager, "dashboard:*", "*"},
		{"p", identity.RoleCSManager, "order:*", "*"},
		{"p", identity.RoleCSManager, "complaint:*", "*"},
		{"p", identity.RoleCSManager, "user:*", "*"},
		{"p", identity.RoleCSManager, "merchant:*", "*"},
		{"p", identity.RoleCSManager, "rider:*", "*"},
		{"p", identity.RoleCSManager, "finance:*", "*"},
		{"p", identity.RoleCSManager, "settings:list", "*"},

		{"p", identity.RoleCSAgent, "dashboard:view", "*"},
		{"p", identity.RoleCSAgent, "order:list", "*"},
		{"p", identity.RoleCSAgent, "order:detail", "*"},
		{"p", identity.RoleCSAgent, "complaint:*", "*"},
		{"p", identity.RoleCSAgent, "user:list", "*"},
		{"p", identity.RoleCSAgent, "user:detail", "*"},

		{"p", identity.RoleRegionStation, "dashboard:*", "*"},
		{"p", identity.RoleRegionStation, "order:*", "*"},
		{"p", identity.RoleRegionStation, "merchant:*", "*"},
		{"p", identity.RoleRegionStation, "rider:*", "*"},
		{"p", identity.RoleRegionStation, "user:*", "*"},
		{"p", identity.RoleRegionStation, "complaint:*", "*"},
		{"p", identity.RoleRegionStation, "finance:*", "*"},
		{"p", identity.RoleRegionStation, "settings:list", "*"},
		{"p", identity.RoleRegionStation, "settings:area", "*"},

		{"p", identity.RoleMerchant, "dashboard:view", "*"},
		{"p", identity.RoleMerchant, "order:list", "*"},
		{"p", identity.RoleMerchant, "order:detail", "*"},
		{"p", identity.RoleMerchant, "merchant:self", "*"},
		{"p", identity.RoleMerchant, "settings:list", "*"},

		{"p", identity.RoleDriver, "dashboard:view", "*"},
		{"p", identity.RoleDriver, "order:list", "*"},
		{"p", identity.RoleDriver, "order:detail", "*"},
		{"p", identity.RoleDriver, "rider:self", "*"},

		{"p", identity.RoleUser, "order:list", "*"},
		{"p", identity.RoleUser, "order:detail", "*"},
	}
}

// PolicyText 生成 CasbinConfig 的 policy 行文本。
func PolicyText(policies [][]string) []string {
	out := make([]string, 0, len(policies))
	for _, p := range policies {
		out = append(out, strings.Join(p, ", "))
	}
	return out
}

// ModelText 返回 model 配置文本。
func ModelText() string { return modelConf }
