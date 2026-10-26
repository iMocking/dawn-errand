package http

import (
	"time"

	"github.com/gin-gonic/gin"

	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/interfaces/middleware"
	"dawn-errand/internal/pkg/response"
)

// GetFinance GET /finance。
func (h *Handler) GetFinance(c *gin.Context) {
	ov, err := h.Finance.Get(c.Request.Context(), scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	withdraws := make([]WithdrawRowDto, 0, len(ov.Withdraws))
	for _, w := range ov.Withdraws {
		withdraws = append(withdraws, toWithdrawRow(w))
	}
	ss := ov.SettlementStats
	incomeDelta := "0%"
	if ov.Stats.YesterdayIncome > 0 {
		incomeDelta = deltaStr((ov.Stats.TodayIncome - ov.Stats.YesterdayIncome) / ov.Stats.YesterdayIncome * 100)
	}
	settleDelta := "0%"
	if ss.YesterdaySettled > 0 {
		settleDelta = deltaStr((ss.TodaySettled - ss.YesterdaySettled) / ss.YesterdaySettled * 100)
	}
	now := time.Now().UTC()
	dateRange := now.AddDate(0, 0, -6).Format("2006.01.02") + " - " + now.Format("2006.01.02")
	response.OK(c, FinanceDto{
		Stats: FinanceStatsDto{
			TodayIncome: ov.Stats.TodayIncome, TodayIncomeDelta: incomeDelta,
			PendingBalance: ov.Stats.PendingBalance, PendingBalanceBreakdown: "含商家结算 " + moneyStr(ov.Stats.PendingBalance),
			WithdrawPendingCount: ov.Stats.WithdrawPending,
			AbnormalBills:        ov.Stats.AbnormalBills, AbnormalBillsAmount: ov.Stats.AbnormalAmount,
		},
		SettlementStats: SettlementStatsDto{
			TodaySettled: ss.TodaySettled, TodaySettledDelta: settleDelta,
			WeekTotal: ss.WeekTotal, WeekBreakdown: "近 7 日结算",
			MonthTotal: ss.MonthTotal, MonthBreakdown: "本月累计",
			PendingAmount:    ss.PendingAmount,
			PendingCountText: thousandSep(ss.PendingCount) + " 笔待处理",
			DateRange:        dateRange, ExportedText: "上次导出 10-21",
		},
		Withdraws: withdraws,
	})
}

// GetReconciliation GET /finance/reconciliation?date=YYYY-MM-DD。
func (h *Handler) GetReconciliation(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	res, err := h.Finance.Reconciliation(c.Request.Context(), date, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, gin.H{
		"date":    res.Date,
		"paid":    gin.H{"orders": res.PaidOrders, "amount": res.OrderAmount},
		"payment": gin.H{"count": res.PaymentCount, "amount": res.PaymentAmount},
		"refund":  gin.H{"count": res.RefundCount, "amount": res.RefundAmount},
		"diff":    res.Diff, "missing_orders": res.MissingOrders,
	})
}

// ListSettlements GET /finance/settlement。
func (h *Handler) ListSettlements(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Finance.SettlementList(c.Request.Context(), finance.SettlementQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: c.Query("status"), Type: c.Query("type"),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]SettlementRowDto, 0, len(rows))
	for _, s := range rows {
		list = append(list, toSettlementRow(s))
	}
	pageResult(c, list, total, page, size)
}

// ListWithdraws GET /finance/withdraws。
func (h *Handler) ListWithdraws(c *gin.Context) {
	page, size := parsePage(c)
	rows, total, err := h.Finance.WithdrawList(c.Request.Context(), finance.WithdrawQuery{
		Page: page, PageSize: size,
		Keyword: c.Query("keyword"), Status: c.Query("status"),
	}, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	list := make([]WithdrawRowDto, 0, len(rows))
	for _, w := range rows {
		list = append(list, toWithdrawRow(w))
	}
	pageResult(c, list, total, page, size)
}

// GetWithdraw GET /finance/withdraw/:id。
func (h *Handler) GetWithdraw(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	w, err := h.Finance.WithdrawDetail(c.Request.Context(), id, scope(c))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, toWithdrawDetail(w))
}

type withdrawAuditBody struct {
	Action string `json:"action"`
	Remark string `json:"remark"`
}

// AuditWithdraw POST /finance/withdraw/:id/audit。
func (h *Handler) AuditWithdraw(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	var in withdrawAuditBody
	if err := bind(c, &in); err != nil {
		response.Err(c, err)
		return
	}
	if in.Action != "pass" && in.Action != "reject" {
		response.ErrStatus(c, 400, 1301, "action 仅支持 pass/reject")
		return
	}
	u := middleware.User(c)
	if err := h.Finance.AuditWithdraw(c.Request.Context(), id, in.Action, in.Remark, u.ID, scope(c)); err != nil {
		response.Err(c, err)
		return
	}
	response.OKMsg(c, "审核完成", gin.H{"id": id, "action": in.Action})
}
