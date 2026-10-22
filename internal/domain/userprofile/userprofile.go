// Package userprofile 用户资料限界上下文（正式会员档案/资产）。
package userprofile

// Profile 用户资料聚合。
type Profile struct {
	UserID           string
	Level            string // 等级，如 Lv.5 乡镇常客
	Nickname         string
	RealName         string
	Verified         string // 已实名 / 未实名
	Region           string
	Address          string
	RegisterChannel  string
	BizTypes         string
	Balance          float64
	RedPacketBalance float64
	RecentActiveAt   int64
	UpdatedAt        int64
}

// ProfileRepository 用户资料仓储端口。
type ProfileRepository interface {
	Save(p *Profile) error
	FindByUserID(userID string) (*Profile, error)
}
