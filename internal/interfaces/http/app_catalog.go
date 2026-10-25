package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	appsvc "dawn-errand/internal/application/app"
	"dawn-errand/internal/domain/address"
	"dawn-errand/internal/domain/catalog"
	"dawn-errand/internal/domain/merchant"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/response"
)

// ===== 用户端：商家目录 =====

// AppListMerchants GET /api/app/client/merchants。
func (h *Handler) AppListMerchants(c *gin.Context) {
	p := middleware.Principal(c)
	page, size := appPage(c)
	lat, lng := appLatLng(c)
	cards, total, err := h.ClientApp.ListMerchants(c.Request.Context(), p.UserID,
		c.Query("keyword"), c.Query("category"), page, size, lat, lng)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(cards))
	for _, card := range cards {
		list = append(list, toMerchantCard(card))
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": size,
		"has_more": int64(page*size) < total,
	})
}

// AppGetMerchant GET /api/app/client/merchants/:id。
func (h *Handler) AppGetMerchant(c *gin.Context) {
	p := middleware.Principal(c)
	lat, lng := appLatLng(c)
	detail, err := h.ClientApp.GetMerchant(c.Request.Context(), p.UserID, c.Param("id"), lat, lng)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"merchant":         toMerchantCard(&appsvc.MerchantCard{Merchant: detail.Merchant, DistanceKM: detail.DistanceKM, DeliveryMinutes: detail.DeliveryMinutes}),
		"categories":       toCategoryList(detail.Categories),
		"dishes":           toDishList(detail.Dishes),
		"distance_km":      detail.DistanceKM,
		"delivery_minutes": detail.DeliveryMinutes,
	})
}

// ===== 用户端：地址簿 =====

type addressBody struct {
	ContactName string  `json:"contact_name"`
	Phone       string  `json:"phone"`
	RegionID    string  `json:"region_id"`
	Detail      string  `json:"detail"`
	Tag         string  `json:"tag"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	IsDefault   bool    `json:"is_default"`
}

func (b addressBody) toInput() appsvc.AddressInput {
	return appsvc.AddressInput{
		ContactName: b.ContactName, Phone: b.Phone, RegionID: b.RegionID,
		Detail: b.Detail, Tag: b.Tag, Lat: b.Lat, Lng: b.Lng, IsDefault: b.IsDefault,
	}
}

// AppListAddresses GET /api/app/client/addresses。
func (h *Handler) AppListAddresses(c *gin.Context) {
	p := middleware.Principal(c)
	rows, err := h.ClientApp.ListAddresses(c.Request.Context(), p.UserID)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, a := range rows {
		list = append(list, toAddressDto(a))
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

// AppCreateAddress POST /api/app/client/addresses。
func (h *Handler) AppCreateAddress(c *gin.Context) {
	var in addressBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	a, err := h.ClientApp.CreateAddress(c.Request.Context(), p.UserID, in.toInput())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "地址已保存", toAddressDto(a))
}

// AppUpdateAddress PUT /api/app/client/addresses/:id。
func (h *Handler) AppUpdateAddress(c *gin.Context) {
	var in addressBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	a, err := h.ClientApp.UpdateAddress(c.Request.Context(), p.UserID, c.Param("id"), in.toInput())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "地址已更新", toAddressDto(a))
}

// AppDeleteAddress DELETE /api/app/client/addresses/:id。
func (h *Handler) AppDeleteAddress(c *gin.Context) {
	p := middleware.Principal(c)
	if err := h.ClientApp.DeleteAddress(c.Request.Context(), p.UserID, c.Param("id")); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "地址已删除", gin.H{})
}

// AppSetDefaultAddress POST /api/app/client/addresses/:id/default。
func (h *Handler) AppSetDefaultAddress(c *gin.Context) {
	p := middleware.Principal(c)
	a, err := h.ClientApp.SetDefaultAddress(c.Request.Context(), p.UserID, c.Param("id"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "默认地址已更新", toAddressDto(a))
}

// ===== 商家端：经营设置 / 菜单 =====

type businessBody struct {
	OpenHours     string  `json:"open_hours"`
	DeliveryRange string  `json:"delivery_range"`
	AutoAccept    bool    `json:"auto_accept"`
	MinPrice      float64 `json:"min_price"`
	DeliveryFee   float64 `json:"delivery_fee"`
	PackingFee    float64 `json:"packing_fee"`
	Announcement  string  `json:"announcement"`
	ServicePhone  string  `json:"service_phone"`
}

// AppGetStoreBusiness GET /api/app/store/business。
func (h *Handler) AppGetStoreBusiness(c *gin.Context) {
	p := middleware.Principal(c)
	m, err := h.StoreApp.GetBusiness(c.Request.Context(), p.MerchantID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toBusinessDto(m))
}

// AppUpdateStoreBusiness PUT /api/app/store/business。
func (h *Handler) AppUpdateStoreBusiness(c *gin.Context) {
	var in businessBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	m, err := h.StoreApp.UpdateBusiness(c.Request.Context(), p.MerchantID, appsvc.BusinessSettings{
		OpenHours: in.OpenHours, DeliveryRange: in.DeliveryRange, AutoAccept: in.AutoAccept,
		MinPrice: in.MinPrice, DeliveryFee: in.DeliveryFee, PackingFee: in.PackingFee,
		Announcement: in.Announcement, ServicePhone: in.ServicePhone,
	}, middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "营业设置已保存", toBusinessDto(m))
}

// AppUpdateStoreStatus PUT /api/app/store/status。
func (h *Handler) AppUpdateStoreStatus(c *gin.Context) {
	var in struct {
		Open bool `json:"open"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	m, err := h.StoreApp.UpdateStatus(c.Request.Context(), p.MerchantID, in.Open, middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "营业状态已更新", gin.H{
		"open":   m.Status == merchant.MerchantStatusOpen,
		"status": appMerchantStatusText(m.Status), "status_code": string(m.Status),
	})
}

// AppStoreMenu GET /api/app/store/menu。
func (h *Handler) AppStoreMenu(c *gin.Context) {
	p := middleware.Principal(c)
	menu, err := h.StoreApp.Menu(c.Request.Context(), p.MerchantID)
	if err != nil {
		response.Err(c, err)
		return
	}
	cats := make([]gin.H, 0, len(menu.Categories))
	for _, mc := range menu.Categories {
		cats = append(cats, gin.H{
			"id": mc.Category.ID, "name": mc.Category.Name, "sort": mc.Category.Sort,
			"dish_count": len(mc.Dishes), "dishes": toDishList(mc.Dishes),
		})
	}
	response.OK(c, gin.H{"categories": cats})
}

// AppCreateCategory POST /api/app/store/categories。
func (h *Handler) AppCreateCategory(c *gin.Context) {
	var in struct {
		Name string `json:"name"`
		Sort int    `json:"sort"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	cat, err := h.StoreApp.CreateCategory(c.Request.Context(), p.MerchantID,
		appsvc.CategoryInput{Name: in.Name, Sort: in.Sort}, middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "分类已创建", toCategoryDto(cat))
}

// AppUpdateCategory PUT /api/app/store/categories/:id。
func (h *Handler) AppUpdateCategory(c *gin.Context) {
	var in struct {
		Name string `json:"name"`
		Sort int    `json:"sort"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	cat, err := h.StoreApp.UpdateCategory(c.Request.Context(), p.MerchantID, c.Param("id"),
		appsvc.CategoryInput{Name: in.Name, Sort: in.Sort}, middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "分类已更新", toCategoryDto(cat))
}

// AppDeleteCategory DELETE /api/app/store/categories/:id。
func (h *Handler) AppDeleteCategory(c *gin.Context) {
	p := middleware.Principal(c)
	if err := h.StoreApp.DeleteCategory(c.Request.Context(), p.MerchantID, c.Param("id"), middleware.User(c).ID); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "分类已删除", gin.H{})
}

type dishBody struct {
	CategoryID  string  `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	OnSale      *bool   `json:"on_sale"`
	Sort        int     `json:"sort"`
}

func (b dishBody) toInput() appsvc.DishInput {
	return appsvc.DishInput{
		CategoryID: b.CategoryID, Name: b.Name, Description: b.Description,
		ImageURL: b.ImageURL, Price: b.Price, Stock: b.Stock, OnSale: b.OnSale, Sort: b.Sort,
	}
}

// AppCreateDish POST /api/app/store/dishes。
func (h *Handler) AppCreateDish(c *gin.Context) {
	var in dishBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	d, err := h.StoreApp.CreateDish(c.Request.Context(), p.MerchantID, in.toInput(), middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "菜品已创建", toDishDto(d))
}

// AppUpdateDish PUT /api/app/store/dishes/:id。
func (h *Handler) AppUpdateDish(c *gin.Context) {
	var in dishBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	d, err := h.StoreApp.UpdateDish(c.Request.Context(), p.MerchantID, c.Param("id"), in.toInput(), middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "菜品已更新", toDishDto(d))
}

// AppSetDishSale POST /api/app/store/dishes/:id/on-sale。
func (h *Handler) AppSetDishSale(c *gin.Context) {
	var in struct {
		OnSale *bool `json:"on_sale"`
		Stock  *int  `json:"stock"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.OnSale == nil {
		response.ErrStatus(c, 400, errcode.CodeBadRequest, "on_sale 必填")
		return
	}
	stock := -1
	if in.Stock != nil {
		stock = *in.Stock
	}
	p := middleware.Principal(c)
	d, err := h.StoreApp.SetDishSale(c.Request.Context(), p.MerchantID, c.Param("id"),
		*in.OnSale, stock, middleware.User(c).ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "菜品状态已更新", toDishDto(d))
}

// AppDeleteDish DELETE /api/app/store/dishes/:id。
func (h *Handler) AppDeleteDish(c *gin.Context) {
	p := middleware.Principal(c)
	if err := h.StoreApp.DeleteDish(c.Request.Context(), p.MerchantID, c.Param("id"), middleware.User(c).ID); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "菜品已删除", gin.H{})
}

// ===== DTO 辅助 =====

func appPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}

func appLatLng(c *gin.Context) (float64, float64) {
	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
	return lat, lng
}

func appMerchantStatusText(s merchant.MerchantStatus) string {
	switch s {
	case merchant.MerchantStatusOpen:
		return "营业中"
	case merchant.MerchantStatusBanned:
		return "已封禁"
	default:
		return "休息中"
	}
}

func toMerchantCard(card *appsvc.MerchantCard) gin.H {
	m := card.Merchant
	return gin.H{
		"id": m.ID, "name": m.Name, "category": m.Category,
		"rating": m.Rating, "monthly_sales": m.MonthlyOrders,
		"distance_km": card.DistanceKM, "delivery_minutes": card.DeliveryMinutes,
		"delivery_fee": m.DeliveryFee, "min_price": m.MinPrice,
		"delivery_range": m.DeliveryRange, "address": m.Address,
		"open_hours": m.OpenHours, "announcement": m.Announcement,
		"status": appMerchantStatusText(m.Status), "status_code": string(m.Status),
	}
}

func toBusinessDto(m *merchant.Merchant) gin.H {
	return gin.H{
		"open":   m.Status == merchant.MerchantStatusOpen,
		"status": appMerchantStatusText(m.Status), "status_code": string(m.Status),
		"open_hours": m.OpenHours, "delivery_range": m.DeliveryRange,
		"auto_accept": m.AutoAccept, "min_price": m.MinPrice,
		"delivery_fee": m.DeliveryFee, "packing_fee": m.PackingFee,
		"announcement": m.Announcement, "service_phone": m.ServicePhone,
	}
}

func toCategoryDto(cat *catalog.Category) gin.H {
	return gin.H{"id": cat.ID, "name": cat.Name, "sort": cat.Sort}
}

func toCategoryList(cats []*catalog.Category) []gin.H {
	out := make([]gin.H, 0, len(cats))
	for _, c := range cats {
		out = append(out, toCategoryDto(c))
	}
	return out
}

func toDishDto(d *catalog.Dish) gin.H {
	return gin.H{
		"id": d.ID, "category_id": d.CategoryID, "name": d.Name,
		"description": d.Description, "image_url": d.ImageURL,
		"price": d.Price, "stock": d.Stock, "on_sale": d.OnSale,
		"sold_out": d.SoldOut(), "monthly_sales": d.MonthlySales, "sort": d.Sort,
	}
}

func toDishList(rows []*catalog.Dish) []gin.H {
	out := make([]gin.H, 0, len(rows))
	for _, d := range rows {
		out = append(out, toDishDto(d))
	}
	return out
}

func toAddressDto(a *address.Address) gin.H {
	return gin.H{
		"id": a.ID, "contact_name": a.ContactName, "phone": a.Phone,
		"region_id": a.RegionID, "detail": a.Detail, "tag": a.Tag,
		"lat": a.Lat, "lng": a.Lng, "is_default": a.IsDefault,
		"created_at": a.CreatedAt, "updated_at": a.UpdatedAt,
	}
}
