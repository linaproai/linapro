// Package monitor implements the monitor-online plugin HTTP controllers.
package monitor

import (
	monitorapi "lina-plugin-monitor-online/backend/api/monitor"
	monitorsvc "lina-plugin-monitor-online/backend/internal/service/monitor"
)

// ControllerV1 is the monitor-online controller.
type ControllerV1 struct {
	monitorSvc monitorsvc.Service // monitor service
}

// NewV1 creates and returns a new monitor-online controller instance.
func NewV1(monitorSvc monitorsvc.Service) monitorapi.IMonitorV1 {
	return &ControllerV1{monitorSvc: monitorSvc}
}
