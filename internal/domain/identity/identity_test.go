package identity

import "testing"

// R5-5：手机号是短信注册/登录的账号主键，格式放宽一步等于放开枚举面与脏数据。
func TestValidatePhone(t *testing.T) {
	ok := []string{"13800138000", "19912345678", "17012345678", "15999999999"}
	bad := []string{
		"", " ", "1380013800", "138001380000", // 长度不符
		"2380013800", "12800138000", "10800138000", // 首位/次位不符
		"+861380013800", "138 0013 8000", "1380013800a", "13800138000 ", // 非纯数字或带空白
	}
	for _, p := range ok {
		if !ValidatePhone(p) {
			t.Errorf("合法手机号被拒: %q", p)
		}
	}
	for _, p := range bad {
		if ValidatePhone(p) {
			t.Errorf("非法手机号被放行: %q", p)
		}
	}
}

// R5-5：平台侧角色判定同时决定「代传证照的可读范围」与后台入口放行，
// user/merchant/driver/system 混进来就是横向越权。
func TestIsPlatformRole(t *testing.T) {
	cases := map[string]bool{
		RoleSuperAdmin:    true,
		RoleCSManager:     true,
		RoleCSAgent:       true,
		RoleRegionStation: true,
		RoleUser:          false,
		RoleMerchant:      false,
		RoleDriver:        false,
		RoleSystem:        false,
		"":                false,
		"admin":           false,
		"Super_Admin":     false,
	}
	for role, want := range cases {
		if got := IsPlatformRole(role); got != want {
			t.Errorf("IsPlatformRole(%q)=%v want=%v", role, got, want)
		}
	}
}

// R5-5：内置角色必须全部有中文显示名，且除 system 外都归入 AllRoles，
// 否则后台筛选下拉与会话角色校验会出现「有角色无名字」的空档。
func TestRoleTablesAreConsistent(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range AllRoles {
		if seen[r] {
			t.Errorf("AllRoles 重复: %s", r)
		}
		seen[r] = true
		if _, ok := RoleDisplayName[r]; !ok {
			t.Errorf("角色 %s 缺少中文显示名", r)
		}
	}
	for _, r := range []string{RoleSuperAdmin, RoleCSManager, RoleCSAgent, RoleRegionStation,
		RoleMerchant, RoleDriver, RoleUser, RoleSystem} {
		if !seen[r] {
			t.Errorf("AllRoles 漏掉角色 %s", r)
		}
	}
	for _, si := range SystemIdentities {
		if si.Email == "" || si.Name == "" {
			t.Errorf("系统身份定义不完整: %+v", si)
		}
	}
}

// R5-5：封禁状态必须即时体现在 IsActive 上，登录与刷新都以此放行。
func TestUserIsActive(t *testing.T) {
	cases := []struct {
		status UserStatus
		want   bool
	}{
		{UserStatusActive, true},
		{UserStatusBanned, false},
		{UserStatus(""), false},
		{UserStatus("active "), false},
	}
	for _, c := range cases {
		u := &User{Status: c.status}
		if got := u.IsActive(); got != c.want {
			t.Errorf("IsActive(%q)=%v want=%v", c.status, got, c.want)
		}
	}
}
