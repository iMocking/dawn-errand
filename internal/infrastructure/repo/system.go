package repo

import (
	"context"
	"encoding/json"

	"dawn-errand/internal/domain/system"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/systemsetting"
)

// SettingRepo 实现 system.SettingRepository。
type SettingRepo struct{ client *ent.Client }

func (r SettingRepo) Get(key system.SettingKey) (*system.SystemSetting, error) {
	e, err := r.client.SystemSetting.Query().
		Where(systemsetting.Key(string(key))).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &system.SystemSetting{
		Key:       system.SettingKey(e.Key),
		Status:    system.SettingStatus(e.Status),
		ValueJSON: string(e.ValueJSON),
		UpdatedBy: e.UpdatedBy,
		UpdatedAt: e.UpdatedAt,
	}, nil
}

func (r SettingRepo) Upsert(s *system.SystemSetting) error {
	ctx := context.Background()
	exists, err := r.client.SystemSetting.Query().Where(systemsetting.Key(string(s.Key))).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err := r.client.SystemSetting.Update().
			Where(systemsetting.Key(string(s.Key))).
			SetStatus(string(s.Status)).SetValueJSON([]byte(s.ValueJSON)).
			SetUpdatedBy(s.UpdatedBy).SetUpdatedAt(s.UpdatedAt).
			Save(ctx)
		return err
	}
	return r.client.SystemSetting.Create().
		SetKey(string(s.Key)).SetStatus(string(s.Status)).
		SetValueJSON([]byte(s.ValueJSON)).
		SetUpdatedBy(s.UpdatedBy).SetUpdatedAt(s.UpdatedAt).
		Exec(ctx)
}

// GetBootstrap 读取初始化状态（无记录视为未初始化）。
func (r SettingRepo) GetBootstrap() (*system.BootstrapState, error) {
	s, err := r.Get(system.KeyBootstrap)
	if err != nil {
		if err == ErrNotFound {
			return &system.BootstrapState{Initialized: false}, nil
		}
		return nil, err
	}
	var st system.BootstrapState
	if err := json.Unmarshal([]byte(s.ValueJSON), &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// MarkInitialized 写入初始化完成标记（已存在则不覆盖）。
func (r SettingRepo) MarkInitialized(adminUserID string, now int64) error {
	cur, err := r.GetBootstrap()
	if err != nil {
		return err
	}
	if cur.Initialized {
		return nil
	}
	b, err := json.Marshal(system.BootstrapState{
		Initialized: true,
		AdminUserID: adminUserID,
		InitAt:      now,
	})
	if err != nil {
		return err
	}
	return r.Upsert(&system.SystemSetting{
		Key:       system.KeyBootstrap,
		Status:    system.SettingEnabled,
		ValueJSON: string(b),
		UpdatedBy: "system",
		UpdatedAt: now,
	})
}
