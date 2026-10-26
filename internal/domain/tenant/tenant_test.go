package tenant

import "testing"

// R3-5 回归：Recognized 是仓储层「默认拒绝」的判据，新增角色必须显式归类，
// 否则会以「未识别」身份落到拒绝分支（安全侧失败），而不是静默全局可见。
func TestScopeRecognized(t *testing.T) {
	cases := []struct {
		role string
		want bool
	}{
		{"super_admin", true},
		{"cs_manager", true},
		{"cs_agent", true},
		{"region_station", true},
		{"merchant", true},
		{"driver", true},
		{"user", true},
		{"system", false},
		{"", false},
		{"root", false},
	}
	for _, c := range cases {
		if got := (Scope{Role: c.role}).Recognized(); got != c.want {
			t.Errorf("Recognized(%q)=%v want=%v", c.role, got, c.want)
		}
	}
}
