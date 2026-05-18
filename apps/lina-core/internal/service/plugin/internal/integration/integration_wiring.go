// This file wires integration service dependencies supplied by the plugin
// facade and source-plugin host service directory.

package integration

import "lina-core/pkg/pluginhost"

// SetBizCtxProvider wires the business-context provider used by route handlers.
func (s *serviceImpl) SetBizCtxProvider(p BizCtxProvider) {
	s.bizCtxSvc = p
}

// SetTopologyProvider wires the cluster-topology provider used by plugin integrations.
func (s *serviceImpl) SetTopologyProvider(t TopologyProvider) {
	s.topology = t
}

// SetDynamicCronExecutor wires the runtime executor used by declared
// dynamic-plugin cron jobs.
func (s *serviceImpl) SetDynamicCronExecutor(executor DynamicCronExecutor) {
	s.dynamicCronExecutor = executor
}

// SetHostServices wires the host-published service directory used by source plugins.
func (s *serviceImpl) SetHostServices(services pluginhost.HostServices) {
	s.hostServices = services
}
