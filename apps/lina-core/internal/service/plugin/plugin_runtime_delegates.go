// This file owns root-composition delegates that break construction-time cycles
// while keeping internal plugin services configured through constructor
// parameters.

package plugin

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/goai"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	"lina-core/pkg/plugin/pluginhost"
)

// RuntimeDelegate is a composition-root cycle breaker for host services that
// must be constructed before the root plugin service can be assembled. It binds
// the startup-owned root Service while only delegating host-facing plugin
// contracts needed before regular runtime wiring becomes reachable.
type RuntimeDelegate struct {
	mu      sync.RWMutex
	service Service
}

// NewRuntimeDelegate creates an unbound startup delegate for root composition.
func NewRuntimeDelegate() *RuntimeDelegate {
	return &RuntimeDelegate{}
}

// BindService connects the completed root plugin service to this startup
// delegate. The caller must bind during process startup before serving traffic.
func (d *RuntimeDelegate) BindService(service Service) error {
	if d == nil {
		return gerror.New("plugin runtime delegate cannot bind through nil receiver")
	}
	if service == nil {
		return gerror.New("plugin runtime delegate cannot bind nil service")
	}
	d.mu.Lock()
	d.service = service
	d.mu.Unlock()
	return nil
}

// Bound reports whether the delegate has been connected to the root plugin
// service. It is intended for startup and test diagnostics only.
func (d *RuntimeDelegate) Bound() bool {
	return d.serviceSnapshot() != nil
}

// DispatchHookEvent dispatches a plugin hook event after binding.
func (d *RuntimeDelegate) DispatchHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	service := d.serviceSnapshot()
	if service == nil {
		return pluginRuntimeDelegateUnboundError()
	}
	return service.DispatchHookEvent(ctx, event, values)
}

// FilterPermissionMenus filters role permission menus after binding.
func (d *RuntimeDelegate) FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	service := d.serviceSnapshot()
	if service == nil {
		return menus
	}
	return service.FilterPermissionMenus(ctx, menus)
}

// FilterMenus filters navigation menus after binding.
func (d *RuntimeDelegate) FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	service := d.serviceSnapshot()
	if service == nil {
		return menus
	}
	return service.FilterMenus(ctx, menus)
}

// ListSourceRouteBindings returns source-plugin route bindings after binding.
func (d *RuntimeDelegate) ListSourceRouteBindings() []pluginhost.SourceRouteBinding {
	service := d.serviceSnapshot()
	if service == nil {
		return nil
	}
	return service.ListSourceRouteBindings()
}

// ProjectDynamicRoutesToOpenAPI projects dynamic routes after binding.
func (d *RuntimeDelegate) ProjectDynamicRoutesToOpenAPI(ctx context.Context, paths goai.Paths) error {
	service := d.serviceSnapshot()
	if service == nil {
		return pluginRuntimeDelegateUnboundError()
	}
	return service.ProjectDynamicRoutesToOpenAPI(ctx, paths)
}

// IsEnabled reports plugin visibility after binding.
func (d *RuntimeDelegate) IsEnabled(ctx context.Context, pluginID string) bool {
	service := d.serviceSnapshot()
	return service != nil && service.IsEnabled(ctx, pluginID)
}

// IsEnabledAuthoritative reports authoritative plugin visibility after binding.
func (d *RuntimeDelegate) IsEnabledAuthoritative(ctx context.Context, pluginID string) bool {
	service := d.serviceSnapshot()
	return service != nil && service.IsEnabledAuthoritative(ctx, pluginID)
}

// IsProviderEnabled reports provider-level plugin availability after binding.
func (d *RuntimeDelegate) IsProviderEnabled(ctx context.Context, pluginID string) bool {
	service := d.serviceSnapshot()
	return service != nil && service.IsProviderEnabled(ctx, pluginID)
}

// OrgProviderEnv returns organization-provider construction inputs after binding.
func (d *RuntimeDelegate) OrgProviderEnv(ctx context.Context, pluginID string) orgspi.ProviderEnv {
	service := d.serviceSnapshot()
	if service == nil {
		return orgspi.ProviderEnv{PluginID: pluginID}
	}
	return service.OrgProviderEnv(ctx, pluginID)
}

// TenantProviderEnv returns tenant-provider construction inputs after binding.
func (d *RuntimeDelegate) TenantProviderEnv(ctx context.Context, pluginID string) tenantspi.ProviderEnv {
	service := d.serviceSnapshot()
	if service == nil {
		return tenantspi.ProviderEnv{PluginID: pluginID}
	}
	return service.TenantProviderEnv(ctx, pluginID)
}

// EnsureTenantPluginDisableAllowed delegates tenant plugin disable guards after binding.
func (d *RuntimeDelegate) EnsureTenantPluginDisableAllowed(ctx context.Context, pluginID string, tenantID int) error {
	service := d.serviceSnapshot()
	if service == nil {
		return pluginRuntimeDelegateUnboundError()
	}
	return service.EnsureTenantPluginDisableAllowed(ctx, pluginID, tenantID)
}

// NotifyTenantPluginDisabled delegates tenant plugin disable notifications after binding.
func (d *RuntimeDelegate) NotifyTenantPluginDisabled(ctx context.Context, pluginID string, tenantID int) {
	service := d.serviceSnapshot()
	if service == nil {
		return
	}
	service.NotifyTenantPluginDisabled(ctx, pluginID, tenantID)
}

// EnsureTenantDeleteAllowed delegates tenant delete guards after binding.
func (d *RuntimeDelegate) EnsureTenantDeleteAllowed(ctx context.Context, tenantID int) error {
	service := d.serviceSnapshot()
	if service == nil {
		return pluginRuntimeDelegateUnboundError()
	}
	return service.EnsureTenantDeleteAllowed(ctx, tenantID)
}

// pluginRuntimeDelegateUnboundError reports missing root plugin service binding.
func pluginRuntimeDelegateUnboundError() error {
	return gerror.New("plugin runtime delegate is not bound")
}

// NotifyTenantDeleted delegates tenant deletion notifications after binding.
func (d *RuntimeDelegate) NotifyTenantDeleted(ctx context.Context, tenantID int) {
	service := d.serviceSnapshot()
	if service == nil {
		return
	}
	service.NotifyTenantDeleted(ctx, tenantID)
}

func (d *RuntimeDelegate) serviceSnapshot() Service {
	if d == nil {
		return nil
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.service
}

// integrationDelegateProvider delegates runtime integration side effects to the
// integration service after root composition completes.
type integrationDelegateProvider struct {
	mu      sync.RWMutex
	service runtime.IntegrationService
}

// BindService connects the integration service owned by the current root
// composition.
func (p *integrationDelegateProvider) BindService(service runtime.IntegrationService) {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.service = service
	p.mu.Unlock()
}

// SyncPluginMenusAndPermissions synchronizes all plugin menus and permissions.
func (p *integrationDelegateProvider) SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error {
	service := p.serviceSnapshot()
	if service == nil {
		return pluginIntegrationDelegateUnboundError()
	}
	return service.SyncPluginMenusAndPermissions(ctx, manifest)
}

// SyncPluginMenus synchronizes manifest-declared menus only.
func (p *integrationDelegateProvider) SyncPluginMenus(ctx context.Context, manifest *catalog.Manifest) error {
	service := p.serviceSnapshot()
	if service == nil {
		return pluginIntegrationDelegateUnboundError()
	}
	return service.SyncPluginMenus(ctx, manifest)
}

// DeletePluginMenusByManifest removes plugin-owned menu rows.
func (p *integrationDelegateProvider) DeletePluginMenusByManifest(ctx context.Context, manifest *catalog.Manifest) error {
	service := p.serviceSnapshot()
	if service == nil {
		return pluginIntegrationDelegateUnboundError()
	}
	return service.DeletePluginMenusByManifest(ctx, manifest)
}

// SyncPluginResourceReferences synchronizes plugin resource reference rows.
func (p *integrationDelegateProvider) SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error {
	service := p.serviceSnapshot()
	if service == nil {
		return pluginIntegrationDelegateUnboundError()
	}
	return service.SyncPluginResourceReferences(ctx, manifest)
}

// DispatchPluginHookEvent dispatches one plugin hook event.
func (p *integrationDelegateProvider) DispatchPluginHookEvent(
	ctx context.Context,
	event pluginhost.ExtensionPoint,
	values map[string]interface{},
) error {
	service := p.serviceSnapshot()
	if service == nil {
		return pluginIntegrationDelegateUnboundError()
	}
	return service.DispatchPluginHookEvent(ctx, event, values)
}

// pluginIntegrationDelegateUnboundError reports missing integration service binding.
func pluginIntegrationDelegateUnboundError() error {
	return gerror.New("plugin integration delegate is not bound")
}

// FilterPermissionMenus filters permission menus through plugin enablement.
func (p *integrationDelegateProvider) FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	service := p.serviceSnapshot()
	if service == nil {
		return menus
	}
	return service.FilterPermissionMenus(ctx, menus)
}

// CanExposeBusinessEntries reports whether plugin business entries can be
// exposed.
func (p *integrationDelegateProvider) CanExposeBusinessEntries(ctx context.Context, pluginID string) bool {
	service := p.serviceSnapshot()
	return service == nil || service.CanExposeBusinessEntries(ctx, pluginID)
}

func (p *integrationDelegateProvider) serviceSnapshot() runtime.IntegrationService {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.service
}

// runtimeCacheChangeNotifierProvider delegates runtime cache publication to the
// root plugin service after the service struct has been assembled.
type runtimeCacheChangeNotifierProvider struct {
	mu      sync.RWMutex
	service runtime.CacheChangeNotifier
}

// BindService connects the root cache notifier.
func (p *runtimeCacheChangeNotifierProvider) BindService(service runtime.CacheChangeNotifier) {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.service = service
	p.mu.Unlock()
}

// MarkRuntimeCacheChanged publishes one runtime cache mutation.
func (p *runtimeCacheChangeNotifierProvider) MarkRuntimeCacheChanged(ctx context.Context, reason string) error {
	service := p.serviceSnapshot()
	if service == nil {
		return gerror.New("plugin runtime cache notifier is not bound")
	}
	return service.MarkRuntimeCacheChanged(ctx, reason)
}

// PublishPluginChange publishes one plugin-scoped cache mutation.
func (p *runtimeCacheChangeNotifierProvider) PublishPluginChange(
	ctx context.Context,
	pluginID string,
	pluginType string,
	reason string,
) error {
	service := p.serviceSnapshot()
	if service == nil {
		return gerror.New("plugin runtime cache notifier is not bound")
	}
	return service.PublishPluginChange(ctx, pluginID, pluginType, reason)
}

func (p *runtimeCacheChangeNotifierProvider) serviceSnapshot() runtime.CacheChangeNotifier {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.service
}

// dependencyValidatorService defines the dependency validation calls exposed by
// the root plugin dependency resolver after the service struct has been assembled.
type dependencyValidatorService interface {
	runtime.DependencyValidator
	ValidateSourcePluginUpgradeCandidate(ctx context.Context, manifest *catalog.Manifest) error
}

// dependencyValidatorProvider delegates candidate validation to the root plugin
// dependency resolver after the service struct has been assembled.
type dependencyValidatorProvider struct {
	mu      sync.RWMutex
	service dependencyValidatorService
}

// BindService connects the root dependency validator.
func (p *dependencyValidatorProvider) BindService(service dependencyValidatorService) {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.service = service
	p.mu.Unlock()
}

// ValidateDynamicPluginCandidate validates one dynamic-plugin candidate.
func (p *dependencyValidatorProvider) ValidateDynamicPluginCandidate(ctx context.Context, manifest *catalog.Manifest) error {
	service := p.serviceSnapshot()
	if service == nil {
		return gerror.New("plugin dependency validator is not bound")
	}
	return service.ValidateDynamicPluginCandidate(ctx, manifest)
}

// ValidateSourcePluginUpgradeCandidate validates one source-plugin upgrade
// candidate before upgrade side effects.
func (p *dependencyValidatorProvider) ValidateSourcePluginUpgradeCandidate(ctx context.Context, manifest *catalog.Manifest) error {
	service := p.serviceSnapshot()
	if service == nil {
		return gerror.New("plugin dependency validator is not bound")
	}
	return service.ValidateSourcePluginUpgradeCandidate(ctx, manifest)
}

func (p *dependencyValidatorProvider) serviceSnapshot() dependencyValidatorService {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.service
}

// wasmRuntimeProvider delegates WASM execution and cache invalidation to the
// startup-configured runtime instance owned by the plugin service.
type wasmRuntimeProvider struct {
	mu      sync.RWMutex
	runtime wasm.Runtime
}

// BindRuntime connects the concrete WASM runtime created after host services
// are available.
func (p *wasmRuntimeProvider) BindRuntime(runtime wasm.Runtime) {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.runtime = runtime
	p.mu.Unlock()
}

// ExecuteBridge delegates one dynamic-plugin bridge call to the configured WASM runtime.
func (p *wasmRuntimeProvider) ExecuteBridge(
	ctx context.Context,
	input wasm.ExecutionInput,
	requestContent []byte,
) (*bridgecontract.BridgeResponseEnvelopeV1, error) {
	runtime := p.runtimeSnapshot()
	if runtime == nil {
		return nil, gerror.New("dynamic wasm runtime is not configured")
	}
	return runtime.ExecuteBridge(ctx, input, requestContent)
}

// InvalidateCache removes one compiled module from the configured WASM runtime cache.
func (p *wasmRuntimeProvider) InvalidateCache(ctx context.Context, artifactPath string) {
	runtime := p.runtimeSnapshot()
	if runtime == nil {
		return
	}
	runtime.InvalidateCache(ctx, artifactPath)
}

// InvalidateAllCache removes all compiled modules from the configured WASM runtime cache.
func (p *wasmRuntimeProvider) InvalidateAllCache(ctx context.Context) {
	runtime := p.runtimeSnapshot()
	if runtime == nil {
		return
	}
	runtime.InvalidateAllCache(ctx)
}

func (p *wasmRuntimeProvider) runtimeSnapshot() wasm.Runtime {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.runtime
}
