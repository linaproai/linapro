// This file defines host logger configuration loading.

package config

import "context"

// LoggerConfig holds logger behavior switches loaded from configuration.
type LoggerConfig struct {
	Structured bool `json:"structured"` // Structured enables GoFrame JSON structured logging.
}

// GetLogger reads logger config from configuration file.
func (s *serviceImpl) GetLogger(ctx context.Context) *LoggerConfig {
	cfg := &LoggerConfig{
		Structured: false,
	}
	mustScanConfig(ctx, "logger", cfg)
	return cfg
}
