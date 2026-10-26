package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testPNG 1x1 像素图片，上传用例的最小合法载荷。
var testPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
	0x0D, 0x49, 0x44, 0x41, 0x54, 0x78, 0xDA, 0x63, 0xFC, 0xCF, 0xC0, 0x50,
	0x0F, 0x00, 0x05, 0x83, 0x02, 0x7F, 0x95, 0x8D, 0xA6, 0x00, 0x00, 0x00,
	0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

// getStatus 请求并只取状态码（私有桶返回二进制，不能按 JSON 信封解析）。
func getStatus(t *testing.T, path, token string) int {
	t.Helper()
	setup()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Code
}

// TestUploadPurposeWhitelist R4-7：上传用途白名单。
// 未知 purpose 必须拒绝——若静默落公开桶，身份证一类 PII 会被静态挂载无鉴权读取；
// 证件类 purpose 走私有桶，图片类走公开桶。
func TestUploadPurposeWhitelist(t *testing.T) {
	owner := mustRegisterAndLogin(t, "upload-guard-a@shuguang.local", "UploadPass123")

	unknown := doMultipart(t, "/api/app/common/uploads", owner.Token, "file", "a.png", testPNG, "purpose", "identity_pic")
	if unknown.Code != 1301 || !strings.Contains(unknown.Message, "purpose") {
		t.Fatalf("未知 purpose 应 1301, got %d msg=%s", unknown.Code, unknown.Message)
	}

	doc := doMultipart(t, "/api/app/common/uploads", owner.Token, "file", "idcard.png", testPNG, "purpose", "doc")
	if doc.Code != 0 {
		t.Fatalf("证件上传 code=%d msg=%s", doc.Code, doc.Message)
	}
	var docBody struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(doc.Body, &docBody); err != nil {
		t.Fatal(err)
	}
	const marker = "/api/uploads/private/"
	if !strings.Contains(docBody.URL, marker+"user/") {
		t.Fatalf("证件应落私有桶 user 前缀, got %s", docBody.URL)
	}
	defer func() {
		rel := docBody.URL[strings.Index(docBody.URL, marker)+len(marker):]
		_ = os.Remove(filepath.Join(handler.Cfg.StorageRootDir, "private", filepath.FromSlash(rel)))
	}()

	img := doMultipart(t, "/api/app/common/uploads", owner.Token, "file", "avatar.png", testPNG, "purpose", "avatar")
	if img.Code != 0 {
		t.Fatalf("图片上传 code=%d msg=%s", img.Code, img.Message)
	}
	var imgBody struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(img.Body, &imgBody); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(imgBody.URL, marker) {
		t.Fatalf("图片用途不应进入私有桶, got %s", imgBody.URL)
	}
	defer func() {
		path := strings.TrimPrefix(imgBody.URL, "http://localhost:8080")
		_ = os.Remove(filepath.Join(handler.Cfg.StorageRootDir, filepath.FromSlash(strings.TrimPrefix(path, "/uploads/"))))
	}()

	// 契约文档历史字段名 scene 仍可用：cert 属证件 → 私有桶。
	scened := doMultipart(t, "/api/app/common/uploads", owner.Token, "file", "cert.png", testPNG, "scene", "cert")
	if scened.Code != 0 {
		t.Fatalf("scene=cert 上传 code=%d msg=%s", scened.Code, scened.Message)
	}
	var scenedBody struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(scened.Body, &scenedBody); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(scenedBody.URL, marker) {
		t.Fatalf("scene=cert 应落私有桶, got %s", scenedBody.URL)
	}
	// URL 尾部直接是归属前缀（user/<uid>/…），不得再多一层 private，否则本人也读不到。
	tail := scenedBody.URL[strings.Index(scenedBody.URL, marker)+len(marker):]
	if strings.HasPrefix(tail, "private/") {
		t.Fatalf("私有 URL 多出一层 private: %s", tail)
	}
	_ = os.Remove(filepath.Join(handler.Cfg.StorageRootDir, "private", filepath.FromSlash(tail)))
}

// TestPrivateUploadOwnership R4-7：私有桶读取归属。
// 普通用户仅本人；平台全局角色（审核需要）可读；未绑定该主体的其他用户拒绝。
func TestPrivateUploadOwnership(t *testing.T) {
	owner := mustRegisterAndLogin(t, "upload-owner@shuguang.local", "UploadPass123")
	stranger := mustRegisterAndLogin(t, "upload-stranger@shuguang.local", "UploadPass123")

	doc := doMultipart(t, "/api/app/common/uploads", owner.Token, "file", "id.png", testPNG, "purpose", "idcard")
	if doc.Code != 0 {
		t.Fatalf("上传 code=%d msg=%s", doc.Code, doc.Message)
	}
	var body struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(doc.Body, &body); err != nil {
		t.Fatal(err)
	}
	const marker = "/api/uploads/private/"
	rel := body.URL[strings.Index(body.URL, marker)+len(marker):]
	defer func() {
		_ = os.Remove(filepath.Join(handler.Cfg.StorageRootDir, "private", filepath.FromSlash(rel)))
	}()

	if code := getStatus(t, marker+rel, owner.Token); code != http.StatusOK {
		t.Fatalf("本人读取私有文件应 200, got %d", code)
	}
	if code := getStatus(t, marker+rel, stranger.Token); code != http.StatusForbidden {
		t.Fatalf("他人读取私有文件应 403, got %d", code)
	}
	if code := getStatus(t, marker+rel, loginAdmin(t)); code != http.StatusOK {
		t.Fatalf("平台管理员读取私有文件应 200, got %d", code)
	}
	if code := getStatus(t, marker+rel, ""); code != http.StatusUnauthorized {
		t.Fatalf("未认证读取私有文件应 401, got %d", code)
	}
}

// TestAdminUploadRouteRoleGuard R4-7：管理端上传路由仅限管理角色。
// 三端 App 走 /api/app/common/uploads，普通用户不应借 /api/uploads 写存储。
func TestAdminUploadRouteRoleGuard(t *testing.T) {
	user := mustRegisterAndLogin(t, "upload-role@shuguang.local", "UploadPass123")
	r := doMultipart(t, "/api/uploads", user.Token, "file", "x.png", testPNG)
	if r.Code != 1203 {
		t.Fatalf("普通用户走管理端上传路由应 1203, got %d msg=%s", r.Code, r.Message)
	}
	admin := loginAdmin(t)
	ok := doMultipart(t, "/api/uploads", admin, "file", "admin.png", testPNG)
	if ok.Code != 0 {
		t.Fatalf("管理员上传 code=%d msg=%s", ok.Code, ok.Message)
	}
	var body struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(ok.Body, &body); err != nil {
		t.Fatal(err)
	}
	path := strings.TrimPrefix(body.URL, "http://localhost:8080")
	_ = os.Remove(filepath.Join(handler.Cfg.StorageRootDir, filepath.FromSlash(strings.TrimPrefix(path, "/uploads/"))))
}
