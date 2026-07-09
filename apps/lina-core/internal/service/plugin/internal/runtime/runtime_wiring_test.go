// This file tests runtime service explicit constructor wiring.

package runtime

import (
	"context"
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/store"
	rolesvc "lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/pluginhost"
)

// TestNewRuntimeWiresRequiredDependencies verifies constructor-owned runtime
// dependencies are available without post-construction setters.
func TestNewRuntimeWiresRequiredDependencies(t *testing.T) {
	service := newRuntimeWiringValidationService(t)
	if service.topology == nil ||
		service.integrationSvc == nil ||
		service.configSvc == nil ||
		service.userCtx == nil ||
		service.sessionStore == nil ||
		service.roleAccess == nil ||
		service.cacheChangeNotifier == nil ||
		service.dependencyValidator == nil {
		t.Fatal("expected constructor to wire all required runtime dependencies")
	}
}

// newRuntimeWiringValidationService builds a runtime service with every required
// production wiring seam configured.
func newRuntimeWiringValidationService(t *testing.T) *serviceImpl {
	t.Helper()

	var (
		catalogSvc   = catalog.New(configsvc.New())
		configSvc    = configsvc.New()
		topology     = runtimeWiringTopology{}
		storeSvc     = store.New(catalogSvc, topology)
		migrationSvc = migration.New(catalogSvc, storeSvc)
	)
	return New(
		catalogSvc,
		storeSvc,
		migrationSvc,
		nil,
		nil,
		locker.New(),
		topology,
		runtimeWiringIntegration{},
		configSvc,
		bizctx.New(),
		session.NewDBStore(),
		testRoleAccessProjector{},
		runtimeWiringCacheNotifier{},
		runtimeWiringDependencyValidator{},
		nil,
		nil,
	).(*serviceImpl)
}

// runtimeWiringTopology provides deterministic single-node topology metadata.
type runtimeWiringTopology struct{}

// Start records no behavior for wiring validation.
func (runtimeWiringTopology) Start(context.Context) {}

// Stop records no behavior for wiring validation.
func (runtimeWiringTopology) Stop(context.Context) {}

// IsEnabled reports single-node mode for wiring validation.
func (runtimeWiringTopology) IsEnabled() bool { return false }

// IsPrimary reports this validation runtime is primary.
func (runtimeWiringTopology) IsPrimary() bool { return true }

// NodeID returns a deterministic validation node ID.
func (runtimeWiringTopology) NodeID() string { return "runtime-wiring-test-node" }

// runtimeWiringIntegration accepts integration calls used by runtime wiring validation.
type runtimeWiringIntegration struct{}

// SyncPluginMenusAndPermissions records no behavior for wiring validation.
func (runtimeWiringIntegration) SyncPluginMenusAndPermissions(context.Context, *catalog.Manifest) error {
	return nil
}

// SyncPluginMenus records no behavior for wiring validation.
func (runtimeWiringIntegration) SyncPluginMenus(context.Context, *catalog.Manifest) error {
	return nil
}

// DeletePluginMenusByManifest records no behavior for wiring validation.
func (runtimeWiringIntegration) DeletePluginMenusByManifest(context.Context, *catalog.Manifest) error {
	return nil
}

// SyncPluginResourceReferences records no behavior for wiring validation.
func (runtimeWiringIntegration) SyncPluginResourceReferences(context.Context, *catalog.Manifest) error {
	return nil
}

// DispatchPluginHookEvent records no behavior for wiring validation.
func (runtimeWiringIntegration) DispatchPluginHookEvent(
	context.Context,
	pluginhost.ExtensionPoint,
	map[string]interface{},
) error {
	return nil
}

// FilterPermissionMenus returns menus unchanged for wiring validation.
func (runtimeWiringIntegration) FilterPermissionMenus(_ context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return menus
}

// CanExposeBusinessEntries reports entries visible for wiring validation.
func (runtimeWiringIntegration) CanExposeBusinessEntries(context.Context, string) bool {
	return true
}

// runtimeWiringCacheNotifier accepts cache change notifications.
type runtimeWiringCacheNotifier struct{}

// MarkRuntimeCacheChanged records no behavior for wiring validation.
func (runtimeWiringCacheNotifier) MarkRuntimeCacheChanged(context.Context, string) error {
	return nil
}

// PublishPluginChange records no behavior for wiring validation.
func (runtimeWiringCacheNotifier) PublishPluginChange(context.Context, string, string, string) error {
	return nil
}

// runtimeWiringDependencyValidator accepts dynamic plugin candidates.
type runtimeWiringDependencyValidator struct{}

// ValidateDynamicPluginCandidate records no behavior for wiring validation.
func (runtimeWiringDependencyValidator) ValidateDynamicPluginCandidate(context.Context, *catalog.Manifest) error {
	return nil
}

// testRoleAccessProjector returns deterministic access projections for tests
// that only need constructor wiring or route-auth behavior.
type testRoleAccessProjector struct {
	rolesvc.Service
	projection *rolesvc.DynamicRouteAccessProjection
	err        error
}

// BuildDynamicRouteAccessProjection returns the configured projection.
func (p testRoleAccessProjector) BuildDynamicRouteAccessProjection(
	context.Context,
	string,
	int,
	int,
) (*rolesvc.DynamicRouteAccessProjection, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.projection == nil {
		return &rolesvc.DynamicRouteAccessProjection{
			Permissions: []string{},
			RoleNames:   []string{},
			DataScope:   datascope.ScopeNone,
		}, nil
	}
	return &rolesvc.DynamicRouteAccessProjection{
		Permissions:          append([]string(nil), p.projection.Permissions...),
		RoleNames:            append([]string(nil), p.projection.RoleNames...),
		DataScope:            p.projection.DataScope,
		DataScopeUnsupported: p.projection.DataScopeUnsupported,
		UnsupportedDataScope: p.projection.UnsupportedDataScope,
		IsSuperAdmin:         p.projection.IsSuperAdmin,
	}, nil
}
