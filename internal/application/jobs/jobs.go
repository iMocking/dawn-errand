// Package jobs 后台定时任务：订单超时、自动确认收货、公告发布。
package jobs

import (
	"context"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/application/app"
	"dawn-errand/internal/pkg/metrics"
)

const lockKey = "jobs:runner"

// Runner 定时任务执行器。
type Runner struct {
	Repos       *application.RepoUnit
	Orders      *app.OrderService
	Locker      application.Locker
	PayTimeout  time.Duration
	AutoConfirm time.Duration
	BatchLimit  int
}

// Result 单次执行结果。
type Result struct {
	Canceled  int
	Confirmed int
	Published int
	Skipped   bool
}

// RunOnce 执行一轮任务（幂等；多实例通过 Redis 锁保证单实例）。
func (r *Runner) RunOnce(ctx context.Context) Result {
	var res Result
	if r.Locker != nil {
		ok, err := r.Locker.TryLock(ctx, lockKey, 50*time.Second)
		if err != nil || !ok {
			res.Skipped = true
			return res
		}
		defer func() { _ = r.Locker.Unlock(ctx, lockKey) }()
	}
	now := time.Now().Unix()
	limit := r.BatchLimit
	if limit <= 0 {
		limit = 100
	}

	payTimeout := r.PayTimeout
	if payTimeout <= 0 {
		payTimeout = 15 * time.Minute
	}
	pending, err := r.Repos.Orders.ListPendingPayBefore(now-int64(payTimeout.Seconds()), limit)
	if err == nil {
		for _, o := range pending {
			if _, err := r.Orders.CancelSystem(ctx, o.ID, "超时未支付，系统自动取消"); err == nil {
				res.Canceled++
				metrics.OrdersTimedOutTotal.Inc()
			}
		}
	}

	autoConfirm := r.AutoConfirm
	if autoConfirm <= 0 {
		autoConfirm = 24 * time.Hour
	}
	delivered, err := r.Repos.Orders.ListDeliveredBefore(now-int64(autoConfirm.Seconds()), limit)
	if err == nil {
		for _, o := range delivered {
			if _, err := r.Orders.AutoConfirm(ctx, o.ID); err == nil {
				res.Confirmed++
				metrics.OrdersAutoConfirmedTotal.Inc()
			}
		}
	}

	due, err := r.Repos.Announcements.ListDueScheduled(now)
	if err == nil {
		for _, a := range due {
			if err := r.Repos.Announcements.MarkPublished(a.ID, now); err == nil {
				res.Published++
			}
		}
	}

	// 数据清理：过期令牌、历史通知、陈旧位置、过期券。
	_ = r.Repos.EmailTokens.DeleteExpired(now)
	_ = r.Repos.PasswordTokens.DeleteExpired(now)
	_ = r.Repos.Notifications.DeleteBefore(now - 180*86400)
	_ = r.Repos.Locations.DeleteStale(now - 7*86400)
	_ = r.Repos.Coupons.ExpireAll(now)
	metrics.JobRunsTotal.WithLabelValues("cleanup", "done").Inc()

	if res.Canceled > 0 || res.Confirmed > 0 || res.Published > 0 {
		metrics.JobRunsTotal.WithLabelValues("order_timeout", "done").Inc()
	}
	return res
}

// Start 启动周期任务，返回停止函数。
func (r *Runner) Start(ctx context.Context, interval time.Duration) func() {
	if interval <= 0 {
		interval = time.Minute
	}
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		// 启动后先跑一轮，避免重启后长时间不处理积压。
		_ = r.RunOnce(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				_ = r.RunOnce(ctx)
			}
		}
	}()
	return func() { close(stop) }
}
