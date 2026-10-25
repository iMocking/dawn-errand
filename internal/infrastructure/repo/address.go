package repo

import (
	"context"

	"dawn-errand/internal/domain/address"
	"dawn-errand/internal/infrastructure/ent"
	entaddress "dawn-errand/internal/infrastructure/ent/useraddress"
)

// AddressRepo implements address.Repository.
type AddressRepo struct{ client *ent.Client }

func (r AddressRepo) Save(a *address.Address) error {
	ctx := context.Background()
	if a.ID == "" {
		a.ID = newULID()
	}
	exists, err := r.client.UserAddress.Query().Where(entaddress.ID(a.ID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		err = r.client.UserAddress.UpdateOneID(a.ID).
			SetUserID(a.UserID).SetContactName(a.ContactName).SetPhone(a.Phone).
			SetRegionID(a.RegionID).SetDetail(a.Detail).SetTag(a.Tag).
			SetLat(a.Lat).SetLng(a.Lng).SetIsDefault(a.IsDefault).
			SetUpdatedAt(nowSec()).Exec(ctx)
		return err
	}
	return r.client.UserAddress.Create().
		SetID(a.ID).SetUserID(a.UserID).SetContactName(a.ContactName).SetPhone(a.Phone).
		SetRegionID(a.RegionID).SetDetail(a.Detail).SetTag(a.Tag).
		SetLat(a.Lat).SetLng(a.Lng).SetIsDefault(a.IsDefault).
		SetCreatedAt(nowSec()).SetUpdatedAt(nowSec()).Exec(ctx)
}

func (r AddressRepo) FindByID(id string) (*address.Address, error) {
	e, err := r.client.UserAddress.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &address.Address{
		ID: e.ID, UserID: e.UserID, ContactName: e.ContactName, Phone: e.Phone,
		RegionID: e.RegionID, Detail: e.Detail, Tag: e.Tag,
		Lat: e.Lat, Lng: e.Lng, IsDefault: e.IsDefault,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}, nil
}

func (r AddressRepo) ListByUser(userID string) ([]*address.Address, error) {
	rows, err := r.client.UserAddress.Query().
		Where(entaddress.UserID(userID)).
		Order(ent.Desc(entaddress.FieldIsDefault), ent.Desc(entaddress.FieldUpdatedAt)).
		All(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]*address.Address, 0, len(rows))
	for _, e := range rows {
		out = append(out, &address.Address{
			ID: e.ID, UserID: e.UserID, ContactName: e.ContactName, Phone: e.Phone,
			RegionID: e.RegionID, Detail: e.Detail, Tag: e.Tag,
			Lat: e.Lat, Lng: e.Lng, IsDefault: e.IsDefault,
			CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
		})
	}
	return out, nil
}

func (r AddressRepo) Delete(id string) error {
	return r.client.UserAddress.DeleteOneID(id).Exec(context.Background())
}

func (r AddressRepo) ClearDefault(userID string) error {
	return r.client.UserAddress.Update().
		Where(entaddress.UserID(userID), entaddress.IsDefault(true)).
		SetIsDefault(false).SetUpdatedAt(nowSec()).
		Exec(context.Background())
}

func (r AddressRepo) CountByUser(userID string) (int64, error) {
	n, err := r.client.UserAddress.Query().Where(entaddress.UserID(userID)).Count(context.Background())
	return int64(n), err
}
