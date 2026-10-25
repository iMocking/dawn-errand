// Package storage 对象存储实现：本地磁盘存储（可切换 OSS）。
package storage

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"

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
		c.AllowedMIME = []string{"image/jpeg", "image/png", "image/webp", "image/gif", "application/pdf"}
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
	contentType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	if !l.allowedMIME(contentType) {
		return "", fmt.Errorf("不支持的文件类型 %s", contentType)
	}
	// 扩展名由 MIME 决定，忽略客户端文件名后缀，避免「图片内容 + .html 后缀」造成存储型 XSS。
	ext := extForMIME(contentType)
	if ext == "" {
		return "", fmt.Errorf("不支持的文件类型 %s", contentType)
	}
	base := filepath.Base(filename) // 防路径穿越
	stem := sanitizeStem(strings.TrimSuffix(base, filepath.Ext(base)))
	if stem == "" {
		stem = "file"
	}
	day := time.Now().UTC().Format("20060102")
	// 磁盘相对路径只含日期/文件名；PathPrefix 仅作为对外 URL 挂载前缀（默认 uploads），
	// 由路由把该前缀挂载到存储根目录，避免出现 uploads/uploads 双层目录。
	rel := path.Join(day, fmt.Sprintf("%s-%d%s", stem, time.Now().UnixNano(), ext))
	prefix := strings.Trim(c.PathPrefix, "/")
	if prefix == "" {
		prefix = "uploads"
	}
	root, err := filepath.Abs(c.RootDir)
	if err != nil {
		return "", err
	}
	full := filepath.Join(root, filepath.FromSlash(rel))
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
	return strings.TrimRight(publicBase, "/") + "/" + prefix + "/" + rel, nil
}

func (l *Local) allowedMIME(mt string) bool {
	c, _ := l.config()
	for _, a := range c.AllowedMIME {
		if strings.EqualFold(strings.TrimSpace(strings.Split(a, ";")[0]), mt) {
			return true
		}
	}
	return false
}

// extForMIME 返回安全的落盘扩展名（只允许白名单类型）。
func extForMIME(mt string) string {
	switch strings.ToLower(mt) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "application/pdf":
		return ".pdf"
	}
	return ""
}

// sanitizeStem 仅保留字母/数字/中文与 -_，其余替换为 _，并限制长度。
func sanitizeStem(s string) string {
	var b strings.Builder
	count := 0
	for _, r := range s {
		if count >= 40 {
			break
		}
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
		count++
	}
	return strings.Trim(b.String(), "_")
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
