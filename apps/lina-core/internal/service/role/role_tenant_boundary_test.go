// This file verifies tenant ownership for roles and role relation rows.

package role

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

const (
	// roleTenantBoundaryStatusNormal is the enabled status used by role tests.
	roleTenantBoundaryStatusNormal = 1
)

// TestCreateWritesTenantOwnershipAndRoleMenuTenant verifies role creation
// persists the current tenant on both role and role-menu rows.
func TestCreateWritesTenantOwnershipAndRoleMenuTenant(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62001)
	svc := New(nil)
	menuID := insertRoleTenantBoundaryMenu(t, ctx, "tenant-create", "system:tenant:create", 62001)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, nil, nil, []int{menuID})
	})

	roleID, err := svc.Create(ctx, CreateInput{
		Name:      uniqueRoleTenantBoundaryName("tenant-role"),
		Key:       uniqueRoleTenantBoundaryName("tenant-role-key"),
		Sort:      1,
		DataScope: roleDataScopeSelf,
		Status:    roleTenantBoundaryStatusNormal,
		MenuIds:   []int{menuID},
	})
	if err != nil {
		t.Fatalf("create tenant role: %v", err)
	}
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, nil, nil)
	})

	roleRow := mustQueryRoleTenantBoundaryRole(t, ctx, roleID)
	if roleRow.TenantId != 62001 {
		t.Fatalf("expected tenant role ownership 62001, got tenant=%d", roleRow.TenantId)
	}
	if count := mustCountRoleTenantBoundaryRoleMenu(t, ctx, roleID, 62001); count != 1 {
		t.Fatalf("expected one tenant role-menu row, got %d", count)
	}
}

// TestAssignUsersWritesCurrentTenantRelation verifies role assignments persist
// the role tenant boundary.
func TestAssignUsersWritesCurrentTenantRelation(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62011)
	svc := New(nil).(*serviceImpl)
	svc.tenantSvc = tenantcapsvc.New(roleTenantBoundaryEnablementReader{})
	roleID := insertRoleTenantBoundaryRole(t, ctx, "tenant-assign", 62011)
	operatorRoleID := insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-assign-operator", 62011, roleDataScopeTenant)
	userID := insertRoleTenantBoundaryUser(t, ctx, "tenant-assign-user", 62011)
	ensureRoleTenantBoundaryMembershipTable(t, ctx)
	pkgtenantcap.RegisterProvider(roleTenantBoundaryProvider{})
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupRoleTestRows(t, ctx, []int{roleID, operatorRoleID}, []int{userID}, nil)
		cleanupRoleTenantBoundaryMembershipRows(t, ctx, []int{userID})
	})
	insertRoleTenantBoundaryUserRole(t, ctx, userID, operatorRoleID, 62011)
	insertRoleTenantBoundaryMembership(t, ctx, userID, 62011, 1)
	svc.bizCtxSvc = roleScopeStaticBizCtx{ctx: &model.Context{UserId: userID, TenantId: 62011}}

	if err := svc.AssignUsers(ctx, roleID, []int{userID}); err != nil {
		t.Fatalf("assign tenant role user: %v", err)
	}
	if count := mustCountRoleTenantBoundaryUserRole(t, ctx, roleID, userID, 62011); count != 1 {
		t.Fatalf("expected one tenant user-role row, got %d", count)
	}
}

// TestTenantRoleRejectsAllDataScope verifies tenant-local roles cannot receive
// cross-tenant all-data scope.
func TestTenantRoleRejectsAllDataScope(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62021)
	svc := New(nil)

	_, err := svc.Create(ctx, CreateInput{
		Name:      uniqueRoleTenantBoundaryName("tenant-all-data-deny"),
		Key:       uniqueRoleTenantBoundaryName("tenant-all-data-deny-key"),
		Sort:      1,
		DataScope: roleDataScopeAll,
		Status:    roleTenantBoundaryStatusNormal,
	})
	if !bizerr.Is(err, CodeTenantRoleAllDataScopeForbidden) {
		t.Fatalf("expected tenant all-data scope denial, got %v", err)
	}
}

// TestPlatformContextRoleRejectsTenantPrimaryUser verifies platform-context
// roles cannot be assigned to tenant-primary users.
func TestPlatformContextRoleRejectsTenantPrimaryUser(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), datascope.PlatformTenantID)
	svc := New(nil).(*serviceImpl)
	adminUserID, _ := mustQueryAdminUserAndRoleID(t, ctx)
	roleID := insertRoleTenantBoundaryRoleWithScope(t, ctx, "platform-role", 0, roleDataScopeAll)
	userID := insertRoleTenantBoundaryUser(t, ctx, "tenant-primary-user", 62022)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, []int{userID}, nil)
	})
	svc.bizCtxSvc = roleScopeStaticBizCtx{ctx: &model.Context{UserId: adminUserID, TenantId: datascope.PlatformTenantID}}

	err := svc.AssignUsers(ctx, roleID, []int{userID})
	if !bizerr.Is(err, CodePlatformRoleAssignmentForbidden) {
		t.Fatalf("expected platform role assignment denial, got %v", err)
	}
}

// TestTenantRoleRejectsPlatformPrimaryUser verifies tenant roles cannot be
// assigned to platform users, which would make platform authority tenant-local.
func TestTenantRoleRejectsPlatformPrimaryUser(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62023)
	svc := New(nil).(*serviceImpl)
	roleID := insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-role-platform-user-deny", 62023, roleDataScopeTenant)
	operatorRoleID := insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-role-platform-user-operator", 62023, roleDataScopeTenant)
	operatorUserID := insertRoleTenantBoundaryUser(t, ctx, "tenant-role-platform-user-operator", 62023)
	platformUserID := insertRoleTenantBoundaryUser(t, ctx, "tenant-role-platform-user", 0)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID, operatorRoleID}, []int{operatorUserID, platformUserID}, nil)
	})
	insertRoleTenantBoundaryUserRole(t, ctx, operatorUserID, operatorRoleID, 62023)
	svc.bizCtxSvc = roleScopeStaticBizCtx{ctx: &model.Context{UserId: operatorUserID, TenantId: 62023}}

	err := svc.AssignUsers(ctx, roleID, []int{platformUserID})
	if !bizerr.Is(err, CodeTenantRoleAssignmentForbidden) {
		t.Fatalf("expected tenant role assignment to platform user to be denied, got %v", err)
	}
	if count := mustCountRoleTenantBoundaryUserRole(t, ctx, roleID, platformUserID, 62023); count != 0 {
		t.Fatalf("expected no tenant role relation for platform user, got %d", count)
	}
}

// TestTenantRoleRequiresActiveMembershipWhenTableExists verifies tenant role
// assignment checks the multi-tenant membership table when it is installed.
func TestTenantRoleRequiresActiveMembershipWhenTableExists(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62024)
	svc := New(nil).(*serviceImpl)
	svc.tenantSvc = tenantcapsvc.New(roleTenantBoundaryEnablementReader{})
	ensureRoleTenantBoundaryMembershipTable(t, ctx)
	roleID := insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-role-membership-deny", 62024, roleDataScopeTenant)
	operatorRoleID := insertRoleTenantBoundaryRoleWithScope(t, ctx, "tenant-role-membership-operator", 62024, roleDataScopeTenant)
	operatorUserID := insertRoleTenantBoundaryUser(t, ctx, "tenant-role-membership-operator", 62024)
	targetUserID := insertRoleTenantBoundaryUser(t, ctx, "tenant-role-membership-target", 62024)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID, operatorRoleID}, []int{operatorUserID, targetUserID}, nil)
		cleanupRoleTenantBoundaryMembershipRows(t, ctx, []int{operatorUserID, targetUserID})
	})
	insertRoleTenantBoundaryUserRole(t, ctx, operatorUserID, operatorRoleID, 62024)
	insertRoleTenantBoundaryMembership(t, ctx, operatorUserID, 62024, 1)
	pkgtenantcap.RegisterProvider(roleTenantBoundaryProvider{})
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })
	svc.bizCtxSvc = roleScopeStaticBizCtx{ctx: &model.Context{UserId: operatorUserID, TenantId: 62024}}

	err := svc.AssignUsers(ctx, roleID, []int{targetUserID})
	if !bizerr.Is(err, CodeTenantRoleAssignmentForbidden) {
		t.Fatalf("expected tenant role assignment without membership to be denied, got %v", err)
	}
	if count := mustCountRoleTenantBoundaryUserRole(t, ctx, roleID, targetUserID, 62024); count != 0 {
		t.Fatalf("expected no tenant role relation without membership, got %d", count)
	}

	insertRoleTenantBoundaryMembership(t, ctx, targetUserID, 62024, 1)
	if err = svc.AssignUsers(ctx, roleID, []int{targetUserID}); err != nil {
		t.Fatalf("expected tenant role assignment with active membership to succeed, got %v", err)
	}
	if count := mustCountRoleTenantBoundaryUserRole(t, ctx, roleID, targetUserID, 62024); count != 1 {
		t.Fatalf("expected one tenant role relation with membership, got %d", count)
	}
}

// TestTenantRoleAccessFiltersRoleMenuByTenant verifies permission resolution
// does not reuse role-menu rows from another tenant for the same role ID.
func TestTenantRoleAccessFiltersRoleMenuByTenant(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62031)
	svc := New(nil).(*serviceImpl)
	roleID := insertRoleTenantBoundaryRole(t, ctx, "tenant-access", 62031)
	userID := insertRoleTenantBoundaryUser(t, ctx, "tenant-access-user", 62031)
	tenantMenuID := insertRoleTenantBoundaryMenu(t, ctx, "tenant-access-menu", "system:tenant:visible", 62031)
	otherMenuID := insertRoleTenantBoundaryMenu(t, ctx, "tenant-access-other-menu", "system:tenant:hidden", 62032)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, []int{userID}, []int{tenantMenuID, otherMenuID})
	})
	insertRoleTenantBoundaryRoleMenu(t, ctx, roleID, tenantMenuID, 62031)
	insertRoleTenantBoundaryRoleMenu(t, ctx, roleID, otherMenuID, 62032)
	insertRoleTenantBoundaryUserRole(t, ctx, userID, roleID, 62031)

	access, err := svc.GetUserAccessContext(ctx, userID)
	if err != nil {
		t.Fatalf("load tenant role access: %v", err)
	}
	if access == nil || !containsRoleTenantBoundaryString(access.Permissions, "system:tenant:visible") {
		t.Fatalf("expected tenant permission in access snapshot, got %#v", access)
	}
	if containsRoleTenantBoundaryString(access.Permissions, "system:tenant:hidden") {
		t.Fatalf("did not expect cross-tenant permission in access snapshot, got %#v", access.Permissions)
	}
}

// TestImpersonationAccessUsesPlatformRoles verifies tenant impersonation keeps
// target-tenant data context while permission grants come from platform roles.
func TestImpersonationAccessUsesPlatformRoles(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62041)
	svc := New(nil).(*serviceImpl)
	roleID := insertRoleTenantBoundaryRoleWithScope(t, ctx, "impersonation-platform-role", datascope.PlatformTenantID, roleDataScopeAll)
	userID := insertRoleTenantBoundaryUser(t, ctx, "impersonation-platform-user", datascope.PlatformTenantID)
	menuID := insertRoleTenantBoundaryMenu(t, ctx, "impersonation-platform-menu", "system:tenant:impersonate:test", datascope.PlatformTenantID)
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, []int{userID}, []int{menuID})
	})
	insertRoleTenantBoundaryRoleMenu(t, ctx, roleID, menuID, datascope.PlatformTenantID)
	insertRoleTenantBoundaryUserRole(t, ctx, userID, roleID, datascope.PlatformTenantID)
	svc.bizCtxSvc = roleScopeStaticBizCtx{ctx: &model.Context{
		UserId:         userID,
		TenantId:       62041,
		ActingAsTenant: true,
		ActingUserId:   userID,
	}}

	access, err := svc.GetUserAccessContext(ctx, userID)
	if err != nil {
		t.Fatalf("load impersonation access: %v", err)
	}
	if access == nil || !containsRoleTenantBoundaryString(access.Permissions, "system:tenant:impersonate:test") {
		t.Fatalf("expected platform role permission during impersonation, got %#v", access)
	}
	if access.DataScope != datascope.ScopeAll {
		t.Fatalf("expected platform role data scope during impersonation, got %d", access.DataScope)
	}
	if datascope.CurrentTenantID(ctx) != 62041 {
		t.Fatalf("expected request tenant to remain target tenant, got %d", datascope.CurrentTenantID(ctx))
	}
}

// roleTenantBoundaryEnablementReader marks multi-tenancy enabled in role tests.
type roleTenantBoundaryEnablementReader struct{}

// IsEnabled reports multi-tenant as enabled.
func (roleTenantBoundaryEnablementReader) IsEnabled(_ context.Context, pluginID string) bool {
	return pluginID == pkgtenantcap.ProviderPluginID
}

// roleTenantBoundaryProvider simulates the plugin-owned membership provider.
type roleTenantBoundaryProvider struct{}

// ResolveTenant returns platform for provider interface completeness.
func (roleTenantBoundaryProvider) ResolveTenant(context.Context, *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant accepts all users for provider interface completeness.
func (roleTenantBoundaryProvider) ValidateUserInTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ListUserTenants returns no tenant rows for provider interface completeness.
func (roleTenantBoundaryProvider) ListUserTenants(context.Context, int) ([]pkgtenantcap.TenantInfo, error) {
	return nil, nil
}

// SwitchTenant accepts all switches for provider interface completeness.
func (roleTenantBoundaryProvider) SwitchTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ApplyUserTenantScope is unused by role tests.
func (roleTenantBoundaryProvider) ApplyUserTenantScope(
	_ context.Context,
	model *gdb.Model,
	_ string,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ApplyUserTenantFilter is unused by role tests.
func (roleTenantBoundaryProvider) ApplyUserTenantFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ pkgtenantcap.TenantID,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenantProjections is unused by role tests.
func (roleTenantBoundaryProvider) ListUserTenantProjections(
	context.Context,
	[]int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	return map[int]*pkgtenantcap.UserTenantProjection{}, nil
}

// ResolveUserTenantAssignment is unused by role tests.
func (roleTenantBoundaryProvider) ResolveUserTenantAssignment(
	context.Context,
	[]pkgtenantcap.TenantID,
	pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	return &pkgtenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is unused by role tests.
func (roleTenantBoundaryProvider) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*pkgtenantcap.UserTenantAssignmentPlan,
) error {
	return nil
}

// EnsureUsersInTenant verifies role assignment targets are active tenant members.
func (roleTenantBoundaryProvider) EnsureUsersInTenant(
	ctx context.Context,
	userIDs []int,
	tenantID pkgtenantcap.TenantID,
) error {
	if len(userIDs) == 0 || tenantID <= pkgtenantcap.PLATFORM {
		return nil
	}
	count, err := dao.SysUser.DB().Model("plugin_multi_tenant_user_membership").Safe().Ctx(ctx).
		WhereIn("user_id", userIDs).
		Where("tenant_id", int(tenantID)).
		Where("status", 1).
		Count()
	if err != nil {
		return err
	}
	if count != len(userIDs) {
		return bizerr.NewCode(pkgtenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return nil
}

// ValidateStartupConsistency is unused by role tests.
func (roleTenantBoundaryProvider) ValidateStartupConsistency(context.Context) ([]string, error) {
	return nil, nil
}

// uniqueRoleTenantBoundaryName builds a stable unique test label.
func uniqueRoleTenantBoundaryName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// insertRoleTenantBoundaryRole inserts one role with explicit tenant ownership.
func insertRoleTenantBoundaryRole(t *testing.T, ctx context.Context, label string, tenantID int) int {
	t.Helper()
	return insertRoleTenantBoundaryRoleWithScope(t, ctx, label, tenantID, roleDataScopeSelf)
}

// insertRoleTenantBoundaryRoleWithScope inserts one role with explicit tenant
// ownership and data scope.
func insertRoleTenantBoundaryRoleWithScope(
	t *testing.T,
	ctx context.Context,
	label string,
	tenantID int,
	dataScope int,
) int {
	t.Helper()

	name := uniqueRoleTenantBoundaryName(label)
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      name,
		Key:       name,
		Sort:      99,
		DataScope: dataScope,
		Status:    roleTenantBoundaryStatusNormal,
		TenantId:  tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant boundary role: %v", err)
	}
	return int(id)
}

// insertRoleTenantBoundaryUser inserts one temporary user.
func insertRoleTenantBoundaryUser(t *testing.T, ctx context.Context, label string, tenantID int) int {
	t.Helper()

	username := uniqueRoleTenantBoundaryName(label)
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "test-password-hash",
		Nickname: username,
		Status:   roleTenantBoundaryStatusNormal,
		TenantId: tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant boundary user: %v", err)
	}
	return int(id)
}

// insertRoleTenantBoundaryMenu inserts one temporary global menu permission row.
func insertRoleTenantBoundaryMenu(t *testing.T, ctx context.Context, label string, permission string, _ int) int {
	t.Helper()

	key := uniqueRoleTenantBoundaryName(label)
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		MenuKey: key,
		Name:    key,
		Perms:   permission,
		Type:    "F",
		Sort:    99,
		Visible: roleTenantBoundaryStatusNormal,
		Status:  roleTenantBoundaryStatusNormal,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant boundary menu: %v", err)
	}
	return int(id)
}

// insertRoleTenantBoundaryRoleMenu inserts one role-menu relation.
func insertRoleTenantBoundaryRoleMenu(t *testing.T, ctx context.Context, roleID int, menuID int, tenantID int) {
	t.Helper()

	if _, err := dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{
		RoleId:   roleID,
		MenuId:   menuID,
		TenantId: tenantID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant boundary role-menu: %v", err)
	}
}

// insertRoleTenantBoundaryUserRole inserts one user-role relation.
func insertRoleTenantBoundaryUserRole(t *testing.T, ctx context.Context, userID int, roleID int, tenantID int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId:   userID,
		RoleId:   roleID,
		TenantId: tenantID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant boundary user-role: %v", err)
	}
}

// ensureRoleTenantBoundaryMembershipTable creates the minimal multi-tenant
// membership table needed by role assignment tests when the plugin schema is not installed.
func ensureRoleTenantBoundaryMembershipTable(t *testing.T, ctx context.Context) {
	t.Helper()
	_, err := dao.SysUser.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS plugin_multi_tenant_user_membership (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tenant_id BIGINT NOT NULL,
    status INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP NULL
)`)
	if err != nil {
		t.Fatalf("ensure role tenant boundary membership table failed: %v", err)
	}
}

// insertRoleTenantBoundaryMembership inserts one active or inactive membership row.
func insertRoleTenantBoundaryMembership(t *testing.T, ctx context.Context, userID int, tenantID int, status int) {
	t.Helper()
	if _, err := dao.SysUser.DB().Model("plugin_multi_tenant_user_membership").Safe().Ctx(ctx).Data(struct {
		UserID   int `orm:"user_id"`
		TenantID int `orm:"tenant_id"`
		Status   int `orm:"status"`
	}{
		UserID:   userID,
		TenantID: tenantID,
		Status:   status,
	}).Insert(); err != nil {
		t.Fatalf("insert role tenant boundary membership: %v", err)
	}
}

// cleanupRoleTenantBoundaryMembershipRows removes temporary membership rows.
func cleanupRoleTenantBoundaryMembershipRows(t *testing.T, ctx context.Context, userIDs []int) {
	t.Helper()
	if len(userIDs) == 0 {
		return
	}
	if _, err := dao.SysUser.DB().Model("plugin_multi_tenant_user_membership").Safe().Ctx(ctx).Unscoped().WhereIn("user_id", userIDs).Delete(); err != nil {
		t.Errorf("cleanup role tenant boundary membership rows failed: %v", err)
	}
}

// mustQueryRoleTenantBoundaryRole loads a role row for assertions.
func mustQueryRoleTenantBoundaryRole(t *testing.T, ctx context.Context, roleID int) *roleTenantBoundaryRoleProjection {
	t.Helper()

	var roleRow *roleTenantBoundaryRoleProjection
	if err := dao.SysRole.Ctx(ctx).Where(do.SysRole{Id: roleID}).Scan(&roleRow); err != nil {
		t.Fatalf("query tenant boundary role: %v", err)
	}
	if roleRow == nil {
		t.Fatalf("expected role %d to exist", roleID)
	}
	return roleRow
}

// mustCountRoleTenantBoundaryRoleMenu counts role-menu rows by tenant.
func mustCountRoleTenantBoundaryRoleMenu(t *testing.T, ctx context.Context, roleID int, tenantID int) int {
	t.Helper()

	count, err := dao.SysRoleMenu.Ctx(ctx).Where(do.SysRoleMenu{RoleId: roleID, TenantId: tenantID}).Count()
	if err != nil {
		t.Fatalf("count tenant boundary role-menu: %v", err)
	}
	return count
}

// mustCountRoleTenantBoundaryUserRole counts user-role rows by tenant.
func mustCountRoleTenantBoundaryUserRole(t *testing.T, ctx context.Context, roleID int, userID int, tenantID int) int {
	t.Helper()

	count, err := dao.SysUserRole.Ctx(ctx).
		Where(do.SysUserRole{RoleId: roleID, UserId: userID, TenantId: tenantID}).
		Count()
	if err != nil {
		t.Fatalf("count tenant boundary user-role: %v", err)
	}
	return count
}

// roleTenantBoundaryRoleProjection is a compact role ownership projection.
type roleTenantBoundaryRoleProjection struct {
	TenantId int `json:"tenantId" orm:"tenant_id"`
}

// containsRoleTenantBoundaryString reports whether a slice contains a value.
func containsRoleTenantBoundaryString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
