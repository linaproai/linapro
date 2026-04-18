// This file exposes source-plugin route binding snapshots on the root plugin facade.

package plugin

import "lina-core/pkg/pluginhost"

// ListSourceRouteBindings returns the source-plugin route bindings captured during registration.
func (s *serviceImpl) ListSourceRouteBindings() []pluginhost.SourceRouteBinding {
	return s.integrationSvc.ListSourceRouteBindings()
}
