// This file resolves plugin-scoped source service directories used by source-plugin callbacks.

package integration

import "lina-core/pkg/plugin/pluginhost"

// sourceServicesForPlugin returns the source-plugin-only service view at the
// callback boundary after the common capability services are scoped.
func (s *serviceImpl) sourceServicesForPlugin(pluginID string) pluginhost.Services {
	if s == nil || s.sourceServices == nil {
		return nil
	}
	return s.sourceServices.SourceServicesForPlugin(pluginID)
}
