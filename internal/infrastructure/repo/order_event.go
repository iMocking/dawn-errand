package repo

import (
	"context"

	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/orderevent"
)

// OrderEventRepo 实现 order.EventRepository。
type OrderEventRepo struct{ client *ent.Client }

func (r OrderEventRepo) Save(e *order.Event) error {
	if e.ID == "" {
		e.ID = newULID()
	}
	return r.client.OrderEvent.Create().
		SetID(e.ID).SetOrderID(e.OrderID).
		SetFromStatus(string(e.FromStatus)).SetToStatus(string(e.ToStatus)).
		SetActorType(string(e.ActorType)).SetActorID(e.ActorID).SetNote(e.Note).
		SetCreatedAt(e.CreatedAt).
		Exec(context.Background())
}

func (r OrderEventRepo) ListByOrder(orderID string) ([]*order.Event, error) {
	rows, err := r.client.OrderEvent.Query().
		Where(orderevent.OrderID(orderID)).
		Order(ent.Asc(orderevent.FieldCreatedAt), ent.Asc(orderevent.FieldID)).
		All(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]*order.Event, 0, len(rows))
	for _, e := range rows {
		out = append(out, &order.Event{
			ID: e.ID, OrderID: e.OrderID,
			FromStatus: order.OrderStatus(e.FromStatus), ToStatus: order.OrderStatus(e.ToStatus),
			ActorType: order.ActorType(e.ActorType), ActorID: e.ActorID,
			Note: e.Note, CreatedAt: e.CreatedAt,
		})
	}
	return out, nil
}
