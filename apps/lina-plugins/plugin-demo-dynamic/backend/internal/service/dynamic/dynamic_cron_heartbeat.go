// This file implements the declared cron heartbeat business logic for the
// dynamic sample plugin.

package dynamicservice

import (
	"strconv"

	"lina-core/pkg/pluginbridge"
)

const (
	cronHeartbeatStateKey = "cron_heartbeat_count"
)

// cronHeartbeatPayload summarizes one successful cron heartbeat execution.
type cronHeartbeatPayload struct {
	// Count is the accumulated execution count persisted in plugin-scoped runtime state.
	Count int `json:"count"`
	// Message describes the cron execution result.
	Message string `json:"message"`
}

// BuildCronHeartbeatPayload executes the declared cron heartbeat task and
// returns a lightweight execution summary.
func (s *serviceImpl) BuildCronHeartbeatPayload() (*cronHeartbeatPayload, error) {
	count, found, err := s.runtimeSvc.StateGetInt(cronHeartbeatStateKey)
	if err != nil {
		return nil, err
	}
	if !found {
		count = 0
	}
	count++
	if err = s.runtimeSvc.StateSetInt(cronHeartbeatStateKey, count); err != nil {
		return nil, err
	}
	if err = s.runtimeSvc.Log(
		int(pluginbridge.LogLevelInfo),
		"declared cron heartbeat executed",
		map[string]string{
			"stateKey": cronHeartbeatStateKey,
			"count":    strconv.Itoa(count),
		},
	); err != nil {
		return nil, err
	}
	return &cronHeartbeatPayload{
		Count:   count,
		Message: "Dynamic plugin cron heartbeat executed successfully.",
	}, nil
}
