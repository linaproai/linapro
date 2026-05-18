// Package pluginlifecycle exposes host plugin lifecycle orchestration through
// the stable source-plugin service contract.
package pluginlifecycle

import "lina-core/pkg/pluginservice/contract"

// service delegates lifecycle orchestration to the host-owned runner.
type service struct {
	runner contract.PluginLifecycleRunner
}

// New creates a source-plugin visible plugin lifecycle service.
func New(runner contract.PluginLifecycleRunner) contract.PluginLifecycleService {
	return &service{runner: runner}
}
