package repo

import (
	"context"

	"dawn-errand/internal/domain/tracking"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/riderlocation"
)

// TrackingRepo 实现 tracking.Repository。
type TrackingRepo struct{ client *ent.Client }

func (r TrackingRepo) Upsert(l *tracking.Location) error {
	ctx := context.Background()
	if l.ID == "" {
		l.ID = newULID()
	}
	exists, err := r.client.RiderLocation.Query().Where(riderlocation.RiderID(l.RiderID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return r.client.RiderLocation.Update().
			Where(riderlocation.RiderID(l.RiderID)).
			SetRegionID(l.RegionID).SetLat(l.Lat).SetLng(l.Lng).
			SetHeading(l.Heading).SetSpeed(l.Speed).SetAccuracy(l.Accuracy).
			SetUpdatedAt(l.UpdatedAt).
			Exec(ctx)
	}
	return r.client.RiderLocation.Create().
		SetID(l.ID).SetRiderID(l.RiderID).SetRegionID(l.RegionID).
		SetLat(l.Lat).SetLng(l.Lng).SetHeading(l.Heading).
		SetSpeed(l.Speed).SetAccuracy(l.Accuracy).SetUpdatedAt(l.UpdatedAt).
		Exec(ctx)
}

func (r TrackingRepo) FindByRider(riderID string) (*tracking.Location, error) {
	e, err := r.client.RiderLocation.Query().Where(riderlocation.RiderID(riderID)).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toLocation(e), nil
}

func (r TrackingRepo) FindByRiders(riderIDs []string) (map[string]*tracking.Location, error) {
	if len(riderIDs) == 0 {
		return map[string]*tracking.Location{}, nil
	}
	rows, err := r.client.RiderLocation.Query().
		Where(riderlocation.RiderIDIn(riderIDs...)).
		All(context.Background())
	if err != nil {
		return nil, err
	}
	out := make(map[string]*tracking.Location, len(rows))
	for _, e := range rows {
		out[e.RiderID] = toLocation(e)
	}
	return out, nil
}

func (r TrackingRepo) DeleteStale(before int64) error {
	_, err := r.client.RiderLocation.Delete().
		Where(riderlocation.UpdatedAtLT(before)).Exec(context.Background())
	return err
}

func toLocation(e *ent.RiderLocation) *tracking.Location {
	return &tracking.Location{
		ID: e.ID, RiderID: e.RiderID, RegionID: e.RegionID,
		Lat: e.Lat, Lng: e.Lng, Heading: e.Heading, Speed: e.Speed,
		Accuracy: e.Accuracy, UpdatedAt: e.UpdatedAt,
	}
}
