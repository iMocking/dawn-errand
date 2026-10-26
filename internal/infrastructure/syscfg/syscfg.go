// Package syscfg 运行期配置的基础设施适配：把「数据库里的系统设置」翻译成各组件
// 需要的配置闭包，并集中健康检查探针。装配层（internal/assembly）只负责连线与注入。
package syscfg

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"

	"dawn-errand/internal/application"
	"dawn-errand/internal/config"
	sysdomain "dawn-errand/internal/domain/system"
	"dawn-errand/internal/infrastructure/crypto"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/repo"
)

// SMTP 读取已保存的 SMTP 配置（解密密码；未配置返回空配置）。
func SMTP(cipher *crypto.Cipher, store *repo.Store) func(context.Context) (*sysdomain.SMTPConfig, error) {
	return func(ctx context.Context) (*sysdomain.SMTPConfig, error) {
		rec, err := store.Repos().Settings.Get(ctx, sysdomain.KeySMTP)
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				return &sysdomain.SMTPConfig{}, nil
			}
			return nil, err
		}
		var cfg sysdomain.SMTPConfig
		if err := json.Unmarshal([]byte(rec.ValueJSON), &cfg); err != nil {
			return nil, err
		}
		if cfg.Password != "" && cfg.Password != "******" {
			if plain, err := cipher.Decrypt(cfg.Password); err == nil {
				cfg.Password = plain
			}
		}
		return &cfg, nil
	}
}

// LocalStorage 读取已保存的本地存储配置；未配置或损坏时回退到启动配置。
func LocalStorage(store *repo.Store, cfg *config.Config) func(context.Context) *sysdomain.LocalStorageConfig {
	return func(ctx context.Context) *sysdomain.LocalStorageConfig {
		def := &sysdomain.LocalStorageConfig{
			RootDir:       cfg.StorageRootDir,
			PublicBaseURL: cfg.StoragePublicURL,
			MaxSizeMB:     cfg.StorageMaxSizeMB,
		}
		rec, err := store.Repos().Settings.Get(ctx, sysdomain.KeyLocal)
		if err != nil {
			return def
		}
		var c sysdomain.LocalStorageConfig
		if err := json.Unmarshal([]byte(rec.ValueJSON), &c); err != nil {
			return def
		}
		if c.RootDir == "" {
			c.RootDir = def.RootDir
		}
		if c.PublicBaseURL == "" {
			c.PublicBaseURL = def.PublicBaseURL
		}
		if c.MaxSizeMB <= 0 {
			c.MaxSizeMB = def.MaxSizeMB
		}
		return &c
	}
}

// Outbox dev 邮件发件箱查询（最近 limit 封，按创建时间倒序）。
func Outbox(client *ent.Client) func(ctx context.Context, limit int) ([]application.MailEntry, error) {
	return func(ctx context.Context, limit int) ([]application.MailEntry, error) {
		if limit < 1 || limit > 200 {
			limit = 50
		}
		rows, err := client.MailOutbox.Query().Order(ent.Desc("created_at")).Limit(limit).All(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]application.MailEntry, 0, len(rows))
		for _, r := range rows {
			out = append(out, application.MailEntry{
				ID: r.ID, ToEmail: r.ToEmail, Subject: r.Subject, Body: r.Body, CreatedAt: r.CreatedAt,
			})
		}
		return out, nil
	}
}

// Ready 就绪探针：数据库与 Redis 都要真的能读写一次才算就绪。
func Ready(client *ent.Client, rdb *redis.Client) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if _, err := client.User.Query().Limit(1).Exist(ctx); err != nil {
			return err
		}
		return rdb.Ping(ctx).Err()
	}
}
