package admin

import (
	"context"
	"strings"
	"testing"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
)

// R5-5 管理端入驻审核单测。审核是一条「状态机 + 开通账号」的组合动作：
// 驳回必须写原因、通过前必须校验资料完整、审核通过后开通失败要能重试。
// integration 只走了「资料齐全→通过」这一条主干，分支靠这里补齐。

type fakeMerchantRepo struct {
	merchant.MerchantRepository
	byID       map[string]*merchant.Merchant
	auditCalls [][3]string // action(status), reason, operator
	statused   []merchant.MerchantStatus
	saved      []*merchant.Merchant
}

func newFakeMerchantRepo() *fakeMerchantRepo {
	return &fakeMerchantRepo{byID: map[string]*merchant.Merchant{}}
}

func (f *fakeMerchantRepo) FindByID(ctx context.Context, id string) (*merchant.Merchant, error) {
	m, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return m, nil
}

func (f *fakeMerchantRepo) UpdateAudit(ctx context.Context, id string, approved bool, reason, operator string, now int64) error {
	m, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	if approved {
		m.AuditStatus = merchant.AuditApproved
		m.Status = merchant.MerchantStatusOpen
	} else {
		m.AuditStatus = merchant.AuditRejected
		m.RejectReason = reason
		m.Status = merchant.MerchantStatusRest
	}
	f.auditCalls = append(f.auditCalls, [3]string{string(m.AuditStatus), reason, operator})
	return nil
}

func (f *fakeMerchantRepo) UpdateStatus(ctx context.Context, id string, status merchant.MerchantStatus, now int64) error {
	m, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	m.Status = status
	f.statused = append(f.statused, status)
	return nil
}

func (f *fakeMerchantRepo) Save(ctx context.Context, m *merchant.Merchant) error {
	f.saved = append(f.saved, m)
	f.byID[m.ID] = m
	return nil
}

type fakeRegionRepo struct {
	tenant.RegionRepository
	regions []*tenant.Region
	err     error
}

func (f *fakeRegionRepo) List(ctx context.Context) ([]*tenant.Region, error) {
	return f.regions, f.err
}

// completeMerchant 造一条资料齐全、待审核的入驻申请。
func completeMerchant(id string) *merchant.Merchant {
	return &merchant.Merchant{
		ID: id, RegionID: "RG1", Name: "老王面馆", Category: "餐饮", Contact: "老王",
		Phone: "13800138000", Address: "学院路 1 号", OpenHours: "09:00-21:00",
		Status: merchant.MerchantStatusRest, AuditStatus: merchant.AuditPending,
		Application: merchant.Application{
			LicenseNo: "91110000MA001", LicenseImage: "/uploads/license.jpg",
			LegalName: "王大明", LegalIDCard: "110101199003077219",
			IDCardFront: "/uploads/front.jpg", IDCardBack: "/uploads/back.jpg",
			StorefrontImages: []string{"/uploads/store.jpg"},
			BankName:         "工商银行", BankAccountName: "王大明", BankAccountNo: "6222021234567890",
			FoodLicenseNo: "JY1101000001", FoodLicenseImage: "/uploads/food.jpg",
			AccountEmail: "laowang@example.com", AccountPhone: "13800138000",
		},
	}
}

func TestMerchantAuditRejectsAndGuards(t *testing.T) {
	ctx := context.Background()
	repo := newFakeMerchantRepo()
	audit := &fakeAudit{}
	s := &MerchantsService{Repos: &application.RepoUnit{Merchants: repo, Audit: audit}}
	global := tenant.Scope{Role: "super_admin"}

	if code := adminCodeOf(t, mustProvisionErr(s.AuditAction(ctx, "M404", "approve", "", "op", global))); code != errcode.CodeMerchantNotFound {
		t.Errorf("申请不存在应 404, got=%d", code)
	}

	m := completeMerchant("M1")
	repo.byID["M1"] = m
	// 驳回必须写原因：否则业务员无从补件。
	if code := adminCodeOf(t, mustProvisionErr(s.AuditAction(ctx, "M1", "reject", "   ", "op", global))); code != errcode.CodeBadRequest {
		t.Errorf("空原因驳回应 400, got=%d", code)
	}
	if m.AuditStatus != merchant.AuditPending {
		t.Error("被拒的驳回改动了审核状态")
	}
	if _, err := s.AuditAction(ctx, "M1", "reject", " 执照模糊 ", "op-1", global); err != nil {
		t.Fatalf("驳回答错: %v", err)
	}
	if m.AuditStatus != merchant.AuditRejected || m.RejectReason != "执照模糊" {
		t.Errorf("驳回未落状态/原因: %+v", m)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "merchant.audit" {
		t.Errorf("驳回未留审计: %+v", audit.entries)
	}
	// 已处理的申请不能再审（防重复开通）。
	if code := adminCodeOf(t, mustProvisionErr(s.AuditAction(ctx, "M1", "approve", "", "op", global))); code != errcode.CodeMerchantStatus {
		t.Errorf("已处理申请再审应拒, got=%d", code)
	}

	// 区域站长不得越区审核。
	m2 := completeMerchant("M2")
	m2.RegionID = "RG2"
	repo.byID["M2"] = m2
	if code := adminCodeOf(t, mustProvisionErr(s.AuditAction(ctx, "M2", "approve", "", "op", tenant.Scope{Role: "region_station", RegionID: "RG1"}))); code != errcode.CodeTenantScopeDenied {
		t.Errorf("跨区域审核应 403, got=%d", code)
	}
}

func TestMerchantAuditApproveValidatesAndProvisions(t *testing.T) {
	ctx := context.Background()
	repo := newFakeMerchantRepo()
	audit := &fakeAudit{}
	s := &MerchantsService{Repos: &application.RepoUnit{Merchants: repo, Audit: audit}}
	global := tenant.Scope{Role: "super_admin"}

	// 资料不全：必须列出缺什么，而不是先落审核状态再报错。
	m := completeMerchant("M1")
	m.Application.BankAccountNo = ""
	repo.byID["M1"] = m
	_, err := s.AuditAction(ctx, "M1", "approve", "", "op", global)
	if code := adminCodeOf(t, err); code != errcode.CodeBadRequest {
		t.Errorf("资料不全应 400, got=%d", code)
	}
	if m.AuditStatus != merchant.AuditPending {
		t.Error("资料校验失败却落成已通过")
	}
	if msg := err.Error(); msg == "" || !strings.Contains(msg, "结算账号") {
		t.Errorf("缺失项未点明: %q", msg)
	}

	// 已开通账号的申请再审：只落审核状态，不再重复开通。
	m2 := completeMerchant("M2")
	m2.OwnerUserID = "U9"
	repo.byID["M2"] = m2
	if res, err := s.AuditAction(ctx, "M2", "approve", "", "op", global); err != nil || res != nil {
		t.Errorf("已开通账号的申请不应再开通: res=%+v err=%v", res, err)
	}
	if m2.AuditStatus != merchant.AuditApproved || m2.Status != merchant.MerchantStatusOpen {
		t.Errorf("通过后状态错误: %+v", m2)
	}

	// 未开通账号且平台未装配开通服务：审核已通过，但必须明确报「服务未配置」，
	// 让人在详情页重试开通，而不是静默返回成功。
	m3 := completeMerchant("M3")
	repo.byID["M3"] = m3
	if code := adminCodeOf(t, mustProvisionErr(s.AuditAction(ctx, "M3", "approve", "", "op", global))); code != errcode.CodeInternal {
		t.Errorf("未装配开通服务应报内部错误, got=%d", code)
	}
	if m3.AuditStatus != merchant.AuditApproved {
		t.Errorf("开通失败时审核状态应已通过（可重试开通）, got=%q", m3.AuditStatus)
	}
	if len(audit.entries) != 1 {
		t.Errorf("审计条数错误: %d", len(audit.entries))
	}
}

func TestMerchantDetailAndStatusScopes(t *testing.T) {
	ctx := context.Background()
	repo := newFakeMerchantRepo()
	audit := &fakeAudit{}
	s := &MerchantsService{Repos: &application.RepoUnit{Merchants: repo, Audit: audit}}
	global := tenant.Scope{Role: "super_admin"}

	m := completeMerchant("M1")
	m.AuditStatus = merchant.AuditApproved
	repo.byID["M1"] = m

	if code := adminCodeOf(t, mustMerchantErr(s.Detail(ctx, "M404", global))); code != errcode.CodeMerchantNotFound {
		t.Errorf("商家不存在应 404, got=%d", code)
	}
	if code := adminCodeOf(t, mustMerchantErr(s.Detail(ctx, "M1", tenant.Scope{Role: "region_station", RegionID: "RG9"}))); code != errcode.CodeTenantScopeDenied {
		t.Errorf("跨区域详情应 403, got=%d", code)
	}
	if got, err := s.Detail(ctx, "M1", global); err != nil || got.ID != "M1" {
		t.Errorf("详情失败: %+v err=%v", got, err)
	}

	if code := adminCodeOf(t, s.UpdateStatus(ctx, "M404", merchant.MerchantStatusOpen, "op", global)); code != errcode.CodeMerchantNotFound {
		t.Errorf("不存在的商家改状态应 404, got=%d", code)
	}
	if code := adminCodeOf(t, s.UpdateStatus(ctx, "M1", merchant.MerchantStatusOpen, "op", tenant.Scope{Role: "region_station", RegionID: "RG9"})); code != errcode.CodeTenantScopeDenied {
		t.Errorf("跨区域改状态应 403, got=%d", code)
	}
	if err := s.UpdateStatus(ctx, "M1", merchant.MerchantStatusRest, "op-1", global); err != nil {
		t.Fatalf("改状态失败: %v", err)
	}
	if m.Status != merchant.MerchantStatusRest || len(repo.statused) != 1 {
		t.Errorf("状态未落库: %+v %v", m.Status, repo.statused)
	}
	if last := audit.entries[len(audit.entries)-1]; last.Action != "merchant.status" {
		t.Errorf("改状态未留审计: %+v", last)
	}
}

func TestMerchantCreateFillsRegion(t *testing.T) {
	ctx := context.Background()
	repo := newFakeMerchantRepo()
	regions := &fakeRegionRepo{regions: []*tenant.Region{{ID: "RG1"}}}
	audit := &fakeAudit{}
	s := &MerchantsService{Repos: &application.RepoUnit{Merchants: repo, Regions: regions, Audit: audit}}

	// 站长创建：未填区域自动落本区域，越区直接拒绝。
	m := completeMerchant("")
	m.RegionID = ""
	if err := s.Create(ctx, m, tenant.Scope{Role: "region_station", RegionID: "RG1"}); err != nil {
		t.Fatalf("站长创建失败: %v", err)
	}
	if m.RegionID != "RG1" {
		t.Errorf("站长创建应自动填本区域: %q", m.RegionID)
	}
	m2 := completeMerchant("")
	m2.RegionID = "RG2"
	if code := adminCodeOf(t, s.Create(ctx, m2, tenant.Scope{Role: "region_station", RegionID: "RG1"})); code != errcode.CodeTenantScopeDenied {
		t.Errorf("越区创建应 403, got=%d", code)
	}

	// 超管未指定区域：取第一个服务区域；没有任何区域时必须拒绝（库表 NotEmpty）。
	m3 := completeMerchant("")
	m3.RegionID = ""
	if err := s.Create(ctx, m3, tenant.Scope{Role: "super_admin"}); err != nil {
		t.Fatalf("超管创建失败: %v", err)
	}
	if m3.RegionID != "RG1" {
		t.Errorf("超管创建应回落首个区域: %q", m3.RegionID)
	}
	empty := &fakeRegionRepo{}
	s2 := &MerchantsService{Repos: &application.RepoUnit{Merchants: repo, Regions: empty, Audit: audit}}
	m4 := completeMerchant("")
	m4.RegionID = ""
	if code := adminCodeOf(t, s2.Create(ctx, m4, tenant.Scope{Role: "super_admin"})); code != errcode.CodeBadRequest {
		t.Errorf("无服务区域应拒绝创建, got=%d", code)
	}
}

func mustMerchantErr(_ *merchant.Merchant, err error) error { return err }
func mustProvisionErr(_ *ProvisionResult, err error) error  { return err }
