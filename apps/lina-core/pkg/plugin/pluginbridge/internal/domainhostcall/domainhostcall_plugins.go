// This file implements guest-side plugin registry, state, and lifecycle
// capability hostcall clients. Plugin config remains owned by the public guest
// package adapter.

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

// pluginLifecycleService adapts plugin lifecycle governance calls to host services.
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
func (s pluginRegistryService) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []plugincap.PluginID) (*capmodel.BatchResult[*plugincap.Projection, plugincap.PluginID], error) {
	out := &capmodel.BatchResult[*plugincap.Projection, plugincap.PluginID]{Items: map[plugincap.PluginID]*plugincap.Projection{}}
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsBatchGet, idsRequest{IDs: pluginIDsToStrings(ids)}, out)
	return out, err
}

// Current returns the projection for the current caller plugin.
func (s pluginRegistryService) Current(_ context.Context, _ capmodel.CapabilityContext) (*plugincap.Projection, error) {
	out := &plugincap.Projection{}
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsCurrent, nil, out)
	return out, err
}

// Search returns bounded plugin governance projections.
func (s pluginRegistryService) Search(_ context.Context, _ capmodel.CapabilityContext, input plugincap.SearchInput) (*capmodel.PageResult[*plugincap.Projection], error) {
	out := &capmodel.PageResult[*plugincap.Projection]{Items: []*plugincap.Projection{}}
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsSearch, input, out)
	return out, err
}

// ListTenantPlugins returns tenant-controllable plugins with tenant enablement.
func (s pluginRegistryService) ListTenantPlugins(_ context.Context, _ capmodel.CapabilityContext, input plugincap.TenantListInput) (*capmodel.PageResult[*plugincap.TenantProjection], error) {
	out := &capmodel.PageResult[*plugincap.TenantProjection]{Items: []*plugincap.TenantProjection{}}
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsListTenant, input, out)
	return out, err
}

// BatchGetCapabilityStatus returns framework capability status projections by stable key.
func (s pluginRegistryService) BatchGetCapabilityStatus(_ context.Context, _ capmodel.CapabilityContext, keys []plugincap.CapabilityKey) (*capmodel.BatchResult[*capmodel.CapabilityStatus, plugincap.CapabilityKey], error) {
	out := &capmodel.BatchResult[*capmodel.CapabilityStatus, plugincap.CapabilityKey]{Items: map[plugincap.CapabilityKey]*capmodel.CapabilityStatus{}}
	err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsBatchGetCapabilityStatus, capabilityKeysRequest{Keys: capabilityKeysToStrings(keys)}, out)
	return out, err
}

// IsEnabled reports whether the plugin is enabled in the current request scope.
func (s pluginStateService) IsEnabled(_ context.Context, pluginID string) bool {
	var out bool
	if err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsIsEnabled, pluginIDRequest{PluginID: pluginID}, &out); err != nil {
		return false
	}
	return out
}

// IsProviderEnabled reports platform-level provider availability.
func (s pluginStateService) IsProviderEnabled(_ context.Context, pluginID string) bool {
	var out bool
	if err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsIsProviderEnabled, pluginIDRequest{PluginID: pluginID}, &out); err != nil {
		return false
	}
	return out
}

// IsEnabledAuthoritative bypasses process-local snapshots for sensitive checks.
func (s pluginStateService) IsEnabledAuthoritative(_ context.Context, pluginID string) bool {
	var out bool
	if err := s.callJSONRequest(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsIsEnabledAuthoritative, pluginIDRequest{PluginID: pluginID}, &out); err != nil {
		return false
	}
	return out
}

// EnsureTenantPluginDisableAllowed runs tenant-plugin disable preconditions.
func (s pluginLifecycleService) EnsureTenantPluginDisableAllowed(_ context.Context, pluginID string, tenantID int) error {
	return s.callJSONRequest(
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisable,
		pluginTenantLifecycleRequest{PluginID: pluginID, TenantID: tenantID},
		nil,
	)
}

// NotifyTenantPluginDisabled runs tenant-plugin disable notifications.
func (s pluginLifecycleService) NotifyTenantPluginDisabled(_ context.Context, pluginID string, tenantID int) {
	if err := s.callJSONRequest(
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled,
		pluginTenantLifecycleRequest{PluginID: pluginID, TenantID: tenantID},
		nil,
	); err != nil {
		return
	}
}

// EnsureTenantDeleteAllowed runs tenant-delete preconditions.
func (s pluginLifecycleService) EnsureTenantDeleteAllowed(_ context.Context, tenantID int) error {
	return s.callJSONRequest(
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleEnsureTenantDelete,
		tenantLifecycleRequest{TenantID: tenantID},
		nil,
	)
}

// NotifyTenantDeleted runs tenant-delete notifications.
func (s pluginLifecycleService) NotifyTenantDeleted(_ context.Context, tenantID int) {
	if err := s.callJSONRequest(
		protocol.HostServicePlugins,
		protocol.HostServiceMethodPluginsLifecycleNotifyTenantDeleted,
		tenantLifecycleRequest{TenantID: tenantID},
		nil,
	); err != nil {
		return
	}
}

// pluginIDRequest carries one plugin identifier for state lookups.
type pluginIDRequest struct {
	PluginID string `json:"pluginId"`
}

// pluginTenantLifecycleRequest carries one tenant-scoped plugin lifecycle target.
type pluginTenantLifecycleRequest struct {
	PluginID string `json:"pluginId"`
	TenantID int    `json:"tenantId"`
}

// tenantLifecycleRequest carries one tenant lifecycle target.
type tenantLifecycleRequest struct {
	TenantID int `json:"tenantId"`
}

// capabilityKeysRequest carries framework capability status keys.
type capabilityKeysRequest struct {
	Keys []string `json:"keys"`
}

// pluginIDsToStrings converts plugin IDs to transport strings.
func pluginIDsToStrings(ids []plugincap.PluginID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

// capabilityKeysToStrings converts capability keys to transport strings.
func capabilityKeysToStrings(keys []plugincap.CapabilityKey) []string {
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, string(key))
	}
	return out
}

var (
	_ plugincap.RegistryService  = (*pluginRegistryService)(nil)
	_ plugincap.StateService     = (*pluginStateService)(nil)
	_ plugincap.LifecycleService = (*pluginLifecycleService)(nil)
)
