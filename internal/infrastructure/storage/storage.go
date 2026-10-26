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
	cfg func(context.Context) *system.LocalStorageConfig
}

// NewLocal 构建本地存储。
func NewLocal(cfg func(context.Context) *system.LocalStorageConfig) *Local {
	return &Local{cfg: cfg}
}

func (l *Local) config(ctx context.Context) (*system.LocalStorageConfig, error) {
	c := l.cfg(ctx)
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

// Put 保存文件到公开桶。校验大小与 MIME，防路径穿越。
func (l *Local) Put(ctx context.Context, data []byte, filename string, contentType string) (string, error) {
	c, ext, err := l.validate(ctx, data, filename, contentType)
	if err != nil {
		return "", err
	}
	base := filepath.Base(filename) // 防路径穿越
	stem := sanitizeStem(strings.TrimSuffix(base, filepath.Ext(base)))
	if stem == "" {
		stem = "file"
	}
	day := time.Now().UTC().Format("20060102")
	// 磁盘落在 <root>/<prefix>/日期/文件名，与路由静态挂载 <root>/<prefix> 一一对应；
	// 私有桶固定落在 <root>/private，不在任何静态挂载内。
	prefix := l.urlPrefix(c)
	rel := path.Join(prefix, day, fmt.Sprintf("%s-%d%s", stem, time.Now().UnixNano(), ext))
	full, err := l.safePath(c.RootDir, rel)
	if err != nil {
		return "", err
	}
	if err := writeObject(full, data); err != nil {
		return "", err
	}
	// URL 与磁盘相对路径一致。
	return strings.TrimRight(l.publicBase(c), "/") + "/" + rel, nil
}

// PutPrivate 保存敏感对象（证件等）到私有桶：不在静态挂载内，仅经鉴权路由读取。
func (l *Local) PutPrivate(ctx context.Context, data []byte, rel, filename, contentType string) (string, string, error) {
	c, ext, err := l.validate(ctx, data, filename, contentType)
	if err != nil {
		return "", "", err
	}
	rel, err = sanitizePrivateRel(rel)
	if err != nil {
		return "", "", err
	}
	full, err := l.safePath(c.RootDir, path.Join("private", rel+ext))
	if err != nil {
		return "", "", err
	}
	if err := writeObject(full, data); err != nil {
		return "", "", err
	}
	relOut := "private/" + rel + ext
	base := strings.TrimRight(l.publicBase(c), "/")
	// URL 不带 private/ 段：鉴权路由按 <主体>/<主体ID>/<对象名> 解析归属前缀，
	// 多一层 private 会让归属校验把首段读成 "private" 而一律拒绝。
	return base + "/api/uploads/private/" + rel + ext, relOut, nil
}

// ReadPrivate 按相对路径读取私有桶对象（拒绝穿越与越界）。
func (l *Local) ReadPrivate(ctx context.Context, rel string) ([]byte, string, error) {
	c, err := l.config(ctx)
	if err != nil {
		return nil, "", err
	}
	rel, err = sanitizePrivateRel(strings.TrimPrefix(strings.TrimSpace(filepath.ToSlash(rel)), "private/"))
	if err != nil {
		return nil, "", err
	}
	full, err := l.safePath(c.RootDir, path.Join("private", rel))
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("文件不存在")
		}
		return nil, "", err
	}
	return data, mime.TypeByExtension(filepath.Ext(full)), nil
}

// validate 公共校验：大小 + MIME + 落盘扩展名白名单。
func (l *Local) validate(ctx context.Context, data []byte, filename, contentType string) (*system.LocalStorageConfig, string, error) {
	c, err := l.config(ctx)
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > c.MaxSizeMB*1024*1024 {
		return nil, "", fmt.Errorf("文件超过大小限制 %dMB", c.MaxSizeMB)
	}
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(filename))
	}
	contentType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	if !l.allowedMIME(ctx, contentType) {
		return nil, "", fmt.Errorf("不支持的文件类型 %s", contentType)
	}
	// 扩展名由 MIME 决定，忽略客户端文件名后缀，避免「图片内容 + .html 后缀」造成存储型 XSS。
	ext := extForMIME(contentType)
	if ext == "" {
		return nil, "", fmt.Errorf("不支持的文件类型 %s", contentType)
	}
	return c, ext, nil
}

// safePath 拼接 root+rel 并断言清洗后仍在 root 内。
func (l *Local) safePath(rootDir, rel string) (string, error) {
	root, err := filepath.Abs(rootDir)
	if err != nil {
		return "", err
	}
	full := filepath.Join(root, filepath.FromSlash(rel))
	if !strings.HasPrefix(filepath.Clean(full), filepath.Clean(root)+string(os.PathSeparator)) {
		return "", fmt.Errorf("非法路径")
	}
	return full, nil
}

func (l *Local) urlPrefix(c *system.LocalStorageConfig) string {
	prefix := strings.Trim(c.PathPrefix, "/")
	if prefix == "" {
		prefix = "uploads"
	}
	return prefix
}

func (l *Local) publicBase(c *system.LocalStorageConfig) string {
	if c.PublicBaseURL == "" {
		return "http://localhost:8080"
	}
	return c.PublicBaseURL
}

func writeObject(full string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o600)
}

// sanitizePrivateRel 私有对象相对路径：仅字母数字/._-与子目录段，拒绝 .. 与绝对路径。
func sanitizePrivateRel(rel string) (string, error) {
	rel = strings.TrimSpace(filepath.ToSlash(rel))
	if strings.HasPrefix(rel, "/") {
		return "", fmt.Errorf("非法路径")
	}
	rel = strings.Trim(rel, "/")
	if rel == "" {
		return "", fmt.Errorf("非法路径")
	}
	for _, seg := range strings.Split(rel, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return "", fmt.Errorf("非法路径")
		}
		for _, r := range seg {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-') {
				return "", fmt.Errorf("非法路径")
			}
		}
	}
	return rel, nil
}

func (l *Local) allowedMIME(ctx context.Context, mt string) bool {
	c, _ := l.config(ctx)
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
	c, err := l.config(ctx)
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
