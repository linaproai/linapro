// This file wires integration service dependencies supplied by the plugin
// facade and source-plugin capability services.

package integration

import (
	"lina-core/pkg/plugin/capability"
	capabilityorgcap "lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/pluginhost"
)

// SetBizCtxProvider wires the business-context provider used by route handlers.
func (s *serviceImpl) SetBizCtxProvider(p BizCtxProvider) {
	s.bizCtxSvc = p
}

// SetTopologyProvider wires the cluster-topology provider used by plugin integrations.
func (s *serviceImpl) SetTopologyProvider(t TopologyProvider) {
	s.topology = t
}

// SetCapabilities wires the runtime-owned capability services used by source plugins.
func (s *serviceImpl) SetCapabilities(capabilities capability.Services) {
	s.capabilities = capabilities
}

// sourceServicesForPlugin returns the source-plugin-only service view at the
// callback boundary after the common capability services are scoped.
func (s *serviceImpl) sourceServicesForPlugin(pluginID string) pluginhost.Services {
	if s == nil || s.capabilities == nil {
		return nil
	}
	services := capability.ServicesForPlugin(s.capabilities, pluginID)
	if sourceServices, ok := services.(pluginhost.Services); ok {
		return sourceServices
	}
	return nil
}

// SetOrganizationCapability wires the runtime-owned organization capability
// used by plugin resource data-scope filters.
func (s *serviceImpl) SetOrganizationCapability(service capabilityorgcap.Service) {
	s.orgSvc = service
}

// SetDynamicJobExecutor wires the dynamic-plugin Jobs declaration executor.
func (s *serviceImpl) SetDynamicJobExecutor(executor DynamicJobExecutor) {
	s.dynamicJobExecutor = executor
}
