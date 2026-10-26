package repo

import (
	"context"

	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/region"
)

// RegionRepo 实现 tenant.RegionRepository。
type RegionRepo struct{ client *ent.Client }

func (r RegionRepo) Save(ctx context.Context, reg *tenant.Region) error {
	if reg.ID == "" {
		reg.ID = newULID()
	}
	exists, err := r.client.Region.Query().Where(region.NameEqualFold(reg.Name)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err := r.client.Region.Update().
			Where(region.NameEqualFold(reg.Name)).
			SetStatus(string(reg.Status)).SetVillages(reg.Villages).SetRadiusKm(reg.RadiusKM).
			SetStartFee(reg.StartFee).SetFreeKm(reg.FreeKM).
			SetNotice(reg.Notice).SetUpdatedAt(reg.UpdatedAt).
			Save(ctx)
		return err
	}
	return r.client.Region.Create().
		SetID(reg.ID).SetName(reg.Name).SetStatus(string(reg.Status)).
		SetVillages(reg.Villages).SetRadiusKm(reg.RadiusKM).
		SetStartFee(reg.StartFee).SetFreeKm(reg.FreeKM).
		SetNotice(reg.Notice).SetCreatedAt(reg.CreatedAt).SetUpdatedAt(reg.UpdatedAt).
		Exec(ctx)
}

func toRegion(e *ent.Region) *tenant.Region {
	if e == nil {
		return nil
	}
	return &tenant.Region{
		ID:        e.ID,
		Name:      e.Name,
		Status:    tenant.RegionStatus(e.Status),
		Villages:  e.Villages,
		RadiusKM:  e.RadiusKm,
		StartFee:  e.StartFee,
		FreeKM:    e.FreeKm,
		Notice:    e.Notice,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func (r RegionRepo) FindByID(ctx context.Context, id string) (*tenant.Region, error) {
	e, err := r.client.Region.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toRegion(e), nil
}

func (r RegionRepo) List(ctx context.Context) ([]*tenant.Region, error) {
	rows, err := r.client.Region.Query().Order(ent.Asc(region.FieldName)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*tenant.Region, 0, len(rows))
	for _, e := range rows {
		out = append(out, toRegion(e))
	}
	return out, nil
}

func (r RegionRepo) ListByIDs(ctx context.Context, ids []string) ([]*tenant.Region, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.client.Region.Query().Where(region.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*tenant.Region, 0, len(rows))
	for _, e := range rows {
		out = append(out, toRegion(e))
	}
	return out, nil
}

func (r RegionRepo) Delete(ctx context.Context, id string) error {
	return r.client.Region.DeleteOneID(id).Exec(ctx)
}
