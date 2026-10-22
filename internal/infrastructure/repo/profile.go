package repo

import (
	"context"

	"dawn-errand/internal/domain/userprofile"
	"dawn-errand/internal/infrastructure/ent"
	entuserprofile "dawn-errand/internal/infrastructure/ent/userprofile"
)

// ProfileRepo 实现 userprofile.ProfileRepository。
type ProfileRepo struct{ client *ent.Client }

func (r ProfileRepo) Save(p *userprofile.Profile) error {
	ctx := context.Background()
	exists, err := r.client.UserProfile.Query().Where(entuserprofile.UserID(p.UserID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err := r.client.UserProfile.Update().
			Where(entuserprofile.UserID(p.UserID)).
			SetLevel(p.Level).SetNickname(p.Nickname).SetRealName(p.RealName).SetVerified(p.Verified).
			SetRegion(p.Region).SetAddress(p.Address).SetRegisterChannel(p.RegisterChannel).SetBizTypes(p.BizTypes).
			SetBalance(p.Balance).SetRedPacketBalance(p.RedPacketBalance).
			SetRecentActiveAt(p.RecentActiveAt).SetUpdatedAt(p.UpdatedAt).
			Save(ctx)
		return err
	}
	return r.client.UserProfile.Create().
		SetUserID(p.UserID).SetLevel(p.Level).SetNickname(p.Nickname).
		SetRealName(p.RealName).SetVerified(p.Verified).SetRegion(p.Region).
		SetAddress(p.Address).SetRegisterChannel(p.RegisterChannel).SetBizTypes(p.BizTypes).
		SetBalance(p.Balance).SetRedPacketBalance(p.RedPacketBalance).
		SetRecentActiveAt(p.RecentActiveAt).SetUpdatedAt(p.UpdatedAt).
		Exec(ctx)
}

func toProfile(e *ent.UserProfile) *userprofile.Profile {
	if e == nil {
		return nil
	}
	return &userprofile.Profile{
		UserID:           e.UserID,
		Level:            e.Level,
		Nickname:         e.Nickname,
		RealName:         e.RealName,
		Verified:         e.Verified,
		Region:           e.Region,
		Address:          e.Address,
		RegisterChannel:  e.RegisterChannel,
		BizTypes:         e.BizTypes,
		Balance:          e.Balance,
		RedPacketBalance: e.RedPacketBalance,
		RecentActiveAt:   e.RecentActiveAt,
		UpdatedAt:        e.UpdatedAt,
	}
}

func (r ProfileRepo) FindByUserID(userID string) (*userprofile.Profile, error) {
	e, err := r.client.UserProfile.Query().
		Where(entuserprofile.UserID(userID)).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toProfile(e), nil
}
