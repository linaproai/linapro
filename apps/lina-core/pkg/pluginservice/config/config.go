// Package config exposes a narrowed host configuration contract to source
// plugins.
package config

import (
	"context"

	internalconfig "lina-core/internal/service/config"
)

// MonitorConfig aliases the host monitor configuration snapshot published to plugins.
type MonitorConfig = internalconfig.MonitorConfig

// Service defines the configuration operations published to source plugins.
type Service interface {
	// GetMonitor returns the effective monitor configuration snapshot.
	GetMonitor(ctx context.Context) *MonitorConfig
}

// serviceAdapter bridges the internal config service into the published plugin contract.
type serviceAdapter struct {
	service internalconfig.Service
}

// New creates and returns the published config service adapter.
func New() Service {
	return &serviceAdapter{service: internalconfig.New()}
}

// GetMonitor returns the effective monitor configuration snapshot.
func (s *serviceAdapter) GetMonitor(ctx context.Context) *MonitorConfig {
	if s == nil || s.service == nil {
		return &MonitorConfig{}
	}
	return s.service.GetMonitor(ctx)
}
