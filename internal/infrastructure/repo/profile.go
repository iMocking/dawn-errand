package repo

import (
	"context"

	"dawn-errand/internal/domain/userprofile"
	"dawn-errand/internal/infrastructure/ent"
	entuserprofile "dawn-errand/internal/infrastructure/ent/userprofile"
)

// ProfileRepo 实现 userprofile.ProfileRepository。
type ProfileRepo struct{ client *ent.Client }

func (r ProfileRepo) Save(ctx context.Context, p *userprofile.Profile) error {
	exists, err := r.client.UserProfile.Query().Where(entuserprofile.UserID(p.UserID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		// 余额字段不经 Save 覆盖：唯一写路径是 AddBalance 原子更新，防 Find-then-Save 双花。
		_, err := r.client.UserProfile.Update().
			Where(entuserprofile.UserID(p.UserID)).
			SetLevel(p.Level).SetNickname(p.Nickname).SetRealName(p.RealName).SetVerified(p.Verified).
			SetRegion(p.Region).SetAddress(p.Address).SetRegisterChannel(p.RegisterChannel).SetBizTypes(p.BizTypes).
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

func (r ProfileRepo) FindByUserID(ctx context.Context, userID string) (*userprofile.Profile, error) {
	e, err := r.client.UserProfile.Query().
		Where(entuserprofile.UserID(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toProfile(e), nil
}

// AddBalance 原子增减余额：单条 UPDATE ... SET balance = balance + delta [WHERE balance >= min]。
// minBalance >= 0 时附加余额下限条件；profile 不存在时 Update 不命中，affected=0。
func (r ProfileRepo) AddBalance(ctx context.Context, userID string, delta, minBalance float64, now int64) (int, error) {
	q := r.client.UserProfile.Update().Where(entuserprofile.UserID(userID))
	if minBalance >= 0 {
		q = q.Where(entuserprofile.BalanceGTE(minBalance))
	}
	res, err := q.AddBalance(delta).SetUpdatedAt(now).Save(ctx)
	return res, err
}
