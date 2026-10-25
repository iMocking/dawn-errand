package http

import (
	"github.com/gin-gonic/gin"

	appsvc "dawn-errand/internal/application/app"
	"dawn-errand/internal/domain/refund"
	"dawn-errand/internal/domain/review"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/response"
)

// ===== 商家/骑手资金 =====

func (h *Handler) appFinanceSummary(c *gin.Context, partyType, partyID string) {
	sum, err := h.FinanceApp.Summary(c.Request.Context(), partyType, partyID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"balance": sum.Balance, "month_income": sum.MonthIncome,
		"today_income": sum.TodayIncome, "pending_withdraw": sum.PendingWithdraw,
		"pending_settlement": sum.PendingSettle, "withdrawable": sum.Balance,
	})
}

// AppStoreFinanceSummary GET /api/app/store/finance/summary。
func (h *Handler) AppStoreFinanceSummary(c *gin.Context) {
	p := middleware.Principal(c)
	h.appFinanceSummary(c, "merchant", p.MerchantID)
}

// AppDriverFinanceSummary GET /api/app/driver/finance/summary。
func (h *Handler) AppDriverFinanceSummary(c *gin.Context) {
	p := middleware.Principal(c)
	h.appFinanceSummary(c, "rider", p.RiderID)
}

func (h *Handler) appFinanceRecords(c *gin.Context, partyType, partyID string) {
	page, size := appPage(c)
	rows, total, err := h.FinanceApp.Records(c.Request.Context(), partyType, partyID, page, size)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		list = append(list, gin.H{
			"id": r.ID, "type": r.Type, "title": financeTypeTitle(r.Type),
			"amount": r.Amount, "incoming": r.Amount > 0,
			"balance_after": r.BalanceAfter, "order_id": r.OrderID,
			"note": r.Note, "created_at": r.CreatedAt,
		})
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": size,
		"has_more": int64(page*size) < total,
	})
}

// AppStoreFinanceRecords GET /api/app/store/finance/records。
func (h *Handler) AppStoreFinanceRecords(c *gin.Context) {
	p := middleware.Principal(c)
	h.appFinanceRecords(c, "merchant", p.MerchantID)
}

// AppDriverFinanceRecords GET /api/app/driver/finance/records。
func (h *Handler) AppDriverFinanceRecords(c *gin.Context) {
	p := middleware.Principal(c)
	h.appFinanceRecords(c, "rider", p.RiderID)
}

func (h *Handler) appSettlements(c *gin.Context, partyType, partyID string) {
	page, size := appPage(c)
	rows, total, err := h.FinanceApp.Settlements(c.Request.Context(), partyType, partyID, page, size)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, s := range rows {
		list = append(list, gin.H{
			"id": s.ID, "amount": s.Amount, "type": string(s.Type),
			"status": string(s.Status), "period": s.Period,
			"settled_at": s.SettledAt, "action": s.Action,
		})
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": size,
		"has_more": int64(page*size) < total,
	})
}

// AppStoreSettlements GET /api/app/store/settlements。
func (h *Handler) AppStoreSettlements(c *gin.Context) {
	p := middleware.Principal(c)
	h.appSettlements(c, "merchant", p.MerchantID)
}

// AppDriverSettlements GET /api/app/driver/settlements。
func (h *Handler) AppDriverSettlements(c *gin.Context) {
	p := middleware.Principal(c)
	h.appSettlements(c, "rider", p.RiderID)
}

func (h *Handler) appWithdrawals(c *gin.Context, partyType, partyID string) {
	page, size := appPage(c)
	rows, total, err := h.FinanceApp.Withdrawals(c.Request.Context(), partyType, partyID, page, size)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, w := range rows {
		list = append(list, gin.H{
			"id": w.ID, "amount": w.Amount, "fee": w.Fee,
			"status": string(w.Status), "kind": w.Kind,
			"bank_info": w.BankInfo, "arrive_expect": w.ArriveExpect,
			"reject_reason": w.RejectReason, "payout_ref": w.PayoutRef,
			"payout_at": w.PayoutAt, "created_at": w.CreatedAt,
		})
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": size,
		"has_more": int64(page*size) < total,
	})
}

// AppStoreWithdrawals GET /api/app/store/withdrawals。
func (h *Handler) AppStoreWithdrawals(c *gin.Context) {
	p := middleware.Principal(c)
	h.appWithdrawals(c, "merchant", p.MerchantID)
}

// AppDriverWithdrawals GET /api/app/driver/withdrawals。
func (h *Handler) AppDriverWithdrawals(c *gin.Context) {
	p := middleware.Principal(c)
	h.appWithdrawals(c, "rider", p.RiderID)
}

type withdrawBody struct {
	Amount      float64 `json:"amount"`
	Channel     string  `json:"channel"`
	AccountName string  `json:"account_name"`
	BankCard    string  `json:"bank_card"`
}

func (h *Handler) appCreateWithdrawal(c *gin.Context, partyType, partyID string) {
	var in withdrawBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	w, err := h.FinanceApp.CreateWithdrawal(c.Request.Context(), partyType, partyID, appsvc.WithdrawRequest{
		Amount: in.Amount, Channel: in.Channel, AccountName: in.AccountName, BankCard: in.BankCard,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "提现申请已提交", gin.H{
		"id": w.ID, "amount": w.Amount, "fee": w.Fee,
		"status": string(w.Status), "arrive_expect": w.ArriveExpect,
		"remain_balance": w.RemainBalance,
	})
}

// AppStoreCreateWithdrawal POST /api/app/store/withdrawals。
func (h *Handler) AppStoreCreateWithdrawal(c *gin.Context) {
	p := middleware.Principal(c)
	h.appCreateWithdrawal(c, "merchant", p.MerchantID)
}

// AppDriverCreateWithdrawal POST /api/app/driver/withdrawals。
func (h *Handler) AppDriverCreateWithdrawal(c *gin.Context) {
	p := middleware.Principal(c)
	h.appCreateWithdrawal(c, "rider", p.RiderID)
}

// AppListCoupons GET /api/app/client/coupons?status=available|used|expired|all。
func (h *Handler) AppListCoupons(c *gin.Context) {
	p := middleware.Principal(c)
	filter := c.DefaultQuery("status", "available")
	rows, err := h.Coupons.List(c.Request.Context(), p.UserID, filter)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, cp := range rows {
		list = append(list, gin.H{
			"id": cp.ID, "title": cp.Title, "amount": cp.Amount,
			"min_spend": cp.MinSpend, "merchant_id": cp.MerchantID,
			"status": string(cp.Status), "expires_at": cp.ExpiresAt,
			"used_order_id": cp.UsedOrderID,
		})
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

// ===== 评价 =====

// AppSubmitReview POST /api/app/client/orders/:id/review。
func (h *Handler) AppSubmitReview(c *gin.Context) {
	var in struct {
		Rating  int      `json:"rating"`
		Tags    []string `json:"tags"`
		Content string   `json:"content"`
		Images  []string `json:"images"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	o, rv, err := h.Reviews.Submit(c.Request.Context(), p.UserID, c.Param("id"), appsvc.ReviewInput{
		Rating: in.Rating, Tags: in.Tags, Content: in.Content, Images: in.Images,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "评价成功", gin.H{
		"order_id": o.ID, "status": string(o.Status),
		"review": reviewDto(rv),
	})
}

// AppStoreReviews GET /api/app/store/reviews。
func (h *Handler) AppStoreReviews(c *gin.Context) {
	p := middleware.Principal(c)
	page, size := appPage(c)
	rows, total, err := h.Reviews.ListStore(c.Request.Context(), p.MerchantID, page, size)
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, rv := range rows {
		list = append(list, reviewDto(rv))
	}
	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": size,
		"has_more": int64(page*size) < total,
	})
}

// AppReplyReview POST /api/app/store/reviews/:id/reply。
func (h *Handler) AppReplyReview(c *gin.Context) {
	var in struct {
		Content string `json:"content"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	if err := h.Reviews.Reply(c.Request.Context(), p.MerchantID, c.Param("id"), in.Content); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "回复成功", gin.H{})
}

// ===== 售后 =====

// AppApplyAfterSale POST /api/app/client/orders/:id/after-sale。
func (h *Handler) AppApplyAfterSale(c *gin.Context) {
	var in struct {
		Reason string  `json:"reason"`
		Amount float64 `json:"amount"`
	}
	_ = c.ShouldBindJSON(&in)
	p := middleware.Principal(c)
	r, err := h.Refunds.Apply(c.Request.Context(), p.UserID, c.Param("id"), in.Reason, in.Amount)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "售后申请已提交", refundDto(r))
}

// AppDecideRefund POST /api/app/store/orders/:id/refund/decision。
func (h *Handler) AppDecideRefund(c *gin.Context) {
	var in struct {
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
	}
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	p := middleware.Principal(c)
	o, err := h.Refunds.Decide(c.Request.Context(), p.MerchantID, c.Param("id"), in.Approve, in.Reason)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "售后已处理", gin.H{
		"order_id": o.ID, "status": string(o.Status),
		"payment_status": o.PaymentStatus,
	})
}

// ===== DTO =====

func financeTypeTitle(t string) string {
	switch t {
	case "order_income":
		return "订单收入"
	case "delivery_income":
		return "配送收入"
	case "withdraw":
		return "提现"
	case "withdraw_refund":
		return "提现驳回退回"
	case "refund_reversal":
		return "售后退回"
	case "refund":
		return "退款"
	}
	return t
}

func reviewDto(r *review.Review) gin.H {
	return gin.H{
		"id": r.ID, "order_id": r.OrderID, "user_id": r.UserID,
		"rating": r.Rating, "tags": r.Tags, "content": r.Content,
		"images": r.Images, "reply": r.Reply, "replied_at": r.RepliedAt,
		"created_at": r.CreatedAt,
	}
}

func refundDto(r *refund.Refund) gin.H {
	return gin.H{
		"id": r.ID, "order_id": r.OrderID, "amount": r.Amount,
		"reason": r.Reason, "status": string(r.Status),
		"created_at": r.CreatedAt, "handled_at": r.HandledAt,
	}
}
