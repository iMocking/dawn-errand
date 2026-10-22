// Package storage 对象存储实现：本地磁盘存储（可切换 OSS）。
package storage

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dawn-errand/internal/domain/system"
)

// Local 本地存储实现 application.StoragePort。
type Local struct {
	cfg func() *system.LocalStorageConfig
}

// NewLocal 构建本地存储。
func NewLocal(cfg func() *system.LocalStorageConfig) *Local {
	return &Local{cfg: cfg}
}

func (l *Local) config() (*system.LocalStorageConfig, error) {
	c := l.cfg()
	if c == nil || c.RootDir == "" {
		return nil, fmt.Errorf("本地存储未配置")
	}
	if len(c.AllowedMIME) == 0 {
		c.AllowedMIME = []string{"image/jpeg", "image/png", "image/webp", "image/gif"}
	}
	if c.MaxSizeMB <= 0 {
		c.MaxSizeMB = 10
	}
	return c, nil
}

// Put 保存文件。校验大小与 MIME，防路径穿越。
func (l *Local) Put(ctx context.Context, data []byte, filename string, contentType string) (string, error) {
	c, err := l.config()
	if err != nil {
		return "", err
	}
	if int64(len(data)) > c.MaxSizeMB*1024*1024 {
		return "", fmt.Errorf("文件超过大小限制 %dMB", c.MaxSizeMB)
	}
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(filename))
	}
	if !l.allowedMIME(contentType) {
		return "", fmt.Errorf("不支持的文件类型 %s", contentType)
	}
	base := filepath.Base(filename) // 防路径穿越
	ext := filepath.Ext(base)
	if ext == "" {
		ext = ".bin"
	}
	day := time.Now().UTC().Format("20060102")
	name := fmt.Sprintf("%s/%s-%d%s", day, strings.TrimSuffix(base, ext), time.Now().UnixNano(), ext)
	root, err := filepath.Abs(c.RootDir)
	if err != nil {
		return "", err
	}
	full := filepath.Join(root, c.PathPrefix, name)
	// 防穿越：确保最终路径仍在 root 内
	if !strings.HasPrefix(filepath.Clean(full), filepath.Clean(root)+string(os.PathSeparator)) {
		return "", fmt.Errorf("非法路径")
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return "", err
	}
	publicBase := c.PublicBaseURL
	if publicBase == "" {
		publicBase = "http://localhost:8080"
	}
	return strings.TrimRight(publicBase, "/") + "/" + c.PathPrefix + "/" + name, nil
}

func (l *Local) allowedMIME(mt string) bool {
	c, _ := l.config()
	for _, a := range c.AllowedMIME {
		if strings.EqualFold(a, mt) {
			return true
		}
	}
	return false
}

// TestConnection 检查根目录可写。
func (l *Local) TestConnection(ctx context.Context) error {
	c, err := l.config()
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(c.RootDir, c.PathPrefix), 0o755)
}

// TestLocal 使用给定配置测试本地存储（配置管理页「测试连接」）。
func TestLocal(ctx context.Context, c *system.LocalStorageConfig) error {
	if c == nil || c.RootDir == "" {
		return fmt.Errorf("本地存储未配置")
	}
	return os.MkdirAll(filepath.Join(c.RootDir, c.PathPrefix), 0o755)
}
