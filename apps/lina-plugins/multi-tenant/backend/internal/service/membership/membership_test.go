// This file verifies membership service database query behavior.

package membership

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	plugincontract "lina-core/pkg/pluginservice/contract"
	pkgtenantcap "lina-core/pkg/tenantcap"
	"lina-plugin-multi-tenant/backend/internal/model/do"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// membershipTestBizCtxService returns a fixed business context snapshot for
// membership tests that need tenant-scoped behavior.
type membershipTestBizCtxService struct {
	current plugincontract.CurrentContext
}

// Current returns the configured test business context snapshot.
func (s membershipTestBizCtxService) Current(context.Context) plugincontract.CurrentContext {
	return s.current
}

// membershipTestService creates a membership service with an explicit request
// context snapshot, avoiding host-internal context-key dependencies in plugin tests.
func membershipTestService(tenantID int, userID int) Service {
	return &serviceImpl{bizCtxSvc: membershipTestBizCtxService{current: plugincontract.CurrentContext{
		TenantID: tenantID,
		UserID:   userID,
	}}}
}

// TestListCountsWithoutProjectedColumns verifies the member list count query
// does not inherit list projection fields that PostgreSQL cannot count.
func TestListCountsWithoutProjectedColumns(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		tenantID = int64(424243)
		username = "membership_count_projection_test"
		password = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	var userID int64
	value, err := shared.Model(ctx, shared.TableSysUser).
		Where("username", username).
		Value("id")
	if err != nil {
		t.Fatalf("query test user failed: %v", err)
	}
	if value != nil && !value.IsNil() {
		userID = value.Int64()
	}
	if userID == 0 {
		userID, err = shared.Model(ctx, shared.TableSysUser).Data(do.SysUser{
			Username: username,
			Password: password,
			Nickname: "Membership Count Projection",
			Status:   1,
			TenantId: tenantID,
		}).InsertAndGetId()
		if err != nil {
			t.Fatalf("insert test user failed: %v", err)
		}
	}
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().Where("user_id", userID).Delete(); err != nil {
			t.Errorf("cleanup test membership failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().Where("id", userID).Delete(); err != nil {
			t.Errorf("cleanup test user failed: %v", err)
		}
	})

	if _, err = shared.Model(ctx, shared.TableMembership).Data(do.UserMembership{
		UserId:   userID,
		TenantId: tenantID,
		Status:   shared.MembershipStatusEnabled,
	}).InsertIgnore(); err != nil {
		t.Fatalf("insert test membership failed: %v", err)
	}

	out, err := New(membershipTestBizCtxService{}).List(ctx, ListInput{
		PageNum:  1,
		PageSize: 10,
		TenantID: tenantID,
		Status:   shared.MembershipStatusEnabled,
	})
	if err != nil {
		t.Fatalf("list tenant members failed: %v", err)
	}
	if out.Total < 1 {
		t.Fatalf("expected at least one member, got total=%d", out.Total)
	}
	if len(out.List) == 0 {
		t.Fatal("expected at least one member row")
	}
}

// TestListUsesCurrentTenantOverRequestedTenant verifies tenant-scoped requests
// cannot read another tenant by changing the query tenant id.
func TestListUsesCurrentTenantOverRequestedTenant(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		tenantAID = int64(424251)
		tenantBID = int64(424252)
		usernameA = "membership_scope_a_test"
		usernameB = "membership_scope_b_test"
		password  = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	userAID := insertMembershipTestUser(t, ctx, usernameA, password, tenantAID)
	userBID := insertMembershipTestUser(t, ctx, usernameB, password, tenantBID)
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().WhereIn("user_id", []int64{userAID, userBID}).Delete(); err != nil {
			t.Errorf("cleanup scoped memberships failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().WhereIn("id", []int64{userAID, userBID}).Delete(); err != nil {
			t.Errorf("cleanup scoped users failed: %v", err)
		}
	})

	insertMembershipTestRow(t, ctx, userAID, tenantAID)
	insertMembershipTestRow(t, ctx, userBID, tenantBID)

	out, err := membershipTestService(int(tenantAID), 0).List(ctx, ListInput{
		PageNum:  1,
		PageSize: 10,
		TenantID: tenantBID,
		Status:   shared.MembershipStatusEnabled,
	})
	if err != nil {
		t.Fatalf("list tenant members failed: %v", err)
	}
	seenA := false
	for _, item := range out.List {
		if item.UserID == userBID {
			t.Fatalf("tenant A context leaked tenant B user: %#v", item)
		}
		if item.UserID == userAID {
			seenA = true
		}
	}
	if !seenA {
		t.Fatalf("expected tenant A user in scoped result, got %#v", out.List)
	}
}

// TestAddRejectsUnavailableTenant verifies membership writes cannot create an
// active membership for missing, suspended, deleted, or platform tenant scopes.
func TestAddRejectsUnavailableTenant(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		suspendedCode = "membership-add-suspended-test"
		username      = "membership_add_unavailable_test"
		password      = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	suspendedTenantID := insertMembershipTestTenant(t, ctx, suspendedCode, shared.TenantStatusSuspended)
	userID := insertMembershipTestUser(t, ctx, username, password, suspendedTenantID)
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().Where("user_id", userID).Delete(); err != nil {
			t.Errorf("cleanup unavailable tenant memberships failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().Where("id", userID).Delete(); err != nil {
			t.Errorf("cleanup unavailable tenant user failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableTenant).Unscoped().Where("id", suspendedTenantID).Delete(); err != nil {
			t.Errorf("cleanup unavailable tenant failed: %v", err)
		}
	})

	_, err := New(membershipTestBizCtxService{}).Add(ctx, AddInput{TenantID: suspendedTenantID, UserID: userID})
	if !bizerr.Is(err, CodeTenantUnavailable) {
		t.Fatalf("expected unavailable tenant error, got %v", err)
	}

	count, err := shared.Model(ctx, shared.TableMembership).
		Where("user_id", userID).
		Where("tenant_id", suspendedTenantID).
		Count()
	if err != nil {
		t.Fatalf("count unavailable tenant membership rows failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected unavailable tenant membership not to be written, got %d", count)
	}
}

// TestUpdateRejectsOtherTenantMembership verifies tenant-scoped updates cannot
// modify a membership owned by another tenant.
func TestUpdateRejectsOtherTenantMembership(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		tenantAID = int64(424261)
		tenantBID = int64(424262)
		usernameB = "membership_update_scope_b_test"
		password  = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	userBID := insertMembershipTestUser(t, ctx, usernameB, password, tenantBID)
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().Where("user_id", userBID).Delete(); err != nil {
			t.Errorf("cleanup update scoped membership failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().Where("id", userBID).Delete(); err != nil {
			t.Errorf("cleanup update scoped user failed: %v", err)
		}
	})

	membershipBID := insertMembershipTestRow(t, ctx, userBID, tenantBID)
	statusDisabled := shared.MembershipStatusDisabled
	err := membershipTestService(int(tenantAID), 99001).Update(ctx, UpdateInput{Id: membershipBID, Status: &statusDisabled})
	if !bizerr.Is(err, CodeMembershipNotFound) {
		t.Fatalf("expected cross-tenant update to be hidden as not found, got %v", err)
	}

	var item *Entity
	if err = shared.Model(ctx, shared.TableMembership).Where("id", membershipBID).Scan(&item); err != nil {
		t.Fatalf("reload tenant B membership failed: %v", err)
	}
	if item == nil {
		t.Fatal("expected tenant B membership to remain present")
	}
	if item.Status != shared.MembershipStatusEnabled {
		t.Fatalf("expected tenant B membership status unchanged, got %d", item.Status)
	}
}

// TestRemoveRejectsOtherTenantMembership verifies tenant-scoped deletes cannot
// remove a membership owned by another tenant.
func TestRemoveRejectsOtherTenantMembership(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		tenantAID = int64(424271)
		tenantBID = int64(424272)
		usernameB = "membership_remove_scope_b_test"
		password  = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	userBID := insertMembershipTestUser(t, ctx, usernameB, password, tenantBID)
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().Where("user_id", userBID).Delete(); err != nil {
			t.Errorf("cleanup remove scoped membership failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().Where("id", userBID).Delete(); err != nil {
			t.Errorf("cleanup remove scoped user failed: %v", err)
		}
	})

	membershipBID := insertMembershipTestRow(t, ctx, userBID, tenantBID)
	err := membershipTestService(int(tenantAID), 99002).Remove(ctx, membershipBID)
	if !bizerr.Is(err, CodeMembershipNotFound) {
		t.Fatalf("expected cross-tenant remove to be hidden as not found, got %v", err)
	}

	count, err := shared.Model(ctx, shared.TableMembership).Where("id", membershipBID).Count()
	if err != nil {
		t.Fatalf("count tenant B membership failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected tenant B membership to remain present, got count=%d", count)
	}
}

// TestTenantAuthorizationOnlyAllowsActiveTenants verifies login and tenant
// switch authorization reject suspended tenants and hide them from candidates.
func TestTenantAuthorizationOnlyAllowsActiveTenants(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		activeCode    = "membership-active-test"
		suspendedCode = "membership-suspended-test"
		username      = "membership_lifecycle_test"
		password      = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	activeTenantID := insertMembershipTestTenant(t, ctx, activeCode, shared.TenantStatusActive)
	suspendedTenantID := insertMembershipTestTenant(t, ctx, suspendedCode, shared.TenantStatusSuspended)
	userID := insertMembershipTestUser(t, ctx, username, password, activeTenantID)
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().Where("user_id", userID).Delete(); err != nil {
			t.Errorf("cleanup lifecycle memberships failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().Where("id", userID).Delete(); err != nil {
			t.Errorf("cleanup lifecycle user failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableTenant).Unscoped().WhereIn("id", []int64{activeTenantID, suspendedTenantID}).Delete(); err != nil {
			t.Errorf("cleanup lifecycle tenants failed: %v", err)
		}
	})

	insertMembershipTestRow(t, ctx, userID, activeTenantID)
	insertMembershipTestRow(t, ctx, userID, suspendedTenantID)

	svc := New(membershipTestBizCtxService{})
	tenants, err := svc.ListUserTenants(ctx, userID)
	if err != nil {
		t.Fatalf("list user tenants failed: %v", err)
	}
	if len(tenants) != 1 || tenants[0].Id != activeTenantID {
		t.Fatalf("expected only active tenant candidate, got %#v", tenants)
	}
	if _, err = svc.GetByUserAndTenant(ctx, userID, suspendedTenantID); !bizerr.Is(err, CodeTenantUnavailable) {
		t.Fatalf("expected suspended tenant unavailable error, got %v", err)
	}
	if _, err = svc.GetByUserAndTenant(ctx, userID, activeTenantID); err != nil {
		t.Fatalf("expected active tenant membership to authorize, got %v", err)
	}
}

// TestGetByUserAndTenantHonorsRequestedTenant verifies provider membership
// checks validate the explicit target tenant even in a tenant-scoped request.
func TestGetByUserAndTenantHonorsRequestedTenant(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		username = "membership_requested_tenant_test"
		password = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	tenantAID := insertMembershipTestTenant(t, ctx, "membership-requested-tenant-a", shared.TenantStatusActive)
	tenantBID := insertMembershipTestTenant(t, ctx, "membership-requested-tenant-b", shared.TenantStatusActive)
	userID := insertMembershipTestUser(t, ctx, username, password, tenantAID)
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().Where("user_id", userID).Delete(); err != nil {
			t.Errorf("cleanup requested-tenant memberships failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().Where("id", userID).Delete(); err != nil {
			t.Errorf("cleanup requested-tenant user failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableTenant).Unscoped().WhereIn("id", []int64{tenantAID, tenantBID}).Delete(); err != nil {
			t.Errorf("cleanup requested-tenant tenants failed: %v", err)
		}
	})

	insertMembershipTestRow(t, ctx, userID, tenantAID)
	svc := membershipTestService(int(tenantAID), int(userID))
	if _, err := svc.GetByUserAndTenant(ctx, userID, tenantAID); err != nil {
		t.Fatalf("expected current tenant membership to authorize: %v", err)
	}
	if _, err := svc.GetByUserAndTenant(ctx, userID, tenantBID); !bizerr.Is(err, CodeMembershipNotFound) {
		t.Fatalf("expected requested tenant without membership to be rejected, got %v", err)
	}
}

// TestCurrentUsesContextIdentity verifies the current membership lookup is
// bound to BizCtx user and tenant instead of caller-supplied query fields.
func TestCurrentUsesContextIdentity(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		tenantBID  = int64(424282)
		tenantCode = "membership-current-ctx-tenant-a"
		usernameA  = "membership_current_scope_a_test"
		usernameB  = "membership_current_scope_b_test"
		password   = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	tenantAID := insertMembershipTestTenant(t, ctx, tenantCode, shared.TenantStatusActive)
	userAID := insertMembershipTestUser(t, ctx, usernameA, password, tenantAID)
	userBID := insertMembershipTestUser(t, ctx, usernameB, password, tenantBID)
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().WhereIn("user_id", []int64{userAID, userBID}).Delete(); err != nil {
			t.Errorf("cleanup current scoped memberships failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().WhereIn("id", []int64{userAID, userBID}).Delete(); err != nil {
			t.Errorf("cleanup current scoped users failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableTenant).Unscoped().Where("id", tenantAID).Delete(); err != nil {
			t.Errorf("cleanup current scoped tenant failed: %v", err)
		}
	})

	membershipAID := insertMembershipTestRow(t, ctx, userAID, tenantAID)
	insertMembershipTestRow(t, ctx, userBID, tenantBID)

	item, err := membershipTestService(int(tenantAID), int(userAID)).Current(ctx)
	if err != nil {
		t.Fatalf("current membership lookup failed: %v", err)
	}
	if item.Id != membershipAID || item.UserID != userAID || item.TenantID != tenantAID {
		t.Fatalf("expected context membership id=%d user=%d tenant=%d, got %#v", membershipAID, userAID, tenantAID, item)
	}
}

// TestReplaceUserTenantAssignmentsCanClearPlatformUserMembership verifies
// platform operators can remove all tenant memberships without tripping the
// platform-user membership guard.
func TestReplaceUserTenantAssignmentsCanClearPlatformUserMembership(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		tenantCode = "membership-clear-platform-test"
		username   = "membership_clear_platform_test"
		password   = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	tenantID := insertMembershipTestTenant(t, ctx, tenantCode, shared.TenantStatusActive)
	userID := insertMembershipTestUser(t, ctx, username, password, shared.PlatformTenantID)
	insertMembershipTestRow(t, ctx, userID, tenantID)
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().Where("user_id", userID).Delete(); err != nil {
			t.Errorf("cleanup clear platform membership failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().Where("id", userID).Delete(); err != nil {
			t.Errorf("cleanup clear platform user failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableTenant).Unscoped().Where("id", tenantID).Delete(); err != nil {
			t.Errorf("cleanup clear platform tenant failed: %v", err)
		}
	})

	err := New(membershipTestBizCtxService{}).ReplaceUserTenantAssignments(ctx, int(userID), &pkgtenantcap.UserTenantAssignmentPlan{
		ShouldReplace: true,
		PrimaryTenant: pkgtenantcap.PLATFORM,
	})
	if err != nil {
		t.Fatalf("clear platform user memberships failed: %v", err)
	}
	count, err := shared.Model(ctx, shared.TableMembership).Where("user_id", userID).Count()
	if err != nil {
		t.Fatalf("count cleared memberships failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected memberships cleared, got %d", count)
	}
}

// TestReplaceUserTenantAssignmentsDefaultMultiModeAllowsMultipleTenants
// verifies the code-owned default cardinality allows one user to join multiple
// tenants without requiring a host config-file value.
func TestReplaceUserTenantAssignmentsDefaultMultiModeAllowsMultipleTenants(t *testing.T) {
	ctx := context.Background()
	configureMembershipTestDB(t, ctx)

	const (
		tenantACode = "membership-multi-default-a"
		tenantBCode = "membership-multi-default-b"
		username    = "membership_multi_default_test"
		password    = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
	)

	tenantAID := insertMembershipTestTenant(t, ctx, tenantACode, shared.TenantStatusActive)
	tenantBID := insertMembershipTestTenant(t, ctx, tenantBCode, shared.TenantStatusActive)
	userID := insertMembershipTestUser(t, ctx, username, password, tenantAID)
	insertMembershipTestRow(t, ctx, userID, tenantAID)
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableMembership).Unscoped().Where("user_id", userID).Delete(); err != nil {
			t.Errorf("cleanup default multi memberships failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableSysUser).Unscoped().Where("id", userID).Delete(); err != nil {
			t.Errorf("cleanup default multi user failed: %v", err)
		}
		if _, err := shared.Model(ctx, shared.TableTenant).Unscoped().WhereIn("id", []int64{tenantAID, tenantBID}).Delete(); err != nil {
			t.Errorf("cleanup default multi tenants failed: %v", err)
		}
	})

	svc := New(membershipTestBizCtxService{})
	err := svc.ReplaceUserTenantAssignments(ctx, int(userID), &pkgtenantcap.UserTenantAssignmentPlan{
		TenantIDs: []pkgtenantcap.TenantID{
			pkgtenantcap.TenantID(tenantAID),
			pkgtenantcap.TenantID(tenantBID),
		},
		ShouldReplace: true,
		PrimaryTenant: pkgtenantcap.TenantID(tenantAID),
	})
	if err != nil {
		t.Fatalf("default multi-cardinality replacement failed: %v", err)
	}
	if count := countMembershipTestRows(t, ctx, userID, tenantAID); count != 1 {
		t.Fatalf("expected tenant A membership retained, got %d", count)
	}
	if count := countMembershipTestRows(t, ctx, userID, tenantBID); count != 1 {
		t.Fatalf("expected tenant B membership inserted, got %d", count)
	}
}

// configureMembershipTestDB points the package test at the local PostgreSQL
// database initialized by the repository test workflow.
func configureMembershipTestDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure membership test database failed: %v", err)
	}
	db := g.DB()
	ensureMembershipTestTables(t, ctx)
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close membership test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore membership test database config failed: %v", err)
		}
	})
}

// ensureMembershipTestTables creates the minimal plugin-owned tables required
// by membership service tests when the local database has not installed the plugin.
func ensureMembershipTestTables(t *testing.T, ctx context.Context) {
	t.Helper()
	statements := []string{
		`CREATE TABLE IF NOT EXISTS plugin_multi_tenant_tenant (
			id BIGSERIAL PRIMARY KEY,
			code VARCHAR(64) NOT NULL UNIQUE,
			name VARCHAR(128) NOT NULL,
			status VARCHAR(32) NOT NULL,
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
			t.Fatalf("ensure membership test table failed: %v", err)
		}
	}
}

// insertMembershipTestUser creates one sys_user test row and returns its id.
func insertMembershipTestUser(t *testing.T, ctx context.Context, username string, password string, tenantID int64) int64 {
	t.Helper()
	userID, err := shared.Model(ctx, shared.TableSysUser).Data(do.SysUser{
		Username: username,
		Password: password,
		Nickname: username,
		Status:   1,
		TenantId: tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test user %s failed: %v", username, err)
	}
	return userID
}

// insertMembershipTestRow creates one enabled membership test row.
func insertMembershipTestRow(t *testing.T, ctx context.Context, userID int64, tenantID int64) int64 {
	t.Helper()
	id, err := shared.Model(ctx, shared.TableMembership).Data(do.UserMembership{
		UserId:   userID,
		TenantId: tenantID,
		Status:   shared.MembershipStatusEnabled,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert membership user=%d tenant=%d failed: %v", userID, tenantID, err)
	}
	return id
}

// insertMembershipTestTenant creates one tenant lifecycle test row and returns its id.
func insertMembershipTestTenant(t *testing.T, ctx context.Context, code string, status shared.TenantStatus) int64 {
	t.Helper()
	tenantID, err := shared.Model(ctx, shared.TableTenant).Data(do.Tenant{
		Code:   code,
		Name:   code,
		Status: string(status),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test tenant %s failed: %v", code, err)
	}
	return tenantID
}

// countMembershipTestRows counts membership rows for one user and tenant.
func countMembershipTestRows(t *testing.T, ctx context.Context, userID int64, tenantID int64) int {
	t.Helper()

	count, err := shared.Model(ctx, shared.TableMembership).
		Where("user_id", userID).
		Where("tenant_id", tenantID).
		Where("status", shared.MembershipStatusEnabled).
		Count()
	if err != nil {
		t.Fatalf("count membership rows failed: %v", err)
	}
	return count
}
