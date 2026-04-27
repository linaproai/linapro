// This file registers built-in cron declarations for the dynamic sample
// plugin through the governed cron host service.

package dynamicservice

import "lina-core/pkg/pluginbridge"

// Cron heartbeat declaration constants define the built-in cron contract
// exported by the dynamic sample plugin.
const (
	cronHeartbeatName        = "heartbeat"
	cronHeartbeatDisplayName = "Dynamic Plugin Heartbeat"
	cronHeartbeatDesc        = "Runs the dynamic plugin built-in job through the Wasm bridge and accumulates heartbeat executions."
	cronHeartbeatPattern     = "# */10 * * * *"
	cronHeartbeatPath        = "/cron-heartbeat"
	cronHeartbeatTimeout     = 30
)

// RegisterCrons publishes all built-in cron declarations for host-side
// discovery.
func (s *serviceImpl) RegisterCrons() error {
	return s.cronSvc.Register(&pluginbridge.CronContract{
		Name:           cronHeartbeatName,
		DisplayName:    cronHeartbeatDisplayName,
		Description:    cronHeartbeatDesc,
		Pattern:        cronHeartbeatPattern,
		Timezone:       pluginbridge.DefaultCronContractTimezone,
		Scope:          pluginbridge.CronScopeAllNode,
		Concurrency:    pluginbridge.CronConcurrencySingleton,
		MaxConcurrency: 1,
		TimeoutSeconds: cronHeartbeatTimeout,
		InternalPath:   cronHeartbeatPath,
	})
}
