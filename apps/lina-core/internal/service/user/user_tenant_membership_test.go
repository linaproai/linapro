// This file verifies tenant membership boundaries in host user management.

package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

const (
	// userTenantMembershipTestMembershipTable is the plugin-owned membership table used by tests.
	userTenantMembershipTestMembershipTable = "plugin_multi_tenant_user_membership"
	// userTenantMembershipTestTenantTable is the plugin-owned tenant table used by tests.
	userTenantMembershipTestTenantTable = "plugin_multi_tenant_tenant"
	// userTenantMembershipTestActive marks an active membership row.
	userTenantMembershipTestActive = 1
)

// userTenantMembershipTestInsertData is a typed insert payload for membership rows.
type userTenantMembershipTestInsertData struct {
	UserID   int64       `orm:"user_id"`
	TenantID int64       `orm:"tenant_id"`
	Status   int         `orm:"status"`
	JoinedAt *gtime.Time `orm:"joined_at"`
}

// userTenantMembershipTestRow is a compact membership join projection.
type userTenantMembershipTestRow struct {
	UserID     int    `orm:"user_id"`
	TenantID   int    `orm:"tenant_id"`
	TenantName string `orm:"tenant_name"`
}

// TestUserListUsesMembershipVisibility verifies tenant users are visible by
// active membership instead of only sys_user.tenant_id.
func TestUserListUsesMembershipVisibility(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 61001)
	ensureUserTenantMembershipTestTables(t, ctx)
	currentUserID := insertUserTenantMembershipTestUser(t, ctx, "membership-list-current", 61001)
	primaryOtherUserID := insertUserTenantMembershipTestUser(t, ctx, "membership-list-primary-other", 61002)
	invisibleUserID := insertUserTenantMembershipTestUser(t, ctx, "membership-list-invisible", 61001)
	roleID := insertUserTenantMembershipTestRole(t, ctx, "membership-list-role", 61001)
	t.Cleanup(func() {
		cleanupUserTenantMembershipRows(t, ctx, []int{currentUserID, primaryOtherUserID, invisibleUserID})
		cleanupUserDeleteTestRows(t, ctx, []int{currentUserID, primaryOtherUserID, invisibleUserID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	insertUserTenantMembershipTestUserRole(t, ctx, currentUserID, roleID, 61001)
	insertUserTenantMembershipTestMembership(t, ctx, currentUserID, 61001, userTenantMembershipTestActive)
	insertUserTenantMembershipTestMembership(t, ctx, primaryOtherUserID, 61001, userTenantMembershipTestActive)
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: currentUserID, TenantId: 61001}})

	out, err := svc.List(ctx, ListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list users by tenant membership: %v", err)
	}
	actual := userDataScopeIDSet(out.List)
	if _, ok := actual[currentUserID]; !ok {
		t.Fatalf("expected current user %d in tenant result, got %v", currentUserID, userDataScopeListIDs(out.List))
	}
	if _, ok := actual[primaryOtherUserID]; !ok {
		t.Fatalf("expected membership user %d in tenant result, got %v", primaryOtherUserID, userDataScopeListIDs(out.List))
	}
	if _, ok := actual[invisibleUserID]; ok {
		t.Fatalf("did not expect non-member user %d in tenant result, got %v", invisibleUserID, userDataScopeListIDs(out.List))
	}
}

// TestUserCreateWritesTenantAndMembership verifies tenant-scoped creation
// stores the primary tenant and creates an active membership row.
func TestUserCreateWritesTenantAndMembership(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 61011)
	ensureUserTenantMembershipTestTables(t, ctx)
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	username := fmt.Sprintf("membership-create-%d", time.Now().UnixNano())

	userID, err := svc.Create(ctx, CreateInput{
		Username: username,
		Password: "admin123",
		Status:   StatusNormal,
	})
	if err != nil {
		t.Fatalf("create tenant user: %v", err)
	}
	t.Cleanup(func() {
		cleanupUserTenantMembershipRows(t, ctx, []int{userID})
		cleanupUserDeleteTestRows(t, ctx, []int{userID})
	})

	if tenantID := mustQueryUserTenantID(t, ctx, userID); tenantID != 61011 {
		t.Fatalf("expected sys_user tenant_id=61011, got %d", tenantID)
	}
	if count := mustCountUserTenantMembership(t, ctx, userID, 61011); count != 1 {
		t.Fatalf("expected one active tenant membership, got %d", count)
	}
}

// TestPlatformUserCreateWritesSelectedTenantMemberships verifies platform
// creation can assign one or more tenant memberships from the drawer field.
func TestPlatformUserCreateWritesSelectedTenantMemberships(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-create-platform-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-create-platform-b")
	operatorID := insertUserTenantMembershipTestUser(t, ctx, "membership-create-platform-operator", 0)
	username := fmt.Sprintf("membership-create-platform-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, []int{operatorID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: operatorID, TenantId: 0, DataScope: 1}})
	userID, err := svc.Create(ctx, CreateInput{
		Username:  username,
		Password:  "admin123",
		Status:    StatusNormal,
		TenantIds: []int{tenantAID, tenantBID, tenantAID},
	})
	if err != nil {
		t.Fatalf("create platform assigned user: %v", err)
	}
	t.Cleanup(func() {
		cleanupUserTenantMembershipRows(t, ctx, []int{userID})
		cleanupUserDeleteTestRows(t, ctx, []int{userID})
	})

	if tenantID := mustQueryUserTenantID(t, ctx, userID); tenantID != tenantAID {
		t.Fatalf("expected primary tenant id %d, got %d", tenantAID, tenantID)
	}
	if count := mustCountUserTenantMembership(t, ctx, userID, tenantAID); count != 1 {
		t.Fatalf("expected one active tenant A membership, got %d", count)
	}
	if count := mustCountUserTenantMembership(t, ctx, userID, tenantBID); count != 1 {
		t.Fatalf("expected one active tenant B membership, got %d", count)
	}
}

// TestPlatformUserUpdateReplacesTenantMemberships verifies edit drawer tenant
// changes replace membership rows.
func TestPlatformUserUpdateReplacesTenantMemberships(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-update-platform-a")
	tenantBID, tenantBName := insertUserTenantMembershipTestTenant(t, ctx, "membership-update-platform-b")
	operatorID := insertUserTenantMembershipTestUser(t, ctx, "membership-update-platform-operator", 0)
	userID := insertUserTenantMembershipTestUser(t, ctx, "membership-update-platform-user", tenantAID)
	roleID := insertUserDataScopeTestRole(t, ctx, "membership-update-platform-role", userDataScopeAll, 1)
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, ctx, []int{operatorID, userID})
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, []int{operatorID, userID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserDeleteTestUserRole(t, ctx, operatorID, roleID)
	insertUserTenantMembershipTestMembership(t, ctx, userID, tenantAID, userTenantMembershipTestActive)
	insertUserTenantMembershipTestMembership(t, ctx, userID, tenantBID, userTenantMembershipTestActive)

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: operatorID, TenantId: 0, DataScope: 1}})
	if err := svc.Update(ctx, UpdateInput{Id: userID, TenantIds: []int{tenantBID}}); err != nil {
		t.Fatalf("update platform user memberships: %v", err)
	}

	if tenantID := mustQueryUserTenantID(t, ctx, userID); tenantID != tenantBID {
		t.Fatalf("expected primary tenant id %d, got %d", tenantBID, tenantID)
	}
	if count := mustCountUserTenantMembership(t, ctx, userID, tenantAID); count != 0 {
		t.Fatalf("expected tenant A membership removed, got %d", count)
	}
	if count := mustCountUserTenantMembership(t, ctx, userID, tenantBID); count != 1 {
		t.Fatalf("expected tenant B membership retained, got %d", count)
	}
	tenantIds, tenantNames, err := svc.GetUserTenantMemberships(ctx, userID)
	if err != nil {
		t.Fatalf("get user tenant memberships: %v", err)
	}
	if len(tenantIds) != 1 || tenantIds[0] != tenantBID {
		t.Fatalf("expected tenant ids [%d], got %v", tenantBID, tenantIds)
	}
	if len(tenantNames) != 1 || tenantNames[0] != tenantBName {
		t.Fatalf("expected tenant names [%q], got %v", tenantBName, tenantNames)
	}
}

// TestTenantBoundOperatorCreateAllowsOwnedTenantAssignments verifies platform
// users with tenant memberships can assign only their own tenant set.
func TestTenantBoundOperatorCreateAllowsOwnedTenantAssignments(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-create-owned-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-create-owned-b")
	operatorID := insertUserTenantMembershipTestUser(t, ctx, "membership-create-owned-operator", 0)
	username := fmt.Sprintf("membership-create-owned-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, ctx, []int{operatorID})
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, []int{operatorID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserTenantMembershipTestMembership(t, ctx, operatorID, tenantAID, userTenantMembershipTestActive)
	insertUserTenantMembershipTestMembership(t, ctx, operatorID, tenantBID, userTenantMembershipTestActive)

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: operatorID, TenantId: 0, DataScope: 2}})
	userID, err := svc.Create(ctx, CreateInput{
		Username:  username,
		Password:  "admin123",
		Status:    StatusNormal,
		TenantIds: []int{tenantAID, tenantBID},
	})
	if err != nil {
		t.Fatalf("create tenant-bound operator assigned user: %v", err)
	}
	t.Cleanup(func() {
		cleanupUserTenantMembershipRows(t, ctx, []int{userID})
		cleanupUserDeleteTestRows(t, ctx, []int{userID})
	})

	if count := mustCountUserTenantMembership(t, ctx, userID, tenantAID); count != 1 {
		t.Fatalf("expected one tenant A membership, got %d", count)
	}
	if count := mustCountUserTenantMembership(t, ctx, userID, tenantBID); count != 1 {
		t.Fatalf("expected one tenant B membership, got %d", count)
	}
}

// TestTenantBoundOperatorCreateRejectsForeignTenantAssignments verifies direct
// requests cannot assign users to tenants outside the operator membership set.
func TestTenantBoundOperatorCreateRejectsForeignTenantAssignments(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-create-foreign-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-create-foreign-b")
	operatorID := insertUserTenantMembershipTestUser(t, ctx, "membership-create-foreign-operator", 0)
	username := fmt.Sprintf("membership-create-foreign-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, ctx, []int{operatorID})
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, []int{operatorID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserTenantMembershipTestMembership(t, ctx, operatorID, tenantAID, userTenantMembershipTestActive)

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: operatorID, TenantId: 0, DataScope: 2}})
	_, err := svc.Create(ctx, CreateInput{
		Username:  username,
		Password:  "admin123",
		Status:    StatusNormal,
		TenantIds: []int{tenantBID},
	})
	if err == nil {
		t.Fatal("expected foreign tenant assignment to fail")
	}
	if !bizerr.Is(err, pkgtenantcap.CodeTenantForbidden) {
		t.Fatalf("expected tenant forbidden error, got %v", err)
	}
	if count := mustCountUserTenantMembership(t, ctx, operatorID, tenantBID); count != 0 {
		t.Fatalf("expected no foreign membership side effect, got %d", count)
	}
}

// TestTenantBoundOperatorCreateRejectsEmptyTenantAssignments verifies a
// tenant-bound platform-context operator cannot create platform-scope users.
func TestTenantBoundOperatorCreateRejectsEmptyTenantAssignments(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-create-empty-a")
	operatorID := insertUserTenantMembershipTestUser(t, ctx, "membership-create-empty-operator", 0)
	username := fmt.Sprintf("membership-create-empty-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, ctx, []int{operatorID})
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID})
		cleanupUserDeleteTestRows(t, ctx, []int{operatorID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserTenantMembershipTestMembership(t, ctx, operatorID, tenantAID, userTenantMembershipTestActive)

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: operatorID, TenantId: 0, DataScope: 2}})
	_, err := svc.Create(ctx, CreateInput{
		Username: username,
		Password: "admin123",
		Status:   StatusNormal,
	})
	if err == nil {
		t.Fatal("expected platform-scope user creation to fail for tenant-bound operator")
	}
	if !bizerr.Is(err, pkgtenantcap.CodeCrossTenantNotAllowed) {
		t.Fatalf("expected cross-tenant error, got %v", err)
	}
}

// TestTenantBoundOperatorUpdateRejectsForeignTenantAssignments verifies edit
// requests cannot replace memberships with tenants outside the operator set.
func TestTenantBoundOperatorUpdateRejectsForeignTenantAssignments(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-update-foreign-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-update-foreign-b")
	operatorID := insertUserTenantMembershipTestUser(t, ctx, "membership-update-foreign-operator", 0)
	userID := insertUserTenantMembershipTestUser(t, ctx, "membership-update-foreign-user", tenantAID)
	roleID := insertUserDataScopeTestRole(t, ctx, "membership-update-foreign-role", userDataScopeTenant, 1)
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, ctx, []int{operatorID, userID})
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, []int{operatorID, userID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserDeleteTestUserRole(t, ctx, operatorID, roleID)
	insertUserTenantMembershipTestMembership(t, ctx, operatorID, tenantAID, userTenantMembershipTestActive)
	insertUserTenantMembershipTestMembership(t, ctx, userID, tenantAID, userTenantMembershipTestActive)

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: operatorID, TenantId: 0, DataScope: 2}})
	err := svc.Update(ctx, UpdateInput{Id: userID, TenantIds: []int{tenantBID}})
	if err == nil {
		t.Fatal("expected foreign tenant update to fail")
	}
	if !bizerr.Is(err, pkgtenantcap.CodeTenantForbidden) {
		t.Fatalf("expected tenant forbidden error, got %v", err)
	}
	if count := mustCountUserTenantMembership(t, ctx, userID, tenantAID); count != 1 {
		t.Fatalf("expected original tenant A membership to remain, got %d", count)
	}
	if count := mustCountUserTenantMembership(t, ctx, userID, tenantBID); count != 0 {
		t.Fatalf("expected no tenant B membership, got %d", count)
	}
}

// TestListUserTenantMembershipsAggregatesTenantNames verifies the user-list
// tenant projection reports all active memberships for each visible user.
func TestListUserTenantMembershipsAggregatesTenantNames(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	userID := insertUserTenantMembershipTestUser(t, ctx, "membership-tenant-names", 0)
	tenantAID, tenantAName := insertUserTenantMembershipTestTenant(t, ctx, "membership-tenant-a")
	tenantBID, tenantBName := insertUserTenantMembershipTestTenant(t, ctx, "membership-tenant-b")
	t.Cleanup(func() {
		cleanupUserTenantMembershipRows(t, ctx, []int{userID})
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, []int{userID})
	})
	insertUserTenantMembershipTestMembership(t, ctx, userID, tenantAID, userTenantMembershipTestActive)
	insertUserTenantMembershipTestMembership(t, ctx, userID, tenantBID, userTenantMembershipTestActive)

	actual, err := (&userTenantMembershipTestProvider{}).ListUserTenantProjections(ctx, []int{userID})
	if err != nil {
		t.Fatalf("list tenant memberships: %v", err)
	}
	item := actual[userID]
	if item == nil {
		t.Fatalf("expected tenant projection for user %d, got %v", userID, actual)
	}
	if len(item.TenantIDs) != 2 || len(item.TenantNames) != 2 {
		t.Fatalf("expected two tenants, got ids=%v names=%v", item.TenantIDs, item.TenantNames)
	}
	if item.TenantNames[0] != tenantAName || item.TenantNames[1] != tenantBName {
		t.Fatalf("expected tenant names %q,%q got %v", tenantAName, tenantBName, item.TenantNames)
	}
}

// TestListUserTenantMembershipsRespectsTenantContext verifies tenant-scoped
// user lists do not reveal memberships from other tenants.
func TestListUserTenantMembershipsRespectsTenantContext(t *testing.T) {
	baseCtx := context.Background()
	ensureUserTenantMembershipTestTables(t, baseCtx)
	userID := insertUserTenantMembershipTestUser(t, baseCtx, "membership-tenant-filter", 0)
	tenantAID, tenantAName := insertUserTenantMembershipTestTenant(t, baseCtx, "membership-filter-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, baseCtx, "membership-filter-b")
	t.Cleanup(func() {
		cleanupUserTenantMembershipRows(t, baseCtx, []int{userID})
		cleanupUserTenantMembershipTestTenants(t, baseCtx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, baseCtx, []int{userID})
	})
	insertUserTenantMembershipTestMembership(t, baseCtx, userID, tenantAID, userTenantMembershipTestActive)
	insertUserTenantMembershipTestMembership(t, baseCtx, userID, tenantBID, userTenantMembershipTestActive)

	actual, err := (&userTenantMembershipTestProvider{}).ListUserTenantProjections(
		datascope.WithTenantForTest(baseCtx, tenantAID),
		[]int{userID},
	)
	if err != nil {
		t.Fatalf("list tenant-scoped memberships: %v", err)
	}
	item := actual[userID]
	if item == nil {
		t.Fatalf("expected tenant-scoped projection for user %d, got %v", userID, actual)
	}
	if len(item.TenantIDs) != 1 || item.TenantIDs[0] != pkgtenantcap.TenantID(tenantAID) {
		t.Fatalf("expected only current tenant id %d, got %v", tenantAID, item.TenantIDs)
	}
	if len(item.TenantNames) != 1 || item.TenantNames[0] != tenantAName {
		t.Fatalf("expected only current tenant name %q, got %v", tenantAName, item.TenantNames)
	}
}

// TestUserListTenantFilterUsesMembershipForPlatformContext verifies platform
// operators can filter the user list by active tenant membership.
func TestUserListTenantFilterUsesMembershipForPlatformContext(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-list-filter-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-list-filter-b")
	userAID := insertUserTenantMembershipTestUser(t, ctx, "membership-list-filter-a-user", 0)
	userBID := insertUserTenantMembershipTestUser(t, ctx, "membership-list-filter-b-user", 0)
	roleID := insertUserDataScopeTestRole(t, ctx, "membership-list-filter-platform-role", userDataScopeAll, 1)
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, ctx, []int{userAID, userBID})
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, []int{userAID, userBID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserDeleteTestUserRole(t, ctx, userAID, roleID)
	insertUserTenantMembershipTestMembership(t, ctx, userAID, tenantAID, userTenantMembershipTestActive)
	insertUserTenantMembershipTestMembership(t, ctx, userBID, tenantBID, userTenantMembershipTestActive)

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: userAID, TenantId: 0, DataScope: 1}})
	out, err := svc.List(ctx, ListInput{PageNum: 1, PageSize: 20, TenantId: &tenantAID})
	if err != nil {
		t.Fatalf("list users by tenant filter: %v", err)
	}

	actual := userDataScopeIDSet(out.List)
	if _, ok := actual[userAID]; !ok {
		t.Fatalf("expected tenant A user %d in platform filtered result, got %v", userAID, userDataScopeListIDs(out.List))
	}
	if _, ok := actual[userBID]; ok {
		t.Fatalf("did not expect tenant B user %d in platform filtered result, got %v", userBID, userDataScopeListIDs(out.List))
	}
}

// TestTenantBoundOperatorListRejectsForeignTenantFilter verifies manual query
// parameters cannot reveal users from tenants outside the operator membership set.
func TestTenantBoundOperatorListRejectsForeignTenantFilter(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-list-foreign-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-list-foreign-b")
	operatorID := insertUserTenantMembershipTestUser(t, ctx, "membership-list-foreign-operator", 0)
	targetID := insertUserTenantMembershipTestUser(t, ctx, "membership-list-foreign-target", 0)
	roleID := insertUserDataScopeTestRole(t, ctx, "membership-list-foreign-role", userDataScopeTenant, 1)
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, ctx, []int{operatorID, targetID})
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, []int{operatorID, targetID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserDeleteTestUserRole(t, ctx, operatorID, roleID)
	insertUserTenantMembershipTestMembership(t, ctx, operatorID, tenantAID, userTenantMembershipTestActive)
	insertUserTenantMembershipTestMembership(t, ctx, targetID, tenantBID, userTenantMembershipTestActive)

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: operatorID, TenantId: 0, DataScope: 2}})
	_, err := svc.List(ctx, ListInput{PageNum: 1, PageSize: 20, TenantId: &tenantBID})
	if err == nil {
		t.Fatal("expected foreign tenant list filter to fail")
	}
	if !bizerr.Is(err, pkgtenantcap.CodeTenantForbidden) {
		t.Fatalf("expected tenant forbidden error, got %v", err)
	}
}

// TestTenantBoundAllScopeOperatorListRejectsForeignTenantFilter verifies
// tenant membership still limits tenant filters for all-data-scope operators.
func TestTenantBoundAllScopeOperatorListRejectsForeignTenantFilter(t *testing.T) {
	ctx := context.Background()
	ensureUserTenantMembershipTestTables(t, ctx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-list-all-scope-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, ctx, "membership-list-all-scope-b")
	operatorID := insertUserTenantMembershipTestUser(t, ctx, "membership-list-all-scope-operator", 0)
	roleID := insertUserDataScopeTestRole(t, ctx, "membership-list-all-scope-role", userDataScopeAll, 1)
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, ctx, []int{operatorID})
		cleanupUserTenantMembershipTestTenants(t, ctx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, ctx, []int{operatorID})
		cleanupUserDeleteTestRoles(t, ctx, []int{roleID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserDeleteTestUserRole(t, ctx, operatorID, roleID)
	insertUserTenantMembershipTestMembership(t, ctx, operatorID, tenantAID, userTenantMembershipTestActive)

	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: operatorID, TenantId: 0, DataScope: 1}})
	_, err := svc.List(ctx, ListInput{PageNum: 1, PageSize: 20, TenantId: &tenantBID})
	if err == nil {
		t.Fatal("expected foreign tenant list filter to fail for tenant-bound all-scope operator")
	}
	if !bizerr.Is(err, pkgtenantcap.CodeTenantForbidden) {
		t.Fatalf("expected tenant forbidden error, got %v", err)
	}
}

// TestUserListTenantContextIgnoresCrossTenantFilter verifies tenant contexts
// keep the current tenant as the visibility authority.
func TestUserListTenantContextIgnoresCrossTenantFilter(t *testing.T) {
	baseCtx := context.Background()
	ensureUserTenantMembershipTestTables(t, baseCtx)
	tenantAID, _ := insertUserTenantMembershipTestTenant(t, baseCtx, "membership-list-context-a")
	tenantBID, _ := insertUserTenantMembershipTestTenant(t, baseCtx, "membership-list-context-b")
	userAID := insertUserTenantMembershipTestUser(t, baseCtx, "membership-list-context-a-user", 0)
	userBID := insertUserTenantMembershipTestUser(t, baseCtx, "membership-list-context-b-user", 0)
	roleID := insertUserTenantMembershipTestRole(t, baseCtx, "membership-list-context-role", tenantAID)
	t.Cleanup(func() {
		pkgtenantcap.RegisterProvider(nil)
		cleanupUserTenantMembershipRows(t, baseCtx, []int{userAID, userBID})
		cleanupUserTenantMembershipTestTenants(t, baseCtx, []int{tenantAID, tenantBID})
		cleanupUserDeleteTestRows(t, baseCtx, []int{userAID, userBID})
		cleanupUserDeleteTestRoles(t, baseCtx, []int{roleID})
	})
	pkgtenantcap.RegisterProvider(&userTenantMembershipTestProvider{})
	insertUserTenantMembershipTestUserRole(t, baseCtx, userAID, roleID, tenantAID)
	insertUserTenantMembershipTestMembership(t, baseCtx, userAID, tenantAID, userTenantMembershipTestActive)
	insertUserTenantMembershipTestMembership(t, baseCtx, userBID, tenantBID, userTenantMembershipTestActive)

	ctx := datascope.WithTenantForTest(baseCtx, tenantAID)
	svc := newUserTestService(userTenantMembershipEnablementReader{}).(*serviceImpl)
	setUserTestBizCtx(svc, userDeleteStaticBizCtx{ctx: &model.Context{UserId: userAID, TenantId: tenantAID}})
	out, err := svc.List(ctx, ListInput{PageNum: 1, PageSize: 20, TenantId: &tenantBID})
	if err != nil {
		t.Fatalf("list users by tenant context with cross-tenant filter: %v", err)
	}

	actual := userDataScopeIDSet(out.List)
	if _, ok := actual[userAID]; !ok {
		t.Fatalf("expected current tenant user %d in result, got %v", userAID, userDataScopeListIDs(out.List))
	}
	if _, ok := actual[userBID]; ok {
		t.Fatalf("did not expect cross-tenant user %d in result, got %v", userBID, userDataScopeListIDs(out.List))
	}
}

// ensureUserTenantMembershipTestTables creates the plugin-owned tables used by
// the host user service tests when the multi-tenant plugin is not installed.
func ensureUserTenantMembershipTestTables(t *testing.T, ctx context.Context) {
	t.Helper()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS plugin_multi_tenant_tenant (
			id BIGSERIAL PRIMARY KEY,
			code VARCHAR(64) NOT NULL UNIQUE,
			name VARCHAR(128) NOT NULL,
			status VARCHAR(32) NOT NULL,
			plan VARCHAR(64),
			deleted_at TIMESTAMP NULL
		)`,
		`CREATE TABLE IF NOT EXISTS plugin_multi_tenant_user_membership (
				id BIGSERIAL PRIMARY KEY,
				user_id BIGINT NOT NULL,
				tenant_id BIGINT NOT NULL,
				status INT NOT NULL DEFAULT 1,
				joined_at TIMESTAMP NULL,
				created_by BIGINT,
				updated_by BIGINT,
				deleted_at TIMESTAMP NULL
		)`,
	}
	for _, statement := range statements {
		if _, err := g.DB().Exec(ctx, statement); err != nil {
			t.Fatalf("ensure tenant membership test tables: %v", err)
		}
	}
}

// userTenantMembershipEnablementReader marks the multi-tenant plugin enabled in tests.
type userTenantMembershipEnablementReader struct{}

// IsEnabled reports multi-tenant as enabled.
func (userTenantMembershipEnablementReader) IsEnabled(_ context.Context, pluginID string) bool {
	return pluginID == pkgtenantcap.ProviderPluginID
}

// userTenantMembershipTestProvider satisfies the tenantcap provider contract for tests.
type userTenantMembershipTestProvider struct{}

// ResolveTenant is unused by these tests.
func (*userTenantMembershipTestProvider) ResolveTenant(context.Context, *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	return nil, nil
}

// ValidateUserInTenant is unused by these tests.
func (*userTenantMembershipTestProvider) ValidateUserInTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ListUserTenants returns the operator's active tenant memberships.
func (*userTenantMembershipTestProvider) ListUserTenants(
	ctx context.Context,
	userID int,
) ([]pkgtenantcap.TenantInfo, error) {
	result := make([]pkgtenantcap.TenantInfo, 0)
	var rows []*userTenantMembershipTestRow
	err := g.DB().Model(userTenantMembershipTestMembershipTable).Safe().Ctx(ctx).
		As("m").
		InnerJoin(userTenantMembershipTestTenantTable+" t", "t.id = m.tenant_id AND t.deleted_at IS NULL").
		Fields("m.user_id, m.tenant_id, t.name AS tenant_name").
		Where("m.user_id", userID).
		Where("m.status", userTenantMembershipTestActive).
		OrderAsc("m.id").
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		result = append(result, pkgtenantcap.TenantInfo{
			ID:   pkgtenantcap.TenantID(row.TenantID),
			Name: row.TenantName,
		})
	}
	return result, nil
}

// SwitchTenant is unused by these tests.
func (*userTenantMembershipTestProvider) SwitchTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ApplyUserTenantScope constrains user rows by active current-tenant membership.
func (*userTenantMembershipTestProvider) ApplyUserTenantScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	tenantID := datascope.CurrentTenantID(ctx)
	if model == nil || tenantID == datascope.PlatformTenantID {
		return model, false, nil
	}
	return model.WhereIn(userIDColumn, activeUserTenantMembershipTestUserModel(ctx, tenantID)), false, nil
}

// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
func (*userTenantMembershipTestProvider) ApplyUserTenantFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	tenantID pkgtenantcap.TenantID,
) (*gdb.Model, bool, error) {
	if model == nil || tenantID <= pkgtenantcap.PLATFORM || datascope.CurrentTenantID(ctx) != datascope.PlatformTenantID {
		return model, false, nil
	}
	return model.WhereIn(userIDColumn, activeUserTenantMembershipTestUserModel(ctx, int(tenantID))), false, nil
}

// ListUserTenantProjections returns tenant ownership labels for visible users.
func (*userTenantMembershipTestProvider) ListUserTenantProjections(
	ctx context.Context,
	userIDs []int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	result := make(map[int]*pkgtenantcap.UserTenantProjection)
	if len(userIDs) == 0 {
		return result, nil
	}
	var rows []*userTenantMembershipTestRow
	model := g.DB().Model(userTenantMembershipTestMembershipTable).Safe().Ctx(ctx).
		As("m").
		InnerJoin(userTenantMembershipTestTenantTable+" t", "t.id = m.tenant_id AND t.deleted_at IS NULL").
		Fields("m.user_id, m.tenant_id, t.name AS tenant_name").
		WhereIn("m.user_id", userIDs).
		Where("m.status", userTenantMembershipTestActive).
		OrderAsc("m.id")
	if tenantID := datascope.CurrentTenantID(ctx); tenantID != datascope.PlatformTenantID {
		model = model.Where("m.tenant_id", tenantID)
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		item := result[row.UserID]
		if item == nil {
			item = &pkgtenantcap.UserTenantProjection{}
			result[row.UserID] = item
		}
		item.TenantIDs = append(item.TenantIDs, pkgtenantcap.TenantID(row.TenantID))
		item.TenantNames = append(item.TenantNames, row.TenantName)
	}
	return result, nil
}

// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
func (*userTenantMembershipTestProvider) ResolveUserTenantAssignment(
	ctx context.Context,
	requested []pkgtenantcap.TenantID,
	mode pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	tenantID := datascope.CurrentTenantID(ctx)
	normalized := normalizeUserTenantMembershipTestTenantIDs(requested)
	if tenantID != datascope.PlatformTenantID {
		if mode == pkgtenantcap.UserTenantAssignmentUpdate {
			return &pkgtenantcap.UserTenantAssignmentPlan{}, nil
		}
		return &pkgtenantcap.UserTenantAssignmentPlan{
			TenantIDs:     []pkgtenantcap.TenantID{pkgtenantcap.TenantID(tenantID)},
			ShouldReplace: true,
			PrimaryTenant: pkgtenantcap.TenantID(tenantID),
		}, nil
	}
	return &pkgtenantcap.UserTenantAssignmentPlan{
		TenantIDs:     normalized,
		ShouldReplace: mode == pkgtenantcap.UserTenantAssignmentUpdate || len(normalized) > 0,
		PrimaryTenant: firstUserTenantMembershipTestTenantIDOrPlatform(normalized),
	}, nil
}

// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
func (*userTenantMembershipTestProvider) ReplaceUserTenantAssignments(
	ctx context.Context,
	userID int,
	plan *pkgtenantcap.UserTenantAssignmentPlan,
) error {
	if plan == nil {
		return nil
	}
	normalized := normalizeUserTenantMembershipTestTenantIDs(plan.TenantIDs)
	if _, err := g.DB().Model(userTenantMembershipTestMembershipTable).Safe().Ctx(ctx).Unscoped().Where("user_id", userID).Delete(); err != nil {
		return err
	}
	for _, tenantID := range normalized {
		if _, err := g.DB().Model(userTenantMembershipTestMembershipTable).Safe().Ctx(ctx).Data(userTenantMembershipTestInsertData{
			UserID:   int64(userID),
			TenantID: int64(tenantID),
			Status:   userTenantMembershipTestActive,
			JoinedAt: gtime.Now(),
		}).Insert(); err != nil {
			return err
		}
	}
	return nil
}

// EnsureUsersInTenant verifies every user has active membership in the tenant.
func (*userTenantMembershipTestProvider) EnsureUsersInTenant(
	ctx context.Context,
	userIDs []int,
	tenantID pkgtenantcap.TenantID,
) error {
	if len(userIDs) == 0 || tenantID == pkgtenantcap.PLATFORM {
		return nil
	}
	count, err := activeUserTenantMembershipTestUserModel(ctx, int(tenantID)).WhereIn("user_id", userIDs).Count()
	if err != nil {
		return err
	}
	if count != len(userIDs) {
		return bizerr.NewCode(pkgtenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return nil
}

// ValidateStartupConsistency is unused by user tests.
func (*userTenantMembershipTestProvider) ValidateStartupConsistency(context.Context) ([]string, error) {
	return nil, nil
}

// activeUserTenantMembershipTestUserModel returns active membership users.
func activeUserTenantMembershipTestUserModel(ctx context.Context, tenantID int) *gdb.Model {
	return g.DB().Model(userTenantMembershipTestMembershipTable).Safe().Ctx(ctx).
		Fields("user_id").
		Where("tenant_id", tenantID).
		Where("status", userTenantMembershipTestActive)
}

// normalizeUserTenantMembershipTestTenantIDs returns positive unique tenant IDs.
func normalizeUserTenantMembershipTestTenantIDs(tenantIDs []pkgtenantcap.TenantID) []pkgtenantcap.TenantID {
	normalized := make([]pkgtenantcap.TenantID, 0, len(tenantIDs))
	seen := make(map[pkgtenantcap.TenantID]struct{}, len(tenantIDs))
	for _, tenantID := range tenantIDs {
		if tenantID <= pkgtenantcap.PLATFORM {
			continue
		}
		if _, ok := seen[tenantID]; ok {
			continue
		}
		seen[tenantID] = struct{}{}
		normalized = append(normalized, tenantID)
	}
	return normalized
}

// firstUserTenantMembershipTestTenantIDOrPlatform returns the host primary tenant.
func firstUserTenantMembershipTestTenantIDOrPlatform(
	tenantIDs []pkgtenantcap.TenantID,
) pkgtenantcap.TenantID {
	if len(tenantIDs) == 0 {
		return pkgtenantcap.PLATFORM
	}
	return tenantIDs[0]
}

// insertUserTenantMembershipTestUser inserts a temporary user with primary tenant.
func insertUserTenantMembershipTestUser(t *testing.T, ctx context.Context, label string, tenantID int) int {
	t.Helper()

	username := fmt.Sprintf("%s-%d", label, time.Now().UnixNano())
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "test-password-hash",
		Nickname: username,
		Status:   1,
		TenantId: tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant membership test user: %v", err)
	}
	return int(id)
}

// insertUserTenantMembershipTestRole inserts one tenant-scoped all-data role.
func insertUserTenantMembershipTestRole(t *testing.T, ctx context.Context, label string, tenantID int) int {
	t.Helper()

	suffix := time.Now().UnixNano()
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      fmt.Sprintf("%s-%d", label, suffix),
		Key:       fmt.Sprintf("%s-%d", label, suffix),
		Sort:      99,
		DataScope: int(userDataScopeTenant),
		Status:    1,
		TenantId:  tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant membership test role: %v", err)
	}
	return int(id)
}

// insertUserTenantMembershipTestUserRole inserts one tenant-scoped role binding.
func insertUserTenantMembershipTestUserRole(t *testing.T, ctx context.Context, userID int, roleID int, tenantID int) {
	t.Helper()

	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
		UserId:   userID,
		RoleId:   roleID,
		TenantId: tenantID,
	}).Insert(); err != nil {
		t.Fatalf("insert tenant membership test user-role relation: %v", err)
	}
}

// insertUserTenantMembershipTestMembership inserts one temporary membership row.
func insertUserTenantMembershipTestMembership(t *testing.T, ctx context.Context, userID int, tenantID int, status int) {
	t.Helper()

	if _, err := g.DB().Model(userTenantMembershipTestMembershipTable).Safe().Ctx(ctx).Data(userTenantMembershipTestInsertData{
		UserID:   int64(userID),
		TenantID: int64(tenantID),
		Status:   status,
		JoinedAt: gtime.Now(),
	}).Insert(); err != nil {
		t.Fatalf("insert tenant membership test row: %v", err)
	}
}

// cleanupUserTenantMembershipRows removes temporary membership rows by user ID.
func cleanupUserTenantMembershipRows(t *testing.T, ctx context.Context, userIDs []int) {
	t.Helper()

	if _, err := g.DB().Model(userTenantMembershipTestMembershipTable).Safe().Ctx(ctx).Unscoped().WhereIn("user_id", userIDs).Delete(); err != nil {
		t.Fatalf("cleanup tenant membership rows: %v", err)
	}
}

// mustQueryUserTenantID returns one user's primary tenant ID.
func mustQueryUserTenantID(t *testing.T, ctx context.Context, userID int) int {
	t.Helper()

	var row *modelUserTenantProjection
	if err := dao.SysUser.Ctx(ctx).Fields(dao.SysUser.Columns().TenantId).Where(do.SysUser{Id: userID}).Scan(&row); err != nil {
		t.Fatalf("query user tenant id: %v", err)
	}
	if row == nil {
		t.Fatalf("expected user %d to exist", userID)
	}
	return row.TenantId
}

// mustCountUserTenantMembership counts active membership rows for one user and tenant.
func mustCountUserTenantMembership(t *testing.T, ctx context.Context, userID int, tenantID int) int {
	t.Helper()

	count, err := g.DB().Model(userTenantMembershipTestMembershipTable).Safe().Ctx(ctx).
		Where("user_id", userID).
		Where("tenant_id", tenantID).
		Where("status", userTenantMembershipTestActive).
		Count()
	if err != nil {
		t.Fatalf("count tenant membership rows: %v", err)
	}
	return count
}

// modelUserTenantProjection is a tiny user tenant test projection.
type modelUserTenantProjection struct {
	TenantId int `json:"tenantId" orm:"tenant_id"`
}

// tenantMembershipTenantInsertData is a typed insert payload for tenant tests.
type tenantMembershipTenantInsertData struct {
	Code   string `orm:"code"`
	Name   string `orm:"name"`
	Status string `orm:"status"`
}

// insertUserTenantMembershipTestTenant inserts a temporary tenant row.
func insertUserTenantMembershipTestTenant(t *testing.T, ctx context.Context, label string) (int, string) {
	t.Helper()

	suffix := time.Now().UnixNano()
	name := fmt.Sprintf("%s-name-%d", label, suffix)
	id, err := g.DB().Model(userTenantMembershipTestTenantTable).Safe().Ctx(ctx).Data(tenantMembershipTenantInsertData{
		Code:   fmt.Sprintf("%s-%d", label, suffix),
		Name:   name,
		Status: "active",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant membership test tenant: %v", err)
	}
	return int(id), name
}

// cleanupUserTenantMembershipTestTenants removes temporary tenant rows by ID.
func cleanupUserTenantMembershipTestTenants(t *testing.T, ctx context.Context, tenantIDs []int) {
	t.Helper()

	if _, err := g.DB().Model(userTenantMembershipTestTenantTable).Safe().Ctx(ctx).Unscoped().WhereIn("id", tenantIDs).Delete(); err != nil {
		t.Fatalf("cleanup tenant membership test tenants: %v", err)
	}
}
