package storage

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"dawn-errand/internal/domain/system"
)

func newTestLocal(t *testing.T) (*Local, string) {
	t.Helper()
	root := t.TempDir()
	l := NewLocal(func(context.Context) *system.LocalStorageConfig {
		return &system.LocalStorageConfig{
			RootDir: root, PathPrefix: "uploads",
			PublicBaseURL: "http://test.local:8080", MaxSizeMB: 1,
		}
	})
	return l, root
}

var pngBytes = []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("x", 128))

func TestPublicPutURLMatchesDisk(t *testing.T) {
	l, root := newTestLocal(t)
	url, err := l.Put(context.Background(), pngBytes, "photo.png", "image/png")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "http://test.local:8080/uploads/") {
		t.Fatalf("公开 URL 前缀异常: %s", url)
	}
	rel := strings.TrimPrefix(url, "http://test.local:8080/")
	disk := filepath.Join(root, filepath.FromSlash(rel))
	if _, err := os.Stat(disk); err != nil {
		t.Fatalf("URL 与磁盘路径不一致 %s: %v", rel, err)
	}
}

func TestPrivatePutReadRoundTrip(t *testing.T) {
	l, _ := newTestLocal(t)
	url, rel, err := l.PutPrivate(context.Background(), pngBytes, "merchant/M0001", "id.png", "image/png")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rel, "private/merchant/M0001") || !strings.HasSuffix(rel, ".png") {
		t.Fatalf("私有 rel 异常: %s", rel)
	}
	if !strings.HasPrefix(url, "http://test.local:8080/api/uploads/private/") {
		t.Fatalf("私有 URL 应指向鉴权路由: %s", url)
	}
	// URL 尾部直接是 <主体>/<主体ID>…：鉴权路由据此解析归属，多一层 private 会让本人也读不到。
	tail := strings.TrimPrefix(url, "http://test.local:8080/api/uploads/private/")
	if strings.HasPrefix(tail, "private/") || len(strings.Split(tail, "/")) < 2 {
		t.Fatalf("私有 URL 归属段异常: %s", tail)
	}
	data, mime, err := l.ReadPrivate(context.Background(), rel)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(pngBytes) || mime != "image/png" {
		t.Fatalf("回读不一致: len=%d mime=%s", len(data), mime)
	}
}

func TestPrivateRejectsTraversal(t *testing.T) {
	l, _ := newTestLocal(t)
	for _, bad := range []string{"../secret", "a/../../b", "..", "/abs/path", "ok/../../etc", ""} {
		if _, _, err := l.PutPrivate(context.Background(), pngBytes, bad, "x.png", "image/png"); err == nil {
			t.Fatalf("非法路径应被拒绝: %q", bad)
		}
	}
	if _, _, err := l.ReadPrivate(context.Background(), "private/../../outside.png"); err == nil {
		t.Fatal("读取穿越应被拒绝")
	}
}

func TestPrivateFileMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows 不实现 POSIX 权限位")
	}
	l, root := newTestLocal(t)
	_, rel, err := l.PutPrivate(context.Background(), pngBytes, "user/U0001", "a.png", "image/png")
	if err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm()&0o077 != 0 {
		t.Fatalf("私有文件权限过宽: %v", fi.Mode())
	}
}
