// This file defines online-session configuration loading and duration migration.

package config

import (
	"context"
	"time"
)

// SessionConfig holds session management configuration.
type SessionConfig struct {
	Timeout         time.Duration `json:"timeout"`         // Timeout is the inactivity threshold for one session.
	CleanupInterval time.Duration `json:"cleanupInterval"` // CleanupInterval is the cleanup job execution interval.
}

// getStaticSessionConfig lazily loads the config-file-backed session settings
// for reuse across the process.
func (s *serviceImpl) getStaticSessionConfig(ctx context.Context) *SessionConfig {
	return processStaticConfigCaches.session.load(func() *SessionConfig {
		cfg := &SessionConfig{
			Timeout:         24 * time.Hour,
			CleanupInterval: 5 * time.Minute,
		}
		cfg.Timeout = mustLoadDurationConfig(ctx, "session.timeout", cfg.Timeout)
		cfg.CleanupInterval = mustLoadDurationConfig(ctx, "session.cleanupInterval", cfg.CleanupInterval)
		cfg.CleanupInterval = mustValidateSecondAlignedDuration("session.cleanupInterval", cfg.CleanupInterval)
		return cfg
	})
}

// GetSession reads session config from configuration file.
func (s *serviceImpl) GetSession(ctx context.Context) *SessionConfig {
	cfg := cloneSessionConfig(s.getStaticSessionConfig(ctx))
	cfg.Timeout = s.applyRuntimeDurationOverride(ctx, RuntimeParamKeySessionTimeout, cfg.Timeout)
	return cfg
}

// GetSessionTimeout returns the runtime-effective online-session timeout.
func (s *serviceImpl) GetSessionTimeout(ctx context.Context) time.Duration {
	cfg := s.getStaticSessionConfig(ctx)
	if cfg == nil {
		return 24 * time.Hour
	}
	return s.applyRuntimeDurationOverride(ctx, RuntimeParamKeySessionTimeout, cfg.Timeout)
}
