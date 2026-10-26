package http

import (
	"fmt"
	"strings"

	"dawn-errand/internal/domain/finance"
)

// ---- 财务 ----

func financeKindColors(kind string) (string, string) {
	switch {
	case strings.Contains(kind, "骑手"):
		return "#E8F0FF", "#2B7FFF"
	case strings.Contains(kind, "商家"):
		return "#FFF3E8", "#FF6B2C"
	}
	return "#FFE8E8", "#FF4D4F"
}

func withdrawStatusText(s finance.WithdrawStatus) string {
	switch s {
	case finance.WithdrawApproved:
		return "已通过"
	case finance.WithdrawRejected:
		return "已驳回"
	}
	return "待审核"
}

// WithdrawRowDto 提现列表行（对齐前端 WithdrawRow）。
type WithdrawRowDto struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	AvatarBg     string  `json:"avatarBg"`
	AvatarColor  string  `json:"avatarColor"`
	Kind         string  `json:"kind"`
	KindBg       string  `json:"kindBg"`
	KindColor    string  `json:"kindColor"`
	BankInfo     string  `json:"bankInfo"`
	RiskTip      string  `json:"riskTip,omitempty"`
	RejectReason string  `json:"rejectReason,omitempty"`
	Amount       float64 `json:"amount"`
	AmountColor  string  `json:"amountColor,omitempty"`
	CreatedAt    int64   `json:"created_at"`
	Status       string  `json:"status"`
}

func toWithdrawRow(w *finance.Withdrawal) WithdrawRowDto {
	kbg, kcol := financeKindColors(w.Kind)
	row := WithdrawRowDto{
		ID: w.ID, Name: w.PartyName, AvatarBg: kbg, AvatarColor: kcol,
		Kind: w.Kind, KindBg: kbg, KindColor: kcol, BankInfo: w.BankInfo,
		Amount: w.Amount, CreatedAt: w.CreatedAt, Status: withdrawStatusText(w.Status),
	}
	if w.RiskWarning != "" {
		row.RiskTip = w.RiskWarning
		row.AmountColor = "#FF4D4F"
	}
	if w.Status == finance.WithdrawRejected {
		row.RejectReason = w.RejectReason
	}
	return row
}

// FinanceStatsDto 财务汇总（对齐前端 FinanceStats）。
type FinanceStatsDto struct {
	TodayIncome             float64 `json:"todayIncome"`
	TodayIncomeDelta        string  `json:"todayIncomeDelta"`
	PendingBalance          float64 `json:"pendingBalance"`
	PendingBalanceBreakdown string  `json:"pendingBalanceBreakdown"`
	WithdrawPendingCount    int64   `json:"withdrawPendingCount"`
	AbnormalBills           int64   `json:"abnormalBills"`
	AbnormalBillsAmount     float64 `json:"abnormalBillsAmount"`
}

// SettlementStatsDto 结算汇总（对齐前端 SettlementStats）。
type SettlementStatsDto struct {
	TodaySettled      float64 `json:"todaySettled"`
	TodaySettledDelta string  `json:"todaySettledDelta"`
	WeekTotal         float64 `json:"weekTotal"`
	WeekBreakdown     string  `json:"weekBreakdown"`
	MonthTotal        float64 `json:"monthTotal"`
	MonthBreakdown    string  `json:"monthBreakdown"`
	PendingAmount     float64 `json:"pendingAmount"`
	PendingCountText  string  `json:"pendingCountText"`
	DateRange         string  `json:"dateRange"`
	ExportedText      string  `json:"exportedText"`
}

// FinanceDto 财务页（对齐前端 FinanceDto）。
type FinanceDto struct {
	Stats           FinanceStatsDto    `json:"stats"`
	SettlementStats SettlementStatsDto `json:"settlementStats"`
	Withdraws       []WithdrawRowDto   `json:"withdraws"`
}

// SettlementRowDto 结算流水行（对齐前端 SettlementRow）。
type SettlementRowDto struct {
	ID        string  `json:"id"`
	Party     string  `json:"party"`
	Type      string  `json:"type"`
	TypeBg    string  `json:"typeBg"`
	TypeColor string  `json:"typeColor"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	Period    string  `json:"period"`
	SettledAt int64   `json:"settled_at"`
	Action    string  `json:"action"`
}

func toSettlementRow(s *finance.Settlement) SettlementRowDto {
	bg, col := financeKindColors(string(s.Type))
	amount := s.Amount
	if s.Type == finance.SettlementRefund {
		amount = -s.Amount
	}
	status := "已到账"
	if s.Status == finance.SettlementProcessing {
		status = "处理中"
	}
	return SettlementRowDto{
		ID: s.ID, Party: s.Party, Type: string(s.Type), TypeBg: bg, TypeColor: col,
		Amount: amount, Status: status, Period: s.Period, SettledAt: s.SettledAt, Action: s.Action,
	}
}

// WithdrawDetailDto 提现详情（对齐前端 WithdrawDetailDto）。
type WithdrawDetailDto struct {
	ID              string          `json:"id"`
	Amount          float64         `json:"amount"`
	Fee             float64         `json:"fee"`
	FeeNote         string          `json:"feeNote"`
	Status          string          `json:"status"`
	Kind            string          `json:"kind"`
	Applicant       string          `json:"applicant"`
	ApplicantNo     string          `json:"applicantNo"`
	CreatedAt       int64           `json:"created_at"`
	WithdrawType    string          `json:"withdrawType"`
	BankInfo        string          `json:"bankInfo"`
	AccountHolder   string          `json:"accountHolder"`
	ArriveExpect    string          `json:"arriveExpect"`
	Purpose         string          `json:"purpose"`
	Balance         float64         `json:"balance"`
	BalanceNote     string          `json:"balanceNote"`
	UsedPercentText string          `json:"usedPercentText"`
	RemainBalance   float64         `json:"remainBalance"`
	RemainNote      string          `json:"remainNote"`
	MonthOrders     int64           `json:"monthOrders"`
	MonthIncome     float64         `json:"monthIncome"`
	LastMonthNote   string          `json:"lastMonthNote"`
	RiskItems       []string        `json:"riskItems"`
	RiskWarning     string          `json:"riskWarning"`
	Timeline        []finance.Event `json:"timeline"`
}

func toWithdrawDetail(w *finance.Withdrawal) WithdrawDetailDto {
	usedPct := "0%"
	if w.Balance > 0 {
		usedPct = pctStr(w.Amount / w.Balance * 100)
	}
	lastMonthNote := "上月同期 " + moneyStr(w.MonthIncome)
	if w.MonthOrders > 0 {
		lastMonthNote = fmt.Sprintf("本月 %d 单 · %s", w.MonthOrders, moneyStr(w.MonthIncome))
	}
	return WithdrawDetailDto{
		ID: w.ID, Amount: w.Amount, Fee: w.Fee, FeeNote: w.FeeNote,
		Status: withdrawStatusText(w.Status), Kind: w.Kind, Applicant: w.Applicant,
		ApplicantNo: w.PartyNo, CreatedAt: w.CreatedAt, WithdrawType: w.WithdrawType,
		BankInfo: w.BankInfo, AccountHolder: w.AccountHolder, ArriveExpect: w.ArriveExpect,
		Purpose: w.Purpose, Balance: w.Balance,
		BalanceNote: "可提现余额", UsedPercentText: usedPct,
		RemainBalance: w.RemainBalance, RemainNote: "提现后余额",
		MonthOrders: w.MonthOrders, MonthIncome: w.MonthIncome,
		LastMonthNote: lastMonthNote,
		RiskItems:     w.RiskItems, RiskWarning: w.RiskWarning, Timeline: w.Timeline,
	}
}
