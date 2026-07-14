// This file verifies auth_impl.go behavior: tenant-aware token transitions
// and login policy driven by managed sys_config runtime parameters.

package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/golang-jwt/jwt/v5"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginhost"
)

// TestSelectTenantConsumesPreTokenOnce verifies pre-login tokens are single-use
// and tenant selection signs a tenant-bound formal JWT.
func TestSelectTenantConsumesPreTokenOnce(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	preToken, err := svc.preTokens.Create(ctx, preTokenRecord{
		UserID:     101,
		Username:   "tenant-user",
		Status:     1,
		ClientType: tokencap.ClientTypeWeb,
	})
	if err != nil {
		t.Fatalf("create pre-token: %v", err)
	}

	out, err := svc.IssueTenantToken(ctx, TenantTokenIssueInput{PreToken: preToken, TenantID: 11})
	if err != nil {
		t.Fatalf("select tenant: %v", err)
	}
	if out.RefreshToken == "" {
		t.Fatal("expected selected tenant refresh token")
	}
	claims, err := svc.parseAccessTokenForTest(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("parse selected token: %v", err)
	}
	if claims.TenantId != 11 || claims.UserId != 101 {
		t.Fatalf("expected selected tenant claims, got tenant=%d user=%d", claims.TenantId, claims.UserId)
	}
	if claims.ClientType != tokencap.ClientTypeWeb {
		t.Fatalf("expected selected tenant clientType %q, got %q", tokencap.ClientTypeWeb, claims.ClientType)
	}

	_, err = svc.IssueTenantToken(ctx, TenantTokenIssueInput{PreToken: preToken, TenantID: 11})
	if !bizerr.Is(err, CodeAuthPreTokenInvalid) {
		t.Fatalf("expected consumed pre-token error, got %v", err)
	}
}

// TestIssueTenantTokenPrimesAccessContextWithSelectedTenant verifies tenant
// selection primes role access with the selected tenant instead of the caller
// context tenant.
func TestIssueTenantTokenPrimesAccessContextWithSelectedTenant(t *testing.T) {
	var (
		ctx     = datascope.WithTenantScope(context.Background(), 99)
		svc     = newTenantAuthTestService()
		roleSvc = &trackingRoleTestService{}
	)
	svc.roleSvc = roleSvc

	preToken, err := svc.preTokens.Create(ctx, preTokenRecord{
		UserID:     101,
		Username:   "tenant-user",
		Status:     1,
		ClientType: tokencap.ClientTypeWeb,
	})
	if err != nil {
		t.Fatalf("create pre-token: %v", err)
	}
	if _, err = svc.IssueTenantToken(ctx, TenantTokenIssueInput{PreToken: preToken, TenantID: 11}); err != nil {
		t.Fatalf("select tenant: %v", err)
	}

	if len(roleSvc.tenantIDs) != 1 || roleSvc.tenantIDs[0] != 11 {
		t.Fatalf("expected role cache prime for tenant 11, got %v", roleSvc.tenantIDs)
	}
}

// TestIssueImpersonationTokenUsesHostSignerAndTenantScopedPrime verifies
// impersonation tokens are host-owned and permission priming receives the
// target tenant plus impersonation business context.
func TestIssueImpersonationTokenUsesHostSignerAndTenantScopedPrime(t *testing.T) {
	var (
		ctx     = context.WithValue(context.Background(), bizctx.ContextKey, &model.Context{ClientType: tokencap.ClientTypeDesktop.String()})
		svc     = newTenantAuthTestService()
		roleSvc = &trackingRoleTestService{}
	)
	svc.roleSvc = roleSvc
	username := fmt.Sprintf("impersonation-admin-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")

	out, err := svc.IssueImpersonationToken(ctx, ImpersonationTokenIssueInput{ActingUserID: userID, TenantID: 42})
	if err != nil {
		t.Fatalf("issue impersonation token: %v", err)
	}
	if out.AccessToken == "" || out.TokenID == "" || out.TenantID != 42 || out.ActingUserID != userID {
		t.Fatalf("unexpected impersonation output: %#v", out)
	}
	claims, err := svc.parseAccessTokenForTest(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("parse impersonation token: %v", err)
	}
	if !claims.IsImpersonation || claims.ActingUserId != userID || claims.UserId != userID || claims.TenantId != 42 || claims.TokenId != out.TokenID {
		t.Fatalf("unexpected impersonation claims: %#v", claims)
	}
	if claims.ClientType != tokencap.ClientTypeDesktop {
		t.Fatalf("expected impersonation clientType %q, got %q", tokencap.ClientTypeDesktop, claims.ClientType)
	}
	if sessionItem, err := svc.sessionStore.Get(ctx, out.TokenID); err != nil || sessionItem == nil || sessionItem.TenantId != 42 || sessionItem.UserId != userID || sessionItem.ClientType != tokencap.ClientTypeDesktop.String() {
		t.Fatalf("expected impersonation session in target tenant, session=%#v err=%v", sessionItem, err)
	}
	if len(roleSvc.tenantIDs) != 1 || roleSvc.tenantIDs[0] != 42 {
		t.Fatalf("expected role cache prime under target tenant, got %v", roleSvc.tenantIDs)
	}
	if len(roleSvc.contexts) != 1 || roleSvc.contexts[0] == nil {
		t.Fatalf("expected impersonation business context, got %#v", roleSvc.contexts)
	}
	if !roleSvc.contexts[0].IsImpersonation ||
		!roleSvc.contexts[0].ActingAsTenant ||
		roleSvc.contexts[0].ActingUserId != userID ||
		roleSvc.contexts[0].TenantId != 42 ||
		roleSvc.contexts[0].ClientType != tokencap.ClientTypeDesktop.String() {
		t.Fatalf("unexpected impersonation business context: %#v", roleSvc.contexts[0])
	}

	if err = svc.RevokeImpersonationToken(ctx, "Bearer "+out.AccessToken, 42); err != nil {
		t.Fatalf("revoke impersonation token: %v", err)
	}
	if _, err = svc.parseAccessTokenForTest(ctx, out.AccessToken); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected revoked impersonation token to be invalid, got %v", err)
	}
}

// TestRevokeImpersonationTokenRejectsNonImpersonationToken verifies plugins
// cannot use the impersonation revoke path to tear down ordinary sessions.
func TestRevokeImpersonationTokenRejectsNonImpersonationToken(t *testing.T) {
	var (
		ctx  = context.Background()
		svc  = newTenantAuthTestService()
		user = &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	)
	accessToken, _, _, err := svc.generateTokenPair(ctx, user, 42, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate tenant token: %v", err)
	}
	if err = svc.RevokeImpersonationToken(ctx, accessToken, 42); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected non-impersonation revoke to be rejected, got %v", err)
	}
}

// TestRevokeImpersonationTokenRequiresOnlineSession verifies impersonation
// revocation does not accept a valid JWT after the authoritative session row is
// gone.
func TestRevokeImpersonationTokenRequiresOnlineSession(t *testing.T) {
	var (
		ctx      = context.WithValue(context.Background(), bizctx.ContextKey, &model.Context{ClientType: tokencap.ClientTypeWeb.String()})
		svc      = newTenantAuthTestService()
		username = fmt.Sprintf("impersonation-session-%d", time.Now().UnixNano())
		userID   = insertAuthTestUser(t, ctx, username, "admin123")
	)

	out, err := svc.IssueImpersonationToken(ctx, ImpersonationTokenIssueInput{ActingUserID: userID, TenantID: 42})
	if err != nil {
		t.Fatalf("issue impersonation token: %v", err)
	}
	if err = svc.sessionStore.Delete(ctx, out.TokenID); err != nil {
		t.Fatalf("delete impersonation session: %v", err)
	}
	if err = svc.RevokeImpersonationToken(ctx, "Bearer "+out.AccessToken, 42); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected missing impersonation session to reject revoke, got %v", err)
	}
}

// TestPreTokenTTLIsShortAndEnforced verifies pre-login tokens use the expected
// short lifetime and expired records cannot be exchanged for a formal JWT.
func TestPreTokenTTLIsShortAndEnforced(t *testing.T) {
	ctx := context.Background()
	store := newMemoryPreTokenStore()
	preToken, err := store.Create(ctx, preTokenRecord{
		UserID:     101,
		Username:   "tenant-user",
		Status:     1,
		ClientType: tokencap.ClientTypeWeb,
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
	var (
		ctx         = context.Background()
		sharedCache = newSharedMemoryKVCache()
		firstSvc    = newTenantAuthTestService()
		secondSvc   = newTenantAuthTestService()
	)
	firstSvc.preTokens = newKVPreTokenStore(sharedCache)
	secondSvc.preTokens = newKVPreTokenStore(sharedCache)

	preToken, err := firstSvc.preTokens.Create(ctx, preTokenRecord{
		UserID:     101,
		Username:   "tenant-user",
		Status:     1,
		ClientType: tokencap.ClientTypeWeb,
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
	var (
		ctx         = context.Background()
		sharedCache = newSharedMemoryKVCache()
		firstStore  = newLayeredRevokeStore(newMemoryRevokeStore(), newKVRevokeStore(sharedCache))
		secondStore = newLayeredRevokeStore(newMemoryRevokeStore(), newKVRevokeStore(sharedCache))
		expiresAt   = time.Now().Add(time.Hour)
	)

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
	var (
		ctx  = context.Background()
		svc  = newTenantAuthTestService()
		user = &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	)
	oldToken, oldTokenID, err := svc.generateToken(ctx, user, 11, tokencap.ClientTypeDesktop)
	if err != nil {
		t.Fatalf("generate old token: %v", err)
	}
	oldClaims, err := svc.parseAccessTokenForTest(ctx, oldToken)
	if err != nil {
		t.Fatalf("parse old token: %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeDesktop.String(), TokenId: oldTokenID, TenantId: 11, UserId: 101, Username: "tenant-user"}); err != nil {
		t.Fatalf("set old session: %v", err)
	}

	out, err := svc.ReissueTenantToken(ctx, TenantTokenReissueInput{CurrentClaims: oldClaims, TenantID: 22})
	if err != nil {
		t.Fatalf("switch tenant: %v", err)
	}
	if _, err = svc.parseAccessTokenForTest(ctx, oldToken); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected old token to be revoked, got %v", err)
	}
	newClaims, err := svc.parseAccessTokenForTest(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("parse new token: %v", err)
	}
	if newClaims.TenantId != 22 {
		t.Fatalf("expected new tenant 22, got %d", newClaims.TenantId)
	}
	if newClaims.ClientType != tokencap.ClientTypeDesktop {
		t.Fatalf("expected switched tenant clientType %q, got %q", tokencap.ClientTypeDesktop, newClaims.ClientType)
	}
	if out.RefreshToken == "" {
		t.Fatal("expected switched tenant refresh token")
	}
}

// TestSwitchTenantFromBearerAcceptsAuthorizationHeader verifies the bearer
// helper accepts an Authorization header value and still validates the current
// session before issuing a replacement token.
func TestSwitchTenantFromBearerAcceptsAuthorizationHeader(t *testing.T) {
	var (
		ctx  = context.Background()
		svc  = newTenantAuthTestService()
		user = &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	)
	oldToken, oldTokenID, err := svc.generateToken(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate old token: %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: oldTokenID, TenantId: 11, UserId: 101, Username: "tenant-user"}); err != nil {
		t.Fatalf("set old session: %v", err)
	}

	out, err := svc.ReissueTenantTokenFromBearer(ctx, "Bearer "+oldToken, 22)
	if err != nil {
		t.Fatalf("switch tenant from bearer header: %v", err)
	}
	claims, err := svc.parseAccessTokenForTest(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("parse reissued token: %v", err)
	}
	if claims.TenantId != 22 || claims.UserId != 101 {
		t.Fatalf("expected tenant 22 user 101 claims, got tenant=%d user=%d", claims.TenantId, claims.UserId)
	}
	if _, err = svc.parseAccessTokenForTest(ctx, oldToken); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected old token to be revoked, got %v", err)
	}
}

// TestLoginSelectTenantSwitchTenantLogoutFlow verifies the tenant auth
// lifecycle from password login through tenant selection, switching, and logout.
func TestLoginSelectTenantSwitchTenantLogoutFlow(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	username := fmt.Sprintf("tenant-flow-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	svc.tenantSvc = registerTenantAuthTestProvider(t, map[int][]tenantcap.TenantInfo{
		userID: {
			{ID: 11, Code: "tenant-a", Name: "Tenant A", Status: "enabled"},
			{ID: 22, Code: "tenant-b", Name: "Tenant B", Status: "enabled"},
		},
	})

	hooks := &recordingAuthHookService{}
	svc.hookSvc = hooks

	loginOut, err := svc.Login(ctx, LoginInput{Username: username, Password: "admin123", ClientType: tokencap.ClientTypeMobile})
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
	if selectOut.RefreshToken == "" {
		t.Fatal("expected selected tenant refresh token")
	}
	selectedClaims, err := svc.parseAccessTokenForTest(ctx, selectOut.AccessToken)
	if err != nil {
		t.Fatalf("parse selected token: %v", err)
	}
	if selectedClaims.TenantId != 11 || selectedClaims.UserId != userID {
		t.Fatalf("expected selected tenant/user claims, got tenant=%d user=%d", selectedClaims.TenantId, selectedClaims.UserId)
	}
	if selectedClaims.ClientType != tokencap.ClientTypeMobile {
		t.Fatalf("expected selected tenant clientType %q, got %q", tokencap.ClientTypeMobile, selectedClaims.ClientType)
	}
	if active, err := svc.sessionStore.TouchOrValidate(ctx, 11, selectedClaims.TokenId, time.Hour); err != nil || !active {
		t.Fatalf("expected selected tenant session, active=%v err=%v", active, err)
	}
	if selectedSession, err := svc.sessionStore.Get(ctx, selectedClaims.TokenId); err != nil || selectedSession == nil || selectedSession.ClientType != tokencap.ClientTypeMobile.String() {
		t.Fatalf("expected selected tenant session clientType %q, session=%#v err=%v", tokencap.ClientTypeMobile, selectedSession, err)
	}

	switchOut, err := svc.ReissueTenantToken(ctx, TenantTokenReissueInput{CurrentClaims: selectedClaims, TenantID: 22})
	if err != nil {
		t.Fatalf("switch tenant: %v", err)
	}
	if switchOut.RefreshToken == "" {
		t.Fatal("expected switched tenant refresh token")
	}
	if _, err = svc.parseAccessTokenForTest(ctx, selectOut.AccessToken); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected selected token revoked after switch, got %v", err)
	}
	switchedClaims, err := svc.parseAccessTokenForTest(ctx, switchOut.AccessToken)
	if err != nil {
		t.Fatalf("parse switched token: %v", err)
	}
	if switchedClaims.TenantId != 22 || switchedClaims.UserId != userID {
		t.Fatalf("expected switched tenant/user claims, got tenant=%d user=%d", switchedClaims.TenantId, switchedClaims.UserId)
	}
	if switchedClaims.ClientType != tokencap.ClientTypeMobile {
		t.Fatalf("expected switched tenant clientType %q, got %q", tokencap.ClientTypeMobile, switchedClaims.ClientType)
	}
	if active, err := svc.sessionStore.TouchOrValidate(ctx, 11, selectedClaims.TokenId, time.Hour); err != nil || active {
		t.Fatalf("expected selected tenant session removed, active=%v err=%v", active, err)
	}
	if active, err := svc.sessionStore.TouchOrValidate(ctx, 22, switchedClaims.TokenId, time.Hour); err != nil || !active {
		t.Fatalf("expected switched tenant session, active=%v err=%v", active, err)
	}
	if switchedSession, err := svc.sessionStore.Get(ctx, switchedClaims.TokenId); err != nil || switchedSession == nil || switchedSession.ClientType != tokencap.ClientTypeMobile.String() {
		t.Fatalf("expected switched tenant session clientType %q, session=%#v err=%v", tokencap.ClientTypeMobile, switchedSession, err)
	}

	if err = svc.Logout(ctx, LogoutInput{
		Username:   username,
		TenantID:   switchedClaims.TenantId,
		TokenID:    switchedClaims.TokenId,
		ClientType: switchedClaims.ClientType,
	}); err != nil {
		t.Fatalf("logout switched tenant token: %v", err)
	}
	if active, err := svc.sessionStore.TouchOrValidate(ctx, 22, switchedClaims.TokenId, time.Hour); err != nil || active {
		t.Fatalf("expected switched tenant session removed after logout, active=%v err=%v", active, err)
	}
	if len(hooks.logoutSucceeded) != 1 || hooks.logoutSucceeded[0].ClientType != tokencap.ClientTypeMobile.String() {
		t.Fatalf("expected logout hook clientType %q, got %#v", tokencap.ClientTypeMobile, hooks.logoutSucceeded)
	}
}

// TestLoginFailureHookUsesRequestedClientType verifies failed auth events carry
// the explicit user-session client type instead of an internal default.
func TestLoginFailureHookUsesRequestedClientType(t *testing.T) {
	var (
		ctx   = context.Background()
		svc   = newTenantAuthTestService()
		hooks = &recordingAuthHookService{}
	)
	svc.hookSvc = hooks

	_, err := svc.Login(ctx, LoginInput{
		Username:   "missing-user-for-client-type-hook",
		Password:   "bad-password",
		ClientType: tokencap.ClientTypeMobile,
	})
	if !bizerr.Is(err, CodeAuthInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if len(hooks.loginFailed) != 1 {
		t.Fatalf("expected one login failed hook, got %#v", hooks.loginFailed)
	}
	if hooks.loginFailed[0].ClientType != tokencap.ClientTypeMobile.String() {
		t.Fatalf("expected failed hook clientType %q, got %q", tokencap.ClientTypeMobile, hooks.loginFailed[0].ClientType)
	}
	if hooks.loginFailed[0].Reason != pluginhost.AuthHookReasonInvalidCredentials {
		t.Fatalf("expected invalid credential reason, got %q", hooks.loginFailed[0].Reason)
	}
}

// TestLoginRejectsPluginAndServiceClientTypes verifies non-user actors are not
// accepted as login session client types.
func TestLoginRejectsPluginAndServiceClientTypes(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	for _, clientType := range []ClientType{"plugin", "service"} {
		if _, err := svc.Login(ctx, LoginInput{
			Username:   "ignored",
			Password:   "ignored",
			ClientType: clientType,
		}); !bizerr.Is(err, CodeAuthClientTypeInvalid) {
			t.Fatalf("expected invalid client type for %q, got %v", clientType, err)
		}
	}
}

// TestLoginRejectsTenantUserWithoutActiveTenant verifies suspended or archived
// tenant-only users cannot silently fall back to a platform token.
func TestLoginRejectsTenantUserWithoutActiveTenant(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	username := fmt.Sprintf("tenant-unavailable-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	if _, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: userID}).
		Data(do.SysUser{TenantId: 11}).
		Update(); err != nil {
		t.Fatalf("set tenant id on auth test user: %v", err)
	}
	svc.tenantSvc = registerTenantAuthTestProvider(t, map[int][]tenantcap.TenantInfo{userID: {}})

	if _, err := svc.Login(ctx, LoginInput{Username: username, Password: "admin123", ClientType: tokencap.ClientTypeWeb}); !bizerr.Is(err, CodeAuthTenantUnavailable) {
		t.Fatalf("expected tenant unavailable login error, got %v", err)
	}
}

// TestLoginRejectsTenantUserWhenTenantServiceUnavailable verifies tenant users
// fail closed when the tenant provider is disabled instead of receiving a
// platform-scoped token.
func TestLoginRejectsTenantUserWhenTenantServiceUnavailable(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	hooks := &recordingAuthHookService{}
	svc.hookSvc = hooks

	username := fmt.Sprintf("tenant-service-disabled-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	if _, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: userID}).
		Data(do.SysUser{TenantId: 11}).
		Update(); err != nil {
		t.Fatalf("set tenant id on auth test user: %v", err)
	}

	providerPluginID := fmt.Sprintf("plugin-test-disabled-auth-tenant-provider-%d", time.Now().UnixNano())
	manager := tenantspi.NewManager()
	if err := manager.RegisterFactory(providerPluginID, func(context.Context, tenantspi.ProviderEnv) (tenantspi.Provider, error) {
		return &tenantAuthTestProvider{tenantsByUser: map[int][]tenantcap.TenantInfo{
			userID: {{ID: 11, Code: "tenant-a", Name: "Tenant A", Status: "enabled"}},
		}}, nil
	}); err != nil {
		t.Fatalf("register disabled auth tenant provider: %v", err)
	}
	svc.tenantSvc = tenantspi.New(manager, tenantAuthProviderRuntime{pluginID: "disabled-provider-not-enabled"}, nil, nil)
	if svc.tenantSvc.Available(ctx) {
		t.Fatal("expected test tenant service to be unavailable")
	}

	out, err := svc.Login(ctx, LoginInput{Username: username, Password: "admin123", ClientType: tokencap.ClientTypeWeb})
	if !bizerr.Is(err, CodeAuthTenantUnavailable) {
		t.Fatalf("expected tenant unavailable login error, got out=%#v err=%v", out, err)
	}
	if out != nil {
		t.Fatalf("expected tenant service outage to reject token issuance, got %#v", out)
	}
	if len(hooks.loginFailed) != 1 {
		t.Fatalf("expected one tenant unavailable failure hook, got %#v", hooks.loginFailed)
	}
	if hooks.loginFailed[0].Reason != authHookReasonTenantUnavailable {
		t.Fatalf("expected tenant unavailable reason %q, got %q", authHookReasonTenantUnavailable, hooks.loginFailed[0].Reason)
	}
	if len(hooks.loginSucceeded) != 0 {
		t.Fatalf("expected no login success hook, got %#v", hooks.loginSucceeded)
	}
}

// TestRefreshTokenIssuesFreshAccessToken verifies refresh tokens can renew an
// access token for the same online session without rotating the session ID.
func TestRefreshTokenIssuesFreshAccessToken(t *testing.T) {
	var (
		ctx      = context.Background()
		svc      = newTenantAuthTestService()
		username = fmt.Sprintf("refresh-user-%d", time.Now().UnixNano())
		userID   = insertAuthTestUser(t, ctx, username, "admin123")
		user     = &entity.SysUser{Id: userID, Username: username, Status: 1}
	)

	accessToken, refreshToken, tokenID, err := svc.generateTokenPair(ctx, user, 11, tokencap.ClientTypeCLI)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}
	if _, err = svc.parseAccessTokenForTest(ctx, accessToken); err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeCLI.String(), TokenId: tokenID, TenantId: 11, UserId: userID, Username: username}); err != nil {
		t.Fatalf("set refresh session: %v", err)
	}

	out, err := svc.Refresh(ctx, RefreshInput{RefreshToken: refreshToken})
	if err != nil {
		t.Fatalf("refresh token: %v", err)
	}
	if out.RefreshToken != refreshToken {
		t.Fatalf("expected refresh token to remain stable")
	}
	claims, err := svc.parseAccessTokenForTest(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("parse refreshed access token: %v", err)
	}
	if claims.TokenId != tokenID || claims.TokenType != tokenKindAccess || claims.UserId != userID || claims.TenantId != 11 {
		t.Fatalf("unexpected refreshed claims: %#v", claims)
	}
	if claims.ClientType != tokencap.ClientTypeCLI {
		t.Fatalf("expected refreshed access clientType %q, got %q", tokencap.ClientTypeCLI, claims.ClientType)
	}
}

// TestRefreshPrimesAccessContextWithRefreshTokenTenant verifies refresh
// token renewal primes role access using the tenant encoded in the JWT.
func TestRefreshPrimesAccessContextWithRefreshTokenTenant(t *testing.T) {
	var (
		ctx     = datascope.WithTenantScope(context.Background(), 99)
		svc     = newTenantAuthTestService()
		roleSvc = &trackingRoleTestService{}
	)
	svc.roleSvc = roleSvc
	var (
		username = fmt.Sprintf("refresh-scope-%d", time.Now().UnixNano())
		userID   = insertAuthTestUser(t, context.Background(), username, "admin123")
		user     = &entity.SysUser{Id: userID, Username: username, Status: 1}
	)

	_, refreshToken, tokenID, err := svc.generateTokenPair(ctx, user, 22, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: tokenID, TenantId: 22, UserId: userID, Username: username}); err != nil {
		t.Fatalf("set refresh session: %v", err)
	}

	if _, err = svc.Refresh(ctx, RefreshInput{RefreshToken: refreshToken}); err != nil {
		t.Fatalf("refresh token: %v", err)
	}
	if len(roleSvc.tenantIDs) != 1 || roleSvc.tenantIDs[0] != 22 {
		t.Fatalf("expected role cache prime for tenant 22, got %v", roleSvc.tenantIDs)
	}
}

// TestRefreshTokenCannotBeUsedAsAccessToken verifies refresh JWTs are rejected
// by the protected API access-token parser.
func TestRefreshTokenCannotBeUsedAsAccessToken(t *testing.T) {
	var (
		ctx  = context.Background()
		svc  = newTenantAuthTestService()
		user = &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	)

	_, refreshToken, _, err := svc.generateTokenPair(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}
	if _, err = svc.parseAccessTokenForTest(ctx, refreshToken); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected refresh token to be rejected as access token, got %v", err)
	}
}

// TestRefreshRejectsRevokedSession verifies a valid refresh JWT is not enough
// when the online session has already been revoked.
func TestRefreshRejectsRevokedSession(t *testing.T) {
	var (
		ctx  = context.Background()
		svc  = newTenantAuthTestService()
		user = &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	)

	_, refreshToken, _, err := svc.generateTokenPair(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}
	if _, err = svc.Refresh(ctx, RefreshInput{RefreshToken: refreshToken}); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected missing session to reject refresh, got %v", err)
	}
}

// TestRefreshRejectsNegativeTenantClaim verifies that a refresh token
// claiming a negative/sentinel tenant ID — which the host signer never
// issues — is treated as forged and the underlying session is torn down.
func TestRefreshRejectsNegativeTenantClaim(t *testing.T) {
	var (
		ctx      = context.Background()
		svc      = newTenantAuthTestService()
		username = fmt.Sprintf("tenant-neg-%d", time.Now().UnixNano())
		userID   = insertAuthTestUser(t, ctx, username, "admin123")
		user     = &entity.SysUser{Id: userID, Username: username, Status: 1}
	)

	// Forge a refresh token whose TenantId sits below PLATFORM. We bypass
	// generateTokenPair because the production signer never emits such a
	// value; the goal is to confirm the parser/refresh path rejects it.
	const forgedTenantID = -1
	tokenID := "forged-negative-tenant-token"
	refreshToken, err := svc.signToken(ctx, user, forgedTenantID, tokenID, tokenKindRefresh, tokencap.ClientTypeWeb, false, 0)
	if err != nil {
		t.Fatalf("sign forged refresh token: %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: tokenID, TenantId: forgedTenantID, UserId: userID, Username: username}); err != nil {
		t.Fatalf("seed forged session: %v", err)
	}

	if _, err = svc.Refresh(ctx, RefreshInput{RefreshToken: refreshToken}); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected negative tenant refresh to be rejected with CodeAuthTokenInvalid, got %v", err)
	}
	if active, sessErr := svc.sessionStore.TouchOrValidate(ctx, forgedTenantID, tokenID, time.Hour); sessErr != nil || active {
		t.Fatalf("expected forged-tenant session removed, active=%v err=%v", active, sessErr)
	}
}

// TestRefreshPreservesSessionOnProviderInfraError verifies that a
// transient infrastructure failure from the tenant provider (e.g., DB
// outage) causes refresh to fail without tearing down the online session.
// Access tokens are short-lived; once infra recovers the next refresh will
// re-evaluate membership and revoke if the eviction turns out to be real.
func TestRefreshPreservesSessionOnProviderInfraError(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	var (
		username = fmt.Sprintf("tenant-infra-%d", time.Now().UnixNano())
		userID   = insertAuthTestUser(t, ctx, username, "admin123")
		user     = &entity.SysUser{Id: userID, Username: username, Status: 1}
	)

	infraErr := errors.New("simulated tenant provider infra failure")
	provider := &tenantAuthTestProvider{
		tenantsByUser: map[int][]tenantcap.TenantInfo{
			userID: {{ID: 11, Code: "tenant-a", Name: "Tenant A", Status: "enabled"}},
		},
		validateErr: infraErr,
	}
	svc.tenantSvc = registerTenantAuthProviderInstance(t, provider)

	_, refreshToken, tokenID, err := svc.generateTokenPair(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: tokenID, TenantId: 11, UserId: userID, Username: username}); err != nil {
		t.Fatalf("set refresh session: %v", err)
	}

	if _, err = svc.Refresh(ctx, RefreshInput{RefreshToken: refreshToken}); !errors.Is(err, infraErr) {
		t.Fatalf("expected infra error to propagate from refresh, got %v", err)
	}
	if active, sessErr := svc.sessionStore.TouchOrValidate(ctx, 11, tokenID, time.Hour); sessErr != nil || !active {
		t.Fatalf("expected session preserved on infra error, active=%v err=%v", active, sessErr)
	}

	// Once infra recovers, the next refresh should succeed without losing
	// the session continuity.
	provider.validateErr = nil
	if _, err = svc.Refresh(ctx, RefreshInput{RefreshToken: refreshToken}); err != nil {
		t.Fatalf("expected refresh to succeed after infra recovery: %v", err)
	}
}

// TestRefreshRejectsAfterTenantMembershipRemoved verifies that revoking a
// user's tenant membership immediately blocks refresh from minting fresh
// tenant-scoped access tokens, even while the refresh JWT and online session
// are still nominally valid.
func TestRefreshRejectsAfterTenantMembershipRemoved(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()

	var (
		username = fmt.Sprintf("tenant-evict-%d", time.Now().UnixNano())
		userID   = insertAuthTestUser(t, ctx, username, "admin123")
		user     = &entity.SysUser{Id: userID, Username: username, Status: 1}
	)

	provider := &tenantAuthTestProvider{tenantsByUser: map[int][]tenantcap.TenantInfo{
		userID: {{ID: 11, Code: "tenant-a", Name: "Tenant A", Status: "enabled"}},
	}}
	svc.tenantSvc = registerTenantAuthProviderInstance(t, provider)

	_, refreshToken, tokenID, err := svc.generateTokenPair(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: tokenID, TenantId: 11, UserId: userID, Username: username}); err != nil {
		t.Fatalf("set refresh session: %v", err)
	}

	// Sanity check: while the user is still a tenant member, refresh succeeds.
	if _, err = svc.Refresh(ctx, RefreshInput{RefreshToken: refreshToken}); err != nil {
		t.Fatalf("baseline refresh before eviction: %v", err)
	}

	// Evict the user from the tenant: the refresh JWT and session still look
	// valid, but membership lookups must now fail.
	provider.tenantsByUser[userID] = nil

	if _, err = svc.Refresh(ctx, RefreshInput{RefreshToken: refreshToken}); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected refresh after tenant eviction to fail with CodeAuthTokenInvalid, got %v", err)
	}
	if active, sessErr := svc.sessionStore.TouchOrValidate(ctx, 11, tokenID, time.Hour); sessErr != nil || active {
		t.Fatalf("expected evicted-tenant session removed, active=%v err=%v", active, sessErr)
	}
}

// TestRevokeSharedStoreInvalidatesAcrossInstances verifies that one auth
// instance can revoke a JWT and another instance rejects it through shared state.
func TestRevokeSharedStoreInvalidatesAcrossInstances(t *testing.T) {
	var (
		ctx         = context.Background()
		sharedCache = newSharedMemoryKVCache()
		firstSvc    = newTenantAuthTestService()
		secondSvc   = newTenantAuthTestService()
	)
	firstSvc.revoked = newKVRevokeStore(sharedCache)
	secondSvc.revoked = newKVRevokeStore(sharedCache)
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	token, tokenID, err := firstSvc.generateToken(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate shared revoke token: %v", err)
	}
	claims, err := firstSvc.parseAccessTokenForTest(ctx, token)
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
	if _, err = secondSvc.parseAccessTokenForTest(ctx, token); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected second instance to reject revoked token, got %v", err)
	}
}

// TestAccessTokenParseRevokeReadFailureFailClosed verifies a valid JWT is rejected
// when the shared token-state store cannot confirm whether it has been revoked.
func TestAccessTokenParseRevokeReadFailureFailClosed(t *testing.T) {
	var (
		ctx  = context.Background()
		svc  = newTenantAuthTestService()
		user = &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	)
	token, _, err := svc.generateToken(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	svc.revoked = &failingRevokeStore{revokedErr: errors.New("simulated redis revoke read failure")}
	if _, err = svc.parseAccessTokenForTest(ctx, token); !bizerr.Is(err, CodeAuthTokenStateUnavailable) {
		t.Fatalf("expected revoke read failure to fail closed, got %v", err)
	}
}

// TestAuthenticateAccessTokenUsesOnlineSessionAuthority verifies complete
// access-token authentication rejects a valid JWT when the authoritative
// sys_online_session-equivalent store has no matching session.
func TestAuthenticateAccessTokenUsesOnlineSessionAuthority(t *testing.T) {
	var (
		ctx  = context.Background()
		svc  = newTenantAuthTestService()
		user = &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	)
	token, tokenID, err := svc.generateToken(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if _, err = svc.parseAccessTokenForTest(ctx, token); err != nil {
		t.Fatalf("low-level access-token parsing should only validate JWT and revoke state: %v", err)
	}
	if _, err = svc.AuthenticateAccessToken(ctx, token); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected missing online session to reject complete auth, got %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: tokenID, TenantId: 11, UserId: 101, Username: "tenant-user"}); err != nil {
		t.Fatalf("set online session: %v", err)
	}
	claims, err := svc.AuthenticateAccessToken(ctx, token)
	if err != nil {
		t.Fatalf("expected valid session to authenticate: %v", err)
	}
	if claims.TokenId != tokenID || claims.TenantId != 11 {
		t.Fatalf("unexpected authenticated claims: %#v", claims)
	}
}

// TestLogoutRevokesCurrentToken verifies logout removes the supplied token from
// the session store contract and writes shared JWT revocation state.
func TestLogoutRevokesCurrentToken(t *testing.T) {
	var (
		ctx         = context.Background()
		store       = newMemorySessionStore()
		sharedCache = newSharedMemoryKVCache()
		svc         = newTenantAuthTestService()
		hooks       = &recordingAuthHookService{}
	)
	svc.hookSvc = hooks
	svc.sessionStore = store
	svc.revoked = newKVRevokeStore(sharedCache)
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	token, tokenID, err := svc.generateToken(ctx, user, 22, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate logout token: %v", err)
	}
	if err = store.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: tokenID, TenantId: 22, UserId: 101, Username: "tenant-user"}); err != nil {
		t.Fatalf("set logout session: %v", err)
	}

	if err = svc.Logout(ctx, LogoutInput{
		Username:   "tenant-user",
		TenantID:   22,
		TokenID:    tokenID,
		ClientType: tokencap.ClientTypeWeb,
	}); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if store.deletedTokenID != tokenID {
		t.Fatalf("expected token revoke, got token=%q", store.deletedTokenID)
	}
	if _, ok, err := sharedCache.Get(ctx, kvcache.OwnerTypeModule, revokeCacheKey(tokenID)); err != nil || !ok {
		t.Fatalf("expected logout shared revoke state, ok=%v err=%v", ok, err)
	}
	if _, err = svc.parseAccessTokenForTest(ctx, token); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected logged-out token to be rejected, got %v", err)
	}
	if len(hooks.logoutSucceeded) != 1 || hooks.logoutSucceeded[0].ClientType != tokencap.ClientTypeWeb.String() {
		t.Fatalf("expected logout hook clientType %q, got %#v", tokencap.ClientTypeWeb, hooks.logoutSucceeded)
	}
}

// TestRevokeSessionWritesSharedRevoke verifies force-logout style token-ID
// revocation publishes shared revoke state before removing the session row.
func TestRevokeSessionWritesSharedRevoke(t *testing.T) {
	var (
		ctx         = context.Background()
		store       = newMemorySessionStore()
		sharedCache = newSharedMemoryKVCache()
		svc         = newTenantAuthTestService()
	)
	svc.sessionStore = store
	svc.revoked = newKVRevokeStore(sharedCache)

	if err := store.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: "force-token", TenantId: 22, UserId: 101, Username: "tenant-user"}); err != nil {
		t.Fatalf("set force logout session: %v", err)
	}
	if err := svc.RevokeSession(ctx, "force-token"); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	if store.deletedTokenID != "force-token" {
		t.Fatalf("expected force logout token delete, got token=%q", store.deletedTokenID)
	}
	if _, ok, err := sharedCache.Get(ctx, kvcache.OwnerTypeModule, revokeCacheKey("force-token")); err != nil || !ok {
		t.Fatalf("expected force logout shared revoke state, ok=%v err=%v", ok, err)
	}
}

// TestLogoutRevokeWriteFailureReturnsStructuredError verifies logout does not
// hide shared token-state write failures.
func TestLogoutRevokeWriteFailureReturnsStructuredError(t *testing.T) {
	var (
		ctx   = context.Background()
		svc   = newTenantAuthTestService()
		store = newMemorySessionStore()
	)
	svc.sessionStore = store
	svc.revoked = &failingRevokeStore{addErr: errors.New("simulated logout revoke write failure")}

	if err := svc.Logout(ctx, LogoutInput{
		Username:   "tenant-user",
		TenantID:   22,
		TokenID:    "logout-failure-token",
		ClientType: tokencap.ClientTypeWeb,
	}); !bizerr.Is(err, CodeAuthTokenStateUnavailable) {
		t.Fatalf("expected logout revoke write failure to be structured, got %v", err)
	}
	if store.deletedTokenID != "" {
		t.Fatalf("expected logout revoke failure to preserve session projection, deleted token=%q", store.deletedTokenID)
	}
}

// TestRevokeSessionWriteFailureReturnsStructuredError verifies force-logout
// style revocation reports shared token-state write failures.
func TestRevokeSessionWriteFailureReturnsStructuredError(t *testing.T) {
	var (
		ctx   = context.Background()
		svc   = newTenantAuthTestService()
		store = newMemorySessionStore()
	)
	svc.sessionStore = store
	svc.revoked = &failingRevokeStore{addErr: errors.New("simulated force logout revoke write failure")}

	if err := svc.RevokeSession(ctx, "force-failure-token"); !bizerr.Is(err, CodeAuthTokenStateUnavailable) {
		t.Fatalf("expected force logout revoke write failure to be structured, got %v", err)
	}
	if store.deletedTokenID != "" {
		t.Fatalf("expected force logout revoke failure to preserve session projection, deleted token=%q", store.deletedTokenID)
	}
}

// TestSwitchTenantRevokeWriteFailureReturnsStructuredError verifies old-token
// revocation write failures abort tenant switching with a stable auth error.
func TestSwitchTenantRevokeWriteFailureReturnsStructuredError(t *testing.T) {
	var (
		ctx  = context.Background()
		svc  = newTenantAuthTestService()
		user = &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	)
	oldToken, oldTokenID, err := svc.generateToken(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate old token: %v", err)
	}
	oldClaims, err := svc.parseAccessTokenForTest(ctx, oldToken)
	if err != nil {
		t.Fatalf("parse old token: %v", err)
	}
	if err = svc.sessionStore.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: oldTokenID, TenantId: 11, UserId: 101, Username: "tenant-user"}); err != nil {
		t.Fatalf("set old session: %v", err)
	}

	svc.revoked = &failingRevokeStore{addErr: errors.New("simulated redis revoke write failure")}
	if _, err = svc.ReissueTenantToken(ctx, TenantTokenReissueInput{CurrentClaims: oldClaims, TenantID: 22}); !bizerr.Is(err, CodeAuthTokenStateUnavailable) {
		t.Fatalf("expected switch tenant revoke write failure to be structured, got %v", err)
	}
}

// TestSwitchTenantRequiresOnlineSession verifies tenant switching validates the
// current token against the authoritative online session store before revoking
// it and issuing a new session.
func TestSwitchTenantRequiresOnlineSession(t *testing.T) {
	var (
		ctx  = context.Background()
		svc  = newTenantAuthTestService()
		user = &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	)
	oldToken, _, err := svc.generateToken(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate old token: %v", err)
	}
	oldClaims, err := svc.parseAccessTokenForTest(ctx, oldToken)
	if err != nil {
		t.Fatalf("parse old token: %v", err)
	}

	if _, err = svc.ReissueTenantToken(ctx, TenantTokenReissueInput{CurrentClaims: oldClaims, TenantID: 22}); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected missing online session to reject tenant switch, got %v", err)
	}
}

// TestMemorySessionStoreUsesGlobalTokenIdentity verifies the auth test helper
// mirrors the production globally unique token_id session-store contract.
func TestMemorySessionStoreUsesGlobalTokenIdentity(t *testing.T) {
	ctx := context.Background()
	store := newMemorySessionStore()
	if err := store.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: "same-token", TenantId: 11, UserId: 101}); err != nil {
		t.Fatalf("set tenant 11 session: %v", err)
	}
	if err := store.Set(ctx, &session.Session{ClientType: tokencap.ClientTypeWeb.String(), TokenId: "same-token", TenantId: 22, UserId: 101}); err != nil {
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

// generateToken generates one test access JWT without creating a refresh token.
func (s *serviceImpl) generateToken(ctx context.Context, user *entity.SysUser, tenantID int, clientType ClientType) (string, string, error) {
	tokenID := guid.S()
	token, err := s.signToken(ctx, user, tenantID, tokenID, tokenKindAccess, clientType, false, 0)
	if err != nil {
		return "", "", err
	}
	return token, tokenID, nil
}

// memoryPreTokenStore keeps pre-login tokens in memory for isolated tests.
type memoryPreTokenStore struct {
	mu      sync.Mutex
	records map[string]preTokenRecord
}

// newMemoryPreTokenStore creates an empty in-memory pre-login token store.
func newMemoryPreTokenStore() *memoryPreTokenStore {
	return &memoryPreTokenStore{records: make(map[string]preTokenRecord)}
}

// Create stores one short-lived pre-login token.
func (s *memoryPreTokenStore) Create(_ context.Context, record preTokenRecord) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := preTokenGeneratedPrefix + guid.S()
	record.ExpiresAt = time.Now().Add(preTokenTTL)
	s.records[token] = record
	return token, nil
}

// Consume returns and deletes one pre-login token.
func (s *memoryPreTokenStore) Consume(_ context.Context, token string) (preTokenRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[token]
	if !ok {
		return preTokenRecord{}, false, nil
	}
	delete(s.records, token)
	if time.Now().After(record.ExpiresAt) {
		return preTokenRecord{}, false, nil
	}
	return record, true, nil
}

// newTenantAuthTestService returns a service with in-memory session state.
func newTenantAuthTestService() *serviceImpl {
	kv := kvcache.New()
	return &serviceImpl{
		configSvc:    configTestService{},
		roleSvc:      roleTestService{},
		sessionStore: newMemorySessionStore(),
		preTokens:    newMemoryPreTokenStore(),
		resetTokens:  newKVPasswordResetStore(kv),
		rateLimit:    newKVRateLimitStore(kv),
		kvCache:      kv,
		revoked:      newMemoryRevokeStore(),
	}
}

// parseAccessTokenForTest exposes low-level access-token parsing to same-package
// tests without adding it to the production Service contract.
func (s *serviceImpl) parseAccessTokenForTest(ctx context.Context, tokenString string) (*Claims, error) {
	return s.parseToken(ctx, tokenString, tokenKindAccess)
}

// configTestService provides JWT settings used by auth unit tests.
type configTestService struct {
	configsvc.Service
}

// GetJwtSecret returns a stable test signing secret.
func (configTestService) GetJwtSecret(context.Context) string {
	return "tenant-auth-test-secret"
}

// GetJwtExpire returns a stable test token lifetime.
func (configTestService) GetJwtExpire(context.Context) (time.Duration, error) {
	return time.Hour, nil
}

// GetSessionTimeout returns a stable online-session lifetime for auth tests.
func (configTestService) GetSessionTimeout(context.Context) (time.Duration, error) {
	return time.Hour, nil
}

// IsLoginIPBlacklisted reports no blacklist entries in auth unit tests.
func (configTestService) IsLoginIPBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}

// GetPublicFrontend returns public frontend settings for unit tests. It prefers
// the real config service so withRuntimeParamValue overrides are visible.
func (configTestService) GetPublicFrontend(ctx context.Context) (*configsvc.PublicFrontendConfig, error) {
	if cfg, err := configsvc.New().GetPublicFrontend(ctx); err == nil && cfg != nil {
		return cfg, nil
	}
	return &configsvc.PublicFrontendConfig{
		Auth: configsvc.PublicFrontendAuthConfig{
			RegisterEnabled:       true,
			ForgetPasswordEnabled: true,
			PanelLayout:           configsvc.PublicFrontendAuthPanelLayoutRight,
		},
	}, nil
}

// roleTestService stubs the token access cache hooks used by auth.
type roleTestService struct {
	role.Service
}

// PrimeTokenAccessContext returns a no-op access snapshot.
func (roleTestService) PrimeTokenAccessContext(context.Context, string, int) (*role.UserAccessContext, error) {
	return &role.UserAccessContext{}, nil
}

// InvalidateTokenAccessContext records no state in auth unit tests.
func (roleTestService) InvalidateTokenAccessContext(context.Context, string) {}

// trackingRoleTestService records the tenant scope used to prime token access
// snapshots.
type trackingRoleTestService struct {
	role.Service
	tenantIDs []int
	contexts  []*model.Context
}

// PrimeTokenAccessContext records the current tenant and returns an empty
// access snapshot.
func (s *trackingRoleTestService) PrimeTokenAccessContext(ctx context.Context, _ string, _ int) (*role.UserAccessContext, error) {
	s.tenantIDs = append(s.tenantIDs, datascope.CurrentTenantID(ctx))
	if businessCtx, ok := ctx.Value(bizctx.ContextKey).(*model.Context); ok {
		s.contexts = append(s.contexts, businessCtx)
	} else {
		s.contexts = append(s.contexts, nil)
	}
	return &role.UserAccessContext{}, nil
}

// InvalidateTokenAccessContext records no state for tenant-scope assertions.
func (s *trackingRoleTestService) InvalidateTokenAccessContext(context.Context, string) {}

// recordingAuthHookService records auth hook payloads for client-type
// propagation assertions.
type recordingAuthHookService struct {
	loginSucceeded  []pluginhost.AuthHookPayloadInput
	loginFailed     []pluginhost.AuthHookPayloadInput
	logoutSucceeded []pluginhost.AuthHookPayloadInput
}

// DispatchHookEvent records auth hook payloads by extension point.
func (s *recordingAuthHookService) DispatchHookEvent(
	_ context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	input := authHookPayloadInputFromValues(values)
	switch event {
	case pluginhost.ExtensionPointAuthLoginSucceeded:
		s.loginSucceeded = append(s.loginSucceeded, input)
	case pluginhost.ExtensionPointAuthLoginFailed:
		s.loginFailed = append(s.loginFailed, input)
	case pluginhost.ExtensionPointAuthLogoutSucceeded:
		s.logoutSucceeded = append(s.logoutSucceeded, input)
	}
	return nil
}

// authHookPayloadInputFromValues restores typed auth hook fields for assertions.
func authHookPayloadInputFromValues(values map[string]interface{}) pluginhost.AuthHookPayloadInput {
	status, _ := pluginhost.HookPayloadIntValue(values, pluginhost.HookPayloadKeyStatus)
	return pluginhost.AuthHookPayloadInput{
		UserName:   pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyUserName),
		Status:     status,
		IP:         pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyIP),
		ClientType: pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyClientType),
		Browser:    pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyBrowser),
		OS:         pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyOS),
		Message:    pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyMessage),
		Reason:     pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyReason),
	}
}

// tenantAuthTestProvider provides deterministic tenant memberships for auth tests.
type tenantAuthTestProvider struct {
	tenantsByUser map[int][]tenantcap.TenantInfo
	// validateErr, when non-nil, is returned by ValidateUserInTenant verbatim
	// before the membership lookup. Used to simulate provider infrastructure
	// failures (e.g., DB timeout) that surface as non-bizerr errors.
	validateErr error
}

// ResolveTenant returns no request-derived tenant in auth tests.
func (p *tenantAuthTestProvider) ResolveTenant(context.Context, *ghttp.Request) (*tenantcap.ResolverResult, error) {
	return &tenantcap.ResolverResult{TenantID: tenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant verifies the user is a member of the requested tenant.
func (p *tenantAuthTestProvider) ValidateUserInTenant(_ context.Context, userID int, tenantID tenantcap.TenantID) error {
	if p.validateErr != nil {
		return p.validateErr
	}
	for _, tenant := range p.tenantsByUser[userID] {
		if tenant.ID == tenantID {
			return nil
		}
	}
	return bizerr.NewCode(CodeAuthTokenInvalid)
}

// ListUserTenants returns the configured user tenants.
func (p *tenantAuthTestProvider) ListUserTenants(_ context.Context, userID int) ([]tenantcap.TenantInfo, error) {
	tenants := p.tenantsByUser[userID]
	result := make([]tenantcap.TenantInfo, len(tenants))
	copy(result, tenants)
	return result, nil
}

// SwitchTenant verifies the target tenant membership.
func (p *tenantAuthTestProvider) SwitchTenant(ctx context.Context, userID int, target tenantcap.TenantID) error {
	return p.ValidateUserInTenant(ctx, userID, target)
}

// registerTenantAuthTestProvider installs a temporary tenant provider.
func registerTenantAuthTestProvider(t *testing.T, tenantsByUser map[int][]tenantcap.TenantInfo) tenantspi.Service {
	t.Helper()
	return registerTenantAuthProviderInstance(t, &tenantAuthTestProvider{tenantsByUser: tenantsByUser})
}

// registerTenantAuthProviderInstance installs a temporary tenant provider
// through the pluginservice lifecycle-style registry.
func registerTenantAuthProviderInstance(t *testing.T, provider *tenantAuthTestProvider) tenantspi.Service {
	t.Helper()
	providerPluginID := fmt.Sprintf("plugin-test-auth-tenant-provider-%d", time.Now().UnixNano())
	manager := tenantspi.NewManager()
	if err := manager.RegisterFactory(providerPluginID, func(context.Context, tenantspi.ProviderEnv) (tenantspi.Provider, error) {
		return provider, nil
	}); err != nil {
		t.Fatalf("register auth tenant provider: %v", err)
	}
	return tenantspi.New(manager, tenantAuthProviderRuntime{pluginID: providerPluginID}, nil, nil)
}

// tenantAuthProviderRuntime marks exactly one test provider plugin enabled.
type tenantAuthProviderRuntime struct {
	pluginID string
}

// IsProviderEnabled reports whether the given test provider plugin is enabled.
func (r tenantAuthProviderRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return pluginID == r.pluginID
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

// failingRevokeStore simulates Redis token-state failures in auth tests.
type failingRevokeStore struct {
	addErr     error
	revokedErr error
}

// Add returns the configured write error.
func (s *failingRevokeStore) Add(context.Context, string, time.Time) error {
	return s.addErr
}

// Revoked returns the configured read error.
func (s *failingRevokeStore) Revoked(context.Context, string) (bool, error) {
	return false, s.revokedErr
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
func (s *sharedMemoryKVCache) Expire(_ context.Context, ownerType kvcache.OwnerType, cacheKey string, ttl time.Duration) (bool, *time.Time, error) {
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
	item.ExpireAt = &expireAt
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

// BatchGetScoped returns requested sessions for auth unit tests.
func (s *memorySessionStore) BatchGetScoped(
	_ context.Context,
	tokenIDs []string,
	_ datascope.Service,
	_ tenantspi.ScopeService,
) ([]*session.Session, error) {
	items := make([]*session.Session, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		if item := s.items[tokenID]; item != nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// BatchGetUserOnlineStatusScoped returns in-memory session counts per user.
func (s *memorySessionStore) BatchGetUserOnlineStatusScoped(
	_ context.Context,
	userIDs []int,
	_ datascope.Service,
	_ tenantspi.ScopeService,
) ([]*session.UserOnlineStatus, error) {
	counts := make(map[int]int, len(userIDs))
	for _, item := range s.items {
		if item != nil {
			counts[item.UserId]++
		}
	}
	statuses := make([]*session.UserOnlineStatus, 0, len(userIDs))
	for _, userID := range userIDs {
		statuses = append(statuses, &session.UserOnlineStatus{
			UserId:       userID,
			SessionCount: counts[userID],
		})
	}
	return statuses, nil
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
	tenantspi.ScopeService,
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
	_ configsvc.Service  = configTestService{}
	_ role.Service       = roleTestService{}
	_ session.Store      = (*memorySessionStore)(nil)
	_ kvcache.Service    = (*sharedMemoryKVCache)(nil)
	_ jwt.Claims         = (*Claims)(nil)
	_ tenantspi.Provider = (*tenantAuthTestProvider)(nil)
)

// --- runtime-parameter login policy (formerly auth_runtime_params_test.go) ---

// TestLoginRejectsBlacklistedIP verifies managed login IP blacklist settings
// are enforced before user lookup.
func TestLoginRejectsBlacklistedIP(t *testing.T) {
	withRuntimeParamValue(t, configsvc.RuntimeParamKeyLoginBlackIPList, "127.0.0.1")

	username := fmt.Sprintf("blacklist-test-%s", t.Name())
	ctx := newRequestContext(t, "127.0.0.1:19120")

	_, err := newRuntimeParamAuthTestService().Login(ctx, LoginInput{
		Username:   username,
		Password:   "ignored",
		ClientType: tokencap.ClientTypeWeb,
	})
	if err == nil {
		t.Fatal("expected blacklisted login attempt to fail")
	}
	if localized := i18nsvc.New(bizctx.New(), configsvc.New(), cachecoord.Default(nil)).LocalizeError(context.Background(), err); localized != "登录IP已被禁止" {
		t.Fatalf("expected blacklisted login error %q, got %q", "登录IP已被禁止", localized)
	}
}

// newRuntimeParamAuthTestService constructs auth with explicit test
// dependencies while still reading runtime params from the real config service.
func newRuntimeParamAuthTestService() Service {
	var (
		configSvc    = configsvc.New()
		sessionStore = session.NewDBStore()
		cacheSvc     = kvcache.New()
	)
	return New(configSvc, runtimeParamAuthTestHooks{}, nil, roleTestService{}, tenantspi.New(nil, nil, nil, nil), sessionStore, cacheSvc)
}

// runtimeParamAuthTestHooks discards auth lifecycle hooks for runtime-parameter
// tests because these cases only verify config-driven auth rejection.
type runtimeParamAuthTestHooks struct{}

// DispatchHookEvent records no hook state in runtime-parameter tests.
func (runtimeParamAuthTestHooks) DispatchHookEvent(context.Context, pluginhost.ExtensionPoint, map[string]interface{}) error {
	return nil
}

// newRequestContext builds one request-backed context carrying the supplied
// remote address for auth service tests.
func newRequestContext(t *testing.T, remoteAddr string) context.Context {
	t.Helper()

	httpReq, err := http.NewRequest(http.MethodPost, "http://localhost/api/v1/auth/login", nil)
	if err != nil {
		t.Fatalf("build http request: %v", err)
	}
	httpReq.RemoteAddr = remoteAddr
	httpReq.Header.Set("User-Agent", "runtime-param-test")

	req := &ghttp.Request{Request: httpReq}
	return req.Context()
}

// withRuntimeParamValue temporarily overrides one protected runtime parameter
// and restores the original sys_config record during cleanup.
func withRuntimeParamValue(t *testing.T, key string, value string) {
	t.Helper()

	ctx := context.Background()
	original, err := queryRuntimeParam(ctx, key)
	if err != nil {
		t.Fatalf("query runtime param %s: %v", key, err)
	}

	if original == nil {
		_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
			Name:   key,
			Key:    key,
			Value:  value,
			Remark: "test override",
		}).Insert()
		if err != nil {
			t.Fatalf("insert runtime param %s: %v", key, err)
		}
		markRuntimeParamChanged(t, ctx)
		t.Cleanup(func() {
			if _, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: key}).Delete(); cleanupErr != nil {
				t.Fatalf("cleanup runtime param %s: %v", key, cleanupErr)
			}
			markRuntimeParamChanged(t, ctx)
		})
		return
	}

	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: original.Id}).
		Data(do.SysConfig{Value: value}).
		Update()
	if err != nil {
		t.Fatalf("update runtime param %s: %v", key, err)
	}
	markRuntimeParamChanged(t, ctx)
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{
				Name:   original.Name,
				Key:    original.Key,
				Value:  original.Value,
				Remark: original.Remark,
			}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore runtime param %s: %v", key, cleanupErr)
		}
		markRuntimeParamChanged(t, ctx)
	})
}

// markRuntimeParamChanged bumps the runtime-parameter revision for tests after
// direct sys_config mutations.
func markRuntimeParamChanged(t *testing.T, ctx context.Context) {
	t.Helper()

	if err := configsvc.New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
}

// queryRuntimeParam loads one sys_config record by protected runtime-parameter key.
func queryRuntimeParam(ctx context.Context, key string) (*entity.SysConfig, error) {
	var runtimeParam *entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Key: key}).
		Scan(&runtimeParam)
	if err != nil {
		return nil, err
	}
	return runtimeParam, nil
}
