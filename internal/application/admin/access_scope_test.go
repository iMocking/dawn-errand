package admin

import (
	"testing"

	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/tenant"
)

// R3-5 回归：对象级隔离必须默认拒绝，未识别角色不得因「没有匹配分支」而看到全部数据。
func TestOrderAccessOKDefaultDeny(t *testing.T) {
	o := &order.Order{ID: "O1", RegionID: "R1", MerchantID: "M1", RiderID: "D1", UserID: "U1"}
	cases := []struct {
		name string
		sc   tenant.Scope
		want bool
	}{
		{"超管全局可见", tenant.Scope{Role: "super_admin"}, true},
		{"平台客服全局可见", tenant.Scope{Role: "cs_agent"}, true},
		{"站长同区域", tenant.Scope{Role: "region_station", RegionID: "R1"}, true},
		{"站长跨区域拒绝", tenant.Scope{Role: "region_station", RegionID: "R2"}, false},
		{"商家本单", tenant.Scope{Role: "merchant", MerchantID: "M1"}, true},
		{"商家他单拒绝", tenant.Scope{Role: "merchant", MerchantID: "M2"}, false},
		{"司机本单", tenant.Scope{Role: "driver", RiderID: "D1"}, true},
		{"会员本单", tenant.Scope{Role: "user", UserID: "U1"}, true},
		{"系统身份拒绝", tenant.Scope{Role: "system"}, false},
		{"空角色拒绝", tenant.Scope{}, false},
		{"伪造角色拒绝", tenant.Scope{Role: "root", UserID: "U1"}, false},
	}
	for _, c := range cases {
		if got := orderAccessOK(c.sc, o); got != c.want {
			t.Errorf("%s: orderAccessOK=%v want=%v", c.name, got, c.want)
		}
	}
}
