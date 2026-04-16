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

// GetSession reads session config from configuration file.
func (s *serviceImpl) GetSession(ctx context.Context) *SessionConfig {
	cfg := &SessionConfig{
		Timeout:         24 * time.Hour,
		CleanupInterval: 5 * time.Minute,
	}
	cfg.Timeout = mustLoadDurationConfig(ctx, "session.timeout", cfg.Timeout)
	cfg.CleanupInterval = mustLoadDurationConfig(ctx, "session.cleanupInterval", cfg.CleanupInterval)
	cfg.CleanupInterval = mustValidateSecondAlignedDuration("session.cleanupInterval", cfg.CleanupInterval)
	return cfg
}
