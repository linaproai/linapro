// This file implements the guest-side capability directory delegation. The root
// pluginbridge.go file owns the public component contract; this file keeps
// client selection details out of the main file.

package pluginbridge

import (
	"context"

	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/cachecap"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/dictcap"
	"lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/routecap"
	"lina-core/pkg/plugin/capability/sessioncap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
	"lina-core/pkg/plugin/pluginbridge/internal/domainhostcall"
	"lina-core/pkg/plugin/pluginbridge/recordstore"
)

// directory implements the guest-side capability directory.
type directory struct{}

// defaultDirectory is the process-default guest-side capability directory.
var defaultDirectory Services = directory{}

// APIDoc returns the API-documentation localization guest client.
func (directory) APIDoc() apidoccap.Service { return domainhostcall.APIDoc(invokeCapabilityJSON) }

// Auth returns the authentication and authorization guest namespace.
func (directory) Auth() authcap.Service { return domainhostcall.Auth(invokeCapabilityJSON) }

// Runtime returns the runtime host service guest client.
func (directory) Runtime() RuntimeHostService { return runtimeCapability() }

// Storage returns the storage domain guest client.
func (directory) Storage() storagecap.Service { return storageCapability() }

// Network returns the outbound network host service guest client.
func (directory) Network() NetworkHostService { return networkCapability() }

// RecordStore returns the governed record store facade for the current dynamic plugin.
func (directory) RecordStore() *recordstore.DB {
	return recordstore.OpenWithHostServiceInvoker(invokeGuestHostService)
}

// Cache returns the cache domain guest client.
func (directory) Cache() cachecap.Service { return cacheCapability() }

// Lock returns the distributed lock domain guest client.
func (directory) Lock() lockcap.Service { return lockCapability() }

// Plugins returns the plugin-domain guest capability namespace.
func (directory) Plugins() plugincap.Service { return pluginDirectory{} }

// HostConfig returns the host config capability guest client.
func (directory) HostConfig() hostconfigcap.Service { return hostConfigCapability() }

// Manifest returns the plugin manifest-resource capability guest client.
func (directory) Manifest() manifestcap.Service { return manifestCapability() }

// Users returns the user-domain capability guest client.
func (directory) Users() usercap.Service { return domainhostcall.Users(invokeCapabilityJSON) }

// BizCtx returns the current request business-context guest client.
func (directory) BizCtx() bizctxcap.Service { return domainhostcall.BizCtx(invokeCapabilityJSON) }

// Dict returns the dictionary-domain guest client.
func (directory) Dict() dictcap.Service { return domainhostcall.Dict(invokeCapabilityJSON) }

// Files returns the file-domain guest client.
func (directory) Files() filecap.Service { return domainhostcall.Files(invokeCapabilityJSON) }

// Jobs returns the scheduled-job domain guest client.
func (directory) Jobs() jobcap.Service { return domainhostcall.Jobs(invokeCapabilityJSON) }

// Notifications returns the notification-domain ordinary read guest client.
func (directory) Notifications() notifycap.Service {
	return domainhostcall.Notifications(invokeCapabilityJSON, invokeGuestHostService)
}

// Org returns the organization capability guest client.
func (directory) Org() orgcap.Service { return domainhostcall.Org(invokeCapabilityJSON) }

// Route returns the current dynamic-route metadata guest client.
func (directory) Route() routecap.Service { return domainhostcall.Route(invokeCapabilityJSON) }

// Sessions returns the online-session domain guest client.
func (directory) Sessions() sessioncap.Service { return domainhostcall.Sessions(invokeCapabilityJSON) }

// Tenant returns the tenant capability guest client.
func (directory) Tenant() tenantcap.Service { return domainhostcall.Tenant(invokeCapabilityJSON) }

// pluginDirectory implements the guest-side plugin-domain namespace.
type pluginDirectory struct{}

// Config returns the plugin-scoped config service exposed by the plugins domain.
func (pluginDirectory) Config() plugincap.ConfigService {
	return domainhostcall.PluginConfig(invokeGuestHostService)
}

// Registry returns the plugin-governance projection guest client.
func (pluginDirectory) Registry() plugincap.RegistryService {
	return domainhostcall.PluginRegistry(invokeCapabilityJSON)
}

// State returns plugin enablement lookup stubs.
func (pluginDirectory) State() plugincap.StateService {
	return domainhostcall.PluginState(invokeCapabilityJSON)
}

// Lifecycle returns plugin lifecycle governance operations.
func (pluginDirectory) Lifecycle() plugincap.LifecycleService {
	return domainhostcall.PluginLifecycle(invokeCapabilityJSON)
}

// BatchGet returns visible plugin projections and opaque missing IDs.
func (pluginDirectory) BatchGet(ctx context.Context, ids []plugincap.PluginID) (*capmodel.BatchResult[*plugincap.PluginInfo, plugincap.PluginID], error) {
	return domainhostcall.PluginRegistry(invokeCapabilityJSON).BatchGet(ctx, ids)
}

// Current returns the projection for the current caller plugin.
func (pluginDirectory) Current(ctx context.Context) (*plugincap.PluginInfo, error) {
	return domainhostcall.PluginRegistry(invokeCapabilityJSON).Current(ctx)
}

// Get returns one visible plugin projection.
func (pluginDirectory) Get(ctx context.Context, id plugincap.PluginID) (*plugincap.PluginInfo, error) {
	return domainhostcall.PluginRegistry(invokeCapabilityJSON).Get(ctx, id)
}

// List returns bounded plugin governance projections.
func (pluginDirectory) List(ctx context.Context, input plugincap.ListInput) (*capmodel.PageResult[*plugincap.PluginInfo], error) {
	return domainhostcall.PluginRegistry(invokeCapabilityJSON).List(ctx, input)
}

// ListTenantPlugins returns tenant-controllable plugins with tenant enablement.
func (pluginDirectory) ListTenantPlugins(ctx context.Context, input plugincap.TenantListInput) (*capmodel.PageResult[*plugincap.TenantPluginInfo], error) {
	return domainhostcall.PluginRegistry(invokeCapabilityJSON).ListTenantPlugins(ctx, input)
}

// IsEnabled delegates to the plugin state guest client.
func (pluginDirectory) IsEnabled(ctx context.Context, pluginID plugincap.PluginID) (bool, error) {
	return domainhostcall.PluginState(invokeCapabilityJSON).IsEnabled(ctx, pluginID)
}

// IsProviderEnabled delegates to the plugin state guest client.
func (pluginDirectory) IsProviderEnabled(ctx context.Context, pluginID plugincap.PluginID) (bool, error) {
	return domainhostcall.PluginState(invokeCapabilityJSON).IsProviderEnabled(ctx, pluginID)
}

// IsEnabledAuthoritative delegates to the plugin state guest client.
func (pluginDirectory) IsEnabledAuthoritative(ctx context.Context, pluginID plugincap.PluginID) (bool, error) {
	return domainhostcall.PluginState(invokeCapabilityJSON).IsEnabledAuthoritative(ctx, pluginID)
}
