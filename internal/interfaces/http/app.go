package http

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/application/admin"
	appsvc "dawn-errand/internal/application/app"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/response"
	"dawn-errand/internal/pkg/ulidgen"
)

// ===== App 面：用户端 / 商家端 / 骑手端上下文 =====

// AppClientMe GET /api/app/client/me。
func (h *Handler) AppClientMe(c *gin.Context) {
	u := middleware.User(c)
	me, err := h.Portal.ClientMe(c.Request.Context(), u.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, h.toClientMe(me, u))
}

// AppUpdateClientProfile PUT /api/app/client/me。
func (h *Handler) AppUpdateClientProfile(c *gin.Context) {
	var in struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	me, err := h.Portal.UpdateClientProfile(c.Request.Context(), u.ID, in.Name, in.Avatar)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "资料已更新", h.toClientMe(me, me.User))
}

// AppListRegions GET /api/app/client/regions。
func (h *Handler) AppListRegions(c *gin.Context) {
	rows, err := h.Portal.ListRegions(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		list = append(list, toRegionDto(r))
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

// AppSetRegion PUT /api/app/client/region。
func (h *Handler) AppSetRegion(c *gin.Context) {
	var in struct {
		RegionID string `json:"region_id"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	me, err := h.Portal.SetClientRegion(c.Request.Context(), u.ID, in.RegionID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "服务区域已更新", h.toClientMe(me, me.User))
}

func (h *Handler) toClientMe(me *appsvc.ClientMe, fallback *identity.User) gin.H {
	u := me.User
	if u == nil {
		u = fallback
	}
	body := gin.H{
		"id": u.ID, "name": u.Name, "avatar": u.Avatar,
		"phone": maskPhone(u.Phone), "role": u.Role,
		"region": toRegionDto(me.Region),
		"level":  "Lv.1 新用户", "verified": "未实名",
		"balance": 0.0, "red_packet_balance": 0.0,
	}
	if me.Profile != nil {
		if me.Profile.Level != "" {
			body["level"] = me.Profile.Level
		}
		if me.Profile.Verified != "" {
			body["verified"] = me.Profile.Verified
		}
		body["nickname"] = me.Profile.Nickname
		body["balance"] = me.Profile.Balance
		body["red_packet_balance"] = me.Profile.RedPacketBalance
	}
	return body
}

// AppStoreMe GET /api/app/store/me。
func (h *Handler) AppStoreMe(c *gin.Context) {
	p := middleware.Principal(c)
	me, err := h.Portal.StoreMe(c.Request.Context(), p.MerchantID)
	if err != nil {
		response.Err(c, err)
		return
	}
	m := me.Merchant
	u := middleware.User(c)
	statusCode := string(m.Status)
	statusText := "休息中"
	switch statusCode {
	case "open":
		statusText = "营业中"
	case "banned":
		statusText = "已封禁"
	}
	auditStatus := string(m.AuditStatus)
	response.OK(c, gin.H{
		"shop": gin.H{
			"id": m.ID, "name": m.Name, "category": m.Category,
			"address": m.Address, "phone": maskPhone(m.Phone),
			"rating": m.Rating, "status": statusText, "status_code": statusCode,
			"audit_status": auditStatus, "reject_reason": m.RejectReason,
			"open_hours": m.OpenHours, "delivery_range": m.DeliveryRange,
		},
		"business": gin.H{
			"open": statusCode == "open", "open_hours": m.OpenHours,
			"delivery_range": m.DeliveryRange, "commission": m.Commission,
			"delivery_method": m.DeliveryMethod, "auto_accept": m.AutoAccept,
			"min_price": m.MinPrice, "delivery_fee": m.DeliveryFee,
			"packing_fee": m.PackingFee, "announcement": m.Announcement,
			"service_phone": m.ServicePhone,
		},
		"today": gin.H{
			"orders": me.Today.TodayOrders, "revenue": me.Today.TodayGMV,
			"on_time_rate": me.Today.OnTimeRate,
		},
		"account": gin.H{
			"user_id": u.ID, "name": u.Name, "phone": maskPhone(u.Phone), "role": u.Role,
		},
	})
}

// AppDriverMe GET /api/app/driver/me。
func (h *Handler) AppDriverMe(c *gin.Context) {
	p := middleware.Principal(c)
	me, err := h.Portal.DriverMe(c.Request.Context(), p.RiderID)
	if err != nil {
		response.Err(c, err)
		return
	}
	r := me.Rider
	statusCode := string(r.Status)
	statusText := "离线"
	switch statusCode {
	case "online":
		statusText = "在线"
	case "busy":
		statusText = "忙碌"
	case "frozen":
		statusText = "已冻结"
	}
	response.OK(c, gin.H{
		"id": r.ID, "name": r.Name, "phone": maskPhone(r.Phone),
		"level": r.Level, "area": r.Area, "vehicle": r.Vehicle,
		"status": statusText, "status_code": statusCode,
		"rating": r.Rating, "audit_status": string(r.AuditStatus),
		"service_type": r.ServiceType, "joined_at": r.JoinedAt,
		"today": gin.H{
			"orders": me.Today.TodayOrders, "gmv": me.Today.TodayGMV,
		},
		"month_income": r.MonthIncome,
	})
}

// AppUpload POST /api/app/common/uploads（multipart/form-data，字段 file）。
func (h *Handler) AppUpload(c *gin.Context) {
	if h.Storage == nil {
		response.ErrStatus(c, http.StatusInternalServerError, errcode.CodeInternal, "存储未配置")
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			response.ErrStatus(c, http.StatusRequestEntityTooLarge, errcode.CodeFileTooLarge, "文件超过大小限制")
			return
		}
		_ = c.Error(err)
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "请选择要上传的文件")
		return
	}
	maxMB := h.Cfg.StorageMaxSizeMB
	if maxMB <= 0 {
		maxMB = 10
	}
	maxBytes := maxMB * 1024 * 1024
	if fh.Size > maxBytes {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeFileTooLarge, "文件超过大小限制")
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "文件读取失败")
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "文件读取失败")
		return
	}
	if int64(len(data)) > maxBytes {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeFileTooLarge, "文件超过大小限制")
		return
	}
	mime := http.DetectContentType(data)
	// purpose（旧文档字段名 scene 兼容）白名单：
	// 证件类一律进私有桶（仅归属主体与平台/区域管理员经鉴权路由可读），图片类进公开桶（静态挂载）。
	// 未知值直接拒绝——若静默落公开桶，一次拼写错误（如 identity_pic）就会把身份证暴露到无需鉴权的 URL 下。
	purpose := strings.ToLower(strings.TrimSpace(c.PostForm("purpose")))
	if purpose == "" {
		purpose = strings.ToLower(strings.TrimSpace(c.PostForm("scene")))
	}
	if purpose == "" {
		purpose = "image" // 兼容未带用途的老调用方，按图片处理
	}
	privatePurpose := map[string]bool{"doc": true, "idcard": true, "id": true, "credential": true, "license": true, "cert": true}
	publicPurpose := map[string]bool{
		"image": true, "avatar": true, "dish": true, "store": true, "menu": true,
		"proof": true, "delivery_proof": true, "complaint": true,
	}
	var url string
	switch {
	case !privatePurpose[purpose] && !publicPurpose[purpose]:
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "未知的 purpose 取值")
		return
	case privatePurpose[purpose]:
		partyType, partyID := uploadParty(principal(c))
		rel := partyType + "/" + partyID + "/" + ulidgen.New()
		url, _, err = h.Storage.PutPrivate(c.Request.Context(), data, rel, fh.Filename, mime)
	default:
		url, err = h.Storage.Put(c.Request.Context(), data, fh.Filename, mime)
	}
	if err != nil {
		_ = c.Error(err)
		msg := err.Error()
		switch {
		case strings.Contains(msg, "大小"):
			response.ErrStatus(c, http.StatusBadRequest, errcode.CodeFileTooLarge, "文件超过大小限制")
		case strings.Contains(msg, "类型"):
			response.ErrStatus(c, http.StatusBadRequest, errcode.CodeFileTypeDenied, "文件类型不支持")
		default:
			response.ErrStatus(c, http.StatusInternalServerError, errcode.CodeInternal, "文件保存失败")
		}
		return
	}
	response.OK(c, gin.H{
		"url": url, "name": fh.Filename, "size": len(data), "mime": mime,
	})
}

// uploadParty 上传主体归属前缀（私有桶访问控制依据）。
func uploadParty(p *tenant.Principal) (string, string) {
	switch {
	case p.MerchantID != "":
		return "merchant", p.MerchantID
	case p.RiderID != "":
		return "rider", p.RiderID
	default:
		return "user", p.UserID
	}
}

// privateFileReadable 私有桶读取归属判定：
// 主体本人（前缀与自身绑定一致）可读；未知前缀一律拒绝；
// 平台角色（全局作用域）可读，用于入驻与售后审核；
// 区域站长限本区域商家/骑手的文件（复用管理用例的作用域校验），外加「归属账号是平台角色」的
// user/* 文件（客服代上传的证照落在 user/<后台账号ID> 下）；普通会员本人的证件对站长仍不可见；
// 其余（其他商家/骑手/会员）拒绝——证件属 PII，不能因为「登录了」就可见。
func (h *Handler) privateFileReadable(c *gin.Context, p *tenant.Principal, sc tenant.Scope, seg, id string) bool {
	switch seg {
	case "user":
		if p.UserID == id {
			return true
		}
	case "merchant":
		if p.MerchantID == id {
			return true
		}
	case "rider":
		if p.RiderID == id {
			return true
		}
	default:
		return false
	}
	if sc.IsGlobal() {
		return true
	}
	if sc.RegionOnly() {
		if seg == "merchant" {
			_, err := h.Merchants.Detail(c.Request.Context(), id, sc)
			return err == nil
		}
		if seg == "rider" {
			_, err := h.Riders.Detail(c.Request.Context(), id, sc)
			return err == nil
		}
		if seg == "user" {
			// 管理端代主体上传的证件落在 user/<上传的后台账号ID> 下，路径本身不含商家信息：
			// 只有当该账号是平台角色时才放行，避免区域站长顺带读到普通会员的身份证。
			owner, err := h.MW.Users.FindByID(id)
			return err == nil && owner != nil && identity.IsPlatformRole(owner.Role)
		}
	}
	return false
}

// PrivateUploadDownload GET /api/uploads/private/*rel：私有桶受控读取。
// 归属校验见 privateFileReadable：主体本人 / 平台角色 / 本区域站长，其余 403。
func (h *Handler) PrivateUploadDownload(c *gin.Context) {
	rel := strings.TrimPrefix(c.Param("rel"), "/")
	if rel == "" {
		response.ErrStatus(c, http.StatusBadRequest, errcode.CodeBadRequest, "缺少文件路径")
		return
	}
	p := principal(c)
	if p == nil {
		response.ErrStatus(c, http.StatusUnauthorized, errcode.CodeUnauthorized, "未认证")
		return
	}
	sc := scope(c)
	segs := strings.Split(rel, "/")
	if len(segs) < 3 {
		response.ErrStatus(c, http.StatusNotFound, errcode.CodeNotFound, "文件不存在")
		return
	}
	if !h.privateFileReadable(c, p, sc, segs[0], segs[1]) {
		response.ErrStatus(c, http.StatusForbidden, errcode.CodeForbidden, "无权访问该文件")
		return
	}
	data, mime, err := h.Storage.ReadPrivate(c.Request.Context(), "private/"+rel)
	if err != nil {
		response.ErrStatus(c, http.StatusNotFound, errcode.CodeNotFound, "文件不存在")
		return
	}
	if mime == "" {
		mime = "application/octet-stream"
	}
	c.Data(http.StatusOK, mime, data)
}

// ===== 后台：商家/骑手账号开通 =====

// ProvisionMerchantAccount POST /api/merchants/:id/account。
func (h *Handler) ProvisionMerchantAccount(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	var in struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	res, err := h.Accounts.ProvisionMerchant(c.Request.Context(), id, admin.ProvisionInput{
		Email: in.Email, Phone: in.Phone, Password: in.Password, Name: in.Name,
	}, middleware.User(c).ID, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "商家账号已开通", toProvisionDto(res))
}

// ProvisionRiderAccount POST /api/riders/:id/account。
func (h *Handler) ProvisionRiderAccount(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	var in struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	res, err := h.Accounts.ProvisionRider(c.Request.Context(), id, admin.ProvisionInput{
		Email: in.Email, Phone: in.Phone, Password: in.Password, Name: in.Name,
	}, middleware.User(c).ID, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "骑手账号已开通", toProvisionDto(res))
}

func toProvisionDto(res *admin.ProvisionResult) gin.H {
	body := gin.H{
		"user_id": res.UserID, "email": res.Email, "phone": maskPhone(res.Phone),
		"name": res.Name, "role": res.Role,
		"merchant_id": res.MerchantID, "rider_id": res.RiderID,
	}
	if res.InitialPassword != "" {
		body["initial_password"] = res.InitialPassword
	}
	return body
}

// ===== 共享 DTO 辅助 =====

func toRegionDto(r *tenant.Region) gin.H {
	if r == nil {
		return nil
	}
	return gin.H{
		"id": r.ID, "name": r.Name, "status": string(r.Status),
		"villages": r.Villages, "radius_km": r.RadiusKM,
		"start_fee": r.StartFee, "free_km": r.FreeKM, "notice": r.Notice,
	}
}

// maskPhone 手机号脱敏：保留前 3 后 4。
func maskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) <= 4 {
		return phone
	}
	if len(phone) < 11 {
		return phone[:1] + strings.Repeat("*", len(phone)-2) + phone[len(phone)-1:]
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// maskAddress 地址脱敏：抢单前仅展示大致位置。
func maskAddress(detail string) string {
	r := []rune(strings.TrimSpace(detail))
	if len(r) <= 8 {
		return string(r)
	}
	return string(r[:8]) + "…"
}
