package repo

import (
	"context"
	"errors"
	"strings"

	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/tenant"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/credential"
	"dawn-errand/internal/infrastructure/ent/emailverificationtoken"
	"dawn-errand/internal/infrastructure/ent/passwordresettoken"
	"dawn-errand/internal/infrastructure/ent/user"
)

// UserRepo 实现 identity.UserRepository。
type UserRepo struct{ client *ent.Client }

func (r UserRepo) Save(u *identity.User) error {
	ctx := context.Background()
	if u.ID == "" {
		u.ID = newULID()
	}
	exists, err := r.client.User.Query().Where(user.EmailEqualFold(u.Email)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err := r.client.User.Update().
			Where(user.EmailEqualFold(u.Email)).
			SetPhone(u.Phone).SetName(u.Name).SetAvatar(u.Avatar).
			SetStatus(string(u.Status)).SetEmailVerified(u.EmailVerified).
			SetRole(u.Role).SetRegionID(u.RegionID).SetMerchantID(u.MerchantID).SetRiderID(u.RiderID).
			SetRegisteredAt(u.RegisteredAt).SetUpdatedAt(u.UpdatedAt).
			Save(ctx)
		return err
	}
	return r.client.User.Create().
		SetID(u.ID).SetEmail(u.Email).SetPhone(u.Phone).SetName(u.Name).
		SetAvatar(u.Avatar).SetStatus(string(u.Status)).SetEmailVerified(u.EmailVerified).
		SetRole(u.Role).SetRegionID(u.RegionID).SetMerchantID(u.MerchantID).SetRiderID(u.RiderID).
		SetRegisteredAt(u.RegisteredAt).SetUpdatedAt(u.UpdatedAt).
		Exec(ctx)
}

func toUser(e *ent.User) *identity.User {
	if e == nil {
		return nil
	}
	return &identity.User{
		ID:            e.ID,
		Email:         e.Email,
		Phone:         e.Phone,
		Name:          e.Name,
		Avatar:        e.Avatar,
		Status:        identity.UserStatus(e.Status),
		EmailVerified: e.EmailVerified,
		Role:          e.Role,
		RegionID:      e.RegionID,
		MerchantID:    e.MerchantID,
		RiderID:       e.RiderID,
		RegisteredAt:  e.RegisteredAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func (r UserRepo) FindByID(id string) (*identity.User, error) {
	e, err := r.client.User.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toUser(e), nil
}

func (r UserRepo) FindByEmail(email string) (*identity.User, error) {
	e, err := r.client.User.Query().Where(user.EmailEqualFold(email)).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toUser(e), nil
}

func (r UserRepo) FindByPhone(phone string) (*identity.User, error) {
	if phone == "" {
		return nil, ErrNotFound
	}
	e, err := r.client.User.Query().Where(user.PhoneEqualFold(phone)).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toUser(e), nil
}

func (r UserRepo) List(q identity.UserQuery, sc tenant.Scope) ([]*identity.User, int64, error) {
	ctx := context.Background()
	bq := r.client.User.Query()
	scopeUser(bq, sc)
	if q.Status != "" {
		bq.Where(user.Status(q.Status))
	}
	if k := strings.TrimSpace(q.Keyword); k != "" {
		bq.Where(user.Or(
			user.NameContainsFold(k),
			user.EmailContainsFold(k),
			user.PhoneContainsFold(k),
		))
	}
	total, err := bq.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(q.Page, q.PageSize)
	rows, err := bq.Order(ent.Desc(user.FieldRegisteredAt)).Offset((page - 1) * size).Limit(size).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*identity.User, 0, len(rows))
	for _, r := range rows {
		out = append(out, toUser(r))
	}
	return out, int64(total), nil
}

func (r UserRepo) UpdateStatus(id string, status identity.UserStatus) error {
	err := r.client.User.UpdateOneID(id).
		SetStatus(string(status)).SetUpdatedAt(nowSec()).
		Exec(context.Background())
	return err
}

func (r UserRepo) CountRegisteredSince(sc tenant.Scope, since int64) (int64, error) {
	bq := r.client.User.Query()
	scopeUser(bq, sc)
	n, err := bq.Where(user.RegisteredAtGTE(since)).Count(context.Background())
	return int64(n), err
}

func (r UserRepo) CountStatus(sc tenant.Scope, status identity.UserStatus) (int64, error) {
	bq := r.client.User.Query()
	scopeUser(bq, sc)
	n, err := bq.Where(user.Status(string(status))).Count(context.Background())
	return int64(n), err
}

// CredentialRepo 实现 identity.CredentialRepository。
type CredentialRepo struct{ client *ent.Client }

func (r CredentialRepo) Save(c *identity.Credential) error {
	ctx := context.Background()
	exists, err := r.client.Credential.Query().Where(credential.UserID(c.UserID)).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		_, err := r.client.Credential.Update().
			Where(credential.UserID(c.UserID)).
			SetPasswordHash(c.PasswordHash).SetSessionVersion(c.SessionVersion).SetUpdatedAt(c.UpdatedAt).
			Save(ctx)
		return err
	}
	return r.client.Credential.Create().
		SetUserID(c.UserID).SetPasswordHash(c.PasswordHash).
		SetSessionVersion(c.SessionVersion).SetUpdatedAt(c.UpdatedAt).
		Exec(ctx)
}

func (r CredentialRepo) FindByUserID(userID string) (*identity.Credential, error) {
	e, err := r.client.Credential.Query().Where(credential.UserID(userID)).Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &identity.Credential{
		UserID:         e.UserID,
		PasswordHash:   e.PasswordHash,
		SessionVersion: e.SessionVersion,
		UpdatedAt:      e.UpdatedAt,
	}, nil
}

func (r CredentialRepo) BumpSessionVersion(userID string) (int64, error) {
	ctx := context.Background()
	cur, err := r.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// 无凭证时创建初始版本为 1。
			err := r.client.Credential.Create().
				SetUserID(userID).SetPasswordHash("").SetSessionVersion(1).SetUpdatedAt(nowSec()).
				Exec(ctx)
			if err != nil {
				return 0, err
			}
			return 1, nil
		}
		return 0, err
	}
	next := cur.SessionVersion + 1
	err = r.client.Credential.Update().
		Where(credential.UserID(userID)).
		SetSessionVersion(next).SetUpdatedAt(nowSec()).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return next, nil
}

// EmailTokenRepo 实现 identity.EmailTokenRepository。
type EmailTokenRepo struct{ client *ent.Client }

func (r EmailTokenRepo) Create(t *identity.EmailVerificationToken) error {
	if t.ID == "" {
		t.ID = newULID()
	}
	return r.client.EmailVerificationToken.Create().
		SetID(t.ID).SetUserID(t.UserID).SetTokenHash(t.TokenHash).
		SetExpiresAt(t.ExpiresAt).SetUsedAt(t.UsedAt).
		Exec(context.Background())
}

func (r EmailTokenRepo) FindByHash(hash string) (*identity.EmailVerificationToken, error) {
	e, err := r.client.EmailVerificationToken.Query().
		Where(emailverificationtoken.TokenHash(hash), emailverificationtoken.UsedAt(0)).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &identity.EmailVerificationToken{
		ID: e.ID, UserID: e.UserID, TokenHash: e.TokenHash,
		ExpiresAt: e.ExpiresAt, UsedAt: e.UsedAt,
	}, nil
}

func (r EmailTokenRepo) MarkUsed(id string) error {
	return r.client.EmailVerificationToken.UpdateOneID(id).SetUsedAt(nowSec()).Exec(context.Background())
}

func (r EmailTokenRepo) DeleteExpired(now int64) error {
	_, err := r.client.EmailVerificationToken.Delete().
		Where(emailverificationtoken.ExpiresAtLT(now)).
		Exec(context.Background())
	return err
}

// PasswordTokenRepo 实现 identity.PasswordTokenRepository。
type PasswordTokenRepo struct{ client *ent.Client }

func (r PasswordTokenRepo) Create(t *identity.PasswordResetToken) error {
	if t.ID == "" {
		t.ID = newULID()
	}
	return r.client.PasswordResetToken.Create().
		SetID(t.ID).SetUserID(t.UserID).SetTokenHash(t.TokenHash).
		SetExpiresAt(t.ExpiresAt).SetUsedAt(t.UsedAt).
		Exec(context.Background())
}

func (r PasswordTokenRepo) FindByHash(hash string) (*identity.PasswordResetToken, error) {
	e, err := r.client.PasswordResetToken.Query().
		Where(passwordresettoken.TokenHash(hash), passwordresettoken.UsedAt(0)).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &identity.PasswordResetToken{
		ID: e.ID, UserID: e.UserID, TokenHash: e.TokenHash,
		ExpiresAt: e.ExpiresAt, UsedAt: e.UsedAt,
	}, nil
}

func (r PasswordTokenRepo) MarkUsed(id string) error {
	return r.client.PasswordResetToken.UpdateOneID(id).SetUsedAt(nowSec()).Exec(context.Background())
}

func (r PasswordTokenRepo) DeleteExpired(now int64) error {
	_, err := r.client.PasswordResetToken.Delete().
		Where(passwordresettoken.ExpiresAtLT(now)).
		Exec(context.Background())
	return err
}

func (r PasswordTokenRepo) DeleteByUser(userID string) error {
	_, err := r.client.PasswordResetToken.Delete().
		Where(passwordresettoken.UserID(userID)).
		Exec(context.Background())
	return err
}

func normalizePage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 200 {
		size = 200
	}
	return page, size
}
