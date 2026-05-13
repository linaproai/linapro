// This file verifies tenant creation orchestration rollback behavior.

package tenant

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginservice/bizctx"
	"lina-plugin-multi-tenant/backend/internal/service/resolverconfig"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
	"lina-plugin-multi-tenant/backend/internal/service/tenantplugin"
)

// failingTenantPluginService simulates a tenant-plugin provisioning failure.
type failingTenantPluginService struct {
	err error
}

// List is unused by tenant creation tests.
func (s failingTenantPluginService) List(context.Context) (*tenantplugin.ListOutput, error) {
	return &tenantplugin.ListOutput{}, nil
}

// SetEnabled is unused by tenant creation tests.
func (s failingTenantPluginService) SetEnabled(context.Context, string, bool) error {
	return nil
}

// ProvisionForTenant returns the configured provisioning failure.
func (s failingTenantPluginService) ProvisionForTenant(context.Context, int64) error {
	return s.err
}

// TestCreateRollsBackTenantWhenProvisioningFails verifies explicit tenant
// domain orchestration does not leave a half-created tenant when provisioning
// fails after the tenant row has been inserted.
func TestCreateRollsBackTenantWhenProvisioningFails(t *testing.T) {
	ctx := context.Background()
	configureTenantDeleteTestDB(t, ctx)

	const tenantCode = "tenant-create-rollback"
	if _, err := shared.Model(ctx, shared.TableTenant).Unscoped().Where("code", tenantCode).Delete(); err != nil {
		t.Fatalf("cleanup stale rollback tenant failed: %v", err)
	}

	svc := &serviceImpl{
		bizCtxSvc:         bizctx.New(nil),
		resolverConfigSvc: resolverconfig.New(),
		tenantPluginSvc:   failingTenantPluginService{err: gerror.New("tenant plugin provisioning failed")},
	}
	if _, err := svc.Create(ctx, CreateInput{Code: tenantCode, Name: "Tenant Create Rollback"}); err == nil {
		t.Fatal("expected tenant create to fail when provisioning fails")
	}

	count, err := shared.Model(ctx, shared.TableTenant).Unscoped().Where("code", tenantCode).Count()
	if err != nil {
		t.Fatalf("count rollback tenant failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected tenant create transaction to roll back inserted row, got count=%d", count)
	}
}
