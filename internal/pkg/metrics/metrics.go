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
)
