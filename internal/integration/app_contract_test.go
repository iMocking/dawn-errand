package integration

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// -update 重写三份基线：新增/删除端点、改字段名、动金额字段都必须显式跑一次，
// 这样契约变更只会以「基线 diff 进了代码评审」的方式发生，而不是悄悄漂走。
var updateContract = flag.Bool("update", false, "重写 App 面契约基线文件 testdata/app_*.txt")

// TestAppRouteContractBaseline 冻结 /api/app 面的端点集合（方法 + 路径）。
// 路由表取自 gin 运行时注册结果，因此基线不可能与代码不同步——不一致就是契约变了。
func TestAppRouteContractBaseline(t *testing.T) {
	setup()
	got := appRouteLines(router)
	path := filepath.Join("testdata", "app_routes.txt")
	if *updateContract {
		writeBaseline(t, path, got)
		return
	}
	want := readBaseline(t, path)
	if missing, extra := diffLines(want, got); len(missing) > 0 || len(extra) > 0 {
		t.Errorf("App 面端点集合与基线不一致（缺 %d / 多 %d）：\n%s",
			len(missing), len(extra), formatDiff(missing, extra))
	}
}

// TestAppResponseKeysAreSnakeCase 逐字段校验 App 面响应键名：外层四键固定，内层只允许 snake_case。
// 后台面（/api/*）用驼峰，三端 App 用下划线是既有口径，混用会让客户端反序列化静默丢字段。
func TestAppResponseKeysAreSnakeCase(t *testing.T) {
	for _, c := range appProbeCases(t) {
		t.Run(strings.TrimPrefix(c.path, "/api/app/"), func(t *testing.T) {
			var doc map[string]any
			if code := doRawJSON(t, c.method, c.path, c.token, &doc); code != 0 {
				t.Fatalf("%s %s code=%d", c.method, c.path, code)
			}
			var bad []string
			walkJSON(doc, "", func(key, path string) {
				if !snakeKeyRe.MatchString(key) {
					bad = append(bad, jpath(path, key))
				}
			})
			if len(bad) > 0 {
				t.Errorf("非 snake_case 键 %d 处：\n%s", len(bad), strings.Join(bad, "\n"))
			}
			assertEnvelope(t, doc)
		})
	}
}

// TestAppMoneyFieldBaseline 冻结金额字段清单，供 R2-6（元 float → 分 int）迁移核对。
// 清单是按真实响应走出来的，不是照文档抄的：漏一个字段就会在这条用例里显形。
func TestAppMoneyFieldBaseline(t *testing.T) {
	var lines []string
	for _, c := range appProbeCases(t) {
		var doc map[string]any
		if code := doRawJSON(t, c.method, c.path, c.token, &doc); code != 0 {
			t.Fatalf("%s %s code=%d", c.method, c.path, code)
		}
		seen := map[string]bool{}
		walkJSON(doc, "", func(key, path string) {
			if moneyKeyRe.MatchString(key) {
				seen[c.method+" "+c.label+" -> "+jpath(path, key)] = true
			}
		})
		for l := range seen {
			lines = append(lines, l)
		}
	}
	sort.Strings(lines)
	path := filepath.Join("testdata", "app_money_fields.txt")
	if *updateContract {
		writeBaseline(t, path, lines)
		return
	}
	want := readBaseline(t, path)
	if len(want) == 0 {
		t.Skip("基线为空，先跑 go test -run TestAppMoneyFieldBaseline -update")
	}
	if missing, extra := diffLines(want, lines); len(missing) > 0 || len(extra) > 0 {
		t.Errorf("金额字段清单与基线不一致（缺 %d / 多 %d）：\n%s",
			len(missing), len(extra), formatDiff(missing, extra))
	}
}

// ---- 辅助 ----

var (
	snakeKeyRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	// 金额字段特征名；刻意不含 total——那是分页总数，混进来会让 R2-6 迁移清单误报。
	moneyKeyRe = regexp.MustCompile(`(^|_)(amount|balance|fee|price|income|cost|money)(_|$)`)
)

type appProbe struct {
	method, path, token string
	// label 是契约里的路由形状（带路径参数的用 :id 占位），基线按它排序；
	// path 是本次真实请求的 URL，含具体 id，不能进基线否则每次都不一致。
	label string
}

// appProbeCases 三端各取一组只读端点：覆盖身份、目录、订单、资金、消息五类响应形状。
func appProbeCases(t *testing.T) []appProbe {
	t.Helper()
	clientTok, _ := loginAs(t, "client.test@dawnerrand.local", "Client123456")
	storeTok, _ := loginAs(t, "store.owner@dawnerrand.local", "Store123456")
	driverTok, _ := loginAs(t, "rider.owner@dawnerrand.local", "Rider123456")
	routes := []struct{ path, tok string }{
		{"/api/app/client/me", clientTok}, {"/api/app/client/regions", clientTok},
		{"/api/app/client/merchants", clientTok}, {"/api/app/client/addresses", clientTok},
		{"/api/app/client/orders", clientTok}, {"/api/app/client/coupons", clientTok},
		{"/api/app/store/me", storeTok}, {"/api/app/store/business", storeTok},
		{"/api/app/store/menu", storeTok}, {"/api/app/store/orders", storeTok},
		{"/api/app/store/reviews", storeTok}, {"/api/app/store/finance/summary", storeTok},
		{"/api/app/store/finance/records", storeTok}, {"/api/app/store/settlements", storeTok},
		{"/api/app/store/withdrawals", storeTok},
		{"/api/app/driver/me", driverTok}, {"/api/app/driver/order-pool", driverTok},
		{"/api/app/driver/orders", driverTok}, {"/api/app/driver/finance/summary", driverTok},
		{"/api/app/driver/finance/records", driverTok}, {"/api/app/driver/settlements", driverTok},
		{"/api/app/driver/withdrawals", driverTok},
		{"/api/app/common/notifications", clientTok},
	}
	out := make([]appProbe, 0, len(routes))
	for _, rt := range routes {
		out = append(out, appProbe{method: http.MethodGet, path: rt.path, token: rt.tok, label: rt.path})
	}
	// 金额字段主要长在订单详情响应里（pay_amount / goods_amount / delivery_fee / 骑手收入），
	// 种子账号名下没有单，故现造一单推进到抢单后，把三端详情纳入探针；不探详情，
	// R2-6 的迁移清单就会缺一整块。
	out = append(out, orderDetailProbes(t, clientTok, storeTok, driverTok)...)
	return out
}

// orderDetailProbes 造一单并推进「支付→接单→备餐→抢单」，返回三端订单详情探针。
// 每个用例各自造单、各自清理：缓存探针会让先跑的用例把单删掉，后跑的拿不到 id 只能报 1507。
// 某一步失败就不追加该端探针——用例不该因为流程走不到那一步而红。
func orderDetailProbes(t *testing.T, clientTok, storeTok, driverTok string) []appProbe {
	t.Helper()
	var out []appProbe
	id := createOrderHTTP(t, clientTok, fmt.Sprintf("contract-%d", time.Now().UnixNano()),
		[]map[string]any{{"dish_id": "D001", "count": 1}})
	t.Cleanup(func() { cleanupOrder(t, id) })
	out = append(out, appProbe{http.MethodGet, "/api/app/client/orders/" + id, clientTok, "/api/app/client/orders/:id"})
	if post(t, "/api/app/client/orders/"+id+"/pay", clientTok, map[string]string{"channel": "mock"}) &&
		post(t, "/api/app/store/orders/"+id+"/accept", storeTok, nil) &&
		post(t, "/api/app/store/orders/"+id+"/ready", storeTok, nil) {
		out = append(out, appProbe{http.MethodGet, "/api/app/store/orders/" + id, storeTok, "/api/app/store/orders/:id"})
		if post(t, "/api/app/driver/orders/"+id+"/grab", driverTok, nil) {
			out = append(out, appProbe{http.MethodGet, "/api/app/driver/orders/" + id, driverTok, "/api/app/driver/orders/:id"})
		}
	}
	return out
}

func post(t *testing.T, path, token string, payload any) bool {
	t.Helper()
	r, _ := do(t, http.MethodPost, path, token, payload)
	return r.Code == 0
}

// doRawJSON 直接拿原始响应体（do() 会解码进固定结构体，丢掉契约外的键，无法用于键名校验）。
func doRawJSON(t *testing.T, method, path, token string, out *map[string]any) int {
	t.Helper()
	setup()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "itest-contract")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var doc map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatalf("%s %s 响应不是 JSON 对象: %v（body=%s）", method, path, err, w.Body.String())
	}
	*out = doc
	code, _ := doc["code"].(float64)
	return int(code)
}

// assertEnvelope 外层键必须恰好是 code/message/body/request_id 四个。
func assertEnvelope(t *testing.T, doc map[string]any) {
	t.Helper()
	want := map[string]bool{"code": true, "message": true, "body": true, "request_id": true}
	for k := range doc {
		if !want[k] {
			t.Errorf("响应外层多出键 %q（契约只允许 code/message/body/request_id）", k)
		}
	}
	for k := range want {
		if _, ok := doc[k]; !ok {
			t.Errorf("响应外层缺键 %q", k)
		}
	}
}

// walkJSON 递归遍历，对每个对象键回调 (key, 父路径)；数组用 [] 表示以避开下标抖动。
func walkJSON(v any, path string, onKey func(key, parent string)) {
	switch t := v.(type) {
	case map[string]any:
		for k, sub := range t {
			onKey(k, path)
			walkJSON(sub, jpath(path, k), onKey)
		}
	case []any:
		for _, sub := range t {
			walkJSON(sub, path+"[]", onKey)
		}
	}
}

func jpath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

func appRouteLines(r *gin.Engine) []string {
	var out []string
	for _, rt := range r.Routes() {
		if strings.HasPrefix(rt.Path, "/api/app") {
			out = append(out, rt.Method+" "+rt.Path)
		}
	}
	sort.Strings(out)
	return out
}

func writeBaseline(t *testing.T, path string, lines []string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("已重写 %s（%d 行）", path, len(lines))
}

func readBaseline(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读基线 %s: %v（首次生成请带 -update）", path, err)
	}
	var out []string
	for _, l := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, strings.TrimRight(l, "\r"))
		}
	}
	return out
}

func diffLines(want, got []string) (missing, extra []string) {
	w := map[string]bool{}
	g := map[string]bool{}
	for _, l := range want {
		w[l] = true
	}
	for _, l := range got {
		g[l] = true
	}
	for _, l := range want {
		if !g[l] {
			missing = append(missing, l)
		}
	}
	for _, l := range got {
		if !w[l] {
			extra = append(extra, l)
		}
	}
	return
}

func formatDiff(missing, extra []string) string {
	var b strings.Builder
	for _, l := range missing {
		b.WriteString("  - 基线有而代码没有: " + l + "\n")
	}
	for _, l := range extra {
		b.WriteString("  + 代码新增而基线没有: " + l + "\n")
	}
	return b.String()
}
