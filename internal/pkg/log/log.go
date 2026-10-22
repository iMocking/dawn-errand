// Package log 提供 zap JSON 日志。
package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 构建 JSON 日志。dev 环境输出到控制台便于本地排查，结构仍为 JSON。
func New(env string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	if env == "dev" {
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	return cfg.Build()
}
