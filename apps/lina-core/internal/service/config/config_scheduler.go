// This file defines scheduler configuration loading for host-managed jobs.

package config

import (
	"context"
	"strings"
)

// SchedulerConfig holds scheduler defaults used by code-owned jobs.
type SchedulerConfig struct {
	DefaultTimezone string `json:"defaultTimezone"` // DefaultTimezone is the IANA timezone used by managed jobs.
}

// GetScheduler reads scheduler config from configuration file.
func (s *serviceImpl) GetScheduler(ctx context.Context) *SchedulerConfig {
	return cloneSchedulerConfig(processStaticConfigCaches.scheduler.load(func() *SchedulerConfig {
		cfg := &SchedulerConfig{
			DefaultTimezone: "UTC",
		}
		mustScanConfig(ctx, "scheduler", cfg)
		cfg.DefaultTimezone = normalizeSchedulerDefaultTimezone(cfg.DefaultTimezone)
		return cfg
	}))
}

// GetSchedulerDefaultTimezone returns the configured default timezone for managed scheduled jobs.
func (s *serviceImpl) GetSchedulerDefaultTimezone(ctx context.Context) string {
	cfg := s.GetScheduler(ctx)
	if cfg == nil {
		return "UTC"
	}
	return normalizeSchedulerDefaultTimezone(cfg.DefaultTimezone)
}

// normalizeSchedulerDefaultTimezone applies the runtime default to blank values.
func normalizeSchedulerDefaultTimezone(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "UTC"
	}
	return trimmed
}
