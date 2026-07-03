// This file implements guest-side plugin registry and enablement lookup
// hostcall clients. Plugin config remains owned by the public guest package
// adapter.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// pluginRegistryService adapts plugin registry projection reads to host services.
type pluginRegistryService struct{ baseService }

// pluginStateService adapts plugin enablement lookups to host services.
type pluginStateService struct{ baseService }

// pluginLifecycleService adapts plugin lifecycle governance operations to host services.
type pluginLifecycleService struct{ baseService }

// PluginRegistry creates the plugin-governance projection guest client.
func PluginRegistry(invoker Invoker) plugincap.RegistryService {
	return pluginRegistryService{baseService: newBaseService(invoker)}
}

// PluginState creates the plugin enablement lookup guest client.
func PluginState(invoker Invoker) plugincap.StateService {
	return pluginStateService{baseService: newBaseService(invoker)}
}

// PluginLifecycle creates the plugin lifecycle governance guest client.
func PluginLifecycle(invoker Invoker) plugincap.LifecycleService {
	return pluginLifecycleService{baseService: newBaseService(invoker)}
}

// BatchGet returns visible plugin projections and opaque missing IDs.
func (s pluginRegistryService) BatchGet(_ context.Context, ids []plugincap.PluginID) (*capmodel.BatchResult[*plugincap.PluginInfo, plugincap.PluginID], error) {
	out := &capmodel.BatchResult[*plugincap.PluginInfo, plugincap.PluginID]{Items: map[plugincap.PluginID]*plugincap.PluginInfo{}}
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsBatchGet, idsRequest{IDs: pluginIDsToStrings(ids)}, out)
	return out, err
}

// Current returns the projection for the current caller plugin.
func (s pluginRegistryService) Current(_ context.Context) (*plugincap.PluginInfo, error) {
	out := &plugincap.PluginInfo{}
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsCurrent, nil, out)
	return out, err
}

// Get returns one visible plugin projection through the registered batch-read method.
func (s pluginRegistryService) Get(ctx context.Context, id plugincap.PluginID) (*plugincap.PluginInfo, error) {
	result, err := s.BatchGet(ctx, []plugincap.PluginID{id})
	if err != nil || result == nil {
		return nil, err
	}
	if item, ok := result.Items[id]; ok {
		return item, nil
	}
	return nil, nil
}

// List returns bounded plugin governance projections.
func (s pluginRegistryService) List(_ context.Context, input plugincap.ListInput) (*capmodel.PageResult[*plugincap.PluginInfo], error) {
	out := &capmodel.PageResult[*plugincap.PluginInfo]{Items: []*plugincap.PluginInfo{}}
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsList, input, out)
	return out, err
}

// ListTenantPlugins returns tenant-controllable plugins with tenant enablement.
func (s pluginRegistryService) ListTenantPlugins(_ context.Context, input plugincap.TenantListInput) (*capmodel.PageResult[*plugincap.TenantPluginInfo], error) {
	out := &capmodel.PageResult[*plugincap.TenantPluginInfo]{Items: []*plugincap.TenantPluginInfo{}}
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsListTenant, input, out)
	return out, err
}

// IsEnabled reports whether the target plugin is enabled in the current scope.
func (s pluginStateService) IsEnabled(_ context.Context, pluginID plugincap.PluginID) (bool, error) {
	var out bool
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsStateIsEnabled, pluginIDRequest{PluginID: string(pluginID)}, &out)
	return out, err
}

// IsProviderEnabled reports whether the target plugin may serve provider calls.
func (s pluginStateService) IsProviderEnabled(_ context.Context, pluginID plugincap.PluginID) (bool, error) {
	var out bool
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsStateIsProviderEnabled, pluginIDRequest{PluginID: string(pluginID)}, &out)
	return out, err
}

// IsEnabledAuthoritative reports persisted plugin enablement bypassing local snapshots.
func (s pluginStateService) IsEnabledAuthoritative(_ context.Context, pluginID plugincap.PluginID) (bool, error) {
	var out bool
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsStateIsEnabledAuthoritative, pluginIDRequest{PluginID: string(pluginID)}, &out)
	return out, err
}

// EnsureTenantPluginDisableAllowed runs tenant-plugin disable preconditions.
func (s pluginLifecycleService) EnsureTenantPluginDisableAllowed(_ context.Context, pluginID string, tenantID int) error {
	return s.callJSONRequest(
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed,
		tenantPluginLifecycleRequest{PluginID: pluginID, TenantID: tenantID},
		nil,
	)
}

// NotifyTenantPluginDisabled runs tenant-plugin disabled notifications.
func (s pluginLifecycleService) NotifyTenantPluginDisabled(_ context.Context, pluginID string, tenantID int) {
	_ = s.callJSONRequest(
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled,
		tenantPluginLifecycleRequest{PluginID: pluginID, TenantID: tenantID},
		nil,
	)
}

// EnsureTenantDeleteAllowed runs tenant-delete preconditions.
func (s pluginLifecycleService) EnsureTenantDeleteAllowed(_ context.Context, tenantID int) error {
	return s.callJSONRequest(
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed,
		tenantIDRequest{TenantID: tenantID},
		nil,
	)
}

// NotifyTenantDeleted runs tenant-deleted notifications.
func (s pluginLifecycleService) NotifyTenantDeleted(_ context.Context, tenantID int) {
	_ = s.callJSONRequest(
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleNotifyTenantDeleted,
		tenantIDRequest{TenantID: tenantID},
		nil,
	)
}

// pluginIDRequest carries one plugin identifier.
type pluginIDRequest struct {
	// PluginID is the plugin identifier.
	PluginID string `json:"pluginId"`
}

// tenantPluginLifecycleRequest carries one plugin and tenant pair.
type tenantPluginLifecycleRequest struct {
	// PluginID is the plugin identifier.
	PluginID string `json:"pluginId"`
	// TenantID is the tenant identifier.
	TenantID int `json:"tenantId"`
}

// tenantIDRequest carries one tenant identifier.
type tenantIDRequest struct {
	// TenantID is the tenant identifier.
	TenantID int `json:"tenantId"`
}

// pluginIDsToStrings converts plugin IDs to transport strings.
func pluginIDsToStrings(ids []plugincap.PluginID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

var (
	_ plugincap.RegistryService  = (*pluginRegistryService)(nil)
	_ plugincap.StateService     = (*pluginStateService)(nil)
	_ plugincap.LifecycleService = (*pluginLifecycleService)(nil)
)
