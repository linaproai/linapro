// This file exposes the dynamic-plugin Jobs domain capability client.

package pluginbridge

import (
	"lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/pluginbridge/internal/domainhostcall"
)

// jobsCapability returns the process-default Jobs domain client.
func jobsCapability() jobcap.Service {
	return domainhostcall.Jobs(invokeCapabilityJSON)
}
