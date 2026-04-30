// This file defines health-probe configuration loading.

package config

import (
	"context"
	"time"
)

// HealthConfig holds operational health-probe configuration.
type HealthConfig struct {
	Timeout time.Duration `json:"timeout"` // Timeout bounds the database probe.
}

// GetHealth reads health probe config from configuration file.
func (s *serviceImpl) GetHealth(ctx context.Context) *HealthConfig {
	return cloneHealthConfig(processStaticConfigCaches.health.load(func() *HealthConfig {
		cfg := &HealthConfig{
			Timeout: 5 * time.Second,
		}
		mustScanConfig(ctx, "health", cfg)
		cfg.Timeout = mustLoadDurationConfig(ctx, "health.timeout", cfg.Timeout)
		cfg.Timeout = mustValidateSecondAlignedDuration("health.timeout", cfg.Timeout)
		return cfg
	}))
}
