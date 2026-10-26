package http

import (
	"strings"

	"dawn-errand/internal/domain/complaint"
)

// ---- 投诉 ----

func complaintCategoryColors(category string) (string, string) {
	switch {
	case strings.Contains(category, "配送"):
		return "#FFF3E8", "#FF6B2C"
	case strings.Contains(category, "商家"):
		return "#FFF8E0", "#D48806"
	}
	return "#E8F0FF", "#2B7FFF"
}

// ComplaintRowDto 投诉工单行（对齐前端 ComplaintRow）。
type ComplaintRowDto struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	Complainant string `json:"complainant"`
	Target      string `json:"target"`
	Extra       string `json:"extra"`
	CreatedAt   int64  `json:"created_at"`
	Description string `json:"description,omitempty"`
	DescTone    string `json:"desc_tone,omitempty"`
	Deadline    string `json:"deadline,omitempty"`
	Handler     string `json:"handler,omitempty"`
	Progress    string `json:"progress,omitempty"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
}

func toComplaintRow(c *complaint.Complaint) ComplaintRowDto {
	row := ComplaintRowDto{
		ID: c.ID, Category: c.Category, Title: c.Title, OrderID: c.OrderID,
		Status: string(c.Status), Complainant: c.Complainant, Target: c.Target,
		Extra: c.Extra, CreatedAt: c.CreatedAt,
	}
	if c.Status == complaint.StatusPending {
		row.Description = c.Description
		row.DescTone = c.DescTone
		row.Deadline = c.Deadline
	}
	if c.Status == complaint.StatusProcessing {
		row.Handler = c.Handler
		row.Progress = c.Progress
		row.UpdatedAt = c.UpdatedAt
	}
	return row
}

// ComplaintFlowStepDto 流程节点（对齐前端 ComplaintFlowStep）。
type ComplaintFlowStepDto struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
	State string `json:"state"`
	Step  int    `json:"step,omitempty"`
}

type ComplaintNoteDto struct {
	Avatar      string `json:"avatar"`
	AvatarBg    string `json:"avatar_bg"`
	AvatarColor string `json:"avatar_color"`
	Name        string `json:"name"`
	Time        int64  `json:"time"`
	Content     string `json:"content"`
	Tone        string `json:"tone"`
}

type ComplaintOrderItemDto struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Tone  string `json:"tone,omitempty"`
}

// ComplaintDetailDto 工单详情（对齐前端 ComplaintDetail）。
type ComplaintDetailDto struct {
	ID             string                  `json:"id"`
	Title          string                  `json:"title"`
	Status         string                  `json:"status"`
	Priority       string                  `json:"priority"`
	CreatedAt      int64                   `json:"created_at"`
	Initiator      string                  `json:"initiator"`
	InitiatorExtra string                  `json:"initiator_extra"`
	Target         string                  `json:"target"`
	OrderID        string                  `json:"order_id"`
	Flow           []ComplaintFlowStepDto  `json:"flow"`
	Notes          []ComplaintNoteDto      `json:"notes"`
	Order          []ComplaintOrderItemDto `json:"order"`
	PayAmount      float64                 `json:"pay_amount"`
}

func toComplaintDetail(c *complaint.Complaint) ComplaintDetailDto {
	flow := make([]ComplaintFlowStepDto, 0, len(c.Flow))
	for _, f := range c.Flow {
		flow = append(flow, ComplaintFlowStepDto{Title: f.Title, Desc: f.Desc, State: f.State, Step: f.Step})
	}
	notes := make([]ComplaintNoteDto, 0, len(c.Notes))
	for _, n := range c.Notes {
		notes = append(notes, ComplaintNoteDto{
			Avatar: n.Avatar, AvatarBg: n.AvatarBg, AvatarColor: n.AvatarColor,
			Name: n.Name, Time: n.Time, Content: n.Content, Tone: n.Tone,
		})
	}
	orderItems := make([]ComplaintOrderItemDto, 0, len(c.Order))
	for _, o := range c.Order {
		orderItems = append(orderItems, ComplaintOrderItemDto{Label: o.Label, Value: o.Value, Tone: o.Tone})
	}
	return ComplaintDetailDto{
		ID: c.ID, Title: c.Title, Status: string(c.Status), Priority: c.Priority,
		CreatedAt: c.CreatedAt, Initiator: c.Initiator, InitiatorExtra: c.InitiatorExtra,
		Target: c.Target, OrderID: c.OrderID, Flow: flow, Notes: notes,
		Order: orderItems, PayAmount: c.PayAmount,
	}
}
