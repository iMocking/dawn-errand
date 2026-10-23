// Package repo 基于 ent 的仓储实现（infrastructure 层）。
package repo

import (
	"context"
	"database/sql"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/merchant"
	"dawn-errand/internal/infrastructure/ent/order"
	"dawn-errand/internal/infrastructure/ent/rider"
	"dawn-errand/internal/infrastructure/ent/user"
)

// Store 仓储入口：持有 ent 客户端，提供事务运行器与全部仓储。
type Store struct {
	client *ent.Client
}

// NewStore 构建 Store。
func NewStore(client *ent.Client) *Store { return &Store{client: client} }

// Client 返回 ent 客户端（供迁移/种子等内部使用）。
func (s *Store) Client() *ent.Client { return s.client }

// Repos 构建绑定当前客户端的仓储集合。
func (s *Store) Repos() *application.RepoUnit {
	return &application.RepoUnit{
		Users:          UserRepo{client: s.client},
		Credentials:    CredentialRepo{client: s.client},
		EmailTokens:    EmailTokenRepo{client: s.client},
		PasswordTokens: PasswordTokenRepo{client: s.client},
		Regions:        RegionRepo{client: s.client},
		Settings:       SettingRepo{client: s.client},
		Merchants:      MerchantRepo{client: s.client},
		Riders:         RiderRepo{client: s.client},
		Orders:         OrderRepo{client: s.client},
		Complaints:     ComplaintRepo{client: s.client},
		Withdrawals:    WithdrawalRepo{client: s.client},
		Settlements:    SettlementRepo{client: s.client},
		Announcements:  AnnouncementRepo{client: s.client},
		FeeRules:       FeeRuleRepo{client: s.client},
		Profiles:       ProfileRepo{client: s.client},
		RealNames:      RealNameRepo{client: s.client},
		Audit:          AuditRepo{client: s.client},
	}
}

// RunInTx 在事务内执行 fn，fn 内使用事务绑定的仓储集合。
func (s *Store) RunInTx(ctx context.Context, fn func(tx *application.RepoUnit) error) error {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return err
	}
	txClient := tx.Client()
	txStore := NewStore(txClient)
	if err := fn(txStore.Repos()); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// ---- 数据隔离辅助 ----

// scopeMerchant 对商家查询应用数据隔离。
func scopeMerchant(q *ent.MerchantQuery, sc tenant.Scope) {
	switch {
	case sc.MerchantOnly():
		q.Where(merchant.ID(sc.MerchantID))
	case sc.RegionOnly():
		q.Where(merchant.RegionID(sc.RegionID))
	}
}

// scopeRider 对骑手查询应用数据隔离。
func scopeRider(q *ent.RiderQuery, sc tenant.Scope) {
	switch {
	case sc.RiderOnly():
		q.Where(rider.ID(sc.RiderID))
	case sc.RegionOnly():
		q.Where(rider.RegionID(sc.RegionID))
	}
}

// scopeOrder 对订单查询应用数据隔离。
func scopeOrder(q *ent.OrderQuery, sc tenant.Scope) {
	switch {
	case sc.UserOnly():
		q.Where(order.UserID(sc.UserID))
	case sc.RiderOnly():
		q.Where(order.RiderID(sc.RiderID))
	case sc.MerchantOnly():
		q.Where(order.MerchantID(sc.MerchantID))
	case sc.RegionOnly():
		q.Where(order.RegionID(sc.RegionID))
	}
}

// scopeUser 对用户查询应用数据隔离（区域站长仅见本区域用户）。
func scopeUser(q *ent.UserQuery, sc tenant.Scope) {
	if sc.RegionOnly() {
		q.Where(user.RegionID(sc.RegionID))
	}
}

// NotFound 通用未找到错误标记（对齐 application.ErrNotFound）。
var ErrNotFound = application.ErrNotFound

// scanSelect 具备 Scan 能力的聚合选择器（*OrderSelect/*WithdrawalSelect 等）。
type scanSelect interface {
	Scan(ctx context.Context, dest any) error
}

// aggQuery 具备 Aggregate 能力的生成查询（Order/Withdrawal/Settlement 等）。
type aggQuery[T scanSelect] interface {
	Aggregate(fns ...ent.AggregateFunc) T
}

// sumFloat64 对生成查询做 SUM 聚合；空集（SUM 为 NULL）时返回 0，
// 避免 "converting NULL to float64" 扫描错误。
func sumFloat64[T scanSelect, Q aggQuery[T]](ctx context.Context, q Q, field string) (float64, error) {
	var rows []struct{ Sum sql.NullFloat64 }
	if err := q.Aggregate(ent.As(ent.Sum(field), "sum")).Scan(ctx, &rows); err != nil {
		return 0, err
	}
	if len(rows) == 0 || !rows[0].Sum.Valid {
		return 0, nil
	}
	return rows[0].Sum.Float64, nil
}
