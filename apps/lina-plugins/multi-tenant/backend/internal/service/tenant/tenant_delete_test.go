// This file verifies tenant deletion guard behavior.

package tenant

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/pluginhost"
	pluginbizctx "lina-core/pkg/pluginservice/bizctx"
	"lina-plugin-multi-tenant/backend/internal/service/resolverconfig"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
	"lina-plugin-multi-tenant/backend/internal/service/tenantplugin"
)

// tenantDeleteTestInsertData is the typed insert payload for tenant deletion tests.
type tenantDeleteTestInsertData struct {
	Code   string `orm:"code"`
	Name   string `orm:"name"`
	Status string `orm:"status"`
}

// tenantDeleteVetoGuard rejects tenant deletion in tests.
type tenantDeleteVetoGuard struct{}

// CanTenantDelete returns a deterministic guard veto.
func (tenantDeleteVetoGuard) CanTenantDelete(_ context.Context, _ int) (bool, string, error) {
	return false, "plugin.test.tenant.delete.vetoed", nil
}

// TestDeleteRunsLifecycleGuardBeforeSoftDelete verifies guard vetoes stop tenant deletion.
func TestDeleteRunsLifecycleGuardBeforeSoftDelete(t *testing.T) {
	ctx := context.Background()
	configureTenantDeleteTestDB(t, ctx)

	const guardPluginID = "tenant-delete-test-guard"
	pluginhost.RegisterLifecycleGuard(guardPluginID, tenantDeleteVetoGuard{})
	t.Cleanup(func() {
		pluginhost.UnregisterLifecycleGuard(guardPluginID)
	})

	tenantID, err := shared.Model(ctx, shared.TableTenant).Data(tenantDeleteTestInsertData{
		Code:   "tenant-delete-guard-test",
		Name:   "Tenant Delete Guard Test",
		Status: string(shared.TenantStatusActive),
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert tenant failed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := shared.Model(ctx, shared.TableTenant).Unscoped().Where("id", tenantID).Delete(); err != nil {
			t.Errorf("cleanup tenant failed: %v", err)
		}
	})

	err = New(pluginbizctx.New(nil), resolverconfig.New(), tenantplugin.New(pluginbizctx.New(nil))).Delete(ctx, tenantID)
	if !bizerr.Is(err, CodeTenantDeleteGuardVetoed) {
		t.Fatalf("expected guard veto error, got %v", err)
	}

	count, err := shared.Model(ctx, shared.TableTenant).Where("id", tenantID).Count()
	if err != nil {
		t.Fatalf("count tenant after veto failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected tenant to remain after guard veto, got count=%d", count)
	}
}

// configureTenantDeleteTestDB points the package test at the local PostgreSQL
// database initialized by the repository test workflow.
func configureTenantDeleteTestDB(t *testing.T, ctx context.Context) {
	t.Helper()

	originalConfig := gdb.GetAllConfig()
	if err := gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{{Link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"}},
	}); err != nil {
		t.Fatalf("configure tenant delete test database failed: %v", err)
	}
	db := g.DB()
	ensureTenantDeleteTestTables(t, ctx)
	t.Cleanup(func() {
		if err := db.Close(ctx); err != nil {
			t.Errorf("close tenant delete test database failed: %v", err)
		}
		if err := gdb.SetConfig(originalConfig); err != nil {
			t.Errorf("restore tenant delete test database config failed: %v", err)
		}
	})
}

// ensureTenantDeleteTestTables creates the minimal tenant table required by
// tenant deletion tests when the local database has not installed the plugin.
func ensureTenantDeleteTestTables(t *testing.T, ctx context.Context) {
	t.Helper()

	statement := `CREATE TABLE IF NOT EXISTS plugin_multi_tenant_tenant (
		"id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		"code" VARCHAR(64) NOT NULL,
		"name" VARCHAR(128) NOT NULL,
		"status" VARCHAR(32) NOT NULL DEFAULT 'active',
		"remark" VARCHAR(512) NOT NULL DEFAULT '',
		"created_by" BIGINT NOT NULL DEFAULT 0,
		"updated_by" BIGINT NOT NULL DEFAULT 0,
		"created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"deleted_at" TIMESTAMP,
		CONSTRAINT uk_plugin_multi_tenant_tenant_code UNIQUE ("code")
	)`
	if _, err := g.DB().Exec(ctx, statement); err != nil {
		t.Fatalf("ensure tenant delete test table failed: %v", err)
	}
}
