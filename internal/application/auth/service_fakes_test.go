package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/config"
	"dawn-errand/internal/domain/identity"
	"dawn-errand/internal/domain/userprofile"
)

// errFakeConflict 模拟仓储在并发写入时的唯一键冲突（不是 application 的哨兵错误）。
var errFakeConflict = errors.New("fake conflict")

// 单元测试替身：只实现被测用例走到的方法，其余端口用「嵌入接口 + 空实现」占位，
// 一旦被测代码越界调用未实现方法就会 panic——隐式依赖会当场暴露，而不是等到上线。
// 这里刻意不做真实哈希（Hasher 用 "argon2:"+明文 的可逆替身），
// 单测关心的是「是否把校验结果当回事」，密码学强度由 passwd 包自己的测试负责。

// --- 用户 ---

type fakeUsers struct {
	identity.UserRepository
	byID    map[string]*identity.User
	byEmail map[string]*identity.User
	byPhone map[string]*identity.User
	saveErr error
	saves   int
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{
		byID:    map[string]*identity.User{},
		byEmail: map[string]*identity.User{},
		byPhone: map[string]*identity.User{},
	}
}

func (f *fakeUsers) Save(ctx context.Context, u *identity.User) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saves++
	f.byID[u.ID] = u
	if u.Email != "" {
		f.byEmail[strings.ToLower(u.Email)] = u
	}
	if u.Phone != "" {
		f.byPhone[u.Phone] = u
	}
	return nil
}

func (f *fakeUsers) FindByID(ctx context.Context, id string) (*identity.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return u, nil
}

func (f *fakeUsers) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	u, ok := f.byEmail[strings.ToLower(email)]
	if !ok {
		return nil, application.ErrNotFound
	}
	return u, nil
}

func (f *fakeUsers) FindByPhone(ctx context.Context, phone string) (*identity.User, error) {
	u, ok := f.byPhone[phone]
	if !ok {
		return nil, application.ErrNotFound
	}
	return u, nil
}

// --- 凭证 ---

type fakeCreds struct {
	identity.CredentialRepository
	byUser map[string]*identity.Credential
	// conflictSave 模拟并发创建：本次 Save 返回错误，但行已被「另一个请求」写入，
	// 用于验证 ensureCredential 的失败后回读分支。
	conflictSave bool
	bumps        int
}

func newFakeCreds() *fakeCreds {
	return &fakeCreds{byUser: map[string]*identity.Credential{}}
}

func (f *fakeCreds) Save(ctx context.Context, c *identity.Credential) error {
	if f.conflictSave {
		f.conflictSave = false
		f.byUser[c.UserID] = c
		return errFakeConflict
	}
	f.byUser[c.UserID] = c
	return nil
}

func (f *fakeCreds) FindByUserID(ctx context.Context, userID string) (*identity.Credential, error) {
	c, ok := f.byUser[userID]
	if !ok {
		return nil, application.ErrNotFound
	}
	return c, nil
}

func (f *fakeCreds) BumpSessionVersion(ctx context.Context, userID string) (int64, error) {
	c, ok := f.byUser[userID]
	if !ok {
		return 0, application.ErrNotFound
	}
	c.SessionVersion++
	f.bumps++
	return c.SessionVersion, nil
}

// --- 邮箱验证令牌 ---

type fakeEmailTokens struct {
	identity.EmailTokenRepository
	byHash    map[string]*identity.EmailVerificationToken
	created   []*identity.EmailVerificationToken
	createErr error
	markUsed  []string
}

func newFakeEmailTokens() *fakeEmailTokens {
	return &fakeEmailTokens{byHash: map[string]*identity.EmailVerificationToken{}}
}

func (f *fakeEmailTokens) Create(ctx context.Context, t *identity.EmailVerificationToken) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, t)
	f.byHash[t.TokenHash] = t
	return nil
}

func (f *fakeEmailTokens) FindByHash(ctx context.Context, hash string) (*identity.EmailVerificationToken, error) {
	t, ok := f.byHash[hash]
	if !ok {
		return nil, application.ErrNotFound
	}
	return t, nil
}

func (f *fakeEmailTokens) MarkUsed(ctx context.Context, id string) error {
	f.markUsed = append(f.markUsed, id)
	for _, t := range f.byHash {
		if t.ID == id {
			t.UsedAt = time.Now().Unix()
		}
	}
	return nil
}

// --- 密码重置令牌 ---

type fakePasswordTokens struct {
	identity.PasswordTokenRepository
	byHash    map[string]*identity.PasswordResetToken
	created   []*identity.PasswordResetToken
	createErr error
	markUsed  []string
	deletedBy []string
}

func newFakePasswordTokens() *fakePasswordTokens {
	return &fakePasswordTokens{byHash: map[string]*identity.PasswordResetToken{}}
}

func (f *fakePasswordTokens) Create(ctx context.Context, t *identity.PasswordResetToken) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, t)
	f.byHash[t.TokenHash] = t
	return nil
}

func (f *fakePasswordTokens) FindByHash(ctx context.Context, hash string) (*identity.PasswordResetToken, error) {
	t, ok := f.byHash[hash]
	if !ok {
		return nil, application.ErrNotFound
	}
	return t, nil
}

func (f *fakePasswordTokens) MarkUsed(ctx context.Context, id string) error {
	f.markUsed = append(f.markUsed, id)
	for _, t := range f.byHash {
		if t.ID == id {
			t.UsedAt = time.Now().Unix()
		}
	}
	return nil
}

func (f *fakePasswordTokens) DeleteByUser(ctx context.Context, userID string) error {
	f.deletedBy = append(f.deletedBy, userID)
	return nil
}

// --- 资料 ---

type fakeProfiles struct {
	userprofile.ProfileRepository
	saved []*userprofile.Profile
}

func (f *fakeProfiles) Save(ctx context.Context, p *userprofile.Profile) error {
	f.saved = append(f.saved, p)
	return nil
}

// --- 审计 ---

type fakeAudit struct {
	entries []application.AuditEntry
}

func (f *fakeAudit) Record(ctx context.Context, e application.AuditEntry) {
	f.entries = append(f.entries, e)
}

// --- 哈希 ---

type fakeHasher struct {
	hashErr error
	hashes  int
}

func (h *fakeHasher) Hash(password string) (string, error) {
	if h.hashErr != nil {
		return "", h.hashErr
	}
	h.hashes++
	return "argon2:" + password, nil
}

func (h *fakeHasher) Verify(hash, password string) bool { return hash == "argon2:"+password }

// --- 令牌 ---

type fakeTokens struct {
	application.TokenManager
	issueErr      error
	verifyErr     error
	accessIssued  []application.TokenClaims
	refreshIssued int
	revokedAll    []string
	revokedOne    [][2]string
}

func (t *fakeTokens) IssueAccessToken(claims application.TokenClaims, ttl time.Duration) (string, error) {
	if t.issueErr != nil {
		return "", t.issueErr
	}
	t.accessIssued = append(t.accessIssued, claims)
	return "at-for-" + claims.UserID, nil
}

func (t *fakeTokens) IssueRefreshToken(ctx context.Context, userID string, ttl time.Duration) (string, string, error) {
	if t.issueErr != nil {
		return "", "", t.issueErr
	}
	t.refreshIssued++
	return "rt-" + userID, "rth", nil
}

func (t *fakeTokens) UserIDForRefreshToken(ctx context.Context, token string) (string, error) {
	return strings.TrimPrefix(token, "rt-"), nil
}

func (t *fakeTokens) VerifyRefreshToken(ctx context.Context, userID, token string) error {
	return t.verifyErr
}

func (t *fakeTokens) RevokeRefreshTokens(ctx context.Context, userID string) error {
	t.revokedAll = append(t.revokedAll, userID)
	return nil
}

func (t *fakeTokens) RevokeRefreshToken(ctx context.Context, userID, token string) error {
	t.revokedOne = append(t.revokedOne, [2]string{userID, token})
	return nil
}

// --- 发信 ---

type mailEntry struct {
	to, subject, body string
}

type fakeMailer struct {
	application.MailerPort
	sent    []mailEntry
	sendErr error
}

func (m *fakeMailer) Send(ctx context.Context, to, subject, html string) error {
	m.sent = append(m.sent, mailEntry{to: to, subject: subject, body: html})
	return m.sendErr
}

// --- 验证码存储 ---

type fakeCodes struct {
	application.CodeStore
	values  map[string]string
	getErr  error
	setErr  error
	deleted []string
}

func newFakeCodes() *fakeCodes { return &fakeCodes{values: map[string]string{}} }

func (c *fakeCodes) Set(ctx context.Context, key, code string, ttl time.Duration) error {
	if c.setErr != nil {
		return c.setErr
	}
	c.values[key] = code
	return nil
}

func (c *fakeCodes) Get(ctx context.Context, key string) (string, error) {
	if c.getErr != nil {
		return "", c.getErr
	}
	return c.values[key], nil
}

func (c *fakeCodes) Delete(ctx context.Context, key string) error {
	c.deleted = append(c.deleted, key)
	delete(c.values, key)
	return nil
}

// --- 限流 ---

type fakeLimiter struct {
	allow bool
	err   error
	keys  []string
}

func (l *fakeLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	l.keys = append(l.keys, key)
	return l.allow, l.err
}

// --- 权限 ---

type fakeEnforcer struct {
	policies map[string][]string
}

func (e *fakeEnforcer) PermissionsForRole(role string) []string {
	if p, ok := e.policies[role]; ok {
		return p
	}
	return []string{"user:read"}
}

// --- 事务 ---

type fakeTx struct {
	unit     *application.RepoUnit
	runCount int
	rollback bool // true 时在回调返回错误后不保留写入（替身不做快照，仅用于断言调用）
}

func (t *fakeTx) RunInTx(ctx context.Context, fn func(*application.RepoUnit) error) error {
	t.runCount++
	return fn(t.unit)
}

// --- 装配 ---

type authUnit struct {
	repos    *application.RepoUnit
	tx       *fakeTx
	users    *fakeUsers
	creds    *fakeCreds
	emails   *fakeEmailTokens
	pwTokens *fakePasswordTokens
	profiles *fakeProfiles
	audit    *fakeAudit
	hasher   *fakeHasher
	tokens   *fakeTokens
	mailer   *fakeMailer
	codes    *fakeCodes
	limiter  *fakeLimiter
	enforcer *fakeEnforcer
	cfg      *config.Config
	svc      *Service
}

func newAuthUnit() *authUnit {
	u := &authUnit{
		users:    newFakeUsers(),
		creds:    newFakeCreds(),
		emails:   newFakeEmailTokens(),
		pwTokens: newFakePasswordTokens(),
		profiles: &fakeProfiles{},
		audit:    &fakeAudit{},
		hasher:   &fakeHasher{},
		tokens:   &fakeTokens{},
		mailer:   &fakeMailer{},
		codes:    newFakeCodes(),
		limiter:  &fakeLimiter{allow: true},
		enforcer: &fakeEnforcer{policies: map[string][]string{}},
		cfg:      &config.Config{AppEnv: "dev", AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: 24 * time.Hour},
	}
	u.repos = &application.RepoUnit{
		Users: u.users, Credentials: u.creds,
		EmailTokens: u.emails, PasswordTokens: u.pwTokens,
		Profiles: u.profiles, Audit: u.audit,
	}
	u.tx = &fakeTx{unit: u.repos}
	u.svc = &Service{
		Repos: u.repos, Tx: u.tx, Hasher: u.hasher, Tokens: u.tokens,
		Mailer: u.mailer, Cfg: u.cfg, Codes: u.codes, Enforcer: u.enforcer, Limiter: u.limiter,
	}
	return u
}

// addUser 造一个已注册用户；返回其凭证。
func (u *authUnit) addUser(id, email, password string, opts ...func(*identity.User)) *identity.User {
	usr := &identity.User{
		ID: id, Email: email, Name: "用户" + id, Status: identity.UserStatusActive,
		Role: identity.RoleUser, EmailVerified: true, RegisteredAt: 1, UpdatedAt: 1,
	}
	for _, opt := range opts {
		opt(usr)
	}
	u.users.byID[usr.ID] = usr
	if usr.Email != "" {
		u.users.byEmail[strings.ToLower(usr.Email)] = usr
	}
	if usr.Phone != "" {
		u.users.byPhone[usr.Phone] = usr
	}
	if password != "" {
		u.creds.byUser[id] = &identity.Credential{UserID: id, PasswordHash: "argon2:" + password, SessionVersion: 1}
	}
	return usr
}
