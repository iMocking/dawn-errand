package http

import (
	"strings"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/response"
)

// ListMerchants GET /merchants。
func (h *Handler) ListMerchants(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Merchants.List(c.Request.Context(), merchant.MerchantQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: mapMerchantStatus(c.Query("status")),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]MerchantRowDto, 0, len(rows))
	for _, m := range rows {
		list = append(list, toMerchantRow(m))
	}
	pageResult(c, list, total, page, size)
}

type createMerchantBody struct {
	Name           string  `json:"name"`
	Category       string  `json:"category"`
	Contact        string  `json:"contact"`
	Phone          string  `json:"phone"`
	Address        string  `json:"address"`
	AddressSub     string  `json:"addressSub"`
	OpenHours      string  `json:"openHours"`
	DeliveryRange  string  `json:"deliveryRange"`
	Commission     float64 `json:"commission"`
	DeliveryMethod string  `json:"deliveryMethod"`
	RegionID       string  `json:"regionId"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	AutoAccept     bool    `json:"autoAccept"`
	MinPrice       float64 `json:"minPrice"`
	DeliveryFee    float64 `json:"deliveryFee"`
	PackingFee     float64 `json:"packingFee"`
	Announcement   string  `json:"announcement"`
	ServicePhone   string  `json:"servicePhone"`

	Application merchant.Application `json:"application"`
}

func merchantFromBody(in createMerchantBody) *merchant.Merchant {
	return &merchant.Merchant{
		RegionID: in.RegionID, Name: strings.TrimSpace(in.Name), Category: in.Category,
		Contact: in.Contact, Phone: in.Phone, Address: in.Address, AddressSub: in.AddressSub,
		OpenHours: in.OpenHours, DeliveryRange: in.DeliveryRange, Commission: in.Commission,
		DeliveryMethod: in.DeliveryMethod, Lat: in.Lat, Lng: in.Lng,
		AutoAccept: in.AutoAccept, MinPrice: in.MinPrice,
		DeliveryFee: in.DeliveryFee, PackingFee: in.PackingFee,
		Announcement: in.Announcement, ServicePhone: in.ServicePhone,
		Application: in.Application, BankAccount: in.Application.BankAccountNo,
		Status: merchant.MerchantStatusRest, Operator: in.Contact,
	}
}

// CreateMerchant POST /merchants。
func (h *Handler) CreateMerchant(c *gin.Context) {
	var in createMerchantBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Phone) == "" {
		response.ErrStatus(c, 400, 1301, "商家名称与联系电话为必填项")
		return
	}
	m := merchantFromBody(in)
	if err := h.Merchants.Create(c.Request.Context(), m, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "已提交入驻申请", gin.H{"id": m.ID, "applyNo": m.ApplyNo, "auditStatus": "pending"})
}

// UpdateMerchantApplication PUT /merchants/:id/application：驳回后补件重提。
func (h *Handler) UpdateMerchantApplication(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	var in createMerchantBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	m := merchantFromBody(in)
	u := middleware.User(c)
	if err := h.Merchants.UpdateApplication(c.Request.Context(), id, m, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "资料已保存并重新提交审核", gin.H{"id": id, "auditStatus": "pending"})
}

// GetMerchantStats GET /merchants/stats。
func (h *Handler) GetMerchantStats(c *gin.Context) {
	st, err := h.Merchants.Stats(c.Request.Context(), scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, MerchantStatsDto{
		Total: st.Total, WeekNew: st.WeekNew, Open: st.Open,
		OpenRate: pctStr(st.OpenRate), Pending: st.Pending, Closed: st.Closed,
		ClosedDetail: "休息中 " + thousandSep(st.Closed) + " 家",
	})
}

// ListMerchantAudit GET /merchants/audit。
func (h *Handler) ListMerchantAudit(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Merchants.AuditList(c.Request.Context(), merchant.AuditQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: mapAuditStatus(c.Query("status")),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]MerchantAuditRowDto, 0, len(rows))
	for _, m := range rows {
		list = append(list, toMerchantAuditRow(m))
	}
	pageResult(c, list, total, page, size)
}

// GetMerchant GET /merchants/:id。
func (h *Handler) GetMerchant(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	m, err := h.Merchants.Detail(c.Request.Context(), id, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toMerchantDetail(m))
}

// GetMerchantWorkspace GET /merchants/:id/workspace：详情页各 Tab 数据。
func (h *Handler) GetMerchantWorkspace(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	ws, err := h.Merchants.Workspace(c.Request.Context(), id, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	orders := make([]gin.H, 0, len(ws.Orders))
	for _, o := range ws.Orders {
		orders = append(orders, orderBrief(o))
	}
	abnormal := make([]gin.H, 0, len(ws.Abnormal))
	for _, o := range ws.Abnormal {
		abnormal = append(abnormal, orderBrief(o))
	}
	cats := make([]gin.H, 0, len(ws.Categories))
	for _, cat := range ws.Categories {
		cats = append(cats, gin.H{"id": cat.ID, "name": cat.Name, "sort": cat.Sort})
	}
	dishes := make([]gin.H, 0, len(ws.Dishes))
	for _, d := range ws.Dishes {
		dishes = append(dishes, gin.H{
			"id": d.ID, "categoryId": d.CategoryID, "name": d.Name,
			"description": d.Description, "imageUrl": d.ImageURL,
			"price": d.Price, "stock": d.Stock, "onSale": d.OnSale,
			"soldOut": d.SoldOut(), "monthlySales": d.MonthlySales,
		})
	}
	reviews := make([]gin.H, 0, len(ws.Reviews))
	for _, rv := range ws.Reviews {
		reviews = append(reviews, gin.H{
			"id": rv.ID, "orderId": rv.OrderID, "userId": rv.UserID,
			"userName": ws.UserNames[rv.UserID], "rating": rv.Rating,
			"tags": rv.Tags, "content": rv.Content, "images": rv.Images,
			"reply": rv.Reply, "repliedAt": rv.RepliedAt, "createdAt": rv.CreatedAt,
		})
	}
	response.OK(c, gin.H{
		"application": ws.Merchant.Application,
		"certs":       merchantCerts(ws.Merchant),
		"auditStatus": merchantAuditText(ws.Merchant.AuditStatus),
		"verifiedAt":  ws.Merchant.VerifiedAt,
		"orders":      orders,
		"abnormal":    abnormal,
		"trend":       ws.Trend,
		"menu":        gin.H{"categories": cats, "dishes": dishes},
		"reviews":     gin.H{"list": reviews, "total": len(reviews), "rating": ws.Merchant.Rating},
	})
}

func orderBrief(o *order.Order) gin.H {
	return gin.H{
		"id": o.ID, "orderNo": o.OrderNo, "status": string(o.Status),
		"type": string(o.Type), "payAmount": o.PayAmount, "createdAt": o.CreatedAt,
		"userName": o.User.Name, "route": o.Route,
	}
}

// ResetMerchantAccount POST /merchants/:id/account/reset：重置商家登录密码。
func (h *Handler) ResetMerchantAccount(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	u := middleware.User(c)
	res, err := h.Accounts.ResetMerchantPassword(c.Request.Context(), id, u.ID, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "密码已重置", gin.H{
		"account": gin.H{
			"userId": res.UserID, "email": res.Email, "phone": res.Phone,
			"initialPassword": res.InitialPassword,
		},
	})
}

type auditMerchantBody struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// AuditMerchant POST /merchants/audit/:id。
func (h *Handler) AuditMerchant(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	var in auditMerchantBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.Action != "approve" && in.Action != "reject" {
		response.ErrStatus(c, 400, 1301, "action 仅支持 approve/reject")
		return
	}
	u := middleware.User(c)
	account, err := h.Merchants.AuditAction(c.Request.Context(), id, in.Action, in.Reason, u.ID, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	body := gin.H{"id": id, "action": in.Action}
	if account != nil {
		body["account"] = gin.H{
			"userId": account.UserID, "email": account.Email, "phone": account.Phone,
			"initialPassword": account.InitialPassword,
		}
	}
	response.OKMsg(c, "操作成功", body)
}
