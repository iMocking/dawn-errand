// Package metrics 提供 Prometheus 指标。
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal HTTP 请求计数（method/path/status）。
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dawn_http_requests_total",
		Help: "HTTP requests total",
	}, []string{"method", "path", "status"})
	// HTTPRequestDuration HTTP 请求耗时直方图。
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "dawn_http_request_duration_seconds",
		Help:    "HTTP request duration seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
	// AuthFailuresTotal 认证失败次数。
	AuthFailuresTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_auth_failures_total",
		Help: "Auth failures total",
	})
	// EmailsSentTotal 邮件发送次数。
	EmailsSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_emails_sent_total",
		Help: "Emails sent total",
	})
	// ConfigChangesTotal 配置变更次数。
	ConfigChangesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dawn_config_changes_total",
		Help: "Config changes total",
	}, []string{"key"})
	// LocationReportsTotal 骑手位置上报次数。
	LocationReportsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_location_reports_total",
		Help: "Rider location reports total",
	})
	// TrackRequestsTotal 配送追踪请求次数。
	TrackRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_track_requests_total",
		Help: "Delivery track requests total",
	})
	// OrdersTimedOutTotal 超时取消订单数。
	OrdersTimedOutTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_orders_timed_out_total",
		Help: "Orders canceled by timeout",
	})
	// OrdersAutoConfirmedTotal 自动确认收货数。
	OrdersAutoConfirmedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_orders_auto_confirmed_total",
		Help: "Orders auto confirmed after delivery",
	})
	// JobRunsTotal 定时任务执行次数。
	JobRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dawn_job_runs_total",
		Help: "Background job runs total",
	}, []string{"job", "result"})
	// OrdersCreatedTotal 下单数。
	OrdersCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_orders_created_total",
		Help: "Orders created total",
	})
	// OrdersPaidTotal 支付成功数。
	OrdersPaidTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_orders_paid_total",
		Help: "Orders paid total",
	})
	// OrdersCompletedTotal 订单完成数。
	OrdersCompletedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_orders_completed_total",
		Help: "Orders completed total",
	})
	// RefundsTotal 退款完成数。
	RefundsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_refunds_total",
		Help: "Refunds completed total",
	})
	// WithdrawalsTotal 提现申请数。
	WithdrawalsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "dawn_withdrawals_total",
		Help: "Withdrawal requests total",
	})
)
