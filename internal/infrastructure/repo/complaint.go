package repo

import (
	"context"
	"strings"

	"dawn-errand/internal/domain/complaint"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/infrastructure/ent"
	entcomplaint "dawn-errand/internal/infrastructure/ent/complaint"
)

// ComplaintRepo 实现 complaint.ComplaintRepository。
type ComplaintRepo struct{ client *ent.Client }

func (r ComplaintRepo) Save(c *complaint.Complaint) error {
	ctx := context.Background()
	if c.ID == "" {
		c.ID = newULID()
	}
	err := r.client.Complaint.Create().
		SetID(c.ID).SetOrderID(c.OrderID).SetRegionID(c.RegionID).SetUserID(c.UserID).
		SetCategory(c.Category).SetTitle(c.Title).SetStatus(string(c.Status)).
		SetPriority(c.Priority).SetComplainant(c.Complainant).SetTarget(c.Target).
		SetExtra(c.Extra).SetDescription(c.Description).SetDescTone(c.DescTone).
		SetDeadline(c.Deadline).SetHandler(c.Handler).SetProgress(c.Progress).
		SetInitiator(c.Initiator).SetInitiatorExtra(c.InitiatorExtra).SetPayAmount(c.PayAmount).
		SetCreatedAt(c.CreatedAt).SetUpdatedAt(c.UpdatedAt).
		SetFlow(mustJSON(c.Flow)).SetNotes(mustJSON(c.Notes)).SetOrderItems(mustJSON(c.Order)).
		Exec(ctx)
	return err
}

func toComplaint(e *ent.Complaint) *complaint.Complaint {
	if e == nil {
		return nil
	}
	c := &complaint.Complaint{
		ID: e.ID, OrderID: e.OrderID, RegionID: e.RegionID, UserID: e.UserID,
		Category: e.Category, Title: e.Title, Status: complaint.Status(e.Status),
		Priority: e.Priority, Complainant: e.Complainant, Target: e.Target,
		Extra: e.Extra, Description: e.Description, DescTone: e.DescTone,
		Deadline: e.Deadline, Handler: e.Handler, Progress: e.Progress,
		Initiator: e.Initiator, InitiatorExtra: e.InitiatorExtra, PayAmount: e.PayAmount,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
	_ = jsonUnmarshal(e.Flow, &c.Flow)
	_ = jsonUnmarshal(e.Notes, &c.Notes)
	_ = jsonUnmarshal(e.OrderItems, &c.Order)
	return c
}

func (r ComplaintRepo) FindByID(id string) (*complaint.Complaint, error) {
	e, err := r.client.Complaint.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toComplaint(e), nil
}

func (r ComplaintRepo) List(q complaint.Query, sc tenant.Scope) ([]*complaint.Complaint, int64, error) {
	ctx := context.Background()
	bq := r.client.Complaint.Query()
	if sc.RegionOnly() {
		bq.Where(entcomplaint.RegionID(sc.RegionID))
	}
	if q.UserID != "" {
		bq.Where(entcomplaint.UserID(q.UserID))
	}
	if s := strings.TrimSpace(q.Status); s != "" {
		bq.Where(entcomplaint.Status(s))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(entcomplaint.Or(
			entcomplaint.TitleContainsFold(k),
			entcomplaint.ComplainantContainsFold(k),
			entcomplaint.TargetContainsFold(k),
			entcomplaint.OrderIDContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(entcomplaint.FieldCreatedAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*complaint.Complaint, 0, len(rows))
	for _, e := range rows {
		out = append(out, toComplaint(e))
	}
	return out, int64(total), nil
}

func (r ComplaintRepo) CountPending() (int64, error) {
	n, err := r.client.Complaint.Query().
		Where(entcomplaint.StatusIn(string(complaint.StatusPending), string(complaint.StatusProcessing))).
		Count(context.Background())
	return int64(n), err
}

func (r ComplaintRepo) CountByUserID(userID string) (int64, error) {
	n, err := r.client.Complaint.Query().
		Where(entcomplaint.UserID(userID)).
		Count(context.Background())
	return int64(n), err
}
