package http

import (
	"github.com/gin-gonic/gin"

	"dawn-errand/internal/application/admin"
	"dawn-errand/internal/domain/coupon"
	"dawn-errand/internal/pkg/response"
)

// IssueCoupon POST /api/coupons。
func (h *Handler) IssueCoupon(c *gin.Context) {
	var in struct {
		UserID     string  `json:"userId"`
		Phone      string  `json:"phone"`
		Title      string  `json:"title"`
		Amount     float64 `json:"amount"`
		MinSpend   float64 `json:"minSpend"`
		MerchantID string  `json:"merchantId"`
		ExpiresAt  int64   `json:"expiresAt"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	coupon, err := h.CouponsAdmin.Issue(c.Request.Context(), admin.IssueCouponInput{
		UserID: in.UserID, Phone: in.Phone, Title: in.Title, Amount: in.Amount,
		MinSpend: in.MinSpend, MerchantID: in.MerchantID, ExpiresAt: in.ExpiresAt,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "发放成功", couponAdminDto(coupon))
}

// ListCoupons GET /api/coupons?page=&pageSize=&status=&userId=。
func (h *Handler) ListCoupons(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.CouponsAdmin.List(c.Request.Context(), page, size,
		c.Query("status"), c.Query("userId"))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, cp := range rows {
		list = append(list, couponAdminDto(cp))
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "pageSize": size,
	})
}

func couponAdminDto(c *coupon.Coupon) gin.H {
	return gin.H{
		"id": c.ID, "userId": c.UserID, "merchantId": c.MerchantID,
		"title": c.Title, "amount": c.Amount, "minSpend": c.MinSpend,
		"status": string(c.Status), "expiresAt": c.ExpiresAt,
		"usedOrderId": c.UsedOrderID, "createdAt": c.CreatedAt,
	}
}
