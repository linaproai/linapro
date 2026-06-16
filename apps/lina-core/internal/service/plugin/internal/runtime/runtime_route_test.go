// This file covers dynamic-route-specific session validation behaviors that
// are easy to regress during runtime auth changes.

package runtime

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	rolesvc "lina-core/internal/service/role"
	"lina-core/internal/service/session"
	_ "lina-core/pkg/dbdriver"
	tokencap "lina-core/pkg/plugin/capability/authcap/token"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	"lina-core/pkg/plugin/pluginhost"
)

// TestDynamicRouteEntrypointStaysSlim verifies scheme E keeps runtime_route.go as the
// dispatcher entrypoint instead of regrowing matcher, auth, envelope, or response logic.
func TestDynamicRouteEntrypointStaysSlim(t *testing.T) {
	root := runtimePackageDir(t)
	routeFile := readRuntimeSourceFile(t, root, "runtime_route.go")
	if lines := countSourceLines(routeFile); lines > 400 {
		t.Fatalf("runtime_route.go must stay at or below 400 lines, got %d", lines)
	}
	forbiddenSnippets := []string{
		"func (s *serviceImpl) matchDynamicRoute(",
		"func (s *serviceImpl) authorizeDynamicRouteRequest(",
		"func (s *serviceImpl) buildDynamicRouteRequestEnvelopeWithIdentity(",
		"func (s *serviceImpl) writeDynamicRouteResponse(",
		"func matchDynamicRoutePath(",
	}
	for _, snippet := range forbiddenSnippets {
		if strings.Contains(routeFile, snippet) {
			t.Fatalf("route.go must not own split route responsibility %q", snippet)
		}
	}

	expectedOwners := map[string]string{
		"runtime_route_match.go":    "func (s *serviceImpl) matchDynamicRoute(",
		"runtime_route_auth.go":     "func (s *serviceImpl) authorizeDynamicRouteRequest(",
		"runtime_route_envelope.go": "func (s *serviceImpl) buildDynamicRouteRequestEnvelopeWithIdentity(",
		"runtime_route_response.go": "func (s *serviceImpl) writeDynamicRouteResponse(",
	}
	for file, snippet := range expectedOwners {
		content := readRuntimeSourceFile(t, root, file)
		if !strings.Contains(content, snippet) {
			t.Fatalf("%s must own split route responsibility %q", file, snippet)
		}
	}
}

// TestTouchDynamicRouteSessionKeepsExistingSessionWhenTimestampDoesNotChange verifies
// that second-level TIMESTAMP precision does not invalidate an existing session.
func TestTouchDynamicRouteSessionKeepsExistingSessionWhenTimestampDoesNotChange(t *testing.T) {
	var (
		ctx      = context.Background()
		service  = &serviceImpl{sessionStore: session.NewDBStore()}
		tenantID = 17
		tokenID  = fmt.Sprintf("plugin-dev-dynamic-route-session-%d", time.Now().UnixNano())
	)

	if _, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete stale online session %s: %v", tokenID, err)
	}
	defer func() {
		if _, err := dao.SysOnlineSession.Ctx(ctx).
			Where(do.SysOnlineSession{TokenId: tokenID}).
			Delete(); err != nil {
			t.Fatalf("failed to cleanup online session %s: %v", tokenID, err)
		}
	}()

	currentSecond := waitForFreshSecond(t)
	_, err := dao.SysOnlineSession.Ctx(ctx).Data(do.SysOnlineSession{
		TokenId:        tokenID,
		TenantId:       tenantID,
		UserId:         1,
		Username:       "admin",
		ClientType:     "web",
		DeptName:       "系统管理",
		Ip:             "127.0.0.1",
		Browser:        "go-test",
		Os:             "darwin",
		LoginTime:      currentSecond,
		LastActiveTime: currentSecond,
	}).Insert()
	if err != nil {
		t.Fatalf("expected test session insert to succeed, got error: %v", err)
	}

	exists, err := service.touchDynamicRouteSession(ctx, tenantID, tokenID)
	if err != nil {
		t.Fatalf("expected first session touch to succeed, got error: %v", err)
	}
	if !exists {
		t.Fatal("expected first session touch to keep the session active")
	}

	// Touch the same session again within the same second to emulate the dynamic
	// route request arriving immediately after login or another protected request.
	exists, err = service.touchDynamicRouteSession(ctx, tenantID, tokenID)
	if err != nil {
		t.Fatalf("expected second session touch to succeed, got error: %v", err)
	}
	if !exists {
		t.Fatal("expected existing session to remain active when TIMESTAMP precision keeps the same second")
	}

	exists, err = service.touchDynamicRouteSession(ctx, tenantID+1, tokenID)
	if err != nil {
		t.Fatalf("expected cross-tenant session touch to be a clean miss, got error: %v", err)
	}
	if exists {
		t.Fatal("expected same token in another tenant to be treated as missing")
	}
}

// TestDynamicRouteIdentitySnapshotFiltersRolesByTokenTenant verifies that the
// runtime permission snapshot cannot reuse role grants from another tenant or
// platform scope when one user has role assignments in multiple tenants.
func TestDynamicRouteIdentitySnapshotFiltersRolesByTokenTenant(t *testing.T) {
	var (
		ctx          = context.Background()
		service      = &serviceImpl{jwtConfig: routeTestJwtConfig{secret: "route-tenant-secret"}, sessionStore: session.NewDBStore()}
		tenantAID    = 61001
		actingUserID = 9001
		tokenID      = fmt.Sprintf("plugin-dev-dynamic-route-tenant-token-%d", time.Now().UnixNano())
		tenantAPerm  = fmt.Sprintf("plugin-dev-dynamic-route:tenant-a:%d", time.Now().UnixNano())
		tenantBPerm  = fmt.Sprintf("plugin-dev-dynamic-route:tenant-b:%d", time.Now().UnixNano())
		platformPerm = fmt.Sprintf("plugin-dev-dynamic-route:platform:%d", time.Now().UnixNano())
	)
	userID := insertDynamicRouteAccessTestUser(t, ctx, tenantAID)
	t.Cleanup(func() {
		cleanupDynamicRouteAccessTestRows(t, ctx, tokenID, userID, nil, nil)
	})

	service.roleAccess = testRoleAccessProjector{projection: &rolesvc.DynamicRouteAccessProjection{
		Permissions: []string{tenantAPerm},
		RoleNames:   []string{"tenant-a"},
		DataScope:   datascope.ScopeSelf,
	}}
	insertDynamicRouteAccessTestSession(t, ctx, tenantAID, tokenID, userID)

	tokenString := signDynamicRouteImpersonationTestToken(t, service.jwtConfig, tokenID, tenantAID, userID, actingUserID)
	request := buildDynamicRouteAccessTestRequest(tokenString)
	identity, failure, err := service.buildDynamicRouteIdentitySnapshot(
		ctx,
		&dynamicRouteMatch{Route: &bridgecontract.RouteContract{Permission: tenantAPerm}},
		request,
	)
	if err != nil {
		t.Fatalf("expected tenant-scoped dynamic route identity to build, got error: %v", err)
	}
	if failure != nil {
		t.Fatalf("expected tenant-scoped dynamic route identity to pass, got failure: %#v", failure)
	}
	if identity == nil {
		t.Fatal("expected tenant-scoped dynamic route identity snapshot")
	}
	if identity.TenantId != int32(tenantAID) {
		t.Fatalf("expected identity tenant %d, got %d", tenantAID, identity.TenantId)
	}
	if identity.ActingUserId != int32(actingUserID) || !identity.ActingAsTenant || !identity.IsImpersonation {
		t.Fatalf("expected impersonation snapshot for acting user %d, got %#v", actingUserID, identity)
	}
	if !containsString(identity.Permissions, tenantAPerm) {
		t.Fatalf("expected tenant A permission in snapshot, got %#v", identity.Permissions)
	}
	if containsString(identity.Permissions, tenantBPerm) {
		t.Fatalf("expected tenant B permission to be filtered out, got %#v", identity.Permissions)
	}
	if containsString(identity.Permissions, platformPerm) {
		t.Fatalf("expected platform permission to be filtered out, got %#v", identity.Permissions)
	}
	if identity.DataScope != int32(datascope.ScopeSelf) {
		t.Fatalf("expected role projection data scope %d, got %d", datascope.ScopeSelf, identity.DataScope)
	}

	identity, failure, err = service.buildDynamicRouteIdentitySnapshot(
		ctx,
		&dynamicRouteMatch{Route: &bridgecontract.RouteContract{Permission: tenantBPerm}},
		request,
	)
	if err != nil {
		t.Fatalf("expected forbidden tenant B route to return a bridge failure, got error: %v", err)
	}
	if identity != nil {
		t.Fatalf("expected forbidden tenant B route not to return identity, got %#v", identity)
	}
	if failure == nil || failure.StatusCode != http.StatusForbidden {
		t.Fatalf("expected tenant B permission to be forbidden, got %#v", failure)
	}
}

// TestParseDynamicRouteTokenRejectsRefreshToken verifies dynamic plugin routes
// only accept access JWTs and cannot be called with refresh tokens.
func TestParseDynamicRouteTokenRejectsRefreshToken(t *testing.T) {
	ctx := context.Background()
	service := &serviceImpl{jwtConfig: routeTestJwtConfig{secret: "route-token-secret"}, sessionStore: session.NewDBStore()}

	testCases := []struct {
		name       string
		tokenType  string
		clientType string
	}{
		{name: "missing token type", tokenType: "", clientType: "web"},
		{name: "refresh token", tokenType: tokencap.KindRefresh, clientType: "web"},
		{name: "missing client type", tokenType: tokencap.KindAccess, clientType: ""},
		{name: "plugin client type", tokenType: tokencap.KindAccess, clientType: "plugin"},
		{name: "service client type", tokenType: tokencap.KindAccess, clientType: "service"},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, dynamicRouteClaims{
				TokenId:    "refresh-token-id",
				TokenType:  testCase.tokenType,
				ClientType: testCase.clientType,
				TenantId:   11,
				UserId:     1,
				Username:   "admin",
				Status:     statusNormal,
			})
			tokenString, err := token.SignedString([]byte("route-token-secret"))
			if err != nil {
				t.Fatalf("sign token: %v", err)
			}
			if _, err = service.parseDynamicRouteToken(ctx, tokenString); err == nil {
				t.Fatal("expected token to be rejected by dynamic route parser")
			}
		})
	}
}

// waitForFreshSecond aligns the test clock with a new second to avoid flaky TIMESTAMP updates.
func waitForFreshSecond(t *testing.T) *time.Time {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		now := time.Now()
		if now.Nanosecond() < int((100 * time.Millisecond).Nanoseconds()) {
			truncated := now.Truncate(time.Second)
			return &truncated
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("failed to align test to the beginning of a second")
	return nil
}

// routeTestJwtConfig provides deterministic JWT signing configuration for
// dynamic route identity tests.
type routeTestJwtConfig struct {
	secret string
}

// GetJwtSecret returns the fixed JWT signing secret for route tests.
func (c routeTestJwtConfig) GetJwtSecret(_ context.Context) string {
	return c.secret
}

// GetSessionTimeout returns the fixed online-session timeout for route tests.
func (c routeTestJwtConfig) GetSessionTimeout(context.Context) (time.Duration, error) {
	return time.Hour, nil
}

// insertDynamicRouteAccessTestUser inserts one temporary user bound to a
// primary tenant for dynamic route access tests.
func insertDynamicRouteAccessTestUser(t *testing.T, ctx context.Context, tenantID int) int {
	t.Helper()

	username := fmt.Sprintf("dynamic-route-access-%d", time.Now().UnixNano())
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "test-password-hash",
		Nickname: username,
		Status:   statusNormal,
		TenantId: tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dynamic route access test user: %v", err)
	}
	return int(id)
}

// insertDynamicRouteAccessTestRole inserts one temporary role for a specific
// tenant boundary.
func insertDynamicRouteAccessTestRole(
	t *testing.T,
	ctx context.Context,
	tenantID int,
	label string,
	dataScope int,
) int {
	t.Helper()

	suffix := time.Now().UnixNano()
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      fmt.Sprintf("dynamic-route-%s-%d", label, suffix),
		Key:       fmt.Sprintf("dynamic-route-%s-%d", label, suffix),
		Sort:      99,
		DataScope: dataScope,
		Status:    statusNormal,
		TenantId:  tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dynamic route access test role: %v", err)
	}
	return int(id)
}

// insertDynamicRouteAccessTestMenu inserts one temporary global button
// permission menu. Tenant boundaries are represented by role-menu grants.
func insertDynamicRouteAccessTestMenu(
	t *testing.T,
	ctx context.Context,
	label string,
	permission string,
) int {
	t.Helper()

	menuKey := fmt.Sprintf("dynamic-route:%s:%d", label, time.Now().UnixNano())
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		ParentId: 0,
		MenuKey:  menuKey,
		Name:     menuKey,
		Perms:    permission,
		Type:     catalog.MenuTypeButton.String(),
		Sort:     99,
		Visible:  statusNormal,
		Status:   statusNormal,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert dynamic route access test menu: %v", err)
	}
	return int(id)
}

// insertDynamicRouteAccessTestGrant binds one user, role, and permission menu
// within the same tenant boundary.
func insertDynamicRouteAccessTestGrant(
	t *testing.T,
	ctx context.Context,
	tenantID int,
	userID int,
	roleID int,
	menuID int,
) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId:   userID,
		RoleId:   roleID,
		TenantId: tenantID,
	}).Insert(); err != nil {
		t.Fatalf("insert dynamic route access test user-role: %v", err)
	}
	if _, err := dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{
		RoleId:   roleID,
		MenuId:   menuID,
		TenantId: tenantID,
	}).Insert(); err != nil {
		t.Fatalf("insert dynamic route access test role-menu: %v", err)
	}
}

// insertDynamicRouteAccessTestSession inserts one active session row for a
// tenant-scoped dynamic route token.
func insertDynamicRouteAccessTestSession(
	t *testing.T,
	ctx context.Context,
	tenantID int,
	tokenID string,
	userID int,
) {
	t.Helper()

	now := time.Now()
	if err := session.NewDBStore().Set(ctx, &session.Session{
		TokenId:        tokenID,
		TenantId:       tenantID,
		UserId:         userID,
		Username:       "dynamic-route-access",
		ClientType:     "web",
		DeptName:       "runtime-test",
		Ip:             "127.0.0.1",
		Browser:        "go-test",
		Os:             "darwin",
		LoginTime:      &now,
		LastActiveTime: &now,
	}); err != nil {
		t.Fatalf("insert dynamic route access test session: %v", err)
	}
}

// signDynamicRouteAccessTestToken signs a dynamic route token carrying the
// tenant under test.
func signDynamicRouteAccessTestToken(
	t *testing.T,
	config JwtConfigProvider,
	tokenID string,
	tenantID int,
	userID int,
) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dynamicRouteClaims{
		TokenId:    tokenID,
		TokenType:  tokencap.KindAccess,
		ClientType: "web",
		TenantId:   tenantID,
		UserId:     userID,
		Username:   "dynamic-route-access",
		Status:     statusNormal,
	})
	tokenString, err := token.SignedString([]byte(config.GetJwtSecret(context.Background())))
	if err != nil {
		t.Fatalf("sign dynamic route access test token: %v", err)
	}
	return tokenString
}

// signDynamicRouteImpersonationTestToken signs a dynamic route token carrying
// tenant impersonation metadata.
func signDynamicRouteImpersonationTestToken(
	t *testing.T,
	config JwtConfigProvider,
	tokenID string,
	tenantID int,
	userID int,
	actingUserID int,
) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dynamicRouteClaims{
		TokenId:         tokenID,
		TokenType:       tokencap.KindAccess,
		ClientType:      "web",
		TenantId:        tenantID,
		UserId:          userID,
		Username:        "dynamic-route-access",
		Status:          statusNormal,
		ActingUserId:    actingUserID,
		ActingAsTenant:  true,
		IsImpersonation: true,
	})
	tokenString, err := token.SignedString([]byte(config.GetJwtSecret(context.Background())))
	if err != nil {
		t.Fatalf("sign dynamic route impersonation test token: %v", err)
	}
	return tokenString
}

// buildDynamicRouteAccessTestRequest creates a GoFrame request carrying one
// bearer token.
func buildDynamicRouteAccessTestRequest(tokenString string) *ghttp.Request {
	request := &ghttp.Request{}
	request.Request = httptest.NewRequest(http.MethodGet, pluginhost.PluginAPINamespacePrefix+"/plugin-dev-dynamic-route/access", nil)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	return request
}

// cleanupDynamicRouteAccessTestRows removes all rows created by the dynamic
// route tenant access test.
func cleanupDynamicRouteAccessTestRows(
	t *testing.T,
	ctx context.Context,
	tokenID string,
	userID int,
	roleIDs []int,
	menuIDs []int,
) {
	t.Helper()

	if userID <= 0 {
		return
	}
	if _, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenID}).
		Delete(); err != nil {
		t.Fatalf("cleanup dynamic route access test session: %v", err)
	}
	if _, err := dao.SysUserRole.Ctx(ctx).
		Where(do.SysUserRole{UserId: userID}).
		Delete(); err != nil {
		t.Fatalf("cleanup dynamic route access test user-role rows: %v", err)
	}
	if len(roleIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).
			WhereIn(dao.SysRoleMenu.Columns().RoleId, intsToInterfaces(roleIDs)).
			Delete(); err != nil {
			t.Fatalf("cleanup dynamic route access test role-menu rows by role: %v", err)
		}
		if _, err := dao.SysRole.Ctx(ctx).
			Unscoped().
			WhereIn(dao.SysRole.Columns().Id, intsToInterfaces(roleIDs)).
			Delete(); err != nil {
			t.Fatalf("cleanup dynamic route access test roles: %v", err)
		}
	}
	if len(menuIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).
			WhereIn(dao.SysRoleMenu.Columns().MenuId, intsToInterfaces(menuIDs)).
			Delete(); err != nil {
			t.Fatalf("cleanup dynamic route access test role-menu rows by menu: %v", err)
		}
		if _, err := dao.SysMenu.Ctx(ctx).
			Unscoped().
			WhereIn(dao.SysMenu.Columns().Id, intsToInterfaces(menuIDs)).
			Delete(); err != nil {
			t.Fatalf("cleanup dynamic route access test menus: %v", err)
		}
	}
	if _, err := dao.SysUser.Ctx(ctx).
		Unscoped().
		Where(do.SysUser{Id: userID}).
		Delete(); err != nil {
		t.Fatalf("cleanup dynamic route access test user: %v", err)
	}
}

// containsString reports whether target appears in values.
func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// intsToInterfaces converts test IDs into interface values for WhereIn.
func intsToInterfaces(values []int) []interface{} {
	items := make([]interface{}, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	return items
}

func runtimePackageDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve runtime package directory failed")
	}
	return filepath.Dir(file)
}

func readRuntimeSourceFile(t *testing.T, root string, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatalf("read %s failed: %v", name, err)
	}
	return string(content)
}

func countSourceLines(content string) int {
	if content == "" {
		return 0
	}
	lines := strings.Count(content, "\n")
	if !strings.HasSuffix(content, "\n") {
		lines++
	}
	return lines
}
