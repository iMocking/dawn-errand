package admin

import (
	"context"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/ulidgen"
)

// MerchantsService 商家管理。
type MerchantsService struct {
	Repos    *application.RepoUnit
	Accounts *AccountsService
}

// List 商家列表。
func (s *MerchantsService) List(ctx context.Context, q merchant.MerchantQuery, sc tenant.Scope) ([]*merchant.Merchant, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Merchants.List(q, sc)
}

// Stats 商家聚合。
func (s *MerchantsService) Stats(ctx context.Context, sc tenant.Scope) (*merchant.Stats, error) {
	return s.Repos.Merchants.Stats(sc)
}

// AuditList 入驻审核列表。
func (s *MerchantsService) AuditList(ctx context.Context, q merchant.AuditQuery, sc tenant.Scope) ([]*merchant.Merchant, int64, error) {
	normalizeQ(&q.Page, &q.PageSize)
	return s.Repos.Merchants.AuditList(q, sc)
}

// Detail 商家详情（对象级隔离）。
func (s *MerchantsService) Detail(ctx context.Context, id string, sc tenant.Scope) (*merchant.Merchant, error) {
	m, err := s.Repos.Merchants.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
		}
		return nil, err
	}
	if !merchantAccessOK(sc, m) {
		return nil, tenantDenied()
	}
	return m, nil
}

// AuditAction 审核入驻：通过时自动开通商家端账号并返回一次性初始密码。
func (s *MerchantsService) AuditAction(ctx context.Context, id, action, reason, operator string, sc tenant.Scope) (*ProvisionResult, error) {
	m, err := s.Repos.Merchants.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return nil, errcode.NewStatus(404, errcode.CodeMerchantNotFound, "申请不存在")
		}
		return nil, err
	}
	if !merchantAccessOK(sc, m) {
		return nil, tenantDenied()
	}
	if m.AuditStatus != merchant.AuditPending {
		return nil, errcode.New(errcode.CodeMerchantStatus, "该申请已处理")
	}
	approved := action == "approve"
	if !approved {
		if strings.TrimSpace(reason) == "" {
			return nil, errcode.New(errcode.CodeBadRequest, "驳回必须填写原因，便于业务员补件")
		}
		if err := s.Repos.Merchants.UpdateAudit(id, false, strings.TrimSpace(reason), operator, time.Now().Unix()); err != nil {
			return nil, err
		}
		s.audit(ctx, id, operator)
		return nil, nil
	}

	if err := validateApplication(m); err != nil {
		return nil, err
	}
	// 先落审核状态，再由平台开通账号；开通失败时审核已通过，可在商家详情重试开通。
	if err := s.Repos.Merchants.UpdateAudit(id, true, "", operator, time.Now().Unix()); err != nil {
		return nil, err
	}
	var account *ProvisionResult
	if m.OwnerUserID == "" {
		if s.Accounts == nil {
			return nil, errcode.New(errcode.CodeInternal, "账号开通服务未配置")
		}
		phone := strings.TrimSpace(m.Application.AccountPhone)
		if phone == "" {
			phone = strings.TrimSpace(m.Phone)
		}
		account, err = s.Accounts.ProvisionMerchant(ctx, id, ProvisionInput{
			Email: m.Application.AccountEmail, Phone: phone, Name: m.Contact,
		}, operator)
		if err != nil {
			return nil, err
		}
	}
	s.audit(ctx, id, operator)
	return account, nil
}

func (s *MerchantsService) audit(ctx context.Context, id, operator string) {
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "merchant.audit", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
}

// validateApplication 入驻资料完整性校验（业务员地推必填）。
func validateApplication(m *merchant.Merchant) error {
	var missing []string
	add := func(ok bool, label string) {
		if !ok {
			missing = append(missing, label)
		}
	}
	a := m.Application
	add(strings.TrimSpace(m.Name) != "", "商家名称")
	add(strings.TrimSpace(m.Category) != "", "经营类型")
	add(strings.TrimSpace(m.Contact) != "", "联系人")
	add(strings.TrimSpace(m.Phone) != "", "联系电话")
	add(strings.TrimSpace(m.Address) != "", "详细地址")
	add(strings.TrimSpace(m.OpenHours) != "", "营业时间")
	add(strings.TrimSpace(a.LicenseNo) != "", "营业执照号")
	add(strings.TrimSpace(a.LicenseImage) != "", "营业执照照片")
	add(strings.TrimSpace(a.LegalName) != "", "法人姓名")
	add(strings.TrimSpace(a.LegalIDCard) != "", "法人身份证号")
	add(strings.TrimSpace(a.IDCardFront) != "", "身份证人像面")
	add(strings.TrimSpace(a.IDCardBack) != "", "身份证国徽面")
	add(len(a.StorefrontImages) > 0, "门头/店内照片")
	add(strings.TrimSpace(a.BankName) != "", "开户银行")
	add(strings.TrimSpace(a.BankAccountName) != "", "开户名")
	add(strings.TrimSpace(a.BankAccountNo) != "", "结算账号")
	if strings.TrimSpace(m.Category) == "餐饮" {
		add(strings.TrimSpace(a.FoodLicenseNo) != "", "食品经营许可证号")
		add(strings.TrimSpace(a.FoodLicenseImage) != "", "食品经营许可证照片")
	}
	if len(missing) > 0 {
		return errcode.New(errcode.CodeBadRequest, "入驻资料不完整，请补充："+strings.Join(missing, "、"))
	}
	if m.Commission < 0 || m.Commission > 100 {
		return errcode.New(errcode.CodeBadRequest, "佣金比例需在 0-100 之间")
	}
	if m.MinPrice < 0 || m.DeliveryFee < 0 || m.PackingFee < 0 {
		return errcode.New(errcode.CodeBadRequest, "金额字段不能为负")
	}
	return nil
}

// UpdateApplication 业务员补录/修改被驳回的入驻资料，并重新提交审核。
func (s *MerchantsService) UpdateApplication(ctx context.Context, id string, in *merchant.Merchant, operator string, sc tenant.Scope) error {
	cur, err := s.Repos.Merchants.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeMerchantNotFound, "申请不存在")
		}
		return err
	}
	if !merchantAccessOK(sc, cur) {
		return tenantDenied()
	}
	if cur.OwnerUserID != "" {
		return errcode.New(errcode.CodeMerchantStatus, "账号已开通，资料请走变更流程")
	}
	if cur.AuditStatus == merchant.AuditApproved {
		return errcode.New(errcode.CodeMerchantStatus, "已通过的申请不可直接修改")
	}
	in.ID, in.ApplyNo, in.RegionID = cur.ID, cur.ApplyNo, cur.RegionID
	in.OwnerUserID = cur.OwnerUserID
	in.AuditStatus = merchant.AuditPending
	in.Status = merchant.MerchantStatusRest
	in.RejectReason = ""
	in.AppliedAt = cur.AppliedAt
	in.SubmittedAt = time.Now().Unix()
	in.LastUpdatedAt = time.Now().Unix()
	in.JoinedAt = cur.JoinedAt
	in.Completeness = 100
	in.VerifiedAt = 0
	in.VerifiedBy = ""
	in.Operator = firstNonEmpty(in.Operator, cur.Operator)
	if err := validateApplication(in); err != nil {
		return err
	}
	return s.Repos.Merchants.Save(in)
}

// UpdateStatus 变更营业状态。
func (s *MerchantsService) UpdateStatus(ctx context.Context, id string, status merchant.MerchantStatus, operator string, sc tenant.Scope) error {
	m, err := s.Repos.Merchants.FindByID(id)
	if err != nil {
		if err == application.ErrNotFound {
			return errcode.NewStatus(404, errcode.CodeMerchantNotFound, "商家不存在")
		}
		return err
	}
	if !merchantAccessOK(sc, m) {
		return tenantDenied()
	}
	if err := s.Repos.Merchants.UpdateStatus(id, status, time.Now().Unix()); err != nil {
		return err
	}
	s.Repos.Audit.Record(ctx, application.AuditEntry{
		UserID: operator, Action: "merchant.status", Object: id, Result: "success",
		RequestID: application.RequestID(ctx), IP: application.ClientIP(ctx),
	})
	return nil
}

// Create 创建商家（含入驻申请）。
func (s *MerchantsService) Create(ctx context.Context, m *merchant.Merchant, sc tenant.Scope) error {
	if sc.RegionOnly() {
		if m.RegionID == "" {
			m.RegionID = sc.RegionID
		}
		if m.RegionID != sc.RegionID {
			return tenantDenied()
		}
	} else if m.RegionID == "" {
		// 超管未指定区域：使用第一个服务区域（ent schema region_id NotEmpty）。
		regions, err := s.Repos.Regions.List()
		if err != nil || len(regions) == 0 {
			return errcode.New(errcode.CodeBadRequest, "未配置服务区域，无法创建商家")
		}
		m.RegionID = regions[0].ID
	}
	now := time.Now().Unix()
	if m.ID == "" {
		m.ID = ulidgen.New()
	}
	if m.ApplyNo == "" {
		m.ApplyNo = "SQ" + time.Now().Format("20060102") + m.ID[:4]
	}
	m.AppliedAt = now
	m.SubmittedAt = now
	if m.Completeness == 0 {
		m.Completeness = 60
		if m.Application.LicenseNo != "" && m.Application.LegalIDCard != "" && len(m.Application.StorefrontImages) > 0 {
			m.Completeness = 100
		}
	}
	if m.AuditStatus == "" {
		m.AuditStatus = merchant.AuditPending
	}
	m.LastUpdatedAt = now
	if m.JoinedAt == 0 {
		m.JoinedAt = now
	}
	return s.Repos.Merchants.Save(m)
}
