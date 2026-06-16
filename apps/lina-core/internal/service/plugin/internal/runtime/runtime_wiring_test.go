// This file tests runtime service explicit constructor wiring.

package runtime

import (
	"context"
	"testing"
	"time"

	"lina-core/internal/model/entity"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/pluginhost"
)

// TestNewRuntimeWiresRequiredDependencies verifies constructor-owned runtime
// dependencies are available without post-construction setters.
func TestNewRuntimeWiresRequiredDependencies(t *testing.T) {
	service := newRuntimeWiringValidationService(t)
	if service.topology == nil ||
		service.menuMgr == nil ||
		service.hookDispatcher == nil ||
		service.jwtConfig == nil ||
		service.uploadSize == nil ||
		service.userCtx == nil ||
		service.sessionStore == nil ||
		service.roleAccess == nil ||
		service.menuFilter == nil ||
		service.cacheChangeNotifier == nil ||
		service.dependencyValidator == nil ||
		service.storageCleanupServices == nil {
		t.Fatal("expected constructor to wire all required runtime dependencies")
	}
}

// newRuntimeWiringValidationService builds a runtime service with every required
// production wiring seam configured.
func newRuntimeWiringValidationService(t *testing.T) *serviceImpl {
	t.Helper()

	catalogSvc := catalog.New(configsvc.New())
	topology := runtimeWiringTopology{}
	storeSvc := store.New(catalogSvc, topology)
	migrationSvc := migration.New(catalogSvc, storeSvc)
	return New(
		catalogSvc,
		storeSvc,
		migrationSvc,
		nil,
		nil,
		nil,
		locker.New(),
		topology,
		runtimeWiringMenuManager{},
		runtimeWiringResourceReferenceManager{},
		runtimeWiringHookDispatcher{},
		runtimeWiringJWTConfig{},
		runtimeWiringUploadSize{},
		runtimeWiringUserContext{},
		session.NewDBStore(),
		testRoleAccessProjector{},
		runtimeWiringPermissionFilter{},
		runtimeWiringCacheNotifier{},
		runtimeWiringDependencyValidator{},
		runtimeWiringStorageCleanupProvider{},
		nil,
	).(*serviceImpl)
}

// runtimeWiringTopology provides deterministic single-node topology metadata.
type runtimeWiringTopology struct{}

// IsClusterModeEnabled reports single-node mode for wiring validation.
func (runtimeWiringTopology) IsClusterModeEnabled() bool { return false }

// IsPrimaryNode reports this validation runtime is primary.
func (runtimeWiringTopology) IsPrimaryNode() bool { return true }

// CurrentNodeID returns a deterministic validation node ID.
func (runtimeWiringTopology) CurrentNodeID() string { return "runtime-wiring-test-node" }

// runtimeWiringMenuManager accepts menu synchronization calls.
type runtimeWiringMenuManager struct{}

// SyncPluginMenusAndPermissions records no behavior for wiring validation.
func (runtimeWiringMenuManager) SyncPluginMenusAndPermissions(context.Context, *catalog.Manifest) error {
	return nil
}

// SyncPluginMenus records no behavior for wiring validation.
func (runtimeWiringMenuManager) SyncPluginMenus(context.Context, *catalog.Manifest) error {
	return nil
}

// DeletePluginMenusByManifest records no behavior for wiring validation.
func (runtimeWiringMenuManager) DeletePluginMenusByManifest(context.Context, *catalog.Manifest) error {
	return nil
}

// runtimeWiringResourceReferenceManager accepts resource reference sync calls.
type runtimeWiringResourceReferenceManager struct{}

// SyncPluginResourceReferences records no behavior for wiring validation.
func (runtimeWiringResourceReferenceManager) SyncPluginResourceReferences(context.Context, *catalog.Manifest) error {
	return nil
}

// runtimeWiringHookDispatcher accepts lifecycle hook dispatches.
type runtimeWiringHookDispatcher struct{}

// DispatchPluginHookEvent records no behavior for wiring validation.
func (runtimeWiringHookDispatcher) DispatchPluginHookEvent(
	context.Context,
	pluginhost.ExtensionPoint,
	map[string]interface{},
) error {
	return nil
}

// runtimeWiringJWTConfig provides deterministic JWT settings.
type runtimeWiringJWTConfig struct{}

// GetJwtSecret returns a test signing secret.
func (runtimeWiringJWTConfig) GetJwtSecret(context.Context) string { return "runtime-wiring-secret" }

// GetSessionTimeout returns a deterministic session timeout.
func (runtimeWiringJWTConfig) GetSessionTimeout(context.Context) (time.Duration, error) {
	return time.Hour, nil
}

// runtimeWiringUploadSize provides deterministic upload size settings.
type runtimeWiringUploadSize struct{}

// GetUploadMaxSize returns a non-zero upload size limit.
func (runtimeWiringUploadSize) GetUploadMaxSize(context.Context) (int64, error) { return 16, nil }

// runtimeWiringUserContext accepts identity injection calls.
type runtimeWiringUserContext struct{}

// SetUser records no behavior for wiring validation.
func (runtimeWiringUserContext) SetUser(context.Context, string, int, string, int, string) {}

// SetTenant records no behavior for wiring validation.
func (runtimeWiringUserContext) SetTenant(context.Context, int) {}

// SetUserAccess records no behavior for wiring validation.
func (runtimeWiringUserContext) SetUserAccess(context.Context, int, bool, int) {}

// runtimeWiringPermissionFilter allows business entries and leaves menus unchanged.
type runtimeWiringPermissionFilter struct{}

// FilterPermissionMenus returns menus unchanged for wiring validation.
func (runtimeWiringPermissionFilter) FilterPermissionMenus(_ context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return menus
}

// CanExposeBusinessEntries reports entries visible for wiring validation.
func (runtimeWiringPermissionFilter) CanExposeBusinessEntries(context.Context, string) bool {
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

// runtimeWiringStorageCleanupProvider records no storage cleanup directory for
// wiring validation.
type runtimeWiringStorageCleanupProvider struct{}

// StorageCleanupServices returns no capability directory for wiring validation.
func (runtimeWiringStorageCleanupProvider) StorageCleanupServices() capability.Services {
	return nil
}
