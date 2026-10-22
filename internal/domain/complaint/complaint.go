// Package complaint 投诉工单限界上下文。
package complaint

import "dawn-errand/internal/domain/tenant"

// Status 工单状态。
type Status string

const (
	StatusPending    Status = "待处理"
	StatusProcessing Status = "处理中"
	StatusResolved   Status = "已完结"
)

// FlowStep 处理流程节点。
type FlowStep struct {
	Title string
	Desc  string
	State string // done / current / pending
	Step  int
}

// Note 沟通记录。
type Note struct {
	Avatar      string
	AvatarBg    string
	AvatarColor string
	Name        string
	Time        int64
	Content     string
	Tone        string // gray / yellow
}

// OrderItem 关联订单键值对。
type OrderItem struct {
	Label string
	Value string
	Tone  string // blue / primary / normal
}

// Complaint 投诉工单聚合根。
type Complaint struct {
	ID             string
	OrderID        string
	RegionID       string
	UserID         string
	Category       string // 配送异常 / 商家投诉 / 退款申请
	Title          string
	Status         Status
	Priority       string
	Complainant    string
	Target         string
	Extra          string
	Description    string
	DescTone       string // red / yellow
	Deadline       string
	Handler        string
	Progress       string
	Initiator      string
	InitiatorExtra string
	PayAmount      float64

	CreatedAt int64
	UpdatedAt int64

	Flow  []FlowStep
	Notes []Note
	Order []OrderItem
}

// Query 工单列表查询。
type Query struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
}

// ComplaintRepository 投诉仓储端口。
type ComplaintRepository interface {
	Save(c *Complaint) error
	FindByID(id string) (*Complaint, error)
	List(q Query, scope tenant.Scope) ([]*Complaint, int64, error)
	CountPending() (int64, error)
	// CountByUserID 统计用户相关工单数。
	CountByUserID(userID string) (int64, error)
}
