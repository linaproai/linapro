// This file verifies host tenant-aware authentication token transitions.

package auth

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/golang-jwt/jwt/v5"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// TestSelectTenantConsumesPreTokenOnce verifies pre-login tokens are single-use
// and tenant selection signs a tenant-bound formal JWT.
func TestSelectTenantConsumesPreTokenOnce(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	preToken, err := svc.preTokens.Create(ctx, preTokenRecord{
		UserID:   101,
		Username: "tenant-user",
		Status:   1,
	})
	if err != nil {
		t.Fatalf("create pre-token: %v", err)
	}

	out, err := svc.IssueTenantToken(ctx, TenantTokenIssueInput{PreToken: preToken, TenantID: 11})
	if err != nil {
		t.Fatalf("select tenant: %v", err)
	}
	claims, err := svc.ParseToken(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("parse selected token: %v", err)
	}
	if claims.TenantId != 11 || claims.UserId != 101 {
		t.Fatalf("expected selected tenant claims, got tenant=%d user=%d", claims.TenantId, claims.UserId)
	}

	_, err = svc.IssueTenantToken(ctx, TenantTokenIssueInput{PreToken: preToken, TenantID: 11})
	if !bizerr.Is(err, CodeAuthPreTokenInvalid) {
		t.Fatalf("expected consumed pre-token error, got %v", err)
	}
}

// TestPreTokenTTLIsShortAndEnforced verifies pre-login tokens use the expected
// short lifetime and expired records cannot be exchanged for a formal JWT.
func TestPreTokenTTLIsShortAndEnforced(t *testing.T) {
	ctx := context.Background()
	store := newMemoryPreTokenStore()
	preToken, err := store.Create(ctx, preTokenRecord{
		UserID:   101,
		Username: "tenant-user",
		Status:   1,
	})
	if err != nil {
		t.Fatalf("create pre-token: %v", err)
	}
	record := store.records[preToken]
	remaining := time.Until(record.ExpiresAt)
	if remaining <= 0 || remaining > preTokenTTL {
		t.Fatalf("expected short pre-token ttl <= %s and > 0, got %s", preTokenTTL, remaining)
	}

	record.ExpiresAt = time.Now().Add(-time.Second)
	store.records[preToken] = record
	svc := newTenantAuthTestService()
	svc.preTokens = store
	if _, err = svc.IssueTenantToken(ctx, TenantTokenIssueInput{PreToken: preToken, TenantID: 11}); !bizerr.Is(err, CodeAuthPreTokenInvalid) {
		t.Fatalf("expected expired pre-token error, got %v", err)
	}
}

// TestPreTokenSharedStoreConsumesAcrossInstances verifies that the shared
// token store enforces single-use semantics across auth service instances.
func TestPreTokenSharedStoreConsumesAcrossInstances(t *testing.T) {
	ctx := context.Background()
	sharedCache := newSharedMemoryKVCache()
	firstSvc := newTenantAuthTestService()
	secondSvc := newTenantAuthTestService()
	firstSvc.preTokens = newKVPreTokenStore(sharedCache)
	secondSvc.preTokens = newKVPreTokenStore(sharedCache)

	preToken, err := firstSvc.preTokens.Create(ctx, preTokenRecord{
		UserID:   101,
		Username: "tenant-user",
		Status:   1,
	})
	if err != nil {
		t.Fatalf("create shared pre-token: %v", err)
	}
	if _, err = secondSvc.IssueTenantToken(ctx, TenantTokenIssueInput{PreToken: preToken, TenantID: 11}); err != nil {
		t.Fatalf("select tenant from second instance: %v", err)
	}
	if _, err = firstSvc.IssueTenantToken(ctx, TenantTokenIssueInput{PreToken: preToken, TenantID: 11}); !bizerr.Is(err, CodeAuthPreTokenInvalid) {
		t.Fatalf("expected first instance to observe consumed pre-token, got %v", err)
	}
}

// TestRevokeLayeredStoreUsesLocalAndSharedState verifies revoke checks use a
// process-local memory layer and converge across instances through shared KV state.
func TestRevokeLayeredStoreUsesLocalAndSharedState(t *testing.T) {
	ctx := context.Background()
	sharedCache := newSharedMemoryKVCache()
	firstStore := newLayeredRevokeStore(newMemoryRevokeStore(), newKVRevokeStore(sharedCache))
	secondStore := newLayeredRevokeStore(newMemoryRevokeStore(), newKVRevokeStore(sharedCache))
	expiresAt := time.Now().Add(time.Hour)

	if err := firstStore.Add(ctx, "revoked-token", expiresAt); err != nil {
		t.Fatalf("add layered revoke: %v", err)
	}
	if revoked, err := firstStore.Revoked(ctx, "revoked-token"); err != nil || !revoked {
		t.Fatalf("expected first store local revoke hit, revoked=%v err=%v", revoked, err)
	}
	if revoked, err := secondStore.Revoked(ctx, "revoked-token"); err != nil || !revoked {
		t.Fatalf("expected second store shared revoke hit, revoked=%v err=%v", revoked, err)
	}
	if err := sharedCache.Delete(ctx, kvcache.OwnerTypeModule, revokeCacheKey("revoked-token")); err != nil {
		t.Fatalf("delete shared revoke state: %v", err)
	}
	if revoked, err := firstStore.Revoked(ctx, "revoked-token"); err != nil || !revoked {
		t.Fatalf("expected first store to keep local revoke after shared delete, revoked=%v err=%v", revoked, err)
	}
	if revoked, err := secondStore.Revoked(ctx, "revoked-token"); err != nil || !revoked {
		t.Fatalf("expected second store to backfill local revoke after shared delete, revoked=%v err=%v", revoked, err)
	}
}

// TestSwitchTenantRevokesOldToken verifies switching tenant invalidates the old
// token and signs a new token for the requested tenant.
func TestSwitchTenantRevokesOldToken(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	oldToken, oldTokenID, err := svc.generateToken(ctx, user, 11)
	if err != nil {
		t.Fatalf("generate old token: %v", err)
	}
	oldClaims, err := svc.ParseToken(ctx, oldToken)
	if err != nil {
		t.Fatalf("parse old token: %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{TokenId: oldTokenID, TenantId: 11, UserId: 101, Username: "tenant-user"}); err != nil {
		t.Fatalf("set old session: %v", err)
	}

	out, err := svc.ReissueTenantToken(ctx, TenantTokenReissueInput{CurrentClaims: oldClaims, TenantID: 22})
	if err != nil {
		t.Fatalf("switch tenant: %v", err)
	}
	if _, err = svc.ParseToken(ctx, oldToken); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected old token to be revoked, got %v", err)
	}
	newClaims, err := svc.ParseToken(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("parse new token: %v", err)
	}
	if newClaims.TenantId != 22 {
		t.Fatalf("expected new tenant 22, got %d", newClaims.TenantId)
	}
}

// TestLoginSelectTenantSwitchTenantLogoutFlow verifies the tenant auth
// lifecycle from password login through tenant selection, switching, and logout.
func TestLoginSelectTenantSwitchTenantLogoutFlow(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	svc.tenantSvc = enabledTenantAuthTestService{}

	username := fmt.Sprintf("tenant-flow-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	registerTenantAuthTestProvider(t, map[int][]pkgtenantcap.TenantInfo{
		userID: {
			{ID: 11, Code: "tenant-a", Name: "Tenant A", Status: "enabled"},
			{ID: 22, Code: "tenant-b", Name: "Tenant B", Status: "enabled"},
		},
	})

	loginOut, err := svc.Login(ctx, LoginInput{Username: username, Password: "admin123"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loginOut.AccessToken != "" {
		t.Fatal("expected two-stage login without formal access token")
	}
	if loginOut.PreToken == "" || len(loginOut.Tenants) != 2 {
		t.Fatalf("expected pre-token and tenant candidates, got preToken=%q tenants=%d", loginOut.PreToken, len(loginOut.Tenants))
	}

	selectOut, err := svc.IssueTenantToken(ctx, TenantTokenIssueInput{PreToken: loginOut.PreToken, TenantID: 11})
	if err != nil {
		t.Fatalf("select tenant: %v", err)
	}
	selectedClaims, err := svc.ParseToken(ctx, selectOut.AccessToken)
	if err != nil {
		t.Fatalf("parse selected token: %v", err)
	}
	if selectedClaims.TenantId != 11 || selectedClaims.UserId != userID {
		t.Fatalf("expected selected tenant/user claims, got tenant=%d user=%d", selectedClaims.TenantId, selectedClaims.UserId)
	}
	if active, err := svc.sessionStore.TouchOrValidate(ctx, 11, selectedClaims.TokenId, time.Hour); err != nil || !active {
		t.Fatalf("expected selected tenant session, active=%v err=%v", active, err)
	}

	switchOut, err := svc.ReissueTenantToken(ctx, TenantTokenReissueInput{CurrentClaims: selectedClaims, TenantID: 22})
	if err != nil {
		t.Fatalf("switch tenant: %v", err)
	}
	if _, err = svc.ParseToken(ctx, selectOut.AccessToken); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected selected token revoked after switch, got %v", err)
	}
	switchedClaims, err := svc.ParseToken(ctx, switchOut.AccessToken)
	if err != nil {
		t.Fatalf("parse switched token: %v", err)
	}
	if switchedClaims.TenantId != 22 || switchedClaims.UserId != userID {
		t.Fatalf("expected switched tenant/user claims, got tenant=%d user=%d", switchedClaims.TenantId, switchedClaims.UserId)
	}
	if active, err := svc.sessionStore.TouchOrValidate(ctx, 11, selectedClaims.TokenId, time.Hour); err != nil || active {
		t.Fatalf("expected selected tenant session removed, active=%v err=%v", active, err)
	}
	if active, err := svc.sessionStore.TouchOrValidate(ctx, 22, switchedClaims.TokenId, time.Hour); err != nil || !active {
		t.Fatalf("expected switched tenant session, active=%v err=%v", active, err)
	}

	svc.Logout(ctx, username, switchedClaims.TenantId, switchedClaims.TokenId)
	if active, err := svc.sessionStore.TouchOrValidate(ctx, 22, switchedClaims.TokenId, time.Hour); err != nil || active {
		t.Fatalf("expected switched tenant session removed after logout, active=%v err=%v", active, err)
	}
}

// TestLoginRejectsTenantUserWithoutActiveTenant verifies suspended or archived
// tenant-only users cannot silently fall back to a platform token.
func TestLoginRejectsTenantUserWithoutActiveTenant(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	svc.tenantSvc = enabledTenantAuthTestService{}

	username := fmt.Sprintf("tenant-unavailable-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	if _, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: userID}).
		Data(do.SysUser{TenantId: 11}).
		Update(); err != nil {
		t.Fatalf("set tenant id on auth test user: %v", err)
	}
	registerTenantAuthTestProvider(t, map[int][]pkgtenantcap.TenantInfo{userID: {}})

	if _, err := svc.Login(ctx, LoginInput{Username: username, Password: "admin123"}); !bizerr.Is(err, CodeAuthTenantUnavailable) {
		t.Fatalf("expected tenant unavailable login error, got %v", err)
	}
}

// TestRevokeSharedStoreInvalidatesAcrossInstances verifies that one auth
// instance can revoke a JWT and another instance rejects it through shared state.
func TestRevokeSharedStoreInvalidatesAcrossInstances(t *testing.T) {
	ctx := context.Background()
	sharedCache := newSharedMemoryKVCache()
	firstSvc := newTenantAuthTestService()
	secondSvc := newTenantAuthTestService()
	firstSvc.revoked = newKVRevokeStore(sharedCache)
	secondSvc.revoked = newKVRevokeStore(sharedCache)
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	token, tokenID, err := firstSvc.generateToken(ctx, user, 11)
	if err != nil {
		t.Fatalf("generate shared revoke token: %v", err)
	}
	claims, err := firstSvc.ParseToken(ctx, token)
	if err != nil {
		t.Fatalf("parse shared revoke token before revoke: %v", err)
	}
	if claims.TokenId != tokenID {
		t.Fatalf("expected generated token id %q, got %q", tokenID, claims.TokenId)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("expected token expiration")
	}
	if err = firstSvc.revoked.Add(ctx, claims.TokenId, claims.ExpiresAt.Time); err != nil {
		t.Fatalf("add shared revoke state: %v", err)
	}
	if _, err = secondSvc.ParseToken(ctx, token); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected second instance to reject revoked token, got %v", err)
	}
}

// TestLogoutRevokesCurrentToken verifies logout removes the supplied token from
// the session store contract.
func TestLogoutRevokesCurrentToken(t *testing.T) {
	ctx := context.Background()
	store := newMemorySessionStore()
	svc := newTenantAuthTestService()
	svc.sessionStore = store

	svc.Logout(ctx, "tenant-user", 22, "token-22")
	if store.deletedTokenID != "token-22" {
		t.Fatalf("expected token revoke, got token=%q", store.deletedTokenID)
	}
}

// TestMemorySessionStoreUsesGlobalTokenIdentity verifies the auth test helper
// mirrors the production globally unique token_id session-store contract.
func TestMemorySessionStoreUsesGlobalTokenIdentity(t *testing.T) {
	ctx := context.Background()
	store := newMemorySessionStore()
	if err := store.Set(ctx, &session.Session{TokenId: "same-token", TenantId: 11, UserId: 101}); err != nil {
		t.Fatalf("set tenant 11 session: %v", err)
	}
	if err := store.Set(ctx, &session.Session{TokenId: "same-token", TenantId: 22, UserId: 101}); err != nil {
		t.Fatalf("replace session by token: %v", err)
	}
	if item, err := store.Get(ctx, "same-token"); err != nil || item == nil || item.TenantId != 22 {
		t.Fatalf("expected latest token session with tenant 22, item=%v err=%v", item, err)
	}
	if active, err := store.TouchOrValidate(ctx, 11, "same-token", time.Hour); err != nil || active {
		t.Fatalf("expected tenant 11 mismatch to be invalid, active=%v err=%v", active, err)
	}
	if active, err := store.TouchOrValidate(ctx, 22, "same-token", time.Hour); err != nil || !active {
		t.Fatalf("expected tenant 22 session to remain active, active=%v err=%v", active, err)
	}
	if err := store.Delete(ctx, "same-token"); err != nil {
		t.Fatalf("delete session by token: %v", err)
	}
	if item, err := store.Get(ctx, "same-token"); err != nil || item != nil {
		t.Fatalf("expected token session deleted, item=%v err=%v", item, err)
	}
}

// newTenantAuthTestService returns a service with in-memory session state.
func newTenantAuthTestService() *serviceImpl {
	return &serviceImpl{
		configSvc:    configTestService{},
		roleSvc:      roleTestService{},
		sessionStore: newMemorySessionStore(),
		preTokens:    newMemoryPreTokenStore(),
		revoked:      newMemoryRevokeStore(),
	}
}

// configTestService provides JWT settings used by auth unit tests.
type configTestService struct{}

// GetJwtSecret returns a stable test signing secret.
func (configTestService) GetJwtSecret(context.Context) string {
	return "tenant-auth-test-secret"
}

// GetJwtExpire returns a stable test token lifetime.
func (configTestService) GetJwtExpire(context.Context) (time.Duration, error) {
	return time.Hour, nil
}

// IsLoginIPBlacklisted reports no blacklist entries in auth unit tests.
func (configTestService) IsLoginIPBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}

// roleTestService stubs the token access cache hooks used by auth.
type roleTestService struct{}

// PrimeTokenAccessContext returns a no-op access snapshot.
func (roleTestService) PrimeTokenAccessContext(context.Context, string, int) (*role.UserAccessContext, error) {
	return &role.UserAccessContext{}, nil
}

// InvalidateTokenAccessContext records no state in auth unit tests.
func (roleTestService) InvalidateTokenAccessContext(context.Context, string) {}

// enabledTenantAuthTestService enables tenant provider validation for auth tests.
type enabledTenantAuthTestService struct{}

// Enabled reports multi-tenancy as enabled.
func (enabledTenantAuthTestService) Enabled(context.Context) bool {
	return true
}

// Current returns the platform tenant for tests that do not carry request context.
func (enabledTenantAuthTestService) Current(context.Context) tenantcapsvc.TenantID {
	return pkgtenantcap.PLATFORM
}

// Apply returns the model unchanged in auth tests.
func (enabledTenantAuthTestService) Apply(_ context.Context, model *gdb.Model, _ string) (*gdb.Model, error) {
	return model, nil
}

// PlatformBypass reports no platform bypass in auth tests.
func (enabledTenantAuthTestService) PlatformBypass(context.Context) bool {
	return false
}

// EnsureTenantVisible accepts all tenants in auth tests.
func (enabledTenantAuthTestService) EnsureTenantVisible(context.Context, tenantcapsvc.TenantID) error {
	return nil
}

// ResolveTenant returns no request-derived tenant in auth tests.
func (enabledTenantAuthTestService) ResolveTenant(context.Context, *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
}

// ListUserTenants returns no tenants in auth tests unless provider lookup is used directly.
func (enabledTenantAuthTestService) ListUserTenants(context.Context, int) ([]pkgtenantcap.TenantInfo, error) {
	return []pkgtenantcap.TenantInfo{}, nil
}

// ReadWithPlatformFallback returns no rows in auth tests.
func (enabledTenantAuthTestService) ReadWithPlatformFallback(context.Context, tenantcapsvc.FallbackScanner[any]) ([]any, error) {
	return nil, nil
}

// ApplyUserTenantScope returns the model unchanged in auth tests.
func (enabledTenantAuthTestService) ApplyUserTenantScope(
	_ context.Context,
	model *gdb.Model,
	_ string,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ApplyUserTenantFilter returns the model unchanged in auth tests.
func (enabledTenantAuthTestService) ApplyUserTenantFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ tenantcapsvc.TenantID,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenantProjections returns no projections in auth tests.
func (enabledTenantAuthTestService) ListUserTenantProjections(
	context.Context,
	[]int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	return map[int]*pkgtenantcap.UserTenantProjection{}, nil
}

// ResolveUserTenantAssignment returns an empty plan in auth tests.
func (enabledTenantAuthTestService) ResolveUserTenantAssignment(
	context.Context,
	[]tenantcapsvc.TenantID,
	pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	return &pkgtenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is a no-op in auth tests.
func (enabledTenantAuthTestService) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*pkgtenantcap.UserTenantAssignmentPlan,
) error {
	return nil
}

// EnsureUsersInTenant accepts all users in auth tests.
func (enabledTenantAuthTestService) EnsureUsersInTenant(context.Context, []int, tenantcapsvc.TenantID) error {
	return nil
}

// ValidateUserMembershipStartupConsistency returns no details in auth tests.
func (enabledTenantAuthTestService) ValidateUserMembershipStartupConsistency(context.Context) ([]string, error) {
	return nil, nil
}

// tenantAuthTestProvider provides deterministic tenant memberships for auth tests.
type tenantAuthTestProvider struct {
	tenantsByUser map[int][]pkgtenantcap.TenantInfo
}

// ResolveTenant returns no request-derived tenant in auth tests.
func (p *tenantAuthTestProvider) ResolveTenant(context.Context, *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant verifies the user is a member of the requested tenant.
func (p *tenantAuthTestProvider) ValidateUserInTenant(_ context.Context, userID int, tenantID pkgtenantcap.TenantID) error {
	for _, tenant := range p.tenantsByUser[userID] {
		if tenant.ID == tenantID {
			return nil
		}
	}
	return bizerr.NewCode(CodeAuthTokenInvalid)
}

// ListUserTenants returns the configured user tenants.
func (p *tenantAuthTestProvider) ListUserTenants(_ context.Context, userID int) ([]pkgtenantcap.TenantInfo, error) {
	tenants := p.tenantsByUser[userID]
	result := make([]pkgtenantcap.TenantInfo, len(tenants))
	copy(result, tenants)
	return result, nil
}

// SwitchTenant verifies the target tenant membership.
func (p *tenantAuthTestProvider) SwitchTenant(ctx context.Context, userID int, target pkgtenantcap.TenantID) error {
	return p.ValidateUserInTenant(ctx, userID, target)
}

// registerTenantAuthTestProvider installs a temporary tenant provider.
func registerTenantAuthTestProvider(t *testing.T, tenantsByUser map[int][]pkgtenantcap.TenantInfo) {
	t.Helper()

	previous := pkgtenantcap.CurrentProvider()
	pkgtenantcap.RegisterProvider(&tenantAuthTestProvider{tenantsByUser: tenantsByUser})
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(previous)
	})
}

// insertAuthTestUser inserts one enabled user and cleans it up after the test.
func insertAuthTestUser(t *testing.T, ctx context.Context, username string, password string) int {
	t.Helper()

	hash, err := newTenantAuthTestService().HashPassword(password)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: hash,
		Nickname: username,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert auth test user: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysUser.Ctx(ctx).Unscoped().Where(do.SysUser{Id: id}).Delete(); cleanupErr != nil {
			t.Fatalf("cleanup auth test user: %v", cleanupErr)
		}
	})
	return int(id)
}

// memorySessionStore is an in-memory session store for auth unit tests.
type memorySessionStore struct {
	items          map[string]*session.Session
	deletedTokenID string
}

// sharedMemoryKVCache is a kvcache backend test double shared by multiple auth
// service instances.
type sharedMemoryKVCache struct {
	mu      sync.Mutex
	items   map[string]*kvcache.Item
	expires map[string]time.Time
}

// newSharedMemoryKVCache creates an empty shared kvcache test double.
func newSharedMemoryKVCache() *sharedMemoryKVCache {
	return &sharedMemoryKVCache{
		items:   make(map[string]*kvcache.Item),
		expires: make(map[string]time.Time),
	}
}

// BackendName returns the test backend name.
func (s *sharedMemoryKVCache) BackendName() kvcache.BackendName {
	return kvcache.BackendName("memory-test")
}

// RequiresExpiredCleanup reports no external cleanup requirement.
func (s *sharedMemoryKVCache) RequiresExpiredCleanup() bool {
	return false
}

// Get returns one unexpired item.
func (s *sharedMemoryKVCache) Get(_ context.Context, ownerType kvcache.OwnerType, cacheKey string) (*kvcache.Item, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := testKVKey(ownerType, cacheKey)
	if s.isExpiredLocked(key) {
		delete(s.items, key)
		delete(s.expires, key)
		return nil, false, nil
	}
	item := s.items[key]
	if item == nil {
		return nil, false, nil
	}
	copied := *item
	return &copied, true, nil
}

// GetInt returns one unexpired integer item.
func (s *sharedMemoryKVCache) GetInt(ctx context.Context, ownerType kvcache.OwnerType, cacheKey string) (int64, bool, error) {
	item, ok, err := s.Get(ctx, ownerType, cacheKey)
	if err != nil || !ok {
		return 0, ok, err
	}
	return item.IntValue, true, nil
}

// Set stores one string item.
func (s *sharedMemoryKVCache) Set(_ context.Context, ownerType kvcache.OwnerType, cacheKey string, value string, ttl time.Duration) (*kvcache.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := testKVKey(ownerType, cacheKey)
	item := &kvcache.Item{Key: cacheKey, ValueKind: kvcache.ValueKindString, Value: value}
	s.items[key] = item
	s.storeExpireLocked(key, item, ttl)
	copied := *item
	return &copied, nil
}

// Delete removes one item.
func (s *sharedMemoryKVCache) Delete(_ context.Context, ownerType kvcache.OwnerType, cacheKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := testKVKey(ownerType, cacheKey)
	delete(s.items, key)
	delete(s.expires, key)
	return nil
}

// Incr increments one integer item.
func (s *sharedMemoryKVCache) Incr(_ context.Context, ownerType kvcache.OwnerType, cacheKey string, delta int64, ttl time.Duration) (*kvcache.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := testKVKey(ownerType, cacheKey)
	if s.isExpiredLocked(key) {
		delete(s.items, key)
		delete(s.expires, key)
	}
	item := s.items[key]
	if item == nil {
		item = &kvcache.Item{Key: cacheKey, ValueKind: kvcache.ValueKindInt}
		s.items[key] = item
	}
	item.ValueKind = kvcache.ValueKindInt
	item.IntValue += delta
	item.Value = strconv.FormatInt(item.IntValue, 10)
	if ttl > 0 {
		s.storeExpireLocked(key, item, ttl)
	}
	copied := *item
	return &copied, nil
}

// Expire updates one item expiration.
func (s *sharedMemoryKVCache) Expire(_ context.Context, ownerType kvcache.OwnerType, cacheKey string, ttl time.Duration) (bool, *gtime.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := testKVKey(ownerType, cacheKey)
	item := s.items[key]
	if item == nil {
		return false, nil, nil
	}
	s.storeExpireLocked(key, item, ttl)
	return true, item.ExpireAt, nil
}

// CleanupExpired removes expired test entries.
func (s *sharedMemoryKVCache) CleanupExpired(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.items {
		if s.isExpiredLocked(key) {
			delete(s.items, key)
			delete(s.expires, key)
		}
	}
	return nil
}

// storeExpireLocked stores expiration metadata. Caller must hold s.mu.
func (s *sharedMemoryKVCache) storeExpireLocked(key string, item *kvcache.Item, ttl time.Duration) {
	if ttl <= 0 {
		item.ExpireAt = nil
		delete(s.expires, key)
		return
	}
	expireAt := time.Now().Add(ttl)
	s.expires[key] = expireAt
	item.ExpireAt = gtime.New(expireAt)
}

// isExpiredLocked reports whether one item is expired. Caller must hold s.mu.
func (s *sharedMemoryKVCache) isExpiredLocked(key string) bool {
	expireAt, ok := s.expires[key]
	return ok && time.Now().After(expireAt)
}

// testKVKey scopes memory entries by owner type and encoded key.
func testKVKey(ownerType kvcache.OwnerType, cacheKey string) string {
	return ownerType.String() + ":" + cacheKey
}

// newMemorySessionStore creates an empty in-memory session store.
func newMemorySessionStore() *memorySessionStore {
	return &memorySessionStore{items: make(map[string]*session.Session)}
}

// Set persists one session in memory.
func (s *memorySessionStore) Set(_ context.Context, sessionItem *session.Session) error {
	s.items[sessionItem.TokenId] = sessionItem
	return nil
}

// Get returns one session by token ID.
func (s *memorySessionStore) Get(_ context.Context, tokenID string) (*session.Session, error) {
	return s.items[tokenID], nil
}

// Delete records and removes one token.
func (s *memorySessionStore) Delete(_ context.Context, tokenID string) error {
	s.deletedTokenID = tokenID
	delete(s.items, tokenID)
	return nil
}

// DeleteByUserId removes matching sessions for a tenant/user pair.
func (s *memorySessionStore) DeleteByUserId(_ context.Context, tenantID int, userID int) error {
	for key, item := range s.items {
		if item.TenantId == tenantID && item.UserId == userID {
			delete(s.items, key)
		}
	}
	return nil
}

// List returns all sessions.
func (s *memorySessionStore) List(context.Context, *session.ListFilter) ([]*session.Session, error) {
	items := make([]*session.Session, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items, nil
}

// ListPage returns all sessions in one page.
func (s *memorySessionStore) ListPage(context.Context, *session.ListFilter, int, int) (*session.ListResult, error) {
	items, err := s.List(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return &session.ListResult{Items: items, Total: len(items)}, nil
}

// ListPageScoped returns all sessions in one page.
func (s *memorySessionStore) ListPageScoped(
	context.Context,
	*session.ListFilter,
	int,
	int,
	datascope.Service,
	tenantcapsvc.Service,
) (*session.ListResult, error) {
	items, err := s.List(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return &session.ListResult{Items: items, Total: len(items)}, nil
}

// Count returns the number of sessions.
func (s *memorySessionStore) Count(context.Context) (int, error) {
	return len(s.items), nil
}

// TouchOrValidate reports whether the token exists for the expected tenant.
func (s *memorySessionStore) TouchOrValidate(_ context.Context, tenantID int, tokenID string, _ time.Duration) (bool, error) {
	item := s.items[tokenID]
	return item != nil && item.TenantId == tenantID, nil
}

// CleanupInactive is a no-op for auth unit tests.
func (s *memorySessionStore) CleanupInactive(context.Context, time.Duration) (int64, error) {
	return 0, nil
}

// Interface guards keep the fakes aligned with auth dependencies.
var (
	_ interface {
		GetJwtSecret(context.Context) string
		GetJwtExpire(context.Context) (time.Duration, error)
	} = configTestService{}
	_ interface {
		PrimeTokenAccessContext(context.Context, string, int) (*role.UserAccessContext, error)
		InvalidateTokenAccessContext(context.Context, string)
	} = roleTestService{}
	_ session.Store         = (*memorySessionStore)(nil)
	_ kvcache.Service       = (*sharedMemoryKVCache)(nil)
	_ jwt.Claims            = (*Claims)(nil)
	_ tenantcapsvc.Service  = enabledTenantAuthTestService{}
	_ pkgtenantcap.Provider = (*tenantAuthTestProvider)(nil)
)
