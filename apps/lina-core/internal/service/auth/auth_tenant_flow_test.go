// This file verifies host tenant-aware authentication token transitions.

package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
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
	claims, err := svc.ParseToken(ctx, out.AccessToken)
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
	ctx := datascope.WithTenantScope(context.Background(), 99)
	svc := newTenantAuthTestService()
	roleSvc := &trackingRoleTestService{}
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
	ctx := context.WithValue(context.Background(), bizctx.ContextKey, &model.Context{ClientType: tokencap.ClientTypeDesktop.String()})
	svc := newTenantAuthTestService()
	roleSvc := &trackingRoleTestService{}
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
	claims, err := svc.ParseToken(ctx, out.AccessToken)
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
	if _, err = svc.ParseToken(ctx, out.AccessToken); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected revoked impersonation token to be invalid, got %v", err)
	}
}

// TestRevokeImpersonationTokenRejectsNonImpersonationToken verifies plugins
// cannot use the impersonation revoke path to tear down ordinary sessions.
func TestRevokeImpersonationTokenRejectsNonImpersonationToken(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	accessToken, _, _, err := svc.generateTokenPair(ctx, user, 42, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate tenant token: %v", err)
	}
	if err = svc.RevokeImpersonationToken(ctx, accessToken, 42); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected non-impersonation revoke to be rejected, got %v", err)
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
	ctx := context.Background()
	sharedCache := newSharedMemoryKVCache()
	firstSvc := newTenantAuthTestService()
	secondSvc := newTenantAuthTestService()
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
	oldToken, oldTokenID, err := svc.generateToken(ctx, user, 11, tokencap.ClientTypeDesktop)
	if err != nil {
		t.Fatalf("generate old token: %v", err)
	}
	oldClaims, err := svc.ParseToken(ctx, oldToken)
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
	if newClaims.ClientType != tokencap.ClientTypeDesktop {
		t.Fatalf("expected switched tenant clientType %q, got %q", tokencap.ClientTypeDesktop, newClaims.ClientType)
	}
	if out.RefreshToken == "" {
		t.Fatal("expected switched tenant refresh token")
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
	selectedClaims, err := svc.ParseToken(ctx, selectOut.AccessToken)
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
	ctx := context.Background()
	svc := newTenantAuthTestService()
	hooks := &recordingAuthHookService{}
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

// TestRefreshTokenIssuesFreshAccessToken verifies refresh tokens can renew an
// access token for the same online session without rotating the session ID.
func TestRefreshTokenIssuesFreshAccessToken(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	username := fmt.Sprintf("refresh-user-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	user := &entity.SysUser{Id: userID, Username: username, Status: 1}

	accessToken, refreshToken, tokenID, err := svc.generateTokenPair(ctx, user, 11, tokencap.ClientTypeCLI)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}
	if _, err = svc.ParseToken(ctx, accessToken); err != nil {
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
	claims, err := svc.ParseToken(ctx, out.AccessToken)
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
	ctx := datascope.WithTenantScope(context.Background(), 99)
	svc := newTenantAuthTestService()
	roleSvc := &trackingRoleTestService{}
	svc.roleSvc = roleSvc
	username := fmt.Sprintf("refresh-scope-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, context.Background(), username, "admin123")
	user := &entity.SysUser{Id: userID, Username: username, Status: 1}

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
	ctx := context.Background()
	svc := newTenantAuthTestService()
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}

	_, refreshToken, _, err := svc.generateTokenPair(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}
	if _, err = svc.ParseToken(ctx, refreshToken); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected refresh token to be rejected as access token, got %v", err)
	}
}

// TestRefreshRejectsRevokedSession verifies a valid refresh JWT is not enough
// when the online session has already been revoked.
func TestRefreshRejectsRevokedSession(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}

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
	ctx := context.Background()
	svc := newTenantAuthTestService()
	username := fmt.Sprintf("tenant-neg-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	user := &entity.SysUser{Id: userID, Username: username, Status: 1}

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

	username := fmt.Sprintf("tenant-infra-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	user := &entity.SysUser{Id: userID, Username: username, Status: 1}

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

	username := fmt.Sprintf("tenant-evict-%d", time.Now().UnixNano())
	userID := insertAuthTestUser(t, ctx, username, "admin123")
	user := &entity.SysUser{Id: userID, Username: username, Status: 1}

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
	ctx := context.Background()
	sharedCache := newSharedMemoryKVCache()
	firstSvc := newTenantAuthTestService()
	secondSvc := newTenantAuthTestService()
	firstSvc.revoked = newKVRevokeStore(sharedCache)
	secondSvc.revoked = newKVRevokeStore(sharedCache)
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	token, tokenID, err := firstSvc.generateToken(ctx, user, 11, tokencap.ClientTypeWeb)
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

// TestParseTokenRevokeReadFailureFailClosed verifies a valid JWT is rejected
// when the shared token-state store cannot confirm whether it has been revoked.
func TestParseTokenRevokeReadFailureFailClosed(t *testing.T) {
	ctx := context.Background()
	svc := newTenantAuthTestService()
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	token, _, err := svc.generateToken(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	svc.revoked = &failingRevokeStore{revokedErr: errors.New("simulated redis revoke read failure")}
	if _, err = svc.ParseToken(ctx, token); !bizerr.Is(err, CodeAuthTokenStateUnavailable) {
		t.Fatalf("expected revoke read failure to fail closed, got %v", err)
	}
}

// TestLogoutRevokesCurrentToken verifies logout removes the supplied token from
// the session store contract and writes shared JWT revocation state.
func TestLogoutRevokesCurrentToken(t *testing.T) {
	ctx := context.Background()
	store := newMemorySessionStore()
	sharedCache := newSharedMemoryKVCache()
	svc := newTenantAuthTestService()
	hooks := &recordingAuthHookService{}
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
	if _, err = svc.ParseToken(ctx, token); !bizerr.Is(err, CodeAuthTokenInvalid) {
		t.Fatalf("expected logged-out token to be rejected, got %v", err)
	}
	if len(hooks.logoutSucceeded) != 1 || hooks.logoutSucceeded[0].ClientType != tokencap.ClientTypeWeb.String() {
		t.Fatalf("expected logout hook clientType %q, got %#v", tokencap.ClientTypeWeb, hooks.logoutSucceeded)
	}
}

// TestRevokeSessionWritesSharedRevoke verifies force-logout style token-ID
// revocation publishes shared revoke state before removing the session row.
func TestRevokeSessionWritesSharedRevoke(t *testing.T) {
	ctx := context.Background()
	store := newMemorySessionStore()
	sharedCache := newSharedMemoryKVCache()
	svc := newTenantAuthTestService()
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
	ctx := context.Background()
	svc := newTenantAuthTestService()
	store := newMemorySessionStore()
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
	ctx := context.Background()
	svc := newTenantAuthTestService()
	store := newMemorySessionStore()
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
	ctx := context.Background()
	svc := newTenantAuthTestService()
	user := &entity.SysUser{Id: 101, Username: "tenant-user", Status: 1}
	oldToken, oldTokenID, err := svc.generateToken(ctx, user, 11, tokencap.ClientTypeWeb)
	if err != nil {
		t.Fatalf("generate old token: %v", err)
	}
	oldClaims, err := svc.ParseToken(ctx, oldToken)
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

// GetSessionTimeout returns a stable online-session lifetime for auth tests.
func (configTestService) GetSessionTimeout(context.Context) (time.Duration, error) {
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

// trackingRoleTestService records the tenant scope used to prime token access
// snapshots.
type trackingRoleTestService struct {
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

// HandleAuthLoginSucceeded records login success payloads.
func (s *recordingAuthHookService) HandleAuthLoginSucceeded(_ context.Context, input pluginhost.AuthHookPayloadInput) error {
	s.loginSucceeded = append(s.loginSucceeded, input)
	return nil
}

// HandleAuthLoginFailed records login failure payloads.
func (s *recordingAuthHookService) HandleAuthLoginFailed(_ context.Context, input pluginhost.AuthHookPayloadInput) error {
	s.loginFailed = append(s.loginFailed, input)
	return nil
}

// HandleAuthLogoutSucceeded records logout success payloads.
func (s *recordingAuthHookService) HandleAuthLogoutSucceeded(_ context.Context, input pluginhost.AuthHookPayloadInput) error {
	s.logoutSucceeded = append(s.logoutSucceeded, input)
	return nil
}

// enabledTenantAuthTestService enables tenant provider validation for auth tests.
type enabledTenantAuthTestService struct{}

// disabledTenantAuthTestService keeps multi-tenancy disabled for tests that
// only need the auth service to satisfy its explicit dependency contract.
type disabledTenantAuthTestService struct{}

// Available reports no active tenant provider for disabled tenancy tests.
func (disabledTenantAuthTestService) Available(context.Context) bool {
	return false
}

// Status returns an unavailable tenant capability status.
func (disabledTenantAuthTestService) Status(context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{}
}

// Current returns platform tenant for disabled tenancy tests.
func (disabledTenantAuthTestService) Current(context.Context) tenantcapsvc.TenantID {
	return tenantcap.PLATFORM
}

// Apply returns the input model unchanged when tenancy is disabled.
func (disabledTenantAuthTestService) Apply(_ context.Context, model *gdb.Model, _ string) (*gdb.Model, error) {
	return model, nil
}

// PlatformBypass reports platform bypass for disabled tenancy tests.
func (disabledTenantAuthTestService) PlatformBypass(context.Context) bool {
	return true
}

// EnsureTenantVisible accepts all tenants when tenancy is disabled.
func (disabledTenantAuthTestService) EnsureTenantVisible(context.Context, tenantcapsvc.TenantID) error {
	return nil
}

// ValidateUserInTenant accepts all users when tenancy is disabled.
func (disabledTenantAuthTestService) ValidateUserInTenant(context.Context, int, tenantcapsvc.TenantID) error {
	return nil
}

// ResolveTenant returns platform tenant for disabled tenancy tests.
func (disabledTenantAuthTestService) ResolveTenant(context.Context, *ghttp.Request) (*tenantcap.ResolverResult, error) {
	return &tenantcap.ResolverResult{TenantID: tenantcap.PLATFORM, Matched: true}, nil
}

// ListUserTenants returns no tenant options when tenancy is disabled.
func (disabledTenantAuthTestService) ListUserTenants(context.Context, int) ([]tenantcap.TenantInfo, error) {
	return []tenantcap.TenantInfo{}, nil
}

// SwitchTenant accepts tenant switches when tenancy is disabled.
func (disabledTenantAuthTestService) SwitchTenant(context.Context, int, tenantcapsvc.TenantID) error {
	return nil
}

// ApplyUserTenantScope returns the model unchanged when tenancy is disabled.
func (disabledTenantAuthTestService) ApplyUserTenantScope(
	_ context.Context,
	model *gdb.Model,
	_ string,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ApplyUserTenantFilter returns the model unchanged when tenancy is disabled.
func (disabledTenantAuthTestService) ApplyUserTenantFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ tenantcapsvc.TenantID,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenantProjections returns no projections when tenancy is disabled.
func (disabledTenantAuthTestService) ListUserTenantProjections(
	context.Context,
	[]int,
) (map[int]*tenantcap.UserTenantProjection, error) {
	return map[int]*tenantcap.UserTenantProjection{}, nil
}

// ResolveUserTenantAssignment returns an empty assignment plan when tenancy is disabled.
func (disabledTenantAuthTestService) ResolveUserTenantAssignment(
	context.Context,
	[]tenantcapsvc.TenantID,
	tenantcap.UserTenantAssignmentMode,
) (*tenantcap.UserTenantAssignmentPlan, error) {
	return &tenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is a no-op when tenancy is disabled.
func (disabledTenantAuthTestService) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*tenantcap.UserTenantAssignmentPlan,
) error {
	return nil
}

// EnsureUsersInTenant accepts all users when tenancy is disabled.
func (disabledTenantAuthTestService) EnsureUsersInTenant(context.Context, []int, tenantcapsvc.TenantID) error {
	return nil
}

// ValidateUserMembershipStartupConsistency returns no details when tenancy is disabled.
func (disabledTenantAuthTestService) ValidateUserMembershipStartupConsistency(context.Context) ([]string, error) {
	return nil, nil
}

// ProvisionAutoEnabledTenantPlugins is a no-op when tenancy is disabled.
func (disabledTenantAuthTestService) ProvisionAutoEnabledTenantPlugins(context.Context) error {
	return nil
}

// Available reports an active tenant provider for enabled tenancy tests.
func (enabledTenantAuthTestService) Available(context.Context) bool {
	return true
}

// Status returns an available tenant capability status.
func (enabledTenantAuthTestService) Status(context.Context) capmodel.CapabilityStatus {
	return capmodel.CapabilityStatus{Available: true, ActiveProvider: tenantcap.ProviderPluginID}
}

// Current returns the platform tenant for tests that do not carry request context.
func (enabledTenantAuthTestService) Current(context.Context) tenantcapsvc.TenantID {
	return tenantcap.PLATFORM
}

// CurrentTenantInfo returns the platform tenant projection for auth tests.
func (enabledTenantAuthTestService) CurrentTenantInfo(context.Context) (*tenantcap.TenantInfo, error) {
	return &tenantcap.TenantInfo{ID: tenantcap.PLATFORM}, nil
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

// ValidateUserInTenant accepts all users in auth tests.
func (enabledTenantAuthTestService) ValidateUserInTenant(context.Context, int, tenantcapsvc.TenantID) error {
	return nil
}

// ResolveTenant returns no request-derived tenant in auth tests.
func (enabledTenantAuthTestService) ResolveTenant(context.Context, *ghttp.Request) (*tenantcap.ResolverResult, error) {
	return &tenantcap.ResolverResult{TenantID: tenantcap.PLATFORM, Matched: true}, nil
}

// ListUserTenants returns no tenants in auth tests unless provider lookup is used directly.
func (enabledTenantAuthTestService) ListUserTenants(context.Context, int) ([]tenantcap.TenantInfo, error) {
	return []tenantcap.TenantInfo{}, nil
}

// BatchGetTenants reports requested tenants as missing in auth tests.
func (enabledTenantAuthTestService) BatchGetTenants(_ context.Context, tenantIDs []tenantcap.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
	return &capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID]{
		Items:      map[tenantcap.TenantID]*tenantcap.TenantInfo{},
		MissingIDs: append([]tenantcap.TenantID(nil), tenantIDs...),
	}, nil
}

// SearchTenants returns an empty tenant page in auth tests.
func (enabledTenantAuthTestService) SearchTenants(context.Context, tenantcap.SearchInput) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	return &capmodel.PageResult[*tenantcap.TenantInfo]{Items: []*tenantcap.TenantInfo{}}, nil
}

// BatchListUserTenants returns no user-tenant projections in auth tests.
func (enabledTenantAuthTestService) BatchListUserTenants(context.Context, []int) (map[int][]tenantcap.TenantInfo, error) {
	return map[int][]tenantcap.TenantInfo{}, nil
}

// EnsureTenantsVisible accepts all tenant batches in auth tests.
func (enabledTenantAuthTestService) EnsureTenantsVisible(context.Context, []tenantcap.TenantID) error {
	return nil
}

// SwitchTenant accepts tenant switches in auth tests.
func (enabledTenantAuthTestService) SwitchTenant(context.Context, int, tenantcapsvc.TenantID) error {
	return nil
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
) (map[int]*tenantcap.UserTenantProjection, error) {
	return map[int]*tenantcap.UserTenantProjection{}, nil
}

// ResolveUserTenantAssignment returns an empty plan in auth tests.
func (enabledTenantAuthTestService) ResolveUserTenantAssignment(
	context.Context,
	[]tenantcapsvc.TenantID,
	tenantcap.UserTenantAssignmentMode,
) (*tenantcap.UserTenantAssignmentPlan, error) {
	return &tenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is a no-op in auth tests.
func (enabledTenantAuthTestService) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*tenantcap.UserTenantAssignmentPlan,
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

// ProvisionAutoEnabledTenantPlugins is a no-op in auth tests.
func (enabledTenantAuthTestService) ProvisionAutoEnabledTenantPlugins(context.Context) error {
	return nil
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
func registerTenantAuthTestProvider(t *testing.T, tenantsByUser map[int][]tenantcap.TenantInfo) tenantcap.Service {
	t.Helper()
	return registerTenantAuthProviderInstance(t, &tenantAuthTestProvider{tenantsByUser: tenantsByUser})
}

// registerTenantAuthProviderInstance installs a temporary tenant provider
// through the pluginservice lifecycle-style registry.
func registerTenantAuthProviderInstance(t *testing.T, provider *tenantAuthTestProvider) tenantcap.Service {
	t.Helper()
	providerPluginID := fmt.Sprintf("plugin-test-auth-tenant-provider-%d", time.Now().UnixNano())
	manager := tenantspi.NewManager()
	if err := manager.RegisterFactory(providerPluginID, func(context.Context, tenantspi.ProviderEnv) (tenantspi.Provider, error) {
		return provider, nil
	}); err != nil {
		t.Fatalf("register auth tenant provider: %v", err)
	}
	return tenantspi.New(manager, tenantAuthProviderRuntime{pluginID: providerPluginID}, nil)
}

// tenantAuthProviderRuntime marks exactly one test provider plugin enabled.
type tenantAuthProviderRuntime struct {
	pluginID string
}

// IsProviderEnabled reports whether the given test provider plugin is enabled.
func (r tenantAuthProviderRuntime) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return pluginID == r.pluginID
}

// TenantProviderEnv returns an empty typed provider environment in auth tests.
func (tenantAuthProviderRuntime) TenantProviderEnv(string) tenantspi.ProviderEnv {
	return tenantspi.ProviderEnv{}
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
	_ interface {
		GetJwtSecret(context.Context) string
		GetJwtExpire(context.Context) (time.Duration, error)
		GetSessionTimeout(context.Context) (time.Duration, error)
	} = configTestService{}
	_ interface {
		PrimeTokenAccessContext(context.Context, string, int) (*role.UserAccessContext, error)
		InvalidateTokenAccessContext(context.Context, string)
	} = roleTestService{}
	_ session.Store        = (*memorySessionStore)(nil)
	_ kvcache.Service      = (*sharedMemoryKVCache)(nil)
	_ jwt.Claims           = (*Claims)(nil)
	_ tenantcapsvc.Service = enabledTenantAuthTestService{}
	_ tenantspi.Provider   = (*tenantAuthTestProvider)(nil)
)
