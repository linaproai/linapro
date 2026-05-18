// This file verifies startup consistency checks for plugin tenant governance.

package plugin

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// TestValidateStartupConsistencyRequiresInjectedTenantCapability verifies
// startup validation fails fast instead of building an implicit tenant service.
func TestValidateStartupConsistencyRequiresInjectedTenantCapability(t *testing.T) {
	var (
		service = newTestService()
		ctx     = context.Background()
	)
	service.SetTenantCapability(nil)

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "requires injected tenant capability service")
}

// TestValidateStartupConsistencyUsesInjectedTenantCapability verifies tenant
// membership checks run through the explicitly wired tenant capability.
func TestValidateStartupConsistencyUsesInjectedTenantCapability(t *testing.T) {
	var (
		service   = newTestService()
		ctx       = context.Background()
		tenantSvc = &startupConsistencyTenantCapability{details: []string{"injected tenant capability used"}}
	)
	service.SetTenantCapability(tenantSvc)

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "injected tenant capability used")
	if tenantSvc.calls != 1 {
		t.Fatalf("expected one injected tenant capability call, got %d", tenantSvc.calls)
	}
}

// TestValidateStartupConsistencyRejectsInvalidPluginGovernance verifies invalid
// scope_nature/install_mode combinations fail before serving requests.
func TestValidateStartupConsistencyRejectsInvalidPluginGovernance(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-startup-invalid-governance"
	)
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Startup Invalid Governance",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: "invalid_scope",
		InstallMode: catalog.InstallModeTenantScoped.String(),
	})

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err,
		"plugin "+pluginID+" has invalid scope_nature invalid_scope",
	)
}

// TestValidateStartupConsistencyRejectsPlatformOnlyTenantScoped verifies
// platform-only plugins must remain globally installed.
func TestValidateStartupConsistencyRejectsPlatformOnlyTenantScoped(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-startup-platform-tenant-scoped"
	)
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Startup Platform Tenant Scoped",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: catalog.ScopeNaturePlatformOnly.String(),
		InstallMode: catalog.InstallModeTenantScoped.String(),
	})

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "platform_only plugin "+pluginID+" must use global install_mode")
}

// TestValidateStartupConsistencyRejectsPlatformUserMembership verifies
// platform users are not allowed to carry active tenant memberships.
func TestValidateStartupConsistencyRejectsPlatformUserMembership(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		username = "startup-platform-member"
		tenantID = 19001
	)
	cleanupStartupConsistencyPlugin(t, ctx, pkgtenantcap.ProviderPluginID)
	cleanupStartupConsistencyUserMembership(t, ctx, username, tenantID)
	pkgtenantcap.RegisterProvider(&startupConsistencyTenantProvider{})
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pkgtenantcap.ProviderPluginID) })
	t.Cleanup(func() { cleanupStartupConsistencyUserMembership(t, ctx, username, tenantID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pkgtenantcap.ProviderPluginID,
		Name:        "Multi Tenant Provider",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: catalog.ScopeNaturePlatformOnly.String(),
		InstallMode: catalog.InstallModeGlobal.String(),
	})
	userID := insertStartupConsistencyUser(t, ctx, username, int(pkgtenantcap.PLATFORM))
	insertStartupConsistencyTenantMembership(t, ctx, userID, tenantID, 1)

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "platform user "+username)
}

// TestValidateStartupConsistencyRejectsEnabledTenantPluginWithoutProvider
// verifies multi-tenant enablement requires a registered tenantcap provider.
func TestValidateStartupConsistencyRejectsEnabledTenantPluginWithoutProvider(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = pkgtenantcap.ProviderPluginID
	)
	pkgtenantcap.RegisterProvider(nil)
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Multi Tenant Provider",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: catalog.ScopeNaturePlatformOnly.String(),
		InstallMode: catalog.InstallModeGlobal.String(),
	})

	err := service.ValidateStartupConsistency(ctx)
	assertStartupConsistencyErrorContains(t, err, "multi-tenant plugin is enabled but tenantcap provider is not registered")
}

// TestValidateStartupConsistencyAllowsEnabledTenantPluginWithProvider verifies
// provider registration satisfies multi-tenant startup consistency.
func TestValidateStartupConsistencyAllowsEnabledTenantPluginWithProvider(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = pkgtenantcap.ProviderPluginID
	)
	pkgtenantcap.RegisterProvider(&startupConsistencyTenantProvider{})
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })
	cleanupStartupConsistencyPlugin(t, ctx, pluginID)
	t.Cleanup(func() { cleanupStartupConsistencyPlugin(t, ctx, pluginID) })

	insertStartupConsistencyPlugin(t, ctx, do.SysPlugin{
		PluginId:    pluginID,
		Name:        "Multi Tenant Provider",
		Version:     "v0.1.0",
		Type:        catalog.TypeSource.String(),
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: catalog.ScopeNaturePlatformOnly.String(),
		InstallMode: catalog.InstallModeGlobal.String(),
	})

	if err := service.ValidateStartupConsistency(ctx); err != nil {
		t.Fatalf("expected registered provider to satisfy startup consistency, got %v", err)
	}
}

// startupConsistencyTenantProvider is a no-op tenant provider for startup tests.
type startupConsistencyTenantProvider struct{}

// ResolveTenant returns a platform match for interface completeness.
func (startupConsistencyTenantProvider) ResolveTenant(context.Context, *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant accepts all users for interface completeness.
func (startupConsistencyTenantProvider) ValidateUserInTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ListUserTenants returns no tenant rows for interface completeness.
func (startupConsistencyTenantProvider) ListUserTenants(context.Context, int) ([]pkgtenantcap.TenantInfo, error) {
	return nil, nil
}

// SwitchTenant accepts all switches for interface completeness.
func (startupConsistencyTenantProvider) SwitchTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ApplyUserTenantScope is unused by startup tests.
func (startupConsistencyTenantProvider) ApplyUserTenantScope(
	_ context.Context,
	model *gdb.Model,
	_ string,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ApplyUserTenantFilter is unused by startup tests.
func (startupConsistencyTenantProvider) ApplyUserTenantFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ pkgtenantcap.TenantID,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenantProjections is unused by startup tests.
func (startupConsistencyTenantProvider) ListUserTenantProjections(
	context.Context,
	[]int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	return map[int]*pkgtenantcap.UserTenantProjection{}, nil
}

// ResolveUserTenantAssignment is unused by startup tests.
func (startupConsistencyTenantProvider) ResolveUserTenantAssignment(
	context.Context,
	[]pkgtenantcap.TenantID,
	pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	return &pkgtenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is unused by startup tests.
func (startupConsistencyTenantProvider) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*pkgtenantcap.UserTenantAssignmentPlan,
) error {
	return nil
}

// EnsureUsersInTenant is unused by startup tests.
func (startupConsistencyTenantProvider) EnsureUsersInTenant(context.Context, []int, pkgtenantcap.TenantID) error {
	return nil
}

// ValidateStartupConsistency checks platform-user membership violations.
func (startupConsistencyTenantProvider) ValidateStartupConsistency(ctx context.Context) ([]string, error) {
	return validateStartupConsistencyTestMemberships(ctx)
}

// startupConsistencyTenantCapability records startup membership validation calls.
type startupConsistencyTenantCapability struct {
	calls   int
	details []string
}

// Enabled is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) Enabled(context.Context) bool {
	return true
}

// Current is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) Current(context.Context) pkgtenantcap.TenantID {
	return pkgtenantcap.PLATFORM
}

// Apply is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) Apply(_ context.Context, model *gdb.Model, _ string) (*gdb.Model, error) {
	return model, nil
}

// PlatformBypass is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) PlatformBypass(context.Context) bool {
	return false
}

// EnsureTenantVisible is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) EnsureTenantVisible(context.Context, pkgtenantcap.TenantID) error {
	return nil
}

// ResolveTenant is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) ResolveTenant(context.Context, *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
}

// ReadWithPlatformFallback is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) ReadWithPlatformFallback(
	ctx context.Context,
	scanner tenantcapsvc.FallbackScanner[any],
) ([]any, error) {
	return scanner(ctx, pkgtenantcap.PLATFORM)
}

// ApplyUserTenantScope is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) ApplyUserTenantScope(
	_ context.Context,
	model *gdb.Model,
	_ string,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenants is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) ListUserTenants(context.Context, int) ([]pkgtenantcap.TenantInfo, error) {
	return nil, nil
}

// ApplyUserTenantFilter is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) ApplyUserTenantFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ pkgtenantcap.TenantID,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenantProjections is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) ListUserTenantProjections(
	context.Context,
	[]int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	return map[int]*pkgtenantcap.UserTenantProjection{}, nil
}

// ResolveUserTenantAssignment is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) ResolveUserTenantAssignment(
	context.Context,
	[]pkgtenantcap.TenantID,
	pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	return &pkgtenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*pkgtenantcap.UserTenantAssignmentPlan,
) error {
	return nil
}

// EnsureUsersInTenant is unused by plugin startup consistency tests.
func (s *startupConsistencyTenantCapability) EnsureUsersInTenant(context.Context, []int, pkgtenantcap.TenantID) error {
	return nil
}

// ValidateUserMembershipStartupConsistency records the injected startup check.
func (s *startupConsistencyTenantCapability) ValidateUserMembershipStartupConsistency(context.Context) ([]string, error) {
	s.calls++
	return s.details, nil
}

// validateStartupConsistencyTestMemberships simulates the plugin-owned startup
// consistency check without making host production code know plugin tables.
func validateStartupConsistencyTestMemberships(ctx context.Context) ([]string, error) {
	rows := make([]struct {
		Id       int    `json:"id" orm:"id"`
		Username string `json:"username" orm:"username"`
	}, 0)
	err := dao.SysUser.Ctx(ctx).
		As("u").
		Fields("u.id, u.username").
		InnerJoin(
			"plugin_multi_tenant_user_membership m",
			"m.user_id = u.id AND m.deleted_at IS NULL AND m.status = 1",
		).
		Where("u.tenant_id", int(pkgtenantcap.PLATFORM)).
		Limit(10).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	details := make([]string, 0, len(rows))
	for _, row := range rows {
		details = append(details, "platform user "+row.Username+"("+gconv.String(row.Id)+") must not have active tenant membership")
	}
	return details, nil
}

// insertStartupConsistencyPlugin inserts one isolated plugin row for validation tests.
func insertStartupConsistencyPlugin(t *testing.T, ctx context.Context, data do.SysPlugin) {
	t.Helper()

	if _, err := dao.SysPlugin.Ctx(ctx).Data(data).Insert(); err != nil {
		t.Fatalf("insert startup consistency plugin: %v", err)
	}
}

// cleanupStartupConsistencyPlugin removes one isolated plugin row.
func cleanupStartupConsistencyPlugin(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	if _, err := dao.SysPlugin.Ctx(ctx).Unscoped().Where(do.SysPlugin{PluginId: pluginID}).Delete(); err != nil {
		t.Fatalf("cleanup startup consistency plugin: %v", err)
	}
}

// insertStartupConsistencyUser inserts one user row for startup validation tests.
func insertStartupConsistencyUser(t *testing.T, ctx context.Context, username string, tenantID int) int64 {
	t.Helper()

	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "startup-consistency-test",
		Nickname: username,
		Status:   1,
		TenantId: tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert startup consistency user: %v", err)
	}
	return id
}

// insertStartupConsistencyTenantMembership inserts one membership row for startup validation tests.
func insertStartupConsistencyTenantMembership(t *testing.T, ctx context.Context, userID int64, tenantID int, status int) {
	t.Helper()

	_, err := dao.SysUser.DB().Model("plugin_multi_tenant_user_membership").Data(startupConsistencyMembershipRow{
		UserID:    userID,
		TenantID:  tenantID,
		Status:    status,
		CreatedBy: 0,
		UpdatedBy: 0,
	}).Insert()
	if err != nil {
		t.Fatalf("insert startup consistency membership: %v", err)
	}
}

// startupConsistencyMembershipRow models the plugin membership columns touched
// by startup consistency tests without importing plugin-internal generated DOs.
type startupConsistencyMembershipRow struct {
	UserID    int64 `orm:"user_id"`
	TenantID  int   `orm:"tenant_id"`
	Status    int   `orm:"status"`
	CreatedBy int   `orm:"created_by"`
	UpdatedBy int   `orm:"updated_by"`
}

// cleanupStartupConsistencyUserMembership removes startup validation user fixtures.
func cleanupStartupConsistencyUserMembership(t *testing.T, ctx context.Context, username string, tenantID int) {
	t.Helper()

	var user *entity.SysUser
	if err := dao.SysUser.Ctx(ctx).Unscoped().Where(do.SysUser{Username: username}).Scan(&user); err != nil {
		t.Fatalf("query startup consistency user cleanup: %v", err)
	}
	if user != nil {
		if _, err := dao.SysUser.DB().Model("plugin_multi_tenant_user_membership").
			Unscoped().
			Where("user_id", user.Id).
			Delete(); err != nil {
			t.Fatalf("cleanup startup consistency membership by user: %v", err)
		}
	}
	if _, err := dao.SysUser.DB().Model("plugin_multi_tenant_user_membership").
		Unscoped().
		Where("tenant_id", tenantID).
		Delete(); err != nil {
		t.Fatalf("cleanup startup consistency membership by tenant: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().Where(do.SysUser{Username: username}).Delete(); err != nil {
		t.Fatalf("cleanup startup consistency user: %v", err)
	}
}

// assertStartupConsistencyErrorContains verifies startup errors use the stable
// bizerr code and include actionable details.
func assertStartupConsistencyErrorContains(t *testing.T, err error, expectedDetails ...string) {
	t.Helper()

	if !bizerr.Is(err, CodePluginStartupConsistencyFailed) {
		t.Fatalf("expected startup consistency bizerr, got %v", err)
	}
	message := err.Error()
	for _, detail := range expectedDetails {
		if !strings.Contains(message, detail) {
			t.Fatalf("expected startup consistency error to include %q, got %q", detail, message)
		}
	}
}
