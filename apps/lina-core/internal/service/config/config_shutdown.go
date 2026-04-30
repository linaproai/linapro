// This file defines graceful-shutdown configuration loading.

package config

import (
	"context"
	"time"
)

// ShutdownConfig holds process graceful-shutdown configuration.
type ShutdownConfig struct {
	Timeout time.Duration `json:"timeout"` // Timeout bounds the full shutdown sequence.
}

// GetShutdown reads graceful-shutdown config from configuration file.
func (s *serviceImpl) GetShutdown(ctx context.Context) *ShutdownConfig {
	return cloneShutdownConfig(processStaticConfigCaches.shutdown.load(func() *ShutdownConfig {
		cfg := &ShutdownConfig{
			Timeout: 30 * time.Second,
		}
		mustScanConfig(ctx, "shutdown", cfg)
		cfg.Timeout = mustLoadDurationConfig(ctx, "shutdown.timeout", cfg.Timeout)
		cfg.Timeout = mustValidateSecondAlignedDuration("shutdown.timeout", cfg.Timeout)
		return cfg
	}))
}
